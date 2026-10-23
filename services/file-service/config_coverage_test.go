package main

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// textConfigReader supplies one INI payload to the configuration boundary.
type textConfigReader string

// Read returns the configured INI bytes without filesystem dependencies.
func (reader textConfigReader) Read() ([]byte, error) { return []byte(reader), nil }

// mustLoadPackagedConfig loads the installed configuration for focused field checks.
func mustLoadPackagedConfig(t *testing.T) serviceConfig {
	t.Helper()
	payload, err := os.ReadFile("../../configs/etc/lite-nas/file-service.conf")
	if err != nil {
		t.Fatal(err)
	}
	config, err := loadServiceConfig(textConfigReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	return config
}

// TestLoadServiceConfigAcceptsPackagedSettings covers the installed config shape.
func TestLoadServiceConfigAcceptsPackagedSettings(t *testing.T) {
	config := mustLoadPackagedConfig(t)
	if config.ChunkSize != 1<<20 {
		t.Fatalf("chunk size = %d", config.ChunkSize)
	}
}

// TestLoadServiceConfigAcceptsPackagedAddress covers the configured listen address.
func TestLoadServiceConfigAcceptsPackagedAddress(t *testing.T) {
	config := mustLoadPackagedConfig(t)
	if config.Address != "127.0.0.1:8082" {
		t.Fatalf("address = %q", config.Address)
	}
}

// TestPackagedConfigDefinesProtectedPaths covers the configuration-owned mutation restrictions.
func TestPackagedConfigDefinesProtectedPaths(t *testing.T) {
	config := mustLoadPackagedConfig(t)
	want := []string{"/bin", "/boot", "/dev", "/etc", "/lib", "/lib64", "/proc", "/root", "/run", "/sbin", "/sys", "/usr", "/var/lib"}
	if !slices.Equal(config.ProtectedPaths, want) {
		t.Fatalf("protected paths = %v, want %v", config.ProtectedPaths, want)
	}
}

// TestPackagedConfigDefinesNonAdminPaths covers the configuration-owned read exclusions.
func TestPackagedConfigDefinesNonAdminPaths(t *testing.T) {
	config := mustLoadPackagedConfig(t)
	want := []string{"/bin", "/boot", "/dev", "/etc", "/lib", "/lib64", "/proc", "/root", "/run", "/sbin", "/snap", "/sys", "/usr", "/var"}
	if !slices.Equal(config.ReadRules.NonAdminDenied, want) {
		t.Fatalf("non-admin denied paths = %v, want %v", config.ReadRules.NonAdminDenied, want)
	}
}

// TestPackagedConfigDefinesReadAndWriteRoots covers explicit storage roots.
func TestPackagedConfigDefinesReadAndWriteRoots(t *testing.T) {
	config := mustLoadPackagedConfig(t)
	if !slices.Equal(config.WriteRoots, []string{"/home"}) || !slices.Equal(config.ReadRules.ReadRoots, []string{"/home"}) {
		t.Fatalf("storage roots = write %v, read %v", config.WriteRoots, config.ReadRules.ReadRoots)
	}
}

// TestLoadServiceConfigRejectsUnsafeOverrides covers service config input validation.
func TestLoadServiceConfigRejectsUnsafeOverrides(t *testing.T) {
	payload, err := os.ReadFile("../../configs/etc/lite-nas/file-service.conf")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		old  string
		new  string
	}{
		{name: "remote bind", old: "127.0.0.1:8082", new: "0.0.0.0:8082"},
		{name: "invalid port", old: "127.0.0.1:8082", new: "127.0.0.1:70000"},
		{name: "invalid chunk", old: "chunk_size_mib=1", new: "chunk_size_mib=abc"},
		{name: "small chunk", old: "chunk_size_mib=1", new: "chunk_size_mib=0"},
		{name: "large chunk", old: "chunk_size_mib=1", new: "chunk_size_mib=17"},
		{name: "relative metadata", old: "metadata_dir=/var/lib/lite-nas/file-service/uploads", new: "metadata_dir=relative"},
		{name: "relative root", old: "write_roots=/home", new: "write_roots=relative"},
		{name: "missing write roots", old: "write_roots=/home", new: "write_roots="},
		{name: "relative read root", old: "read_roots=/home", new: "read_roots=relative"},
		{name: "relative protected root", old: "protected_paths=/bin", new: "protected_paths=relative"},
		{name: "malformed mask", old: "known_path_directories=/home/*/.litenas", new: "known_path_directories=/home/**/.litenas"},
		{name: "invalid docs switch", old: "enabled=false", new: "enabled=maybe"},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := strings.Replace(string(payload), test.old, test.new, 1)
			if _, err := loadServiceConfig(textConfigReader(changed)); err == nil {
				t.Fatal("loadServiceConfig() accepted unsafe override")
			}
		})
	}
}

// TestConfigPathsReturnsNilForEmptyRoots covers an omitted root list.
func TestConfigPathsReturnsNilForEmptyRoots(t *testing.T) {
	if got := configPaths("  "); got != nil {
		t.Fatalf("configPaths(empty) = %v", got)
	}
}

// TestConfigPathsTrimsConfiguredRoots covers comma-separated path handling.
func TestConfigPathsTrimsConfiguredRoots(t *testing.T) {
	got := configPaths(" /home , /srv ")
	if len(got) != 2 || got[0] != "/home" || got[1] != "/srv" {
		t.Fatalf("configPaths() = %v", got)
	}
}
