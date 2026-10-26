package main

import (
	"testing"

	"lite-nas/shared/fileio"
)

// TestPackagedWatchdogPolicyLoads verifies the shipped protected-unit list and
// policy sections stay in sync with the parser before any privileged start.
// Requirements: task.md/SVC-005.
func TestPackagedWatchdogPolicyLoads(t *testing.T) {
	reader, err := fileio.NewFileReader("../../configs/etc/lite-nas/service-watchdog.conf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := loadWatchdogConfig(reader); err != nil {
		t.Fatalf("load packaged policy: %v", err)
	}
}
