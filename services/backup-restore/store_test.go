package backuprestore

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Requirements: backup-restore/FR-001.
func TestBackupMetadataUsesAllocatedVersionAndCaller(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(source, []byte("enabled=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	version, err := store.BackupWithMetadata(Spec{Application: "example", Selections: []Selection{{Path: source}}, RestartUnits: []string{"example.service"}}, "named", "description", "admin")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := store.OpenVersion("example", version)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	compressed, err := gzip.NewReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	header, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "metadata.json" {
		t.Fatalf("first member = %q", header.Name)
	}
	var metadata Metadata
	if err := json.NewDecoder(reader).Decode(&metadata); err != nil {
		t.Fatal(err)
	}
	if metadata.Application != "example" || metadata.Version != version || metadata.Name != "named" || metadata.Description != "description" || metadata.CreatedBy != "admin" || metadata.CreatedAt == "" || len(metadata.RestartUnits) != 1 || metadata.RestartUnits[0] != "example.service" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

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
