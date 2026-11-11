package main

import (
	"strings"
	"testing"
)

// mountConfigReader supplies an INI document to configuration tests.
type mountConfigReader string

// Read returns the test document without filesystem dependencies.
func (reader mountConfigReader) Read() ([]byte, error) {
	return []byte(reader), nil
}

// Requirements: security-monitoring/FR-002.
func TestLoadMountConfigUsesDefaultExclusions(t *testing.T) {
	t.Parallel()
	cfg, err := loadMountConfig(mountConfigReader("[mounts]\npoll_interval=5s\n[auth]\nservice_login=mount-monitor\n"))
	if err != nil {
		t.Fatalf("load mount config: %v", err)
	}
	if got := strings.Join(cfg.ExcludedMountPrefixes, ","); got != defaultExcludedMountPrefixes {
		t.Fatalf("excluded prefixes = %q", got)
	}
}

// Requirements: security-monitoring/FR-002.
func TestLoadMountConfigAcceptsEmptyExclusions(t *testing.T) {
	t.Parallel()
	cfg, err := loadMountConfig(mountConfigReader("[mounts]\npoll_interval=5s\nexcluded_mount_prefixes=\n[auth]\nservice_login=mount-monitor\n"))
	if err != nil {
		t.Fatalf("load mount config: %v", err)
	}
	if len(cfg.ExcludedMountPrefixes) != 0 {
		t.Fatalf("excluded prefixes = %v", cfg.ExcludedMountPrefixes)
	}
}

// Requirements: security-monitoring/FR-002.
func TestLoadMountConfigRejectsUnsafeExclusion(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"/", "relative", "/run/docker/../user"} {
		t.Run(value, func(t *testing.T) {
			_, err := loadMountConfig(mountConfigReader("[mounts]\npoll_interval=5s\nexcluded_mount_prefixes=" + value + "\n[auth]\nservice_login=mount-monitor\n"))
			if err == nil {
				t.Fatalf("unsafe exclusion %q accepted", value)
			}
		})
	}
}
