package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	contract "lite-nas/shared/contracts/securityautomation"
)

// Requirements: security-automation/IR-001.
// TestCompleteMitigationOperandsPromptsForMissingValues checks rule-driven prompting.
func TestCompleteMitigationOperandsPromptsForMissingValues(t *testing.T) {
	rule := contract.Rule{RequiredOperands: []string{":requester_ip_address", ":username"}}
	provided := map[string]string{"requester_ip_address": "192.0.2.1"}
	var prompt string
	operands, err := completeMitigationOperands(rule, provided, func(label string) (string, error) {
		prompt = label
		return "alice", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if prompt != "username: " || operands["username"] != "alice" {
		t.Fatalf("prompt %q, operands %#v", prompt, operands)
	}
}

// Requirements: security-automation/IR-001.
// TestReadMitigationOperandAcceptsBoundedLine checks the displayed prompt and accepted value.
func TestReadMitigationOperandAcceptsBoundedLine(t *testing.T) {
	var output bytes.Buffer
	value, err := readMitigationOperand(strings.NewReader("alice\n"), &output, "username: ")
	if err != nil {
		t.Fatal(err)
	}
	if value != "alice" || output.String() != "username: " {
		t.Fatalf("prompt %q, value %q", output.String(), value)
	}
}

// Requirements: security-automation/IR-001.
// TestReadMitigationOperandRejectsBadInput checks terminal failures and input bounds.
func TestReadMitigationOperandRejectsBadInput(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		output bool
	}{
		{"closed input", "", false},
		{"oversized input", strings.Repeat("x", 257) + "\n", false},
		{"closed output", "alice\n", true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			var writer io.Writer = &output
			if test.output {
				writer = errorWriter{}
			}
			_, err := readMitigationOperand(strings.NewReader(test.input), writer, "username: ")
			if err == nil {
				t.Fatal("invalid terminal input accepted")
			}
		})
	}
}

// errorWriter models an unavailable terminal output stream.
type errorWriter struct{}

// Write fails a prompt before any operator input is consumed.
func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("output closed") }

// Requirements: security-automation/IR-001.
// TestCompleteMitigationOperandsRejectsInvalidValues checks invalid inputs never reach the effect RPC.
func TestCompleteMitigationOperandsRejectsInvalidValues(t *testing.T) {
	cases := []struct {
		name     string
		required []string
		provided map[string]string
		answer   string
	}{
		{"unknown", []string{"username"}, map[string]string{"device_id": "abc"}, ""},
		{"option value", []string{"username"}, nil, "--remove"},
		{"blank value", []string{"username"}, nil, ""},
		{"control value", []string{"username"}, nil, "a\nb"},
		{"duplicate rule operand", []string{"username", ":username"}, nil, "alice"},
		{"invalid rule operand", []string{"--bad"}, nil, "alice"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := completeMitigationOperands(contract.Rule{RequiredOperands: test.required}, test.provided, func(string) (string, error) { return test.answer, nil })
			if err == nil {
				t.Fatal("invalid mitigation operands accepted")
			}
		})
	}
}

// Requirements: security-automation/IR-001.
// TestCompleteMitigationOperandsPropagatesPromptFailure checks failed terminal input stops the command.
func TestCompleteMitigationOperandsPropagatesPromptFailure(t *testing.T) {
	expected := errors.New("terminal closed")
	_, err := completeMitigationOperands(contract.Rule{RequiredOperands: []string{"username"}}, nil, func(string) (string, error) { return "", expected })
	if !errors.Is(err, expected) {
		t.Fatalf("prompt error = %v", err)
	}
}
