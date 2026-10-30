package main

import "testing"

// Requirements: backup-restore/SR-003.
func TestValidApplicationRejectsUnsafeOperands(t *testing.T) {
	cases := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "normal", value: "lite-nas", valid: true},
		{name: "option shaped", value: "--all", valid: false},
		{name: "path", value: "../etc", valid: false},
		{name: "empty", value: "", valid: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := validApplication(testCase.value); got != testCase.valid {
				t.Fatalf("validApplication(%q) = %v", testCase.value, got)
			}
		})
	}
}

// Requirements: backup-restore/SR-003.
func TestValidVersionRejectsUnsafeOperands(t *testing.T) {
	cases := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "normal", value: "20261030T120000.000000000Z", valid: true},
		{name: "option shaped", value: "--latest", valid: false},
		{name: "path", value: "../etc", valid: false},
		{name: "empty", value: "", valid: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := validVersion(testCase.value); got != testCase.valid {
				t.Fatalf("validVersion(%q) = %v", testCase.value, got)
			}
		})
	}
}
