package main

import (
	"os"
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
