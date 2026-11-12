package controllers

import (
	"context"
	"testing"
	"time"

	zfsmetricsdto "lite-nas/services/web-gateway/dto/zfs_metrics"
	"lite-nas/shared/metrics"
)

type stubZFSMetricsService struct {
	snapshot metrics.ZFSSnapshot
	history  []metrics.ZFSSnapshot
	err      error
}

// GetSnapshot returns the configured snapshot or one injected test error.
func (s stubZFSMetricsService) GetSnapshot(context.Context) (metrics.ZFSSnapshot, error) {
	if s.err != nil {
		return metrics.ZFSSnapshot{}, s.err
	}

	return s.snapshot, nil
}

// GetHistory returns the configured history or one injected test error.
func (s stubZFSMetricsService) GetHistory(context.Context) ([]metrics.ZFSSnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.history, nil
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestZFSMetricsControllerGetSnapshotWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	runSnapshotEnvelopeTest(
		t,
		zfsSnapshotFixture(123),
		func(snapshot metrics.ZFSSnapshot) func(context.Context, *struct{}) (*zfsmetricsdto.ZFSSnapshotOutput, error) {
			return NewZFSMetricsController(stubZFSMetricsService{snapshot: snapshot}).GetSnapshot
		},
		zfsSnapshotOutputSuccess,
		zfsSnapshotTimestampIsZero,
		zfsSnapshotOutputData,
	)
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestZFSMetricsControllerGetHistoryWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	runHistoryEnvelopeTest(
		t,
		[]metrics.ZFSSnapshot{
			zfsSnapshotFixture(123),
			zfsSnapshotFixture(124),
		},
		func(history []metrics.ZFSSnapshot) func(context.Context, *struct{}) (*zfsmetricsdto.ZFSHistoryOutput, error) {
			return NewZFSMetricsController(stubZFSMetricsService{history: history}).GetHistory
		},
		zfsHistoryOutputSuccess,
		zfsHistoryTimestampIsZero,
		zfsHistoryOutputData,
	)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestZFSMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewZFSMetricsController(stubZFSMetricsService{err: backendFailure()})

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	assertBackendFailureMapped(t, got, err, "GetSnapshot()")
}

// zfsSnapshotFixture returns one representative ZFS metrics snapshot for controller tests.
func zfsSnapshotFixture(unixSeconds int64) metrics.ZFSSnapshot {
	return metrics.ZFSSnapshot{
		Timestamp: time.Unix(unixSeconds, 0).UTC(),
		Pools: []metrics.ZFSPoolSnapshot{
			{
				Name:   "tank",
				Health: metrics.ZFSPoolHealthOnline,
				Errors: "No known data errors",
				Scan:   "scrub repaired 0B in 00:01:23 with 0 errors",
				Root: metrics.ZFSVdevSnapshot{
					Type: metrics.ZFSVdevKindPool,
					Name: "tank",
					Children: []metrics.ZFSVdevSnapshot{
						{
							Type: metrics.ZFSVdevKindMirror,
							Name: "mirror-0",
							Children: []metrics.ZFSVdevSnapshot{
								{Type: metrics.ZFSVdevKindDevice, Name: "sda", Path: "/dev/sda"},
								{Type: metrics.ZFSVdevKindDevice, Name: "sdb", Path: "/dev/sdb"},
							},
						},
					},
				},
				Usage: &metrics.ZFSUsage{
					SizeBytes:      1000,
					AllocatedBytes: 400,
					FreeBytes:      600,
					CapacityPct:    40,
				},
				IOStat: &metrics.ZFSIOStat{
					Operations: metrics.ZFSIOStatValues{Read: 10, Write: 20},
					Bandwidth:  metrics.ZFSIOStatValues{Read: 100, Write: 200},
				},
			},
		},
	}
}

// zfsSnapshotOutputSuccess reads the success flag from one ZFS snapshot output envelope.
func zfsSnapshotOutputSuccess(output *zfsmetricsdto.ZFSSnapshotOutput) bool {
	return output.Body.Success
}

// zfsSnapshotTimestampIsZero reports whether one ZFS snapshot output missed its timestamp.
func zfsSnapshotTimestampIsZero(output *zfsmetricsdto.ZFSSnapshotOutput) bool {
	return output.Body.Timestamp.IsZero()
}

// zfsSnapshotOutputData extracts the ZFS snapshot data payload from one snapshot envelope.
func zfsSnapshotOutputData(output *zfsmetricsdto.ZFSSnapshotOutput) metrics.ZFSSnapshot {
	return output.Body.Data
}

// zfsHistoryOutputSuccess reads the success flag from one ZFS history output envelope.
func zfsHistoryOutputSuccess(output *zfsmetricsdto.ZFSHistoryOutput) bool {
	return output.Body.Success
}

// zfsHistoryTimestampIsZero reports whether one ZFS history output missed its timestamp.
func zfsHistoryTimestampIsZero(output *zfsmetricsdto.ZFSHistoryOutput) bool {
	return output.Body.Timestamp.IsZero()
}

// zfsHistoryOutputData extracts the ZFS history data payload from one history envelope.
func zfsHistoryOutputData(output *zfsmetricsdto.ZFSHistoryOutput) []metrics.ZFSSnapshot {
	return output.Body.Data
}
