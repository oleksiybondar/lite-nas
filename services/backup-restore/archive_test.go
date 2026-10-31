package backuprestore

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Requirements: backup-restore/FR-002.
func TestWriteArchiveIncludesSymlinkAndTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.conf")
	link := filepath.Join(root, "linked.conf")
	if err := os.WriteFile(target, []byte("value=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	if err := WriteArchive(&archive, Spec{Selections: []Selection{{Path: link}}}); err != nil {
		t.Fatal(err)
	}
	paths, err := ArchivePaths(&archive)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{strings.TrimPrefix(link, "/"), strings.TrimPrefix(target, "/")}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("archive paths = %v, want %v", paths, want)
	}
}

// Requirements: backup-restore/FR-002.
func TestLoadSpecAcceptsFileAndDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.conf")
	content := "[backup]\nplugin=file\n[file.config]\npath=/etc/example.conf\n[directory.rules]\npath=/etc/example.d\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := LoadSpec(path, "app")
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Selections) != 2 {
		t.Fatalf("selection count = %d, want 2", len(spec.Selections))
	}
}

// Requirements: backup-restore/FR-004, backup-restore/SR-003.
func TestLoadSpecValidatesRestartUnits(t *testing.T) {
	cases := []struct {
		name, units string
		valid       bool
	}{
		{"service", "lite-nas-auth.service", true},
		{"empty", "", true},
		{"option", "--host=evil.service", false},
		{"path", "../evil.service", false},
		{"duplicate", "lite-nas-auth.service,lite-nas-auth.service", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "app.conf")
			content := "[backup]\nplugin=file\nrestart_units=" + testCase.units + "\n[file.main]\npath=/etc/example.conf\n"
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := LoadSpec(path, "app")
			if (err == nil) != testCase.valid {
				t.Fatalf("restart units %q: %v", testCase.units, err)
			}
		})
	}
}

// Requirements: backup-restore/FR-002.
func TestWriteArchiveIncludesSymlinkChain(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.conf")
	middle := filepath.Join(root, "middle.conf")
	link := filepath.Join(root, "linked.conf")
	if err := os.WriteFile(target, []byte("value=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, middle); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(middle, link); err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	if err := WriteArchive(&archive, Spec{Selections: []Selection{{Path: link}}}); err != nil {
		t.Fatal(err)
	}
	paths, err := ArchivePaths(&archive)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 {
		t.Fatalf("archive paths = %v, want three entries", paths)
	}
}

// Requirements: backup-restore/SR-003.
func TestArchivePathsRejectsCorruptedGzipTrailer(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("value=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	if err := WriteArchive(&archive, Spec{Selections: []Selection{{Path: path}}}); err != nil {
		t.Fatal(err)
	}
	data := archive.Bytes()
	data[len(data)-1] ^= 0xff
	if _, err := ArchivePaths(bytes.NewReader(data)); err == nil {
		t.Fatal("corrupted archive was accepted")
	}
}
