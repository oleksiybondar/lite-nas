package list

import (
	"fmt"
	"strconv"
	"strings"

	"lite-nas/shared/parsers/zfs/tsvbase"
)

// PoolUsage is normalized `zpool list -H` usage data for one pool.
type PoolUsage struct {
	SizeBytes      uint64
	AllocatedBytes uint64
	FreeBytes      uint64
	CapacityPct    float64
}

// Header* constants map `zpool list` columns to semantic field names.
const (
	HeaderName   = "name"
	HeaderSize   = "size"
	HeaderAlloc  = "alloc"
	HeaderFree   = "free"
	HeaderCap    = "cap"
	HeaderHealth = "health"
)

// DefaultHeaders defines expected TSV columns for `zpool list -H -p -o ...`.
var DefaultHeaders = []string{
	HeaderName,
	HeaderSize,
	HeaderAlloc,
	HeaderFree,
	HeaderCap,
	HeaderHealth,
}

// Parse parses `zpool list -H` output (tab-separated) into normalized pool usage by pool name.
// Recommended command form: `zpool list -H -p -o name,size,alloc,free,cap,health`.
func Parse(input string) (map[string]PoolUsage, error) {
	result := make(map[string]PoolUsage)

	rows, err := tsvbase.ParseRows(input, DefaultHeaders)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		name, usage, err := parseUsageRow(row)
		if err != nil {
			return nil, err
		}
		result[name] = usage
	}

	return result, nil
}

// parseUsageRow converts one parsed TSV row into PoolUsage.
func parseUsageRow(row map[string]string) (string, PoolUsage, error) {
	name := row[HeaderName]

	sizeBytes, err := parsePoolSizeField(name, HeaderSize, row[HeaderSize])
	if err != nil {
		return "", PoolUsage{}, err
	}
	allocatedBytes, err := parsePoolSizeField(name, HeaderAlloc, row[HeaderAlloc])
	if err != nil {
		return "", PoolUsage{}, err
	}
	freeBytes, err := parsePoolSizeField(name, HeaderFree, row[HeaderFree])
	if err != nil {
		return "", PoolUsage{}, err
	}
	capacityPct, err := parsePoolPctField(name, HeaderCap, row[HeaderCap])
	if err != nil {
		return "", PoolUsage{}, err
	}

	return name, PoolUsage{
		SizeBytes:      sizeBytes,
		AllocatedBytes: allocatedBytes,
		FreeBytes:      freeBytes,
		CapacityPct:    capacityPct,
	}, nil
}

// parsePoolSizeField parses one size field and adds pool/field context to errors.
func parsePoolSizeField(poolName string, fieldName string, fieldValue string) (uint64, error) {
	parsed, err := parseSizeToBytes(fieldValue)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value for pool %q: %w", strings.ToUpper(fieldName), poolName, err)
	}
	return parsed, nil
}

// parsePoolPctField parses one percentage field and adds pool/field context to errors.
func parsePoolPctField(poolName string, fieldName string, fieldValue string) (float64, error) {
	parsed, err := parsePct(fieldValue)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value for pool %q: %w", strings.ToUpper(fieldName), poolName, err)
	}
	return parsed, nil
}

// parseSizeToBytes parses either raw byte values or human-readable byte units.
func parseSizeToBytes(value string) (uint64, error) {
	trimmed := strings.TrimSpace(strings.ToUpper(value))
	if trimmed == "" || trimmed == "-" {
		return 0, nil
	}

	if isDigitsOnly(trimmed) {
		return strconv.ParseUint(trimmed, 10, 64)
	}

	numberPart, unitPart := splitNumberAndUnit(trimmed)
	number, err := strconv.ParseFloat(numberPart, 64)
	if err != nil {
		return 0, err
	}
	multiplier, err := unitMultiplier(unitPart)
	if err != nil {
		return 0, err
	}

	return uint64(number * multiplier), nil
}

// isDigitsOnly reports whether value contains only ASCII decimal digits.
func isDigitsOnly(value string) bool {
	return strings.IndexFunc(value, func(r rune) bool { return r < '0' || r > '9' }) == -1
}

// splitNumberAndUnit splits a size token into numeric and unit parts.
func splitNumberAndUnit(value string) (string, string) {
	unitIndex := len(value)
	for unitIndex > 0 {
		last := value[unitIndex-1]
		if (last >= '0' && last <= '9') || last == '.' {
			break
		}
		unitIndex--
	}

	return strings.TrimSpace(value[:unitIndex]), strings.TrimSpace(value[unitIndex:])
}

// unitMultiplier returns the byte multiplier for one supported size unit.
func unitMultiplier(unit string) (float64, error) {
	switch unit {
	case "", "B":
		return 1, nil
	case "K":
		return 1024, nil
	case "M":
		return 1024 * 1024, nil
	case "G":
		return 1024 * 1024 * 1024, nil
	case "T":
		return 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unsupported size unit %q", unit)
	}
}

// parsePct parses percentage values, tolerating optional trailing `%`.
func parsePct(value string) (float64, error) {
	trimmed := strings.TrimSpace(strings.TrimSuffix(value, "%"))
	if trimmed == "" || trimmed == "-" {
		return 0, nil
	}
	return strconv.ParseFloat(trimmed, 64)
}
