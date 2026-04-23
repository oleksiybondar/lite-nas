package parser

import (
	"errors"
	"strconv"
	"strings"

	"lite-nas/shared/metrics"
)

// CPUStatParser parses Linux /proc/stat CPU data into raw CPU metrics.
//
// Design tradeoffs:
//
//   - The parser accepts string input instead of []byte. This keeps the parsing
//     logic simpler and easier to maintain. The additional string conversion at
//     the service boundary is acceptable because /proc/stat is small and polled
//     infrequently.
//   - The parser intentionally uses simple string splitting instead of scanner-
//     based or more advanced parsing. /proc/stat is a small, stable,
//     machine-friendly text format, so explicit string operations are easier to
//     read and test.
//   - Parsing is organized into small stages:
//   - split raw text into lines
//   - filter relevant CPU lines
//   - validate required aggregated CPU line
//   - parse filtered lines into domain values
//
// This favors clarity and separation of responsibility over micro-optimizing
// for single-pass low-level text processing.
type CPUStatParser struct{}

// Parse parses the contents of /proc/stat and returns a raw CPU sample.
//
// The input must contain the aggregated "cpu" line. Per-core lines such as
// "cpu0", "cpu1", and so on are also collected.
//
// The returned raw sample keeps only the counters needed for later CPU usage
// calculation:
//   - Total: sum of all numeric counters on the line
//   - Idle: idle + iowait
func (CPUStatParser) Parse(data string) (metrics.CPURawSample, error) {
	lines := strings.Split(data, "\n")
	cpuLines := filterCPUStatLines(lines)

	if !hasTotalCPULine(cpuLines) {
		return metrics.CPURawSample{}, errors.New("missing aggregated cpu line")
	}

	return parseCPULines(cpuLines)
}

// filterCPUStatLines returns only aggregated and per-core CPU lines from
// /proc/stat.
//
// Non-CPU lines are discarded before deeper parsing so later stages can focus
// only on relevant input.
func filterCPUStatLines(lines []string) []string {
	cpuLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if isCPUStatLine(line) {
			cpuLines = append(cpuLines, line)
		}
	}

	return cpuLines
}

// hasTotalCPULine reports whether the filtered CPU lines include the required
// aggregated "cpu" line.
func hasTotalCPULine(lines []string) bool {
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "cpu" {
			return true
		}
	}

	return false
}

// isCPUStatLine reports whether a raw /proc/stat line is an aggregated or
// per-core CPU line.
func isCPUStatLine(line string) bool {
	return strings.HasPrefix(line, "cpu")
}

// parseCPULines parses filtered CPU lines into a raw CPU sample.
func parseCPULines(lines []string) (metrics.CPURawSample, error) {
	var sample metrics.CPURawSample

	for _, line := range lines {
		fields := strings.Fields(line)

		coreSample, err := parseCPUStatLine(fields)
		if err != nil {
			return metrics.CPURawSample{}, err
		}

		if fields[0] == "cpu" {
			sample.Total = coreSample
		} else {
			sample.Cores = append(sample.Cores, coreSample)
		}
	}

	return sample, nil
}

// parseCPUStatLine parses one already-filtered CPU line from /proc/stat.
//
// Supported formats include:
//
//	cpu  user nice system idle iowait irq softirq steal guest guest_nice
//	cpu0 user nice system idle iowait irq softirq steal guest guest_nice
func parseCPUStatLine(fields []string) (metrics.CPUCoreRawSample, error) {
	if len(fields) < 5 {
		return metrics.CPUCoreRawSample{}, errors.New("cpu line has too few fields")
	}

	counters, err := parseCPUCounters(fields[1:])
	if err != nil {
		return metrics.CPUCoreRawSample{}, err
	}

	return metrics.CPUCoreRawSample{
		Total: countTotalCounters(counters),
		Idle:  countIdleCounters(counters),
	}, nil
}

// parseCPUCounters parses numeric CPU counters from one CPU stat line.
//
// The input must not include the leading label such as "cpu" or "cpu0".
func parseCPUCounters(values []string) ([]uint64, error) {
	counters := make([]uint64, 0, len(values))

	for _, valueText := range values {
		value, err := strconv.ParseUint(valueText, 10, 64)
		if err != nil {
			return nil, err
		}

		counters = append(counters, value)
	}

	return counters, nil
}

// countTotalCounters returns the sum of all CPU counters.
func countTotalCounters(counters []uint64) uint64 {
	var total uint64

	for _, counter := range counters {
		total += counter
	}

	return total
}

// countIdleCounters returns total non-busy CPU time used for later CPU usage
// calculation.
//
// Idle time includes both the idle and iowait counters, following standard
// Linux CPU usage calculation practice.
func countIdleCounters(counters []uint64) uint64 {
	idle := counters[3]

	if len(counters) > 4 {
		idle += counters[4]
	}

	return idle
}
