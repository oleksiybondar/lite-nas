package iostat

import (
	"fmt"
	"strconv"
	"strings"

	"lite-nas/shared/parsers/zfs/tsvbase"
)

// IOStatValues stores read/write metrics.
type IOStatValues struct {
	Read  uint64
	Write uint64
}

// PoolIOStat is normalized `zpool iostat -H` data for one pool.
type PoolIOStat struct {
	Operations IOStatValues
	Bandwidth  IOStatValues
}

// Header* constants map `zpool iostat` columns to semantic field names.
const (
	HeaderPool       = "pool"
	HeaderAlloc      = "alloc"
	HeaderFree       = "free"
	HeaderReadOps    = "read_ops"
	HeaderWriteOps   = "write_ops"
	HeaderReadBytes  = "read_bytes"
	HeaderWriteBytes = "write_bytes"
)

// DefaultHeaders defines expected TSV columns for `zpool iostat -H -p`.
var DefaultHeaders = []string{
	HeaderPool,
	HeaderAlloc,
	HeaderFree,
	HeaderReadOps,
	HeaderWriteOps,
	HeaderReadBytes,
	HeaderWriteBytes,
}

// Parse parses `zpool iostat -H` output (tab-separated) into normalized pool IO stats by pool name.
// Recommended command form: `zpool iostat -H -p`.
func Parse(input string) (map[string]PoolIOStat, error) {
	result := make(map[string]PoolIOStat)

	rows, err := tsvbase.ParseRows(input, DefaultHeaders)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		name, ioStat, err := parseIOStatRow(row)
		if err != nil {
			return nil, err
		}
		result[name] = ioStat
	}

	return result, nil
}

// parseIOStatRow converts one parsed TSV row into PoolIOStat.
func parseIOStatRow(row map[string]string) (string, PoolIOStat, error) {
	name := row[HeaderPool]

	readOps, err := parseUintField(row[HeaderReadOps], "read operations", name)
	if err != nil {
		return "", PoolIOStat{}, err
	}
	writeOps, err := parseUintField(row[HeaderWriteOps], "write operations", name)
	if err != nil {
		return "", PoolIOStat{}, err
	}
	readBandwidth, err := parseBandwidthField(row[HeaderReadBytes], "read bandwidth", name)
	if err != nil {
		return "", PoolIOStat{}, err
	}
	writeBandwidth, err := parseBandwidthField(row[HeaderWriteBytes], "write bandwidth", name)
	if err != nil {
		return "", PoolIOStat{}, err
	}

	return name, PoolIOStat{
		Operations: IOStatValues{Read: readOps, Write: writeOps},
		Bandwidth:  IOStatValues{Read: readBandwidth, Write: writeBandwidth},
	}, nil
}

// parseUintField parses a uint64 row value and includes field context in errors.
func parseUintField(value string, field string, poolName string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s for pool %q: %w", field, poolName, err)
	}
	return parsed, nil
}

// parseBandwidthField parses a bandwidth value and includes field context in errors.
func parseBandwidthField(value string, field string, poolName string) (uint64, error) {
	parsed, err := parseSizeToBytes(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s for pool %q: %w", field, poolName, err)
	}
	return parsed, nil
}

// parseSizeToBytes parses `-p` byte values used by `zpool iostat`.
func parseSizeToBytes(value string) (uint64, error) {
	trimmed := strings.TrimSpace(strings.ToUpper(value))
	if trimmed == "" || trimmed == "-" {
		return 0, nil
	}
	return strconv.ParseUint(trimmed, 10, 64)
}
