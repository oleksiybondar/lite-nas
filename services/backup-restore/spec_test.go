package backuprestore

import (
	"os"
	"path/filepath"
	"testing"
)

// Requirements: backup-restore/FR-002.
func TestLoadSpecsLoadsOnlyConfigurationFiles(t *testing.T) {
	directory := t.TempDir()
	configuration := filepath.Join(directory, "example.conf")
	content := "[backup]\nplugin=file\n[file.main]\npath=/etc/example.conf\n"
	if err := os.WriteFile(configuration, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "notes.txt"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(directory, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	specs, err := LoadSpecs(directory)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(specs); got != 1 {
		t.Fatalf("loaded spec count = %d, want 1", got)
	}
}

// Requirements: backup-restore/SR-003.
func TestLoadSpecsRejectsInvalidApplicationFilename(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "bad_name.conf"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpecs(directory); err == nil {
		t.Fatal("invalid application filename was accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestLoadSpecRejectsUnsupportedPlugin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.conf")
	content := "[backup]\nplugin=command\n[file.main]\npath=/etc/example.conf\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpec(path, "example"); err == nil {
		t.Fatal("unsupported plugin was accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestLoadSpecRejectsSelectionWithoutAbsolutePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.conf")
	content := "[backup]\nplugin=file\n[file.main]\npath=relative.conf\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpec(path, "example"); err == nil {
		t.Fatal("relative selection path was accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestLoadSpecRejectsInvalidSelectionDefinitions(t *testing.T) {
	cases := []struct {
		name, selections string
	}{
		{name: "missing selection", selections: ""},
		{name: "unknown section", selections: "[other]\npath=/etc/example.conf\n"},
		{name: "duplicate path", selections: "[file.one]\npath=/etc/example.conf\n[file.two]\npath=/etc/example.conf\n"},
		{name: "filesystem root", selections: "[file.main]\npath=/\n"},
		{name: "overlapping directory", selections: "[directory.main]\npath=/etc/example.d\n[file.child]\npath=/etc/example.d/child.conf\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "app.conf")
			content := "[backup]\nplugin=file\n" + testCase.selections
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadSpec(path, "example"); err == nil {
				t.Fatal("invalid selection definitions were accepted")
			}
		})
	}
}

// Requirements: backup-restore/SR-003.
func TestValidApplicationRejectsOptionShapedName(t *testing.T) {
	if validApplication("--backup") {
		t.Fatal("option-shaped application name was accepted")
	}
}
