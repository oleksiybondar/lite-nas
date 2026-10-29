package main

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestReceiverForwardsSelectedSocketRecord sends a real datagram through a
// Unix socket so kernel-attached sender credentials are exercised.
// Requirements: security-monitoring/FR-001.
func TestReceiverForwardsSelectedSocketRecord(t *testing.T) {
	t.Parallel()
	receiver, client, output := mustReceiver(t, currentSender())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	receiver.Start(ctx)
	if _, err := client.Write([]byte(kernelDatagram)); err != nil {
		t.Fatalf("send datagram: %v", err)
	}
	select {
	case record := <-output:
		if record.SourceID != "kernel" {
			t.Fatalf("source = %q", record.SourceID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("receiver did not forward the datagram")
	}
}

// TestActivatedSocketRequiresSystemdHandoff verifies that the service never
// adopts descriptor 3 unless systemd passed it to this exact process.
// Requirements: security-monitoring/FR-001.
func TestActivatedSocketRequiresSystemdHandoff(t *testing.T) {
	t.Setenv("LISTEN_PID", "1")
	t.Setenv("LISTEN_FDS", "1")
	if _, err := activatedDatagramSocket(); err == nil {
		t.Fatal("socket adopted without a handoff for this process")
	}
}

// Requirements: security-monitoring/FR-001.
func TestReceiverStopsOnCancellation(t *testing.T) {
	t.Parallel()
	receiver, _, _ := mustReceiver(t, currentSender())
	ctx, cancel := context.WithCancel(context.Background())
	receiver.Start(ctx)
	cancel()
	select {
	case <-receiver.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("receiver did not stop after cancellation")
	}
}

// TestReceiverRejectsUnexpectedSender verifies that a local process other
// than rsyslog cannot inject records into detection.
// Requirements: security-monitoring/FR-001.
func TestReceiverRejectsUnexpectedSender(t *testing.T) {
	t.Parallel()
	receiver, _, _ := mustReceiver(t, map[uint32]bool{uint32(os.Getuid()) + 1: true})
	if _, err := receiver.accept([]byte(authDatagram), credentialsFor(uint32(os.Getuid())), 0); err == nil {
		t.Fatal("datagram from an unexpected sender accepted")
	}
}

// Requirements: security-monitoring/FR-001.
func TestReceiverRejectsMissingCredentials(t *testing.T) {
	t.Parallel()
	receiver, _, _ := mustReceiver(t, currentSender())
	if _, err := receiver.accept([]byte(authDatagram), nil, 0); err == nil {
		t.Fatal("datagram without sender credentials accepted")
	}
}

// Requirements: security-monitoring/FR-001.
func TestReceiverIgnoresUnselectedRecord(t *testing.T) {
	t.Parallel()
	receiver, _, _ := mustReceiver(t, currentSender())
	datagram := `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"daemon","input":"imuxsock","program":"cron","message":"job"}`
	records, err := receiver.accept([]byte(datagram), credentialsFor(uint32(os.Getuid())), 0)
	if err != nil || len(records) != 0 {
		t.Fatalf("records = %d, err = %v", len(records), err)
	}
}
