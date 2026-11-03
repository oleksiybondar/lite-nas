package main

import (
	"strings"
	"testing"
	"time"
)

const mountInfoFixture = "42 1 8:1 / / rw - ext4 /dev/sda1 rw\n43 42 8:2 / /media/My\\040Drive rw - vfat /dev/sdb1 rw\n"

// Requirements: security-monitoring/FR-002.
func TestParseMountInfoDecodesPath(t *testing.T) {
	t.Parallel()
	mounts, err := parseMountInfo(strings.NewReader(mountInfoFixture))
	if err != nil {
		t.Fatalf("parse mountinfo: %v", err)
	}
	if mounts[43].Point != "/media/My Drive" {
		t.Fatalf("point = %q", mounts[43].Point)
	}
}

// Requirements: security-monitoring/FR-002.
func TestParseMountInfoRejectsMalformedSnapshot(t *testing.T) {
	t.Parallel()
	_, err := parseMountInfo(strings.NewReader(mountInfoFixture + "bad record\n"))
	if err == nil {
		t.Fatal("malformed snapshot accepted")
	}
}

// Requirements: security-monitoring/FR-002.
func TestChangedMountsEmitsUnmount(t *testing.T) {
	t.Parallel()
	before, err := parseMountInfo(strings.NewReader(mountInfoFixture))
	if err != nil {
		t.Fatalf("parse before: %v", err)
	}
	after, err := parseMountInfo(strings.NewReader(strings.Split(mountInfoFixture, "\n")[0] + "\n"))
	if err != nil {
		t.Fatalf("parse after: %v", err)
	}
	changes := changedMounts(before, after)
	if len(changes) != 1 || changes[0].action != "unmounted" {
		t.Fatalf("changes = %v", changes)
	}
}

// Requirements: security-monitoring/FR-002.
// TestChangedMountsIgnoresUserRuntimeMounts verifies session mounts produce no lifecycle events.
func TestChangedMountsIgnoresUserRuntimeMounts(t *testing.T) {
	t.Parallel()
	for _, point := range []string{"/run/user", "/run/user/1001", "/run/user/1001/gvfs"} {
		t.Run(point, func(t *testing.T) {
			before := map[uint64]mountRecord{}
			after := map[uint64]mountRecord{42: {ID: 42, Point: point}}
			if got := changedMounts(before, after); len(got) != 0 {
				t.Fatalf("changes for %q = %v", point, got)
			}
		})
	}
}

// Requirements: security-monitoring/FR-002.
// TestChangedMountsIgnoresPortScanNamespace verifies the test helper's nsfs mount is silent.
func TestChangedMountsIgnoresPortScanNamespace(t *testing.T) {
	t.Parallel()
	namespace := mountRecord{ID: 42, Point: "/run/netns/litenas-scan-test", FSType: "nsfs"}
	for name, snapshots := range map[string][2]map[uint64]mountRecord{
		"mounted":   {{}, {42: namespace}},
		"unmounted": {{42: namespace}, {}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := changedMounts(snapshots[0], snapshots[1]); len(got) != 0 {
				t.Fatalf("changes = %v", got)
			}
		})
	}
}

// Requirements: security-monitoring/FR-002.
// TestChangedMountsReportsOtherNamespaceMounts keeps unrelated nsfs events observable.
func TestChangedMountsReportsOtherNamespaceMounts(t *testing.T) {
	t.Parallel()
	after := map[uint64]mountRecord{42: {ID: 42, Point: "/run/netns/other", FSType: "nsfs"}}
	if got := changedMounts(map[uint64]mountRecord{}, after); len(got) != 1 {
		t.Fatalf("changes = %v", got)
	}
}

// Requirements: security-monitoring/FR-002.
func TestDiffMountsUsesFirstSnapshotAsBaseline(t *testing.T) {
	t.Parallel()
	mounts, err := parseMountInfo(strings.NewReader(mountInfoFixture))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	snapshots := make(chan mountSnapshot, 1)
	intents := make(chan mountIntent, 1)
	snapshots <- mountSnapshot{mounts: mounts, at: time.Now()}
	close(snapshots)
	diffMounts(snapshots, intents, nil)
	if len(intents) != 0 {
		t.Fatalf("startup intents = %d", len(intents))
	}
}
