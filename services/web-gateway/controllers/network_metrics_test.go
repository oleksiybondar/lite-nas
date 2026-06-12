package controllers

import (
	"context"
	"testing"

	networkmetricsdto "lite-nas/services/web-gateway/dto/network_metrics"
	"lite-nas/services/web-gateway/testutil/networkmetricstest"
	"lite-nas/shared/metrics"
)

type stubNetworkMetricsService struct {
	snapshot metrics.NetworkMetricsSnapshot
	history  []metrics.NetworkMetricsSnapshot
	err      error
}

// GetSnapshot returns the configured snapshot or one injected test error.
func (s stubNetworkMetricsService) GetSnapshot(context.Context) (metrics.NetworkMetricsSnapshot, error) {
	if s.err != nil {
		return metrics.NetworkMetricsSnapshot{}, s.err
	}

	return s.snapshot, nil
}

// GetHistory returns the configured history or one injected test error.
func (s stubNetworkMetricsService) GetHistory(context.Context) ([]metrics.NetworkMetricsSnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.history, nil
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestNetworkMetricsControllerGetSnapshotWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	runSnapshotEnvelopeTest(
		t,
		networkmetricstest.Snapshot(123),
		func(snapshot metrics.NetworkMetricsSnapshot) func(context.Context, *struct{}) (*networkmetricsdto.NetworkSnapshotOutput, error) {
			return NewNetworkMetricsController(stubNetworkMetricsService{snapshot: snapshot}).GetSnapshot
		},
		networkSnapshotOutputSuccess,
		networkSnapshotTimestampIsZero,
		networkSnapshotOutputData,
	)
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestNetworkMetricsControllerGetHistoryWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	runHistoryEnvelopeTest(
		t,
		[]metrics.NetworkMetricsSnapshot{
			networkmetricstest.Snapshot(123),
			networkmetricstest.Snapshot(124),
		},
		func(history []metrics.NetworkMetricsSnapshot) func(context.Context, *struct{}) (*networkmetricsdto.NetworkHistoryOutput, error) {
			return NewNetworkMetricsController(stubNetworkMetricsService{history: history}).GetHistory
		},
		networkHistoryOutputSuccess,
		networkHistoryTimestampIsZero,
		networkHistoryOutputData,
	)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestNetworkMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewNetworkMetricsController(stubNetworkMetricsService{err: backendFailure()})

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	assertBackendFailureMapped(t, got, err, "GetSnapshot()")
}

// networkSnapshotOutputSuccess reads the success flag from one network snapshot output envelope.
func networkSnapshotOutputSuccess(output *networkmetricsdto.NetworkSnapshotOutput) bool {
	return output.Body.Success
}

// networkSnapshotTimestampIsZero reports whether one network snapshot output missed its timestamp.
func networkSnapshotTimestampIsZero(output *networkmetricsdto.NetworkSnapshotOutput) bool {
	return output.Body.Timestamp.IsZero()
}

// networkSnapshotOutputData extracts the network snapshot data payload from one snapshot envelope.
func networkSnapshotOutputData(output *networkmetricsdto.NetworkSnapshotOutput) metrics.NetworkMetricsSnapshot {
	return output.Body.Data
}

// networkHistoryOutputSuccess reads the success flag from one network history output envelope.
func networkHistoryOutputSuccess(output *networkmetricsdto.NetworkHistoryOutput) bool {
	return output.Body.Success
}

// networkHistoryTimestampIsZero reports whether one network history output missed its timestamp.
func networkHistoryTimestampIsZero(output *networkmetricsdto.NetworkHistoryOutput) bool {
	return output.Body.Timestamp.IsZero()
}

// networkHistoryOutputData extracts the network history data payload from one history envelope.
func networkHistoryOutputData(output *networkmetricsdto.NetworkHistoryOutput) []metrics.NetworkMetricsSnapshot {
	return output.Body.Data
}
