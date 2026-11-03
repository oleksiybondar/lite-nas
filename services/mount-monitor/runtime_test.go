package main

import (
	"regexp"
	"testing"

	mountcontract "lite-nas/shared/contracts/mountmonitor"
	"lite-nas/shared/testutil/loggingmanagertest"
)

// Requirements: security-monitoring/FR-002.
func TestMountAlertEventIDMatchesLoggingManagerContract(t *testing.T) {
	t.Parallel()
	eventID := mountAlertEventID(mountcontract.LifecycleEvent{
		EventID: "sysmnt_00000001", MountID: 42, MountPoint: "/mnt/litenas-test", Source: "/dev/loop0", FSType: "ext4",
	}, make(map[uint64]string))
	loggingmanagertest.MustValidateEventID(t, eventID)
}

// Requirements: security-monitoring/FR-002.
func TestUnmountRetainsMountedAlertID(t *testing.T) {
	t.Parallel()
	alerts := make(map[uint64]string)
	mounted := mountcontract.LifecycleEvent{EventID: "sysmnt_00000001", MountID: 42}
	mountAlertEventID(mounted, alerts)
	unmounted := mountcontract.LifecycleEvent{EventID: "sysmnt_00000002", MountID: 42, Action: mountcontract.Unmounted}
	if got := mountAlertEventID(unmounted, alerts); got != mounted.EventID {
		t.Fatalf("unmount alert ID = %q, want %q", got, mounted.EventID)
	}
}

// TestReadableEventID verifies the producer's domain prefix and decimal counter.
func TestReadableEventID(t *testing.T) {
	t.Parallel()
	if got := newMountEventID(); !regexp.MustCompile(`^sysmnt_[0-9]{8}$`).MatchString(got) {
		t.Fatalf("unexpected event ID: %q", got)
	}
}
