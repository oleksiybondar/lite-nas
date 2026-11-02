package backuprestore

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Requirements: backup-restore/FR-002.
func TestPackagedBeforeStartupSpecSelectsHostScript(t *testing.T) {
	path := "../../configs/etc/lite-nas/backup.d/before-startup.conf"
	got, err := LoadSpec(path, "before-startup")
	if err != nil {
		t.Fatal(err)
	}
	want := Spec{
		Application: "before-startup",
		SourcePath:  path,
		Plugin:      "file",
		Selections:  []Selection{{Path: "/etc/lite-nas/before-startup.sh"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("before-startup spec = %#v, want %#v", got, want)
	}
}

// Requirements: backup-restore/FR-002, journal-manager/BR-001.
func TestPackagedJournalManagerSpecSelectsConfigurationAndRestart(t *testing.T) {
	path := "../../configs/etc/lite-nas/backup.d/journal-manager.conf"
	got, err := LoadSpec(path, "journal-manager")
	if err != nil {
		t.Fatal(err)
	}
	want := Spec{
		Application:  "journal-manager",
		SourcePath:   path,
		Plugin:       "file",
		Selections:   []Selection{{Path: "/etc/lite-nas/journal-manager.conf"}},
		RestartUnits: []string{"lite-nas-journal-manager.service"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("journal-manager spec = %#v, want %#v", got, want)
	}
}

// Requirements: backup-restore/FR-002.
func TestPackagedLogrotateSpecSelectsHostPolicy(t *testing.T) {
	path := "../../configs/etc/lite-nas/backup.d/logrotate.conf"
	got, err := LoadSpec(path, "logrotate")
	if err != nil {
		t.Fatal(err)
	}
	want := Spec{
		Application: "logrotate",
		SourcePath:  path,
		Plugin:      "file",
		Selections: []Selection{
			{Path: "/etc/logrotate.conf"},
			{Path: "/etc/logrotate.d", Directory: true},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("logrotate spec = %#v, want %#v", got, want)
	}
}

// Requirements: backup-restore/FR-002.
func TestPackagedHostSecuritySpecsSelectConfiguration(t *testing.T) {
	cases := []struct {
		name       string
		selections []Selection
	}{
		{name: "pam", selections: []Selection{{Path: "/etc/pam.d", Directory: true}, {Path: "/etc/security", Directory: true}}},
		{name: "auditd", selections: []Selection{{Path: "/etc/audit", Directory: true}}},
		{name: "ssh", selections: []Selection{{Path: "/etc/ssh/sshd_config"}, {Path: "/etc/ssh/sshd_config.d", Directory: true, Optional: true}}},
		{name: "journald", selections: []Selection{{Path: "/etc/systemd/journald.conf", Optional: true}, {Path: "/etc/systemd/journald.conf.d", Directory: true, Optional: true}}},
		{name: "nginx", selections: []Selection{{Path: "/etc/nginx", Directory: true, Optional: true}, {Path: "/etc/lite-nas/certificates/nginx", Directory: true, Optional: true}}},
		{name: "privilege-policy", selections: []Selection{{Path: "/etc/sudoers", Optional: true}, {Path: "/etc/sudoers.d", Directory: true, Optional: true}, {Path: "/etc/polkit-1/rules.d", Directory: true, Optional: true}}},
		{name: "kernel-hardening", selections: []Selection{{Path: "/etc/sysctl.conf", Optional: true}, {Path: "/etc/sysctl.d", Directory: true, Optional: true}, {Path: "/etc/modprobe.d", Directory: true, Optional: true}, {Path: "/etc/modules", Optional: true}, {Path: "/etc/modules-load.d", Directory: true, Optional: true}}},
		{name: "storage-recovery", selections: []Selection{{Path: "/etc/fstab"}, {Path: "/etc/crypttab", Optional: true}}},
		{name: "time-sync", selections: []Selection{{Path: "/etc/systemd/timesyncd.conf", Optional: true}, {Path: "/etc/systemd/timesyncd.conf.d", Directory: true, Optional: true}, {Path: "/etc/chrony", Directory: true, Optional: true}, {Path: "/etc/chrony.conf", Optional: true}}},
		{name: "package-trust", selections: []Selection{{Path: "/etc/apt/apt.conf", Optional: true}, {Path: "/etc/apt/apt.conf.d", Directory: true, Optional: true}, {Path: "/etc/apt/sources.list", Optional: true}, {Path: "/etc/apt/sources.list.d", Directory: true, Optional: true}, {Path: "/etc/apt/preferences", Optional: true}, {Path: "/etc/apt/preferences.d", Directory: true, Optional: true}, {Path: "/etc/apt/trusted.gpg", Optional: true}, {Path: "/etc/apt/trusted.gpg.d", Directory: true, Optional: true}, {Path: "/etc/apt/keyrings", Directory: true, Optional: true}}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := "../../configs/etc/lite-nas/backup.d/" + testCase.name + ".conf"
			spec, err := LoadSpec(path, testCase.name)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(spec.Selections, testCase.selections) {
				t.Fatalf("selections = %#v, want %#v", spec.Selections, testCase.selections)
			}
		})
	}
}

// Requirements: backup-restore/FR-002.
func TestPackagedGRUBSpecSelectsSourceConfiguration(t *testing.T) {
	path := "../../configs/etc/lite-nas/backup.d/grub.conf"
	spec, err := LoadSpec(path, "grub")
	if err != nil {
		t.Fatal(err)
	}
	want := []Selection{
		{Path: "/etc/default/grub"},
		{Path: "/etc/default/grub.d", Directory: true, Optional: true},
		{Path: "/etc/grub.d", Directory: true},
	}
	if !reflect.DeepEqual(spec.Selections, want) {
		t.Fatalf("GRUB selections = %#v, want %#v", spec.Selections, want)
	}
}

// Requirements: backup-restore/SR-003.
func TestLoadSpecRejectsInvalidOptionalValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "example.conf")
	content := "[backup]\nplugin=file\n[file.config]\npath=/etc/example.conf\noptional=sometimes\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpec(path, "example"); err == nil {
		t.Fatal("invalid optional value was accepted")
	}
}

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
