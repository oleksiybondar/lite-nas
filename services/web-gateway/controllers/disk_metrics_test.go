package controllers

import (
	"context"
	"testing"
	"time"

	diskmetricsdto "lite-nas/services/web-gateway/dto/disk_metrics"
	"lite-nas/shared/metrics"
)

type stubDiskMetricsService struct {
	snapshot metrics.DiskMetricsSnapshot
	history  []metrics.DiskMetricsSnapshot
	err      error
}

// GetSnapshot returns the configured snapshot or one injected test error.
func (s stubDiskMetricsService) GetSnapshot(context.Context) (metrics.DiskMetricsSnapshot, error) {
	if s.err != nil {
		return metrics.DiskMetricsSnapshot{}, s.err
	}

	return s.snapshot, nil
}

// GetHistory returns the configured history or one injected test error.
func (s stubDiskMetricsService) GetHistory(context.Context) ([]metrics.DiskMetricsSnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.history, nil
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestDiskMetricsControllerGetSnapshotWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	runSnapshotEnvelopeTest(
		t,
		diskSnapshotFixture(123),
		func(snapshot metrics.DiskMetricsSnapshot) func(context.Context, *struct{}) (*diskmetricsdto.DiskSnapshotOutput, error) {
			return NewDiskMetricsController(stubDiskMetricsService{snapshot: snapshot}).GetSnapshot
		},
		diskSnapshotOutputSuccess,
		diskSnapshotTimestampIsZero,
		diskSnapshotOutputData,
	)
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestDiskMetricsControllerGetHistoryWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	runHistoryEnvelopeTest(
		t,
		[]metrics.DiskMetricsSnapshot{
			diskSnapshotFixture(123),
			diskSnapshotFixture(124),
		},
		func(history []metrics.DiskMetricsSnapshot) func(context.Context, *struct{}) (*diskmetricsdto.DiskHistoryOutput, error) {
			return NewDiskMetricsController(stubDiskMetricsService{history: history}).GetHistory
		},
		diskHistoryOutputSuccess,
		diskHistoryTimestampIsZero,
		diskHistoryOutputData,
	)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestDiskMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewDiskMetricsController(stubDiskMetricsService{err: backendFailure()})

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	assertBackendFailureMapped(t, got, err, "GetSnapshot()")
}

// diskSnapshotFixture returns one representative disk metrics snapshot for controller tests.
func diskSnapshotFixture(unixSeconds int64) metrics.DiskMetricsSnapshot {
	return metrics.DiskMetricsSnapshot{
		Timestamp: time.Unix(unixSeconds, 0),
		Devices: []metrics.DiskDeviceSnapshot{
			{
				Node:           "sda",
				Kind:           "disk",
				Major:          8,
				Minor:          0,
				SizeBytes:      1000,
				ConnectionType: "SATA",
			},
		},
		Filesystems: map[string]metrics.DiskFilesystemSummary{
			"ext4": {Type: "local_block", MountCount: 1},
		},
		Mounts: map[string]metrics.DiskMountSnapshot{
			"/": {
				Mountpoint: "/",
				Source:     "/dev/sda1",
				Filesystem: "ext4",
				ReadOnly:   false,
			},
		},
	}
}

// diskSnapshotOutputSuccess reads the success flag from one disk snapshot output envelope.
func diskSnapshotOutputSuccess(output *diskmetricsdto.DiskSnapshotOutput) bool {
	return output.Body.Success
}

// diskSnapshotTimestampIsZero reports whether one disk snapshot output missed its timestamp.
func diskSnapshotTimestampIsZero(output *diskmetricsdto.DiskSnapshotOutput) bool {
	return output.Body.Timestamp.IsZero()
}

// diskSnapshotOutputData extracts the disk snapshot data payload from one snapshot envelope.
func diskSnapshotOutputData(output *diskmetricsdto.DiskSnapshotOutput) metrics.DiskMetricsSnapshot {
	return output.Body.Data
}

// diskHistoryOutputSuccess reads the success flag from one disk history output envelope.
func diskHistoryOutputSuccess(output *diskmetricsdto.DiskHistoryOutput) bool {
	return output.Body.Success
}

// diskHistoryTimestampIsZero reports whether one disk history output missed its timestamp.
func diskHistoryTimestampIsZero(output *diskmetricsdto.DiskHistoryOutput) bool {
	return output.Body.Timestamp.IsZero()
}

// diskHistoryOutputData extracts the disk history data payload from one history envelope.
func diskHistoryOutputData(output *diskmetricsdto.DiskHistoryOutput) []metrics.DiskMetricsSnapshot {
	return output.Body.Data
}
