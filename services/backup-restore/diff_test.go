package backuprestore

import (
	"os"
	"path/filepath"
	"testing"
)

// Requirements: backup-restore/FR-010.
func TestStoreDiffReportsChangedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	spec := Spec{Application: "example", Selections: []Selection{{Path: path}}}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("after"), 0o600); err != nil {
		t.Fatal(err)
	}
	differences, err := store.Diff(spec, version)
	if err != nil {
		t.Fatal(err)
	}
	if len(differences) != 1 || differences[0].Path != path || differences[0].Status != "changed" {
		t.Fatalf("differences = %v", differences)
	}
}

// Requirements: backup-restore/FR-010.
func TestStoreDiffOmitsUnchangedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	spec := Spec{Application: "example", Selections: []Selection{{Path: path}}}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	differences, err := store.Diff(spec, version)
	if err != nil {
		t.Fatal(err)
	}
	if len(differences) != 0 {
		t.Fatalf("differences = %v, want none", differences)
	}
}
