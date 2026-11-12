package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"lite-nas/services/zfs-metrics/config"
	"lite-nas/services/zfs-metrics/modules"
	"lite-nas/services/zfs-metrics/state"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/metrics"
	sharedmodules "lite-nas/shared/modules"
	"lite-nas/shared/testutil/metricsruntimetest"
)

// TestServeSnapshotsPublishesZFSUpdate verifies the queued ZFS snapshot keeps
// its existing event subject and payload type.
func TestServeSnapshotsPublishesZFSUpdate(t *testing.T) {
	input := make(chan metrics.ZFSSnapshot, 1)
	input <- metrics.ZFSSnapshot{Timestamp: time.Unix(100, 0)}
	close(input)
	client := &metricsruntimetest.RecordingClient{}
	err := serveSnapshots(t.Context(), input, nil, state.NewHistoryStore(1), client,
		&metricsruntimetest.RecordingLogger{})
	if err != nil {
		t.Fatalf("serveSnapshots() error = %v", err)
	}
	if got := client.PublishCalls[0].Subject; got != zfsmetricscontract.SnapshotEventSubject {
		t.Fatalf("published subject = %q", got)
	}
}

// TestServeSnapshotsReportsZFSFailure verifies failed outbound publication
// remains observable after moving the loop to a worker.
func TestServeSnapshotsReportsZFSFailure(t *testing.T) {
	input := make(chan metrics.ZFSSnapshot, 1)
	input <- metrics.ZFSSnapshot{}
	close(input)
	log := &metricsruntimetest.RecordingLogger{}
	client := &metricsruntimetest.RecordingClient{PublishErr: errors.New("broker unavailable")}
	if err := serveSnapshots(t.Context(), input, nil, state.NewHistoryStore(1), client, log); err != nil {
		t.Fatalf("serveSnapshots() error = %v", err)
	}
	if len(log.Errors) != 1 {
		t.Fatalf("error log count = %d, want 1", len(log.Errors))
	}
}

// TestServeSnapshotsStopsOnCancellation verifies the worker returns the
// existing cancellation result.
func TestServeSnapshotsStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := serveSnapshots(ctx, make(chan metrics.ZFSSnapshot), nil,
		state.NewHistoryStore(1), &metricsruntimetest.RecordingClient{},
		&metricsruntimetest.RecordingLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v", err)
	}
}

// TestHistoryRPCReadsPublisherStore verifies the query handler observes one
// snapshot committed by the publisher's store.
func TestHistoryRPCReadsPublisherStore(t *testing.T) {
	store := state.NewHistoryStore(1)
	store.Add(metrics.ZFSSnapshot{Timestamp: time.Unix(100, 0)})
	server := &metricsruntimetest.RecordingServer{}
	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}
	response := metricsruntimetest.MustInvokeRPCHandler[zfsmetricscontract.GetHistoryResponse](
		t, server, zfsmetricscontract.HistoryRPCSubject)
	if len(response.Items) != 1 {
		t.Fatalf("history RPC items = %d, want 1", len(response.Items))
	}
}

// TestSnapshotRPCReportsAvailability verifies the query handler exposes the
// state owned by the publisher worker.
func TestSnapshotRPCReportsAvailability(t *testing.T) {
	for _, test := range []struct {
		name      string
		populate  bool
		available bool
	}{
		{name: "empty"},
		{name: "stored", populate: true, available: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := state.NewHistoryStore(1)
			if test.populate {
				store.Add(metrics.ZFSSnapshot{Timestamp: time.Unix(100, 0)})
			}
			server := &metricsruntimetest.RecordingServer{}
			if err := registerRPCHandlers(server, store); err != nil {
				t.Fatalf("registerRPCHandlers() error = %v", err)
			}
			response := metricsruntimetest.MustInvokeRPCHandler[zfsmetricscontract.GetSnapshotResponse](
				t, server, zfsmetricscontract.SnapshotRPCSubject)
			if response.Available != test.available {
				t.Fatalf("snapshot availability = %t, want %t", response.Available, test.available)
			}
		})
	}
}

// TestRegisterRPCHandlersReturnsRegistrationFailure verifies startup retains
// the original first and second handler error paths.
func TestRegisterRPCHandlersReturnsRegistrationFailure(t *testing.T) {
	for _, subject := range []string{zfsmetricscontract.SnapshotRPCSubject, zfsmetricscontract.HistoryRPCSubject} {
		t.Run(subject, func(t *testing.T) {
			want := errors.New("register failed")
			server := &metricsruntimetest.RecordingServer{RegisterRPCErrors: map[string]error{subject: want}}
			err := registerRPCHandlers(server, state.NewHistoryStore(1))
			if !errors.Is(err, want) {
				t.Fatalf("registerRPCHandlers() error = %v", err)
			}
		})
	}
}

// TestHandlePollErrorReportsOpenChannel verifies the error consumer keeps its
// existing logging behavior.
func TestHandlePollErrorReportsOpenChannel(t *testing.T) {
	log := &metricsruntimetest.RecordingLogger{}
	handlePollError(log, errors.New("poll failed"), true)
	if len(log.Errors) != 1 {
		t.Fatalf("poll error log count = %d, want 1", len(log.Errors))
	}
}

// TestHandlePollErrorIgnoresClosedChannel verifies no error is emitted when
// the polling error source reports closure.
func TestHandlePollErrorIgnoresClosedChannel(t *testing.T) {
	log := &metricsruntimetest.RecordingLogger{}
	handlePollError(log, nil, false)
	if len(log.Errors) != 0 {
		t.Fatalf("poll error log count = %d, want 0", len(log.Errors))
	}
}

// TestRunWithDependenciesStopsOnCancellation verifies worker startup and
// publisher completion through the production composition path.
func TestRunWithDependenciesStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := runWithDependencies(ctx, func(string, string) (modules.Infra, error) {
		return modules.Infra{
			CoreInfra: sharedmodules.CoreInfra{
				Logger: sharedlogger.NewNop(),
				Client: &metricsruntimetest.RecordingClient{},
				Server: &metricsruntimetest.RecordingServer{},
			},
			Config: config.Config{Metrics: config.MetricsConfig{HistorySize: 1, PollInterval: time.Hour, ZpoolPath: "/usr/sbin/zpool"}},
		}, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runWithDependencies() error = %v, want context canceled", err)
	}
}
