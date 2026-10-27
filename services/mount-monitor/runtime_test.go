package main

import (
	"testing"

	mountcontract "lite-nas/shared/contracts/mountmonitor"
	"lite-nas/shared/testutil/loggingmanagertest"
)

// Requirements: task.md/SVC-002.
func TestMountAlertEventIDMatchesLoggingManagerContract(t *testing.T) {
	t.Parallel()
	eventID := mountAlertEventID(mountcontract.LifecycleEvent{
		MountID: 42, MountPoint: "/mnt/litenas-test", Source: "/dev/loop0", FSType: "ext4",
	})
	loggingmanagertest.MustValidateEventID(t, eventID)
}
