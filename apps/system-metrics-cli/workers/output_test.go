package workers

import (
	"bytes"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: system-metrics-cli/FR-002, system-metrics-cli/FR-003, system-metrics-cli/IR-002
func TestOutputWriterWritesCurrentSnapshotInHumanReadableFormat(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var output bytes.Buffer

	snapshot := metrics.SystemSnapshot{
		Timestamp: time.Unix(1700000000, 0).UTC(),
		CPU: metrics.CPUSample{
			TotalUsagePct: 62.5,
			PerCoreUsage:  []float64{60, 65},
		},
		Mem: metrics.MemSample{
			TotalBytes: 16 * gibibyte,
			UsedBytes:  5 * gibibyte,
		},
	}

	err := writer.WriteCurrent(outputWriterBuffer(&output), snapshot, CurrentSelection{CPU: true, RAM: true})
	if err != nil {
		t.Fatalf("WriteCurrent() error = %v", err)
	}

	want := "CPU Load: 62.5%\n----\nCore0: 60%\nCore1: 65%\n\n----\n\nRAM: 16GB\nUsed: 5GB\nAvailable: 11GB\n"
	if output.String() != want {
		t.Fatalf("WriteCurrent() output = %q, want %q", output.String(), want)
	}
}

// Requirements: system-metrics-cli/FR-003, system-metrics-cli/IR-002
func TestOutputWriterWritesOnlySelectedCurrentSection(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var output bytes.Buffer

	snapshot := metrics.SystemSnapshot{
		CPU: metrics.CPUSample{
			TotalUsagePct: 42,
			PerCoreUsage:  []float64{40},
		},
	}

	err := writer.WriteCurrent(outputWriterBuffer(&output), snapshot, CurrentSelection{CPU: true})
	if err != nil {
		t.Fatalf("WriteCurrent() error = %v", err)
	}

	want := "CPU Load: 42%\n----\nCore0: 40%\n"
	if output.String() != want {
		t.Fatalf("WriteCurrent() output = %q, want %q", output.String(), want)
	}
}

// Requirements: system-metrics-cli/FR-005, system-metrics-cli/IR-002
func TestOutputWriterWritesHistoryAsIndentedJSON(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var output bytes.Buffer

	history := []metrics.SystemSnapshot{
		{
			Timestamp: time.Unix(1700000000, 0).UTC(),
			CPU: metrics.CPUSample{
				TotalUsagePct: 25,
			},
		},
	}

	err := writer.WriteHistory(outputWriterBuffer(&output), history)
	if err != nil {
		t.Fatalf("WriteHistory() error = %v", err)
	}

	want := "[\n  {\n    \"Timestamp\": \"2023-11-14T22:13:20Z\",\n    \"CPU\": {\n      \"TotalUsagePct\": 25,\n      \"PerCoreUsage\": null\n    },\n    \"Mem\": {\n      \"TotalBytes\": 0,\n      \"UsedBytes\": 0,\n      \"UsedPct\": 0\n    }\n  }\n]\n"
	if output.String() != want {
		t.Fatalf("WriteHistory() output = %q, want %q", output.String(), want)
	}
}

func outputWriterBuffer(buffer *bytes.Buffer) *bytes.Buffer {
	return buffer
}
