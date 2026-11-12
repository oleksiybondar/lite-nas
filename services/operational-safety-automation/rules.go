package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	detection "lite-nas/shared/automationdetection"
	contract "lite-nas/shared/contracts/operationalsafety"
)

var (
	ruleIDPattern       = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	prefixPattern       = regexp.MustCompile(`^[A-Za-z0-9_]{1,20}$`)
	outputPrefixPattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,10}$`)
)

// selector matches an alert ID and any configured conjunction of message, value, or state fields.
type selector struct {
	IDPrefix string `json:"id_prefix"`
	Category string `json:"category,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  struct {
		Equals   string `json:"equals,omitempty"`
		Contains string `json:"contains,omitempty"`
		Regex    string `json:"regex,omitempty"`
	} `json:"message,omitempty"`
	Value struct {
		NumGTE     *float64 `json:"num_gte,omitempty"`
		NumLTE     *float64 `json:"num_lte,omitempty"`
		TextEquals string   `json:"text_equals,omitempty"`
		BoolEquals *bool    `json:"bool_equals,omitempty"`
	} `json:"value,omitempty"`
	State string                          `json:"state,omitempty"`
	Level contract.OperationalSafetyLevel `json:"level,omitempty"`
}

// ruleSpec selects an emit or CPU frequency action after a configured FSM match.
type ruleSpec struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Enabled     *bool      `json:"enabled"`
	CorrelateBy string     `json:"correlate_by"`
	Steps       []selector `json:"steps"`
	FSM         struct {
		Type      detection.PluginName `json:"type"`
		Window    string               `json:"window"`
		Threshold int                  `json:"threshold,omitempty"`
		MaxKeys   int                  `json:"max_keys"`
		MaxIdle   string               `json:"max_idle"`
	} `json:"fsm,omitempty"`
	Action struct {
		Kind       string                                     `json:"kind"`
		Cooldown   string                                     `json:"cooldown,omitempty"`
		CapsKHz    map[contract.OperationalSafetyLevel]uint64 `json:"caps_khz,omitempty"`
		Level      contract.OperationalSafetyLevel            `json:"level,omitempty"`
		ActiveKHz  uint64                                     `json:"active_khz,omitempty"`
		NormalKHz  uint64                                     `json:"normal_khz,omitempty"`
		ParentRule string                                     `json:"parent_rule,omitempty"`
	} `json:"action"`
	Output struct {
		IDPrefix string `json:"id_prefix"`
		Category string `json:"category"`
		Severity string `json:"severity"`
		Message  string `json:"message"`
	} `json:"output"`
}

// compiledRule holds startup-validated selectors and an initialized detector.
type compiledRule struct {
	spec     ruleSpec
	patterns []*regexp.Regexp
	recovery *detection.StagedRecovery
	stage    *detection.ChainStage
	detector *detection.Plugin
	cooldown time.Duration
	enabled  bool
}

// loadRules accepts an empty packaged rule set and validates each configured instance.
func loadRules(path string) (map[string]*compiledRule, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(data) > 1024*1024 {
		return nil, nil, errors.New("rules file exceeds 1 MiB")
	}
	var file struct {
		Rules []ruleSpec `json:"rules"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return nil, nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, nil, errors.New("trailing rules JSON")
	}
	if len(file.Rules) > 128 {
		return nil, nil, errors.New("too many operational safety rules")
	}
	rules := make(map[string]*compiledRule, len(file.Rules))
	order := make([]string, 0, len(file.Rules))
	for _, spec := range file.Rules {
		if !ruleIDPattern.MatchString(spec.ID) || rules[spec.ID] != nil || spec.Enabled == nil || len(spec.Name) == 0 || len(spec.Name) > 128 || len(spec.Steps) == 0 || len(spec.Steps) > 64 {
			return nil, nil, fmt.Errorf("invalid operational safety rule %q", spec.ID)
		}
		rule, err := compileRule(spec)
		if err != nil {
			return nil, nil, fmt.Errorf("rule %q: %w", spec.ID, err)
		}
		rules[spec.ID] = rule
		order = append(order, spec.ID)
	}
	if err := validateRuleGraph(rules); err != nil {
		return nil, nil, err
	}
	if err := validateStageGraph(rules); err != nil {
		return nil, nil, err
	}
	sort.Strings(order)
	return rules, order, nil
}

// validateStageGraph ties each protective stage to the immediately less restrictive rule.
func validateStageGraph(rules map[string]*compiledRule) error {
	outputPrefixes := make(map[string]string, len(rules))
	for id, rule := range rules {
		prefix := rule.spec.Output.IDPrefix
		if previous := outputPrefixes[prefix]; previous != "" {
			return fmt.Errorf("rules %q and %q share output prefix %q", previous, id, prefix)
		}
		outputPrefixes[prefix] = id
		if rule.stage == nil || rule.spec.Action.Level == contract.LevelWarning {
			continue
		}
		parent := rules[rule.spec.Action.ParentRule]
		if parent == nil || parent.stage == nil || parent.spec.Action.ActiveKHz != rule.spec.Action.NormalKHz {
			return fmt.Errorf("CPU stage %q has no compatible parent", id)
		}
		if rule.spec.Action.Level == contract.LevelError && parent.spec.Action.Level != contract.LevelWarning || rule.spec.Action.Level == contract.LevelCritical && parent.spec.Action.Level != contract.LevelError {
			return fmt.Errorf("CPU stage %q has the wrong parent level", id)
		}
	}
	return nil
}

// compileRule validates selectors, effects, and all configured FSM instances at startup.
func compileRule(spec ruleSpec) (*compiledRule, error) {
	return compileRuleWithClock(spec, time.Now)
}

// compileRuleWithClock shares the runtime clock with detector instances in deterministic tests.
func compileRuleWithClock(spec ruleSpec, clock func() time.Time) (*compiledRule, error) {
	if !outputPrefixPattern.MatchString(spec.Output.IDPrefix) || !printASCII(spec.Output.Category) || len(spec.Output.Category) > 128 || strings.TrimSpace(spec.Output.Message) == "" || len(spec.Output.Message) > 256 || !validSeverity(spec.Output.Severity) {
		return nil, errors.New("invalid output alert")
	}
	if spec.CorrelateBy != "event_id" && spec.CorrelateBy != "source" && spec.CorrelateBy != "category" && !ruleIDPattern.MatchString(spec.CorrelateBy) {
		return nil, errors.New("invalid correlation field")
	}
	rule := &compiledRule{spec: spec, enabled: *spec.Enabled}
	for _, step := range spec.Steps {
		if !prefixPattern.MatchString(step.IDPrefix) || len(step.Category) > 128 || len(step.Source) > 128 || len(step.Message.Equals) > 256 || len(step.Message.Contains) > 256 || len(step.Message.Regex) > 256 || !validState(step.State) || !validLevel(step.Level) {
			return nil, errors.New("invalid alert selector")
		}
		if step.Value.NumGTE != nil && step.Value.NumLTE != nil && *step.Value.NumGTE > *step.Value.NumLTE {
			return nil, errors.New("inverted numeric value range")
		}
		var pattern *regexp.Regexp
		if step.Message.Regex != "" {
			var err error
			pattern, err = regexp.Compile(step.Message.Regex)
			if err != nil {
				return nil, err
			}
			if spec.CorrelateBy != "event_id" && spec.CorrelateBy != "source" && spec.CorrelateBy != "category" && pattern.SubexpIndex(spec.CorrelateBy) < 0 {
				return nil, errors.New("regex lacks correlation capture")
			}
		} else if spec.CorrelateBy != "event_id" && spec.CorrelateBy != "source" && spec.CorrelateBy != "category" {
			return nil, errors.New("correlation capture requires regex")
		}
		rule.patterns = append(rule.patterns, pattern)
	}
	switch spec.Action.Kind {
	case "emit", "emit_normal":
		if spec.Action.Cooldown != "" || len(spec.Action.CapsKHz) != 0 || spec.Action.Level != "" || spec.Action.ActiveKHz != 0 || spec.Action.NormalKHz != 0 || spec.Action.ParentRule != "" {
			return nil, errors.New("emit rule cannot set CPU policy")
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
		if err := detection.ValidateConfig(config, clock); err != nil {
			return nil, err
		}
		if spec.FSM.Type == detection.CountWithinWindow && len(spec.Steps) != 1 {
			return nil, errors.New("count detector needs one step")
		}
		rule.detector, err = detection.New(config, clock)
		if err != nil {
			return nil, err
		}
	case "cpu_frequency":
		if spec.Action.Level != "" || spec.Action.ActiveKHz != 0 || spec.Action.NormalKHz != 0 || spec.Action.ParentRule != "" {
			return nil, errors.New("CPU frequency rule cannot set chain stage fields")
		}
		if len(spec.Steps) != 1 {
			return nil, errors.New("CPU frequency rule needs one trigger selector")
		}
		if spec.Steps[0].Level == contract.LevelNormal {
			return nil, errors.New("CPU trigger must request a protective level")
		}
		var err error
		rule.cooldown, err = time.ParseDuration(spec.Action.Cooldown)
		if err != nil || rule.cooldown <= 0 {
			return nil, errors.New("invalid recovery cooldown")
		}
		rule.recovery, err = detection.NewStagedRecovery(rule.cooldown)
		if err != nil {
			return nil, err
		}
		caps := spec.Action.CapsKHz
		if len(caps) != 4 || caps[contract.LevelNormal] == 0 || caps[contract.LevelWarning] == 0 || caps[contract.LevelError] == 0 || caps[contract.LevelCritical] == 0 || caps[contract.LevelNormal] < caps[contract.LevelWarning] || caps[contract.LevelWarning] < caps[contract.LevelError] || caps[contract.LevelError] < caps[contract.LevelCritical] {
			return nil, errors.New("CPU caps must be positive and ordered normal >= warning >= error >= critical")
		}
	case "cpu_stage":
		if len(spec.Steps) == 0 || spec.Action.ActiveKHz == 0 || spec.Action.NormalKHz <= spec.Action.ActiveKHz || len(spec.Action.CapsKHz) != 0 {
			return nil, errors.New("invalid CPU stage caps or selectors")
		}
		if spec.Action.Level != contract.LevelWarning && spec.Action.Level != contract.LevelError && spec.Action.Level != contract.LevelCritical {
			return nil, errors.New("invalid CPU stage level")
		}
		if spec.Action.Level == contract.LevelWarning && spec.Action.ParentRule != "" || spec.Action.Level != contract.LevelWarning && !ruleIDPattern.MatchString(spec.Action.ParentRule) {
			return nil, errors.New("invalid CPU stage parent")
		}
		for _, step := range spec.Steps {
			if step.Level != "" {
				return nil, errors.New("CPU stage level belongs to its action")
			}
		}
		var err error
		rule.cooldown, err = time.ParseDuration(spec.Action.Cooldown)
		if err != nil || rule.cooldown <= 0 {
			return nil, errors.New("invalid CPU stage cooldown")
		}
		rule.stage, err = detection.NewChainStage(rule.cooldown)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown action kind")
	}
	return rule, nil
}

// printASCII follows the System Logging Manager's category character contract.
func printASCII(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 32 || value[index] > 126 {
			return false
		}
	}
	return true
}

// validSeverity accepts only System Logging Manager severity values.
func validSeverity(value string) bool {
	return value == "info" || value == "warning" || value == "error" || value == "critical"
}

// validState accepts the existing system alert lifecycle values.
func validState(value string) bool {
	return value == "" || value == "high" || value == "low" || value == "normal" || value == "active" || value == "failure"
}

// validLevel accepts only configured CPU effect levels.
func validLevel(value contract.OperationalSafetyLevel) bool {
	return value == "" || value == contract.LevelNormal || value == contract.LevelWarning || value == contract.LevelError || value == contract.LevelCritical
}

// validateRuleGraph prevents emitted alerts from recursively triggering their own rule or a cycle.
func validateRuleGraph(rules map[string]*compiledRule) error {
	edges := make(map[string][]string, len(rules))
	for sourceID, source := range rules {
		for targetID, target := range rules {
			for _, step := range target.spec.Steps {
				if strings.HasPrefix(source.spec.Output.IDPrefix, step.IDPrefix) || strings.HasPrefix(step.IDPrefix, source.spec.Output.IDPrefix) {
					edges[sourceID] = append(edges[sourceID], targetID)
					break
				}
			}
		}
	}
	marks := make(map[string]uint8)
	var visit func(string) bool
	visit = func(id string) bool {
		if marks[id] == 1 {
			return true
		}
		if marks[id] == 2 {
			return false
		}
		marks[id] = 1
		for _, next := range edges[id] {
			if visit(next) {
				return true
			}
		}
		marks[id] = 2
		return false
	}
	for id := range rules {
		if visit(id) {
			return errors.New("operational safety alert rules form a publication cycle")
		}
	}
	return nil
}
