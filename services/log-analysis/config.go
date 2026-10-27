package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	"lite-nas/services/log-analysis/plugins"
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// analysisConfig uses the shared INI bootstrap sections and a JSON rule file,
// matching the existing Resources Monitor rules convention.
type analysisConfig struct {
	Messaging    sharedconfig.MessagingConfig
	Auth         sharedconfig.AuthConfig
	Logging      sharedconfig.LoggingConfig
	RulesFile    string
	PollInterval time.Duration
}

func loadAnalysisConfig(reader fileio.Reader) (analysisConfig, error) {
	return sharedconfig.LoadConfigWithSharedSections(reader, func(file *ini.File, shared sharedconfig.SharedConfig) (analysisConfig, error) {
		section := file.Section("analysis")
		rules := section.Key("rules_file").String()
		if !filepath.IsAbs(rules) || filepath.Clean(rules) != rules {
			return analysisConfig{}, errors.New("analysis rules_file must be an absolute clean path")
		}
		interval, err := time.ParseDuration(section.Key("poll_interval").MustString("1s"))
		if err != nil || interval < 100*time.Millisecond || interval > time.Minute {
			return analysisConfig{}, errors.New("analysis poll_interval must be 100ms-1m")
		}
		if shared.Auth.ServiceLogin == "" {
			return analysisConfig{}, errors.New("analysis auth service_login is required")
		}
		return analysisConfig{Messaging: shared.Messaging, Auth: shared.Auth, Logging: shared.Logging, RulesFile: rules, PollInterval: interval}, nil
	})
}

// sourceSpec binds one physical rsyslog output file to a recognized parser.
type sourceSpec struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Format string `json:"format"`
}

// instanceSpec selects one compiled plugin type and its validated parameters.
type instanceSpec struct {
	ID        string `json:"id"`
	Enabled   bool   `json:"enabled"`
	Type      string `json:"type"`
	Threshold int    `json:"threshold"`
	Window    string `json:"window"`
	Cooldown  string `json:"cooldown"` // Quiet period before an aggregated event is normalized.
	Idle      string `json:"idle"`
	KeyBy     string `json:"key_by"`
	MaxKeys   int    `json:"max_keys"`
}

// ruleSpec selects observations for one initialized plugin instance.
type ruleSpec struct {
	ID              string   `json:"id"`
	Enabled         bool     `json:"enabled"`
	Source          string   `json:"source"`
	Category        string   `json:"category"`
	Pattern         string   `json:"pattern"`
	SuccessPattern  string   `json:"success_pattern,omitempty"`
	Require         []string `json:"require,omitempty"`
	Exclude         []string `json:"exclude,omitempty"`
	Instance        string   `json:"instance"`
	Severity        string   `json:"severity"`
	EmailOnCritical bool     `json:"email_on_critical"`
	Lockdown        bool     `json:"lockdown"`
}

// rulesFile is the complete startup construction plan. Nothing in this file
// is evaluated as Go code or a command.
type rulesFile struct {
	EventIDPrefix string         `json:"event_id_prefix"`
	Sources       []sourceSpec   `json:"sources"`
	Instances     []instanceSpec `json:"instances"`
	Rules         []ruleSpec     `json:"rules"`
}

// compiledRule holds safe RE2 patterns and an already initialized plugin.
type compiledRule struct {
	spec    ruleSpec
	match   *regexp.Regexp
	success *regexp.Regexp
	require []*regexp.Regexp
	exclude []*regexp.Regexp
	plugin  *plugins.Plugin
}

// loadRulesFile rejects unknown fields, duplicates, unsafe paths, missing
// instance references and malformed patterns before any listener starts.
func loadRulesFile(path string) (rulesFile, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- validated administrator configuration path
	if err != nil {
		return rulesFile{}, err
	}
	if len(data) > 1024*1024 {
		return rulesFile{}, errors.New("analysis rules file exceeds 1 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var spec rulesFile
	if err := decoder.Decode(&spec); err != nil {
		return rulesFile{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return rulesFile{}, errors.New("analysis rules file contains trailing JSON")
	}
	if len(spec.Sources) == 0 || len(spec.Sources) > 16 || len(spec.Instances) == 0 || len(spec.Instances) > 64 || len(spec.Rules) == 0 || len(spec.Rules) > 128 {
		return rulesFile{}, errors.New("analysis rules file exceeds configured bounds")
	}
	if !validEventIDPrefix(spec.EventIDPrefix) {
		return rulesFile{}, errors.New("analysis rules event_id_prefix must be 1-4 alphanumeric characters")
	}
	sources := make(map[string]bool)
	for _, source := range spec.Sources {
		if !validAnalysisID(source.ID) || sources[source.ID] || !filepath.IsAbs(source.Path) || filepath.Clean(source.Path) != source.Path || source.Format != "rsyslog_auth" {
			return rulesFile{}, fmt.Errorf("invalid source %q", source.ID)
		}
		sources[source.ID] = true
	}
	instances := make(map[string]bool)
	instanceIDs := make(map[string]bool)
	for _, instance := range spec.Instances {
		if !validAnalysisID(instance.ID) || instanceIDs[instance.ID] {
			return rulesFile{}, fmt.Errorf("invalid instance %q", instance.ID)
		}
		instanceIDs[instance.ID] = true
		instances[instance.ID] = instance.Enabled
		if !instance.Enabled {
			continue
		}
		if _, err := instance.pluginConfig(spec.EventIDPrefix); err != nil {
			return rulesFile{}, fmt.Errorf("instance %s: %w", instance.ID, err)
		}
	}
	rules := make(map[string]bool)
	for _, rule := range spec.Rules {
		if !validAnalysisID(rule.ID) || rules[rule.ID] {
			return rulesFile{}, fmt.Errorf("invalid rule %q", rule.ID)
		}
		rules[rule.ID] = true
		if !rule.Enabled {
			continue
		}
		if !sources[rule.Source] || !instances[rule.Instance] || !validAnalysisID(rule.Category) {
			return rulesFile{}, fmt.Errorf("rule %s has missing source, instance or category", rule.ID)
		}
		if rule.Severity != "info" && rule.Severity != "warning" && rule.Severity != "error" && rule.Severity != "critical" {
			return rulesFile{}, fmt.Errorf("rule %s has invalid severity", rule.ID)
		}
		if rule.Lockdown && rule.Severity != "critical" {
			return rulesFile{}, fmt.Errorf("rule %s lockdown requires critical severity", rule.ID)
		}
		if rule.EmailOnCritical && rule.Severity != "critical" {
			return rulesFile{}, fmt.Errorf("rule %s email requires critical severity", rule.ID)
		}
		if _, err := compileRulePatterns(rule); err != nil {
			return rulesFile{}, fmt.Errorf("rule %s: %w", rule.ID, err)
		}
	}
	return spec, nil
}

func validAnalysisID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func (spec instanceSpec) pluginConfig(eventIDPrefix string) (plugins.Config, error) {
	window, err := time.ParseDuration(spec.Window)
	if err != nil {
		return plugins.Config{}, err
	}
	cooldown, err := time.ParseDuration(spec.Cooldown)
	if err != nil {
		return plugins.Config{}, err
	}
	idle, err := time.ParseDuration(spec.Idle)
	if err != nil {
		return plugins.Config{}, err
	}
	return plugins.Config{
		ID: spec.ID, Type: spec.Type, Threshold: spec.Threshold, Window: window,
		Cooldown: cooldown, Idle: idle, KeyBy: spec.KeyBy, MaxKeys: spec.MaxKeys,
		EventIDPrefix: eventIDPrefix,
	}, nil
}

// validEventIDPrefix ensures the configured security marker leaves room for
// the random suffix while respecting logging-manager's ten-character prefix.
func validEventIDPrefix(value string) bool {
	if len(value) < 1 || len(value) > 4 {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func compileRulePatterns(rule ruleSpec) (compiledRule, error) {
	compile := func(text string) (*regexp.Regexp, error) {
		if text == "" || len(text) > 256 || strings.ContainsAny(text, "\r\n\x00") {
			return nil, errors.New("pattern must be 1-256 single-line bytes")
		}
		return regexp.Compile(text)
	}
	match, err := compile(rule.Pattern)
	if err != nil {
		return compiledRule{}, err
	}
	compiled := compiledRule{spec: rule, match: match}
	if rule.SuccessPattern != "" {
		compiled.success, err = compile(rule.SuccessPattern)
		if err != nil {
			return compiledRule{}, err
		}
	}
	if len(rule.Require) > 8 || len(rule.Exclude) > 8 {
		return compiledRule{}, errors.New("too many conditions")
	}
	for _, pattern := range rule.Require {
		parsed, err := compile(pattern)
		if err != nil {
			return compiledRule{}, err
		}
		compiled.require = append(compiled.require, parsed)
	}
	for _, pattern := range rule.Exclude {
		parsed, err := compile(pattern)
		if err != nil {
			return compiledRule{}, err
		}
		compiled.exclude = append(compiled.exclude, parsed)
	}
	return compiled, nil
}
