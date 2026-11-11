package main

import (
	"os"
	"path/filepath"
	"testing"
)

// mustPolicy creates one writable test policy with consistent kernel-style limits.
func mustPolicy(t *testing.T, root, name, old string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	for field, value := range map[string]string{
		"scaling_max_freq": old, "scaling_min_freq": "600000", "cpuinfo_min_freq": "600000", "cpuinfo_max_freq": "2400000",
	} {
		if err := os.WriteFile(filepath.Join(dir, field), []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// Requirements: operational-safety-automation/CPU-effect.
func TestCPUCapHelperUpdatesEveryPolicy(t *testing.T) {
	root := t.TempDir()
	mustPolicy(t, root, "policy0", "1800000")
	mustPolicy(t, root, "policy4", "1800000")
	if err := runCPUCapHelperAt(root, "1500000", "0"); err != nil {
		t.Fatal(err)
	}
	first, err := readPolicyNumber(filepath.Join(root, "policy0"), "scaling_max_freq")
	if err != nil {
		t.Fatal(err)
	}
	second, err := readPolicyNumber(filepath.Join(root, "policy4"), "scaling_max_freq")
	if err != nil {
		t.Fatal(err)
	}
	if first != 1500000 || second != 1500000 {
		t.Fatalf("policy caps = %d, %d", first, second)
	}
}

// Requirements: operational-safety-automation/CPU-effect.
func TestCPUCapHelperRejectsOptionShapedAndMalformedCaps(t *testing.T) {
	for _, cap := range []string{"--max", "1500000kHz", "-1", "01500000"} {
		t.Run(cap, func(t *testing.T) {
			if err := runCPUCapHelperAt(t.TempDir(), cap, "0"); err == nil {
				t.Fatal("malformed cap accepted")
			}
		})
	}
}

// Requirements: operational-safety-automation/CPU-effect.
func TestCPUCapHelperRefusesUnownedRelaxation(t *testing.T) {
	root := t.TempDir()
	mustPolicy(t, root, "policy0", "1500000")
	if err := runCPUCapHelperAt(root, "1800000", "0"); err == nil {
		t.Fatal("unowned relaxation accepted")
	}
}

// Requirements: operational-safety-automation/CPU-effect.
func TestManualCPUCapHelperAllowsVerifiedRestoreAfterRestart(t *testing.T) {
	root := t.TempDir()
	mustPolicy(t, root, "policy0", "1500000")
	if err := runCPUCapHelperAtMode(root, "1900000", "0", true); err != nil {
		t.Fatal(err)
	}
	got, err := readPolicyNumber(filepath.Join(root, "policy0"), "scaling_max_freq")
	if err != nil {
		t.Fatal(err)
	}
	if got != 1900000 {
		t.Fatalf("manual cap = %d", got)
	}
}

// Requirements: operational-safety-automation/CPU-effect.
func TestCPUCapHelperValidatesEveryPolicyBeforeWriting(t *testing.T) {
	root := t.TempDir()
	mustPolicy(t, root, "policy0", "1800000")
	mustPolicy(t, root, "policy4", "1800000")
	if err := os.WriteFile(filepath.Join(root, "policy4", "cpuinfo_max_freq"), []byte("1200000"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runCPUCapHelperAt(root, "1500000", "0"); err == nil {
		t.Fatal("unsupported policy accepted")
	}
	got, err := readPolicyNumber(filepath.Join(root, "policy0"), "scaling_max_freq")
	if err != nil {
		t.Fatal(err)
	}
	if got != 1800000 {
		t.Fatalf("first policy changed before all were validated: %d", got)
	}
}
