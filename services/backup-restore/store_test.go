package backuprestore

import (
	"os"
	"path/filepath"
	"testing"
)

// Requirements: backup-restore/FR-001, backup-restore/FR-002.
func TestStoreBackupCreatesListedVersion(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(source, []byte("enabled=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	version, err := store.Backup(Spec{Application: "example", Selections: []Selection{{Path: source}}})
	if err != nil {
		t.Fatal(err)
	}
	versions, err := store.Versions("example")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 || versions[0] != version {
		t.Fatalf("versions = %v, want [%s]", versions, version)
	}
}

// Requirements: backup-restore/FR-009.
func TestBackedUpApplicationsListsStoredHistory(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(source, []byte("enabled=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	if _, err := store.Backup(Spec{Application: "example", Selections: []Selection{{Path: source}}}); err != nil {
		t.Fatal(err)
	}
	applications, err := store.BackedUpApplications()
	if err != nil {
		t.Fatal(err)
	}
	if len(applications) != 1 || applications[0] != "example" {
		t.Fatalf("backed-up applications = %v", applications)
	}
}
