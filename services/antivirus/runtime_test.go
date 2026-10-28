package main

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"

	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	mountcontract "lite-nas/shared/contracts/mountmonitor"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
	"lite-nas/shared/mountinfo"
	"lite-nas/shared/testutil/loggingmanagertest"
)

// recordingAlertPublisher captures the alert destination used by publication tests.
type recordingAlertPublisher struct {
	messaging.Client
	subject  string
	message  string
	severity enum.Severity
	status   enum.Status
}

// Publish records the destination without requiring a live NATS connection.
func (publisher *recordingAlertPublisher) Publish(_ context.Context, subject string, value any) error {
	publisher.subject = subject
	if alert, ok := value.(loggingcontract.AlertPayload); ok {
		publisher.message = alert.Message
		publisher.severity = alert.Severity
	}
	return nil
}

// Request records the state transition and accepts it for publication tests.
func (publisher *recordingAlertPublisher) Request(_ context.Context, _ string, value any, response any) error {
	if update, ok := value.(loggingcontract.UpdateAlertStateInput); ok {
		publisher.status = update.Status
	}
	if ack, ok := response.(*loggingcontract.OKResponse); ok {
		ack.OK = true
	}
	return nil
}

// mustPublishScanResult records the expected successful result publication.
func mustPublishScanResult(t *testing.T, status scanStatus) *recordingAlertPublisher {
	t.Helper()
	client := &recordingAlertPublisher{}
	alertSent := false
	result := scanResult{mount: mountinfo.Record{ID: 704, Point: "/mnt/usb"}, status: status}
	if err := publishScanResult(context.Background(), result, &alertSent, client); err != nil {
		t.Fatalf("publishScanResult() error = %v", err)
	}
	return client
}

// Requirements: security-monitoring/FR-003.
func TestCleanScanPublishesNormalInformationalEvent(t *testing.T) {
	t.Parallel()
	client := mustPublishScanResult(t, scanClean)
	if client.status != enum.StatusNormal {
		t.Fatalf("clean scan status = %q", client.status)
	}
}

// Requirements: security-monitoring/FR-003.
func TestCleanScanUsesInformationalSeverity(t *testing.T) {
	t.Parallel()
	client := mustPublishScanResult(t, scanClean)
	if client.severity != enum.SeverityInfo {
		t.Fatalf("clean scan severity = %q", client.severity)
	}
}

// Requirements: security-monitoring/FR-003.
func TestMalwareUsesErrorSeverity(t *testing.T) {
	t.Parallel()
	client := mustPublishScanResult(t, scanInfected)
	if client.severity != enum.SeverityError {
		t.Fatalf("malware severity = %q", client.severity)
	}
}

// Requirements: security-monitoring/FR-003.
func TestIncompleteScanUsesWarningSeverity(t *testing.T) {
	t.Parallel()
	client := mustPublishScanResult(t, scanIncomplete)
	if client.severity != enum.SeverityWarning {
		t.Fatalf("incomplete scan severity = %q", client.severity)
	}
}

// Requirements: security-monitoring/FR-003.
func TestScanEventIDMatchesLoggingManagerContract(t *testing.T) {
	t.Parallel()
	eventID := newScanEventID()
	loggingmanagertest.MustValidateEventID(t, eventID)
}

// Requirements: security-monitoring/FR-003.
func TestIncompleteScanAlertUsesSecurityLoggingManager(t *testing.T) {
	t.Parallel()
	client := mustPublishScanResult(t, scanIncomplete)
	if client.subject != securitycontract.AlertSubject {
		t.Fatalf("alert subject = %q, want %q", client.subject, securitycontract.AlertSubject)
	}
}

// Requirements: security-monitoring/FR-003.
func TestIncompleteScanAlertExplainsFailure(t *testing.T) {
	t.Parallel()
	client := &recordingAlertPublisher{}
	alertSent := false
	result := scanResult{
		mount:  mountinfo.Record{ID: 704, Point: "/mnt/usb"},
		status: scanIncomplete, err: errors.New("clamscan exited with status 2"),
	}
	if err := publishScanResult(context.Background(), result, &alertSent, client); err != nil {
		t.Fatalf("publishScanResult() error = %v", err)
	}
	if !strings.Contains(client.message, "clamscan exited with status 2") {
		t.Fatalf("alert message = %q", client.message)
	}
}

// Requirements: security-monitoring/FR-003.
func TestEligibleMountRejectsPrefixSibling(t *testing.T) {
	t.Parallel()
	cfg := antivirusConfig{FSTypes: map[string]bool{"vfat": true}, MountPrefixes: []string{"/media/usb"}}
	mount := mountinfo.Record{ID: 4, Point: "/media/usb-evil/data", FSType: "vfat"}
	if eligibleMount(cfg, mount) {
		t.Fatal("prefix sibling was eligible")
	}
}

// Requirements: security-monitoring/FR-003.
func TestEligibleMountAcceptsAllowlistedFilesystem(t *testing.T) {
	t.Parallel()
	cfg := antivirusConfig{FSTypes: map[string]bool{"vfat": true}, MountPrefixes: []string{"/media"}}
	mount := mountinfo.Record{ID: 4, Point: "/media/usb", FSType: "vfat"}
	if !eligibleMount(cfg, mount) {
		t.Fatal("eligible mount was rejected")
	}
}

// Requirements: security-monitoring/FR-003.
func TestFindMountByIdentityUsesReceiverNamespaceMountID(t *testing.T) {
	t.Parallel()
	event := mountcontract.LifecycleEvent{
		MountID: 4, MountPoint: "/media/usb", Source: "/dev/sdb1", FSType: "vfat",
	}
	mounts := map[uint64]mountinfo.Record{
		704: {ID: 704, Point: event.MountPoint, Source: event.Source, FSType: event.FSType},
	}

	got, found := findMountByIdentity(mounts, event)
	if !found || got.ID != 704 {
		t.Fatalf("findMountByIdentity() = (%+v, %t), want receiver mount ID 704", got, found)
	}
}

// Requirements: security-monitoring/FR-003.
func TestDecodeMountEventRejectsUnknownField(t *testing.T) {
	t.Parallel()
	payload := []byte(`{"version":1,"event_id":"m1","action":"mounted","mount_id":4,"mount_point":"/media/usb","source":"/dev/sdb1","fs_type":"vfat","observed_at":"2026-10-26T18:00:00Z","unexpected":true}`)
	_, err := decodeMountEvent(messaging.Envelope{Payload: payload}, validator.New())
	if err == nil {
		t.Fatal("unknown NATS field accepted")
	}
}

// Requirements: security-monitoring/FR-003.
func TestDecodeMountEventRejectsOversize(t *testing.T) {
	t.Parallel()
	_, err := decodeMountEvent(messaging.Envelope{Payload: make([]byte, maxMountEventBytes+1)}, validator.New())
	if err == nil {
		t.Fatal("oversized NATS payload accepted")
	}
}

// Requirements: security-monitoring/FR-003.
func TestScanMountCancelsDisappearedMount(t *testing.T) {
	t.Parallel()
	result := scanMount(context.Background(), mountinfo.Record{ID: 0, Point: "/tmp", FSType: "vfat"}, 1)
	if result.status != scanCancelled {
		t.Fatalf("status = %s", result.status)
	}
}

// TestReadableEventID verifies the producer's domain prefix and decimal counter.
func TestReadableEventID(t *testing.T) {
	t.Parallel()
	if got := newScanEventID(); !regexp.MustCompile(`^secav_[0-9]{8}$`).MatchString(got) {
		t.Fatalf("unexpected event ID: %q", got)
	}
}
