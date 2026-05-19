package workers

import (
	"bytes"
	"strings"
	"testing"

	"lite-nas/shared/metrics"
)

// Requirements: zfs-metrics-cli/FR-004
func TestWriteCurrentRendersPoolBlock(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	snapshot := currentSnapshotFixture()

	var output bytes.Buffer
	if err := writer.WriteCurrent(&output, snapshot); err != nil {
		t.Fatalf("WriteCurrent() error = %v", err)
	}

	assertContainsTokens(t, output.String(), []string{
		"----",
		"name: tank",
		"size: 100GB",
		"status: ONLINE",
		"Used: 73.2%",
		"Errors: R:1,W:2,C:3",
		"IO: R:5MBps, W:7MBps",
	})
}

// Requirements: zfs-metrics-cli/FR-003
func TestWriteHistoryRendersPrettyJSON(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	history := []metrics.ZFSSnapshot{{Pools: []metrics.ZFSPoolSnapshot{{Name: "tank"}}}}

	var output bytes.Buffer
	if err := writer.WriteHistory(&output, history); err != nil {
		t.Fatalf("WriteHistory() error = %v", err)
	}

	got := output.String()
	if !strings.Contains(got, "\n  {") {
		t.Fatalf("WriteHistory() output = %q, want pretty JSON", got)
	}
	if !strings.Contains(got, "\"Name\": \"tank\"") {
		t.Fatalf("WriteHistory() output = %q, want pool name", got)
	}
}

func currentSnapshotFixture() metrics.ZFSSnapshot {
	return metrics.ZFSSnapshot{
		Pools: []metrics.ZFSPoolSnapshot{
			{
				Name:   "tank",
				Health: metrics.ZFSPoolHealthOnline,
				Root: metrics.ZFSVdevSnapshot{
					Errors: metrics.ZFSIOErrors{Read: 1, Write: 2, Checksum: 3},
				},
				Usage: &metrics.ZFSUsage{
					SizeBytes:   100 * gibibyte,
					CapacityPct: 73.2,
				},
				IOStat: &metrics.ZFSIOStat{
					Bandwidth: metrics.ZFSIOStatValues{
						Read:  5 * megabyte,
						Write: 7 * megabyte,
					},
				},
			},
		},
	}
}

func assertContainsTokens(t *testing.T, output string, tokens []string) {
	t.Helper()

	for _, token := range tokens {
		if !strings.Contains(output, token) {
			t.Fatalf("output missing token %q\noutput:\n%s", token, output)
		}
	}
}
