package workers

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"lite-nas/shared/metrics"
)

const gibibyte = 1024 * 1024 * 1024

// OutputWriter renders the selected CLI output format.
type OutputWriter interface {
	WriteCurrent(writer io.Writer, snapshot metrics.SystemSnapshot, selection CurrentSelection) error
	WriteHistory(writer io.Writer, history []metrics.SystemSnapshot) error
}

type outputWriter struct{}

// NewOutputWriter creates an output rendering worker.
func NewOutputWriter() OutputWriter {
	return outputWriter{}
}

// WriteCurrent renders a human-readable snapshot view.
func (outputWriter) WriteCurrent(
	writer io.Writer,
	snapshot metrics.SystemSnapshot,
	selection CurrentSelection,
) error {
	sections := make([]string, 0, 2)

	if selection.CPU {
		sections = append(sections, renderCPUSection(snapshot))
	}

	if selection.RAM {
		sections = append(sections, renderRAMSection(snapshot))
	}

	_, err := fmt.Fprintf(writer, "%s\n", strings.Join(sections, "\n\n----\n\n"))
	return err
}

// WriteHistory renders history as pretty-printed JSON.
func (outputWriter) WriteHistory(writer io.Writer, history []metrics.SystemSnapshot) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(history)
}

func renderCPUSection(snapshot metrics.SystemSnapshot) string {
	lines := []string{
		fmt.Sprintf("CPU Load: %s%%", formatPercent(snapshot.CPU.TotalUsagePct)),
		"----",
	}

	for coreIndex, usage := range snapshot.CPU.PerCoreUsage {
		lines = append(lines, fmt.Sprintf("Core%d: %s%%", coreIndex, formatPercent(usage)))
	}

	return strings.Join(lines, "\n")
}

func renderRAMSection(snapshot metrics.SystemSnapshot) string {
	availableBytes := snapshot.Mem.TotalBytes - snapshot.Mem.UsedBytes

	return strings.Join([]string{
		fmt.Sprintf("RAM: %sGB", formatGigabytes(snapshot.Mem.TotalBytes)),
		fmt.Sprintf("Used: %sGB", formatGigabytes(snapshot.Mem.UsedBytes)),
		fmt.Sprintf("Available: %sGB", formatGigabytes(availableBytes)),
	}, "\n")
}

func formatPercent(value float64) string {
	return strings.TrimSuffix(strings.TrimSuffix(strconv.FormatFloat(value, 'f', 1, 64), "0"), ".")
}

func formatGigabytes(value uint64) string {
	gigabytes := float64(value) / gibibyte
	return strings.TrimSuffix(strings.TrimSuffix(strconv.FormatFloat(gigabytes, 'f', 1, 64), "0"), ".")
}
