package workers

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"lite-nas/shared/metrics"
)

// parseDiskStatsFile reads one diskstats file and parses it into device I/O
// counters keyed by runtime node name.
func parseDiskStatsFile(path string) (map[string]metrics.DiskDeviceIOCounters, error) {
	content, err := readTrustedFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return parseDiskStats(string(content))
}

// parseDiskStats parses Linux /proc/diskstats content into device I/O counters
// keyed by runtime node name.
func parseDiskStats(content string) (map[string]metrics.DiskDeviceIOCounters, error) {
	result := make(map[string]metrics.DiskDeviceIOCounters)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		node, counters, err := parseDiskStatsLine(line)
		if err != nil {
			return nil, err
		}

		result[node] = counters
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan diskstats: %w", err)
	}

	return result, nil
}

// parseDiskStatsLine parses one /proc/diskstats line.
func parseDiskStatsLine(line string) (string, metrics.DiskDeviceIOCounters, error) {
	fields := strings.Fields(line)
	if len(fields) < 14 {
		return "", metrics.DiskDeviceIOCounters{}, fmt.Errorf("diskstats line has %d fields, want at least 14", len(fields))
	}

	values, err := parseUint64Fields(fields[3:])
	if err != nil {
		return "", metrics.DiskDeviceIOCounters{}, err
	}

	counters := metrics.DiskDeviceIOCounters{
		ReadsCompleted:   values[0],
		ReadsMerged:      values[1],
		SectorsRead:      values[2],
		ReadTimeMS:       values[3],
		WritesCompleted:  values[4],
		WritesMerged:     values[5],
		SectorsWritten:   values[6],
		WriteTimeMS:      values[7],
		IOInProgress:     values[8],
		IOTimeMS:         values[9],
		WeightedIOTimeMS: values[10],
	}

	if len(values) >= 15 {
		counters.DiscardsCompleted = uint64Pointer(values[11])
		counters.DiscardsMerged = uint64Pointer(values[12])
		counters.SectorsDiscarded = uint64Pointer(values[13])
		counters.DiscardTimeMS = uint64Pointer(values[14])
	}
	if len(values) >= 17 {
		counters.FlushesCompleted = uint64Pointer(values[15])
		counters.FlushTimeMS = uint64Pointer(values[16])
	}

	return fields[2], counters, nil
}

// parseUint64Fields converts one string slice into uint64 values.
func parseUint64Fields(fields []string) ([]uint64, error) {
	values := make([]uint64, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse uint %q: %w", field, err)
		}

		values = append(values, value)
	}

	return values, nil
}
