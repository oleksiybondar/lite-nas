package main

import (
	"testing"

	"gopkg.in/ini.v1"

	"lite-nas/shared/fileio"
)

// TestPackagedNetworkSecurityConfigLoads verifies the shipped INI policy and
// shared transport sections satisfy startup validation.
// Requirements: network-security/FR-002.
func TestPackagedNetworkSecurityConfigLoads(t *testing.T) {
	reader, err := fileio.NewFileReader("../../configs/etc/lite-nas/network-security.conf")
	if err != nil {
		t.Fatalf("open packaged config: %v", err)
	}
	_, err = loadSecurityConfig(reader)
	if err != nil {
		t.Fatalf("load packaged config: %v", err)
	}
}

// TestPatternConfigRejectsInvalidPolicy checks startup validation for new FSMs.
// Requirements: network-security/FR-005, network-security/FR-006.
func TestPatternConfigRejectsInvalidPolicy(t *testing.T) {
	for _, section := range []string{
		"[horizontal_scan]\nthreshold=1\n",
		"[horizontal_scan]\nwindow=0s\n",
		"[repeated_attempts]\nmax_pairs=0\n",
		"[repeated_attempts]\ncooldown=bad\n",
	} {
		t.Run(section, func(t *testing.T) {
			file, err := ini.Load([]byte(section))
			if err != nil {
				t.Fatalf("load fixture: %v", err)
			}
			name := "horizontal_scan"
			if file.HasSection("repeated_attempts") {
				name = "repeated_attempts"
			}
			if _, err := loadPatternConfig(file, name); err == nil {
				t.Fatal("invalid detector policy was accepted")
			}
		})
	}
}
