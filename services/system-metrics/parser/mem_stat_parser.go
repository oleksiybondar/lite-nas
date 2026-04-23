package parser

import (
	"errors"
	"strconv"
	"strings"

	"lite-nas/shared/metrics"
)

// MemStatParser parses Linux /proc/meminfo data into memory metrics.
//
// Design tradeoffs:
//
//   - The parser accepts string input instead of []byte. This keeps the parsing
//     logic smaller and easier to read. The additional string conversion at the
//     service boundary is acceptable because /proc/meminfo is small and polled
//     infrequently.
//   - The parser intentionally uses simple string splitting instead of scanner-
//     based or more advanced parsing. /proc/meminfo is a small, stable,
//     machine-friendly text format, so explicit string operations are easier to
//     understand and test.
//   - Parsing is organized into small stages:
//   - split raw text into lines
//   - filter relevant memory lines
//   - validate required lines
//   - parse filtered lines into raw numeric values
//   - build the final memory sample
//
// This favors clarity and separation of responsibility over micro-optimizing
// text parsing.
type MemStatParser struct{}

// Parse parses the contents of /proc/meminfo and returns a memory sample.
//
// The input must contain the required lines:
//   - MemTotal:
//   - MemAvailable:
//
// Values in /proc/meminfo are expressed in kB and are converted to bytes.
//
// Used memory is calculated as:
//
//	used = total - available
//
// Used percentage is calculated as:
//
//	usedPct = used / total * 100
func (MemStatParser) Parse(data string) (metrics.MemSample, error) {
	lines := strings.Split(data, "\n")
	memLines := filterMemStatLines(lines)

	if !hasRequiredMemStatLines(memLines) {
		return metrics.MemSample{}, errors.New("missing required meminfo lines")
	}

	totalKB, availableKB, err := parseMemStatLines(memLines)
	if err != nil {
		return metrics.MemSample{}, err
	}

	return buildMemSample(totalKB, availableKB), nil
}

// filterMemStatLines returns only the memory lines required for building the
// final memory sample.
func filterMemStatLines(lines []string) []string {
	memLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if isMemStatLine(line) {
			memLines = append(memLines, line)
		}
	}

	return memLines
}

// hasRequiredMemStatLines reports whether the filtered meminfo lines contain
// both required fields: MemTotal and MemAvailable.
func hasRequiredMemStatLines(lines []string) bool {
	var hasTotal bool
	var hasAvailable bool

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			hasTotal = true
		case "MemAvailable:":
			hasAvailable = true
		}
	}

	return hasTotal && hasAvailable
}

// isMemStatLine reports whether a raw /proc/meminfo line is relevant for
// building the final memory sample.
func isMemStatLine(line string) bool {
	return strings.HasPrefix(line, "MemTotal:") || strings.HasPrefix(line, "MemAvailable:")
}

// parseMemStatLines parses filtered meminfo lines and returns total and
// available memory in kB.
func parseMemStatLines(lines []string) (uint64, uint64, error) {
	var totalKB uint64
	var availableKB uint64

	for _, line := range lines {
		key, valueKB, err := parseMemStatLine(line)
		if err != nil {
			return 0, 0, err
		}

		if key == "MemTotal:" {
			totalKB = valueKB
		} else {
			availableKB = valueKB
		}
	}

	return totalKB, availableKB, nil
}

// parseMemStatLine parses one already-filtered /proc/meminfo line.
//
// Supported formats include:
//
//	MemTotal:      16367456 kB
//	MemAvailable:   9213456 kB
func parseMemStatLine(line string) (string, uint64, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", 0, errors.New("meminfo line has too few fields")
	}

	valueKB, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return "", 0, err
	}

	return fields[0], valueKB, nil
}

// buildMemSample converts memory values in kB into the final memory sample.
func buildMemSample(totalKB uint64, availableKB uint64) metrics.MemSample {
	totalBytes := totalKB * 1024
	availableBytes := availableKB * 1024
	usedBytes := totalBytes - availableBytes

	var usedPct float64
	if totalBytes > 0 {
		usedPct = float64(usedBytes) / float64(totalBytes) * 100
	}

	return metrics.MemSample{
		TotalBytes: totalBytes,
		UsedBytes:  usedBytes,
		UsedPct:    usedPct,
	}
}
