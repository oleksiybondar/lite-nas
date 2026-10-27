package main

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/messaging"
	"lite-nas/shared/mountinfo"
	"lite-nas/shared/testutil/loggingmanagertest"
)

// Requirements: task.md/SVC-003.
func TestScanEventIDMatchesLoggingManagerContract(t *testing.T) {
	t.Parallel()
	eventID, err := newScanEventID()
	if err != nil {
		t.Fatalf("generate scan event ID: %v", err)
	}
	loggingmanagertest.MustValidateEventID(t, eventID)
}

// Requirements: task.md/SVC-003.
func TestEligibleMountRejectsPrefixSibling(t *testing.T) {
	t.Parallel()
	cfg := antivirusConfig{FSTypes: map[string]bool{"vfat": true}, MountPrefixes: []string{"/media/usb"}}
	mount := mountinfo.Record{ID: 4, Point: "/media/usb-evil/data", FSType: "vfat"}
	if eligibleMount(cfg, mount) {
		t.Fatal("prefix sibling was eligible")
	}
}

// Requirements: task.md/SVC-003.
func TestEligibleMountAcceptsAllowlistedFilesystem(t *testing.T) {
	t.Parallel()
	cfg := antivirusConfig{FSTypes: map[string]bool{"vfat": true}, MountPrefixes: []string{"/media"}}
	mount := mountinfo.Record{ID: 4, Point: "/media/usb", FSType: "vfat"}
	if !eligibleMount(cfg, mount) {
		t.Fatal("eligible mount was rejected")
	}
}

// Requirements: task.md/SVC-003.
func TestDecodeMountEventRejectsUnknownField(t *testing.T) {
	t.Parallel()
	payload := []byte(`{"version":1,"event_id":"m1","action":"mounted","mount_id":4,"mount_point":"/media/usb","source":"/dev/sdb1","fs_type":"vfat","observed_at":"2026-10-26T18:00:00Z","unexpected":true}`)
	_, err := decodeMountEvent(messaging.Envelope{Payload: payload}, validator.New())
	if err == nil {
		t.Fatal("unknown NATS field accepted")
	}
}

// Requirements: task.md/SVC-003.
func TestDecodeMountEventRejectsOversize(t *testing.T) {
	t.Parallel()
	_, err := decodeMountEvent(messaging.Envelope{Payload: make([]byte, maxMountEventBytes+1)}, validator.New())
	if err == nil {
		t.Fatal("oversized NATS payload accepted")
	}
}

// Requirements: task.md/SVC-003.
func TestScanMountRejectsMissingIdentity(t *testing.T) {
	t.Parallel()
	result := scanMount(context.Background(), mountinfo.Record{ID: 0, Point: "/tmp", FSType: "vfat"}, 1)
	if result.status != scanIncomplete {
		t.Fatalf("status = %s", result.status)
	}
}
