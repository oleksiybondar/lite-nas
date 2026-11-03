package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"lite-nas/services/security-automation/detection"
	contract "lite-nas/shared/contracts/securityautomation"
)

var (
	automationID = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	eventPrefix  = regexp.MustCompile(`^[A-Za-z0-9_]{1,20}$`)
)

var actionOperands = map[string][]string{
	"block_source_ip":         {"requester_ip_address"},
	"block_scan_source_ip":    {"requester_ip_address"},
	"revoke_web_sessions":     {"username", "requester_ip_address"},
	"usb_block_device":        {"device_id"},
	"login_terminate_session": {"session_id"},
}

// eventSelector selects one validated security or system creation event.
type eventSelector struct {
	Type     string `json:"type"`
	IDPrefix string `json:"id_prefix"`
	Category string `json:"category,omitempty"`
	Source   string `json:"source,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  struct {
		Equals   string `json:"equals,omitempty"`
		Contains string `json:"contains,omitempty"`
		Regex    string `json:"regex,omitempty"`
	} `json:"message"`
}

// automationRuleSpec is one configured detector and its currently approved action selector.
type automationRuleSpec struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Explanation string `json:"explanation"`
	Enabled     *bool  `json:"enabled"`
	FSM         struct {
		Type      detection.PluginName `json:"type"`
		Window    string               `json:"window"`
		Threshold int                  `json:"threshold,omitempty"`
		MaxKeys   int                  `json:"max_keys"`
		MaxIdle   string               `json:"max_idle"`
	} `json:"fsm"`
	CorrelateBy string          `json:"correlate_by"`
	Steps       []eventSelector `json:"steps"`
	Action      struct {
		Kind             string   `json:"kind"`
		Name             string   `json:"name,omitempty"`
		RequiredOperands []string `json:"required_operands,omitempty"`
		Severity         string   `json:"severity,omitempty"`
		Email            *bool    `json:"email,omitempty"`
		Mitigate         string   `json:"mitigate,omitempty"`
	} `json:"action"`
}

// compiledAutomationRule holds startup-compiled selectors and one initialized FSM for enabled rules.
type compiledAutomationRule struct {
	spec           automationRuleSpec
	detectorConfig detection.Config
	regexes        []*regexp.Regexp
	plugin         *detection.Plugin
	view           contract.Rule
	enabled        atomic.Bool
}

// loadAutomationRules rejects unknown fields and invalid selectors before any NATS callback starts.
func loadAutomationRules(path string) (map[string]*compiledAutomationRule, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- administrator-owned path is validated in INI startup config
	if err != nil {
		return nil, err
	}
	if len(data) > 1024*1024 {
		return nil, errors.New("automation rules file exceeds 1 MiB")
	}
	var file struct {
		Rules []automationRuleSpec `json:"rules"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("trailing automation rules JSON")
	}
	if len(file.Rules) == 0 || len(file.Rules) > 128 {
		return nil, errors.New("automation rules file must define 1-128 rules")
	}
	rules := make(map[string]*compiledAutomationRule, len(file.Rules))
	for _, spec := range file.Rules {
		if !automationID.MatchString(spec.ID) || rules[spec.ID] != nil || spec.Enabled == nil || len(spec.Name) == 0 || len(spec.Name) > 128 || len(spec.Description) == 0 || len(spec.Description) > 256 || len(spec.Explanation) > 1024 || len(spec.Steps) == 0 || len(spec.Steps) > 64 {
			return nil, fmt.Errorf("invalid automation rule %q", spec.ID)
		}
		compiled, err := compileAutomationRule(spec)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", spec.ID, err)
		}
		rules[spec.ID] = compiled
	}
	return rules, nil
}

// compileAutomationRule validates every selector without constructing an FSM instance.
func compileAutomationRule(spec automationRuleSpec) (*compiledAutomationRule, error) {
	if !automationID.MatchString(spec.CorrelateBy) || (spec.Action.Kind != "alert_only" && spec.Action.Kind != "alert" && spec.Action.Kind != "builtin" && spec.Action.Kind != "command_profile") {
		return nil, errors.New("invalid correlation or action")
	}
	if spec.Action.Kind == "alert" {
		if spec.Action.Name != "" || len(spec.Action.RequiredOperands) != 0 || !validSeverity(spec.Action.Severity) || spec.Action.Severity == "" || spec.Action.Email == nil || len(spec.Action.Mitigate) == 0 || len(spec.Action.Mitigate) > 4096 || strings.TrimSpace(spec.Action.Mitigate) != spec.Action.Mitigate {
			return nil, errors.New("alert action requires severity, email policy, and mitigation guidance without command operands")
		}
	} else if spec.Action.Mitigate != "" {
		return nil, errors.New("only alert actions can set operator mitigation guidance")
	} else if spec.Action.Severity != "" || spec.Action.Email != nil {
		return nil, errors.New("only alert actions can set severity or email policy")
	}
	if spec.Action.Kind == "alert_only" && (spec.Action.Name != "" || len(spec.Action.RequiredOperands) != 0) {
		return nil, errors.New("alert-only action cannot take command operands")
	}
	if spec.Action.Kind != "alert_only" && spec.Action.Kind != "alert" && spec.Action.Name == "" {
		return nil, errors.New("mitigation action requires a name")
	}
	if spec.Action.Kind == "builtin" || spec.Action.Kind == "command_profile" {
		expected, known := actionOperands[spec.Action.Name]
		if !known || spec.Action.Kind == "builtin" && spec.Action.Name != "block_source_ip" && spec.Action.Name != "block_scan_source_ip" && spec.Action.Name != "revoke_web_sessions" || spec.Action.Kind == "command_profile" && (spec.Action.Name == "block_source_ip" || spec.Action.Name == "block_scan_source_ip" || spec.Action.Name == "revoke_web_sessions") || len(expected) != len(spec.Action.RequiredOperands) {
			return nil, errors.New("invalid action profile or operands")
		}
		for index, name := range expected {
			if spec.Action.RequiredOperands[index] != name {
				return nil, errors.New("action operands do not match profile")
			}
		}
	}
	window, err := time.ParseDuration(spec.FSM.Window)
	if err != nil {
		return nil, err
	}
	idle, err := time.ParseDuration(spec.FSM.MaxIdle)
	if err != nil {
		return nil, err
	}
	config := detection.Config{Name: spec.FSM.Type, Steps: len(spec.Steps), Threshold: spec.FSM.Threshold, Window: window, MaxKeys: spec.FSM.MaxKeys, MaxIdle: idle}
	if err := detection.ValidateConfig(config, time.Now); err != nil {
		return nil, err
	}
	if spec.FSM.Type == detection.CountWithinWindow && len(spec.Steps) != 1 {
		return nil, errors.New("count detector needs exactly one step")
	}
	compiled := &compiledAutomationRule{spec: spec, detectorConfig: config, view: contract.Rule{ID: spec.ID, Name: spec.Name, Description: spec.Description, Explanation: spec.Explanation, ActionKind: spec.Action.Kind, ActionName: spec.Action.Name, Enabled: *spec.Enabled, ConfiguredEnabled: *spec.Enabled, RequiredOperands: spec.Action.RequiredOperands, State: contract.StateNormal}}
	compiled.enabled.Store(*spec.Enabled)
	for _, selector := range spec.Steps {
		if (selector.Type != "security" && selector.Type != "system") || !eventPrefix.MatchString(selector.IDPrefix) || len(selector.Category) > 128 || len(selector.Source) > 128 || len(selector.Message.Regex) > 256 || len(selector.Message.Equals) > 256 || len(selector.Message.Contains) > 256 || !validSeverity(selector.Severity) {
			return nil, errors.New("invalid event selector")
		}
		matches := 0
		for _, value := range []string{selector.Message.Equals, selector.Message.Contains, selector.Message.Regex} {
			if value != "" {
				matches++
			}
		}
		if matches != 1 {
			return nil, errors.New("event selector requires exactly one message matcher")
		}
		var parsed *regexp.Regexp
		if selector.Message.Regex != "" {
			parsed, err = regexp.Compile(selector.Message.Regex)
			if err != nil {
				return nil, err
			}
			if spec.CorrelateBy == "account_tty" && (parsed.SubexpIndex("username") < 0 || parsed.SubexpIndex("tty") < 0) {
				return nil, errors.New("account-and-TTY selector requires both captures")
			}
			if spec.CorrelateBy == "device_identity" && (parsed.SubexpIndex("device_port") < 0 || parsed.SubexpIndex("device_product") < 0) {
				return nil, errors.New("device selector requires port and product captures")
			}
			if parsed.SubexpIndex(spec.CorrelateBy) < 0 && !fieldCorrelation(spec.CorrelateBy) && spec.CorrelateBy != "account_tty" && spec.CorrelateBy != "device_identity" {
				return nil, fmt.Errorf("event regex lacks correlation capture %q", spec.CorrelateBy)
			}
		} else if !fieldCorrelation(spec.CorrelateBy) {
			return nil, fmt.Errorf("message matcher lacks correlation capture %q", spec.CorrelateBy)
		}
		compiled.regexes = append(compiled.regexes, parsed)
	}
	return compiled, nil
}

// fieldCorrelation identifies validated alert fields that can key a non-regex selector.
func fieldCorrelation(name string) bool {
	return name == "event_id" || name == "source" || name == "category"
}

// validSeverity accepts only the logging manager's established severity values.
func validSeverity(value string) bool {
	return value == "" || value == "info" || value == "warning" || value == "error" || value == "critical"
}
