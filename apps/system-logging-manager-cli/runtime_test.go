package main

import (
	"testing"

	"lite-nas/shared/testutil/loggingclitest"
)

// TestMainHelp checks the system CLI's executable help output.
func TestMainHelp(t *testing.T) {
	loggingclitest.AssertMainHelp(t, "system-logging-manager-cli", main)
}

// TestRunHelp checks that the system CLI handles help before infra setup.
func TestRunHelp(t *testing.T) {
	loggingclitest.AssertRunHelp(t, run)
}

// TestRunRejectsUnknownOption checks the system CLI's argument boundary.
func TestRunRejectsUnknownOption(t *testing.T) {
	loggingclitest.AssertUnknownOption(t, run)
}

// TestLoadInfraRejectsMissingConfig checks the system CLI's config boundary.
func TestLoadInfraRejectsMissingConfig(t *testing.T) {
	loggingclitest.AssertMissingConfig(t, appName, loadInfra)
}
