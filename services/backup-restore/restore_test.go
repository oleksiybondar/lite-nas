package backuprestore

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// Requirements: backup-restore/FR-004.
func TestStoreRestoreReplacesSelectedFile(t *testing.T) {
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
	if err := store.Restore(spec, version, ReplaceMode); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "before" {
		t.Fatalf("restored content = %q", content)
	}
}

// Requirements: backup-restore/FR-004, backup-restore/SR-003.
func TestRestoreRejectsWrongDirectoryArchiveTypeBeforeDeleting(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "settings")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(directory, "keep.conf")
	if err := os.WriteFile(extra, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	archiveDir := filepath.Join(store.Root, "example")
	if err := os.MkdirAll(archiveDir, 0o700); err != nil {
		t.Fatal(err)
	}
	archive, err := os.Create(filepath.Join(archiveDir, "1.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(archive)
	written := tar.NewWriter(compressed)
	if err := written.WriteHeader(&tar.Header{Name: directory[1:], Typeflag: tar.TypeReg, Mode: 0o600}); err != nil {
		t.Fatal(err)
	}
	if err := written.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	spec := Spec{Application: "example", Selections: []Selection{{Path: directory, Directory: true}}}
	_ = store.Restore(spec, "1", ReplaceMode)
	if _, err := os.Lstat(extra); err != nil {
		t.Fatalf("live file was deleted: %v", err)
	}
}

// Requirements: backup-restore/FR-004.
func TestStoreRestoreRemovesAbsentDirectoryEntry(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "settings")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	spec := Spec{Application: "example", Selections: []Selection{{Path: directory, Directory: true}}}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(directory, "extra.conf")
	if err := os.WriteFile(extra, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Restore(spec, version, ReplaceMode); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(extra); !os.IsNotExist(err) {
		t.Fatalf("extra file remains: %v", err)
	}
}

// Requirements: backup-restore/FR-004.
func TestStoreRestoreSymlinkChain(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.conf")
	middle := filepath.Join(root, "middle.conf")
	link := filepath.Join(root, "linked.conf")
	if err := os.WriteFile(target, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, middle); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(middle, link); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	spec := Spec{Application: "example", Selections: []Selection{{Path: link}}}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("after"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Restore(spec, version, ReplaceMode); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(link)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "before" {
		t.Fatalf("restored linked content = %q", content)
	}
}
