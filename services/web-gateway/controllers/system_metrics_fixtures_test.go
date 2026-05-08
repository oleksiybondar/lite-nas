package controllers

import (
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

func systemSnapshotFixture(
	unixSeconds int64,
	totalUsagePct float64,
	perCoreUsage []float64,
	totalBytes uint64,
	usedBytes uint64,
	usedPct float64,
) metrics.SystemSnapshot {
	return metrics.SystemSnapshot{
		Timestamp: time.Unix(unixSeconds, 0).UTC(),
		CPU: metrics.CPUSample{
			TotalUsagePct: totalUsagePct,
			PerCoreUsage:  perCoreUsage,
		},
		Mem: metrics.MemSample{
			TotalBytes: totalBytes,
			UsedBytes:  usedBytes,
			UsedPct:    usedPct,
		},
	}
}

func assertSuccessfulSystemMetricsEnvelope(t *testing.T, success bool, timestampIsZero bool) {
	t.Helper()

	if !success {
		t.Fatalf("Success = false, want true")
	}

	if timestampIsZero {
		t.Fatalf("Timestamp is zero, want populated timestamp")
	}
}
