package backuprestore

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

const ufwArchivePath = "/ufw.json"

// UFWRule stores one UFW-managed CLI rule in replayable operand form.
type UFWRule struct {
	Kind      string   `json:"kind"`
	Action    string   `json:"action"`
	Arguments []string `json:"arguments"`
}

// UFWDefaults stores the three policy directions exposed by the UFW CLI.
type UFWDefaults struct {
	Incoming string `json:"incoming"`
	Outgoing string `json:"outgoing"`
	Routed   string `json:"routed"`
}

// UFWState is the editable, versioned JSON contract for CLI-managed firewall state.
type UFWState struct {
	SchemaVersion int         `json:"schemaVersion"`
	Enabled       bool        `json:"enabled"`
	Logging       string      `json:"logging"`
	Defaults      UFWDefaults `json:"defaults"`
	Rules         []UFWRule   `json:"rules"`
}

// UFWPlugin owns command execution while the backup service owns archive storage.
type UFWPlugin struct {
	run           func(...string) ([]byte, error)
	defaults      string
	configuration string
}

// NewUFWPlugin binds the installed UFW executable to a bounded, shell-free runner.
func NewUFWPlugin() *UFWPlugin {
	return &UFWPlugin{run: runUFWCommand, defaults: "/etc/default/ufw", configuration: "/etc/ufw/ufw.conf"}
}

// runUFWCommand executes a fixed UFW binary with separately validated arguments.
func runUFWCommand(arguments ...string) ([]byte, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("missing UFW command")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "/usr/sbin/ufw", arguments...)
	command.Env = []string{"PATH=/usr/sbin:/usr/bin:/sbin:/bin", "LC_ALL=C"}
	var output limitedUFWOutput
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("ufw %s failed: %w: %s", arguments[0], err, strings.TrimSpace(output.String()))
	}
	if output.exceeded {
		return nil, fmt.Errorf("ufw output exceeds limit")
	}
	return output.Bytes(), nil
}

// limitedUFWOutput prevents an external command from exhausting service memory.
type limitedUFWOutput struct {
	bytes.Buffer
	exceeded bool
}

// Write accepts up to one MiB of UFW output and records overflow.
func (output *limitedUFWOutput) Write(data []byte) (int, error) {
	if output.Len()+len(data) > 1<<20 {
		output.exceeded = true
		return len(data), nil
	}
	return output.Buffer.Write(data)
}

// Export reads UFW-managed rules and policy without mutating the firewall.
func (plugin *UFWPlugin) Export() ([]byte, error) {
	if plugin == nil || plugin.run == nil {
		return nil, fmt.Errorf("UFW plugin is unavailable")
	}
	defaultConfig, err := readBoundedUFWFile(plugin.defaults)
	if err != nil {
		return nil, err
	}
	runtimeConfig, err := readBoundedUFWFile(plugin.configuration)
	if err != nil {
		return nil, err
	}
	state, err := parseUFWSettings(defaultConfig, runtimeConfig)
	if err != nil {
		return nil, err
	}
	output, err := plugin.run("show", "added")
	if err != nil {
		return nil, err
	}
	state.Rules, err = parseUFWAdded(output)
	if err != nil {
		return nil, err
	}
	if err := plugin.validateState(state); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, err
	}
	if len(data)+1 > 1<<20 {
		return nil, fmt.Errorf("UFW JSON exceeds limit")
	}
	return append(data, '\n'), nil
}

// readBoundedUFWFile limits host configuration input before allocation.
func readBoundedUFWFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 1<<20+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("UFW configuration exceeds limit")
	}
	return data, nil
}

// parseUFWSettings reads only CLI-managed defaults and state from UFW files.
func parseUFWSettings(defaultConfig, runtimeConfig []byte) (UFWState, error) {
	defaults, err := parseUFWAssignments(defaultConfig)
	if err != nil {
		return UFWState{}, err
	}
	runtime, err := parseUFWAssignments(runtimeConfig)
	if err != nil {
		return UFWState{}, err
	}
	state := UFWState{SchemaVersion: 1, Rules: []UFWRule{}}
	if state.Defaults.Incoming, err = parseUFWPolicy(defaults["DEFAULT_INPUT_POLICY"]); err != nil {
		return UFWState{}, err
	}
	if state.Defaults.Outgoing, err = parseUFWPolicy(defaults["DEFAULT_OUTPUT_POLICY"]); err != nil {
		return UFWState{}, err
	}
	if state.Defaults.Routed, err = parseUFWPolicy(defaults["DEFAULT_FORWARD_POLICY"]); err != nil {
		return UFWState{}, err
	}
	switch runtime["ENABLED"] {
	case "yes":
		state.Enabled = true
	case "no":
		state.Enabled = false
	default:
		return UFWState{}, fmt.Errorf("unsupported UFW enabled state")
	}
	state.Logging = runtime["LOGLEVEL"]
	if !validUFWLogging(state.Logging) {
		return UFWState{}, fmt.Errorf("unsupported UFW logging level")
	}
	return state, nil
}

// parseUFWAssignments reads literal key/value assignments without evaluating shell code.
func parseUFWAssignments(data []byte) (map[string]string, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
	}
	return values, scanner.Err()
}

// parseUFWPolicy maps UFW's saved policy value to its CLI operand.
func parseUFWPolicy(value string) (string, error) {
	switch value {
	case "ACCEPT":
		return "allow", nil
	case "DROP":
		return "deny", nil
	case "REJECT":
		return "reject", nil
	default:
		return "", fmt.Errorf("unsupported UFW default policy %q", value)
	}
}

// validUFWLogging accepts the levels documented by the UFW CLI.
func validUFWLogging(value string) bool {
	switch value {
	case "off", "low", "medium", "high", "full":
		return true
	default:
		return false
	}
}

// parseUFWAdded parses the CLI's normalized rule report and rejects ambiguity.
func parseUFWAdded(data []byte) ([]UFWRule, error) {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "Added user rules") {
		return nil, fmt.Errorf("unexpected UFW added-rules report")
	}
	rules := make([]UFWRule, 0, len(lines)-1)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "(None)" && len(lines) == 2 {
			return rules, nil
		}
		if !strings.HasPrefix(line, "ufw ") {
			return nil, fmt.Errorf("unexpected UFW rule line")
		}
		words, err := splitUFWCommand(strings.TrimPrefix(line, "ufw "))
		if err != nil {
			return nil, err
		}
		rule, err := ruleFromUFWWords(words)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// splitUFWCommand recognizes whitespace and single-quoted operands from UFW output.
func splitUFWCommand(line string) ([]string, error) {
	var words []string
	var word strings.Builder
	quoted := false
	started := false
	runes := []rune(line)
	for index, char := range runes {
		switch {
		case char == '\'':
			if !quoted && started && word.Len() != 0 || quoted && index+1 < len(runes) && runes[index+1] != ' ' && runes[index+1] != '\t' {
				return nil, fmt.Errorf("ambiguous UFW rule quoting")
			}
			quoted = !quoted
			started = true
		case char == ' ' || char == '\t':
			if quoted {
				word.WriteRune(char)
			} else if started {
				words = append(words, word.String())
				word.Reset()
				started = false
			}
		case char == '\\' || char == '\n' || char == '\r' || char == 0:
			return nil, fmt.Errorf("unsupported UFW rule character")
		default:
			word.WriteRune(char)
			started = true
		}
	}
	if quoted {
		return nil, fmt.Errorf("ambiguous UFW rule quoting")
	}
	if started {
		words = append(words, word.String())
	}
	return words, nil
}

// ruleFromUFWWords converts a reported command to the JSON rule shape.
func ruleFromUFWWords(words []string) (UFWRule, error) {
	rule := UFWRule{Kind: "rule"}
	if len(words) > 0 && words[0] == "route" {
		rule.Kind = "route"
		words = words[1:]
	}
	if len(words) < 2 {
		return UFWRule{}, fmt.Errorf("incomplete UFW rule")
	}
	rule.Action = words[0]
	rule.Arguments = append([]string(nil), words[1:]...)
	if err := validateUFWRule(rule); err != nil {
		return UFWRule{}, err
	}
	return rule, nil
}

// validateUFWRule rejects commands and option-shaped operands outside rule syntax.
func validateUFWRule(rule UFWRule) error {
	if rule.Kind != "rule" && rule.Kind != "route" {
		return fmt.Errorf("unsupported UFW rule kind")
	}
	switch rule.Action {
	case "allow", "deny", "reject", "limit":
	default:
		return fmt.Errorf("unsupported UFW rule action")
	}
	if len(rule.Arguments) == 0 || len(rule.Arguments) > 64 {
		return fmt.Errorf("invalid UFW rule operand count")
	}
	for _, argument := range rule.Arguments {
		if argument == "" || len(argument) > 256 || strings.HasPrefix(argument, "-") || strings.ContainsAny(argument, "\x00\n\r") {
			return fmt.Errorf("invalid UFW rule operand")
		}
	}
	return nil
}

// commandArguments constructs only the two supported CLI rule forms.
func (rule UFWRule) commandArguments() []string {
	arguments := make([]string, 0, len(rule.Arguments)+2)
	if rule.Kind == "route" {
		arguments = append(arguments, "route")
	}
	return append(arguments, append([]string{rule.Action}, rule.Arguments...)...)
}

// decodeUFWState validates a merged JSON document and rejects unknown fields.
func decodeUFWState(data []byte) (UFWState, error) {
	if len(data) == 0 || len(data) > 1<<20 {
		return UFWState{}, fmt.Errorf("invalid UFW JSON size")
	}
	var state UFWState
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return UFWState{}, err
	}
	for _, field := range []string{"schemaVersion", "enabled", "logging", "defaults", "rules"} {
		if len(fields[field]) == 0 || bytes.Equal(fields[field], []byte("null")) {
			return UFWState{}, fmt.Errorf("missing UFW JSON field %s", field)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return UFWState{}, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return UFWState{}, fmt.Errorf("trailing UFW JSON")
	}
	return state, nil
}

// validateState checks the full JSON shape and UFW's rule parser before mutation.
func (plugin *UFWPlugin) validateState(state UFWState) error {
	if plugin == nil || plugin.run == nil || state.SchemaVersion != 1 || !validUFWLogging(state.Logging) || state.Rules == nil || len(state.Rules) > 4096 {
		return fmt.Errorf("invalid UFW JSON state")
	}
	for _, policy := range []string{state.Defaults.Incoming, state.Defaults.Outgoing, state.Defaults.Routed} {
		if policy != "allow" && policy != "deny" && policy != "reject" {
			return fmt.Errorf("invalid UFW default policy")
		}
	}
	seen := make(map[string]bool, len(state.Rules))
	for _, rule := range state.Rules {
		if err := validateUFWRule(rule); err != nil {
			return err
		}
		key, err := json.Marshal(rule)
		if err != nil {
			return err
		}
		if seen[string(key)] {
			return fmt.Errorf("duplicate UFW rule")
		}
		seen[string(key)] = true
		if _, err := plugin.run(append([]string{"--dry-run"}, rule.commandArguments()...)...); err != nil {
			return fmt.Errorf("UFW rejected rule: %w", err)
		}
	}
	return nil
}

// Apply replaces CLI-managed rules and defaults after complete prevalidation.
func (plugin *UFWPlugin) Apply(data []byte) error {
	state, err := decodeUFWState(data)
	if err != nil {
		return err
	}
	if err := plugin.validateState(state); err != nil {
		return err
	}
	currentData, err := plugin.Export()
	if err != nil {
		return err
	}
	current, err := decodeUFWState(currentData)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(state, current) {
		return nil
	}
	if err := plugin.replace(state, current); err != nil {
		liveData, exportErr := plugin.Export()
		if exportErr != nil {
			return fmt.Errorf("UFW apply failed: %w; rollback inspection failed: %v", err, exportErr)
		}
		live, decodeErr := decodeUFWState(liveData)
		if decodeErr != nil {
			return fmt.Errorf("UFW apply failed: %w; rollback decode failed: %v", err, decodeErr)
		}
		rollbackErr := plugin.replace(current, live)
		if rollbackErr != nil {
			return fmt.Errorf("UFW apply failed: %w; rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("UFW apply failed and prior rules were restored: %w", err)
	}
	return nil
}

// Merge restores snapshot rules while preserving live rules absent from the snapshot.
func (plugin *UFWPlugin) Merge(data []byte) error {
	state, err := decodeUFWState(data)
	if err != nil {
		return err
	}
	if err := plugin.validateState(state); err != nil {
		return err
	}
	currentData, err := plugin.Export()
	if err != nil {
		return err
	}
	current, err := decodeUFWState(currentData)
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(state.Rules))
	for _, rule := range state.Rules {
		key, err := json.Marshal(rule)
		if err != nil {
			return err
		}
		seen[string(key)] = true
	}
	for _, rule := range current.Rules {
		key, err := json.Marshal(rule)
		if err != nil {
			return err
		}
		if !seen[string(key)] {
			state.Rules = append(state.Rules, rule)
			seen[string(key)] = true
		}
	}
	merged, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return plugin.Apply(merged)
}

// replace replays one validated state while deleting only CLI-managed rules.
func (plugin *UFWPlugin) replace(state, current UFWState) error {
	for index := len(current.Rules) - 1; index >= 0; index-- {
		rule := current.Rules[index]
		arguments := []string{"--force"}
		if rule.Kind == "route" {
			arguments = append(arguments, "route", "delete")
		} else {
			arguments = append(arguments, "delete")
		}
		arguments = append(arguments, rule.Action)
		arguments = append(arguments, rule.Arguments...)
		if _, err := plugin.run(arguments...); err != nil {
			return fmt.Errorf("UFW rule replacement stopped during deletion: %w", err)
		}
	}
	for _, direction := range []struct{ name, policy string }{{"incoming", state.Defaults.Incoming}, {"outgoing", state.Defaults.Outgoing}, {"routed", state.Defaults.Routed}} {
		if _, err := plugin.run("default", direction.policy, direction.name); err != nil {
			return fmt.Errorf("UFW rule replacement stopped during policy update: %w", err)
		}
	}
	for _, rule := range state.Rules {
		if _, err := plugin.run(rule.commandArguments()...); err != nil {
			return fmt.Errorf("UFW rule replacement stopped during rule addition: %w", err)
		}
	}
	if _, err := plugin.run("logging", state.Logging); err != nil {
		return fmt.Errorf("UFW rule replacement stopped during logging update: %w", err)
	}
	command := "disable"
	if state.Enabled {
		command = "enable"
	}
	if _, err := plugin.run("--force", command); err != nil {
		return fmt.Errorf("UFW rule replacement stopped during state update: %w", err)
	}
	return nil
}
