package workers

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"lite-nas/shared/metrics"
)

const (
	gibibyte = 1024 * 1024 * 1024
	megabyte = 1024 * 1024
)

// OutputWriter renders CLI output for current and history modes.
type OutputWriter interface {
	WriteCurrent(writer io.Writer, snapshot metrics.ZFSSnapshot) error
	WriteHistory(writer io.Writer, history []metrics.ZFSSnapshot) error
}

type outputWriter struct{}

// NewOutputWriter creates an output rendering worker.
func NewOutputWriter() OutputWriter {
	return outputWriter{}
}

// WriteCurrent renders human-readable pool blocks for the latest snapshot.
func (outputWriter) WriteCurrent(writer io.Writer, snapshot metrics.ZFSSnapshot) error {
	if len(snapshot.Pools) == 0 {
		_, err := fmt.Fprintln(writer, "No pools in snapshot.")
		return err
	}

	sections := make([]string, 0, len(snapshot.Pools))
	for _, pool := range snapshot.Pools {
		sections = append(sections, strings.Join([]string{
			"----",
			fmt.Sprintf("name: %s", pool.Name),
			fmt.Sprintf("size: %s", formatSize(pool.Usage)),
			fmt.Sprintf("status: %s", pool.Health),
			fmt.Sprintf("Used: %s", formatUsed(pool.Usage)),
			fmt.Sprintf("Errors: %s", formatErrors(pool.Root.Errors)),
			fmt.Sprintf("IO: %s", formatIO(pool.IOStat)),
		}, "\n"))
	}

	_, err := fmt.Fprintf(writer, "%s\n", strings.Join(sections, "\n"))
	return err
}

// WriteHistory renders snapshot history as pretty-printed JSON.
func (outputWriter) WriteHistory(writer io.Writer, history []metrics.ZFSSnapshot) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(history)
}

func formatSize(usage *metrics.ZFSUsage) string {
	if usage == nil {
		return "n/a"
	}

	sizeGB := float64(usage.SizeBytes) / gibibyte
	return fmt.Sprintf("%sGB", trimFloat(sizeGB))
}

func formatUsed(usage *metrics.ZFSUsage) string {
	if usage == nil {
		return "n/a"
	}

	return fmt.Sprintf("%s%%", trimFloat(usage.CapacityPct))
}

func formatErrors(errors metrics.ZFSIOErrors) string {
	if errors.Read == 0 && errors.Write == 0 && errors.Checksum == 0 {
		return "no errors"
	}

	return fmt.Sprintf("R:%d,W:%d,C:%d", errors.Read, errors.Write, errors.Checksum)
}

func formatIO(ioStat *metrics.ZFSIOStat) string {
	if ioStat == nil {
		return "n/a"
	}

	readMBps := float64(ioStat.Bandwidth.Read) / megabyte
	writeMBps := float64(ioStat.Bandwidth.Write) / megabyte
	return fmt.Sprintf("R:%sMBps, W:%sMBps", trimFloat(readMBps), trimFloat(writeMBps))
}

func trimFloat(value float64) string {
	return strings.TrimSuffix(strings.TrimSuffix(strconv.FormatFloat(value, 'f', 1, 64), "0"), ".")
}
