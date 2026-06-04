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

func (s stubZFSMetricsService) GetSnapshot(context.Context) (metrics.ZFSSnapshot, error) {
	if s.err != nil {
		return metrics.ZFSSnapshot{}, s.err
	}

	return s.snapshot, nil
}

func (s stubZFSMetricsService) GetHistory(context.Context) ([]metrics.ZFSSnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.history, nil
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestZFSMetricsControllerGetSnapshotWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	snapshot := zfsSnapshotFixture(123)
	controller := NewZFSMetricsController(stubZFSMetricsService{snapshot: snapshot})
	assertSnapshotWrapped(
		t,
		snapshot,
		controller.GetSnapshot,
		func(output *zfsmetricsdto.ZFSSnapshotOutput) bool { return output.Body.Success },
		func(output *zfsmetricsdto.ZFSSnapshotOutput) bool { return output.Body.Timestamp.IsZero() },
		func(output *zfsmetricsdto.ZFSSnapshotOutput) metrics.ZFSSnapshot { return output.Body.Data },
	)
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestZFSMetricsControllerGetHistoryWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	history := []metrics.ZFSSnapshot{
		zfsSnapshotFixture(123),
		zfsSnapshotFixture(124),
	}
	controller := NewZFSMetricsController(stubZFSMetricsService{history: history})
	assertHistoryWrapped(
		t,
		history,
		controller.GetHistory,
		func(output *zfsmetricsdto.ZFSHistoryOutput) bool { return output.Body.Success },
		func(output *zfsmetricsdto.ZFSHistoryOutput) bool { return output.Body.Timestamp.IsZero() },
		func(output *zfsmetricsdto.ZFSHistoryOutput) []metrics.ZFSSnapshot { return output.Body.Data },
	)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestZFSMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewZFSMetricsController(stubZFSMetricsService{err: backendFailure()})

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	assertBackendFailureMapped(t, got, err, "GetSnapshot()")
}

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
