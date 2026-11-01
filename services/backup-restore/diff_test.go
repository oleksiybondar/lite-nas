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

// Requirements: backup-restore/FR-010.
func TestStoreDiffReportsSnapshotOnlyAndLiveOnlyDirectoryEntries(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "settings.d")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	snapshotPath := filepath.Join(directory, "snapshot.conf")
	if err := os.WriteFile(snapshotPath, []byte("saved=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	spec := Spec{Application: "example", Selections: []Selection{{Path: directory, Directory: true}}}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(snapshotPath); err != nil {
		t.Fatal(err)
	}
	livePath := filepath.Join(directory, "live.conf")
	if err := os.WriteFile(livePath, []byte("live=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	differences, err := store.Diff(spec, version)
	if err != nil {
		t.Fatal(err)
	}
	if len(differences) != 3 || differences[0].Status != "changed" || differences[1].Status != "live-only" || differences[2].Status != "snapshot-only" {
		t.Fatalf("directory differences = %+v", differences)
	}
}
