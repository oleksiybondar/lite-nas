// Package loggingclitest provides reusable checks for logging CLI wrappers.
package loggingclitest

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lite-nas/shared/loggingmanagercli"
)

// AssertMainHelp checks that a CLI entry point prints its app-specific usage and returns normally.
func AssertMainHelp(t *testing.T, appName string, mainFn func()) {
	t.Helper()
	previousArgs := os.Args
	previousStdout := os.Stdout
	output, err := os.CreateTemp(t.TempDir(), "help-output-")
	if err != nil {
		t.Fatalf("create help output: %v", err)
	}
	os.Args = []string{appName, "--help"}
	os.Stdout = output
	t.Cleanup(func() {
		os.Args = previousArgs
		os.Stdout = previousStdout
		_ = output.Close()
	})

	mainFn()
	contents, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatalf("read help output: %v", err)
	}
	if !strings.Contains(string(contents), "Usage: "+appName) {
		t.Fatalf("help output = %q", contents)
	}
}

// AssertRunHelp checks that help bypasses messaging setup and signals normal CLI termination.
func AssertRunHelp(t *testing.T, runFn func(context.Context, []string) error) {
	t.Helper()
	err := runFn(context.Background(), []string{"--help"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run(--help) error = %v, want context.Canceled", err)
	}
}

// AssertUnknownOption checks that argument validation rejects an unknown option.
func AssertUnknownOption(t *testing.T, runFn func(context.Context, []string) error) {
	t.Helper()
	err := runFn(context.Background(), []string{"--unknown"})
	if err == nil || !strings.Contains(err.Error(), "unknown argument") {
		t.Fatalf("run(--unknown) error = %v, want unknown argument", err)
	}
}

// AssertMissingConfig checks that an app's infra adapter reports an absent config file.
func AssertMissingConfig(t *testing.T, appName string, loadFn func(string, string) (func(), loggingmanagercli.MessagingClient, error)) {
	t.Helper()
	_, _, err := loadFn(filepath.Join(t.TempDir(), "missing.ini"), appName)
	if err == nil {
		t.Fatal("loadInfra(missing config) succeeded")
	}
}
