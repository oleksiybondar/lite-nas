package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	contract "lite-nas/shared/contracts/securityautomation"
)

// completeMitigationOperands prompts for missing rule operands and rejects names outside the rule contract.
func completeMitigationOperands(rule contract.Rule, provided map[string]string, ask func(string) (string, error)) (map[string]string, error) {
	if len(rule.RequiredOperands) > 16 {
		return nil, errors.New("rule requires too many operands")
	}
	if provided == nil {
		provided = make(map[string]string, len(rule.RequiredOperands))
	}
	required := make(map[string]bool, len(rule.RequiredOperands))
	for _, raw := range rule.RequiredOperands {
		name := strings.TrimPrefix(raw, ":")
		if !operandNamePattern.MatchString(name) || required[name] {
			return nil, fmt.Errorf("invalid required operand %q", raw)
		}
		required[name] = true
	}
	for name := range provided {
		if !required[name] {
			return nil, fmt.Errorf("unexpected mitigation operand %q", name)
		}
	}
	for _, raw := range rule.RequiredOperands {
		name := strings.TrimPrefix(raw, ":")
		if _, exists := provided[name]; exists {
			continue
		}
		value, err := ask(name + ": ")
		if err != nil {
			return nil, err
		}
		if !validMitigationOperand(value) {
			return nil, fmt.Errorf("invalid value for mitigation operand %q", name)
		}
		provided[name] = value
	}
	return provided, nil
}

// validMitigationOperand applies the CLI transport bounds before the service validates the operand's command position.
func validMitigationOperand(value string) bool {
	if len(value) == 0 || len(value) > 256 || value[0] == '-' {
		return false
	}
	for _, char := range value {
		if char < 32 || char == 127 {
			return false
		}
	}
	return true
}

// promptMitigationOperand reads one bounded value directly from the operator's terminal.
func promptMitigationOperand(label string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("interactive terminal required for missing mitigation operands: %w", err)
	}
	defer tty.Close()
	return readMitigationOperand(tty, tty, label)
}

// readMitigationOperand reads one bounded terminal line for a required rule operand.
func readMitigationOperand(input io.Reader, output io.Writer, label string) (string, error) {
	if _, err := fmt.Fprint(output, label); err != nil {
		return "", err
	}
	var value [257]byte
	for length := 0; ; length++ {
		var character [1]byte
		if _, err := input.Read(character[:]); err != nil {
			return "", err
		}
		if character[0] == '\n' {
			return string(value[:length]), nil
		}
		if length >= 256 {
			return "", errors.New("mitigation operand is too long")
		}
		value[length] = character[0]
	}
}
