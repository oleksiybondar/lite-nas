package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeIOModuleFixtureFile(t *testing.T, path string, data string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", filepath.Base(path), err)
	}
}

func loadIOModuleFixture(t *testing.T) IO {
	t.Helper()

	baseDir := t.TempDir()
	cpuPath := filepath.Join(baseDir, "stat")
	memPath := filepath.Join(baseDir, "meminfo")

	writeIOModuleFixtureFile(t, cpuPath, "cpu data")
	writeIOModuleFixtureFile(t, memPath, "mem data")

	module, err := NewIOModule(cpuPath, memPath)
	if err != nil {
		t.Fatalf("NewIOModule() error = %v", err)
	}

	return module
}
