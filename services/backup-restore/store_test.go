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

// Requirements: backup-restore/FR-012.
// TestVersionByNameRequiresExactMatch prevents a case-insensitive name lookup.
func TestVersionByNameRequiresExactMatch(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(source, []byte("settings"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	if _, err := store.BackupWithMetadata(Spec{Application: "example", Selections: []Selection{{Path: source}}}, "Release", "", "admin"); err != nil {
		t.Fatal(err)
	}
	version, err := store.VersionByName("example", "release")
	if err != nil {
		t.Fatal(err)
	}
	if version != "" {
		t.Fatalf("case-mismatched version = %q, want none", version)
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

// Requirements: backup-restore/FR-001.
func TestStoreMetadataReadsSnapshotDetails(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(source, []byte("enabled=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	version, err := store.BackupWithMetadata(Spec{Application: "example", Selections: []Selection{{Path: source}}}, "named", "description", "admin")
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := store.Metadata("example", version)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Name != "named" {
		t.Fatalf("metadata name = %q, want named", metadata.Name)
	}
}

// Requirements: backup-restore/FR-009.
func TestStoreDeleteRemovesSnapshot(t *testing.T) {
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
	if err := store.Delete("example", version); err != nil {
		t.Fatal(err)
	}
	versions, err := store.Versions("example")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 0 {
		t.Fatalf("versions after delete = %v, want none", versions)
	}
}

// Requirements: backup-restore/SR-003.
func TestStoreDeleteRejectsMissingSnapshot(t *testing.T) {
	if err := (Store{Root: t.TempDir()}).Delete("example", "1"); err == nil {
		t.Fatal("missing snapshot deletion succeeded")
	}
}

// Requirements: backup-restore/SR-003.
func TestStoreDeleteRejectsSymlinkArchive(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "example")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "target.tar.gz")
	if err := os.WriteFile(target, []byte("archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(directory, "1.tar.gz")); err != nil {
		t.Fatal(err)
	}
	if err := (Store{Root: root}).Delete("example", "1"); err == nil {
		t.Fatal("symlink archive was deleted")
	}
}

// Requirements: backup-restore/SR-003.
func TestStoreMetadataRejectsInvalidGzip(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "example")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "1.tar.gz"), []byte("invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (Store{Root: root}).Metadata("example", "1"); err == nil {
		t.Fatal("invalid gzip metadata archive was accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestStoreMetadataRejectsMismatchedArchiveIdentity(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "example")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	archive, err := os.Create(filepath.Join(directory, "1.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(archive)
	writer := tar.NewWriter(compressed)
	content := `{"application":"other","version":"1","created_at":"2026-01-01T00:00:00Z"}`
	if err := writer.WriteHeader(&tar.Header{Name: "metadata.json", Typeflag: tar.TypeReg, Mode: 0o600, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := (Store{Root: root}).Metadata("example", "1"); err == nil {
		t.Fatal("mismatched archive identity was accepted")
	}
}
