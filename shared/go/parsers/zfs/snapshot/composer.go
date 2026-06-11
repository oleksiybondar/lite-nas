package snapshot

import (
	"strconv"
	"strings"
	"time"

	"lite-nas/shared/metrics"
	iostatparser "lite-nas/shared/parsers/zfs/iostat"
	listparser "lite-nas/shared/parsers/zfs/list"
	statusparser "lite-nas/shared/parsers/zfs/status"
)

// Compose builds a normalized ZFS snapshot from parsed status, list, and iostat outputs.
func Compose(
	collectedAt time.Time,
	statusDoc statusparser.StatusDocument,
	usageByPool map[string]listparser.PoolUsage,
	ioByPool map[string]iostatparser.PoolIOStat,
) metrics.ZFSSnapshot {
	pools := make([]metrics.ZFSPoolSnapshot, 0, len(statusDoc.Pools))

	for _, pool := range statusDoc.Pools {
		pools = append(pools, composePoolSnapshot(pool, usageByPool, ioByPool))
	}

	return metrics.ZFSSnapshot{
		Timestamp: collectedAt,
		Pools:     pools,
	}
}

// composePoolSnapshot assembles one pool snapshot from normalized parser outputs.
func composePoolSnapshot(
	pool statusparser.PoolBlock,
	usageByPool map[string]listparser.PoolUsage,
	ioByPool map[string]iostatparser.PoolIOStat,
) metrics.ZFSPoolSnapshot {
	poolSnapshot := metrics.ZFSPoolSnapshot{
		Name:   pool.PoolName,
		Health: normalizeHealth(pool.Metadata.State),
		Errors: pool.ErrorsSummary,
		Scan:   pool.Metadata.Scan,
		Root:   defaultRootNode(pool.PoolName),
	}

	applyRootNode(&poolSnapshot, pool)
	applyUsage(&poolSnapshot, pool.PoolName, usageByPool)
	applyIOStat(&poolSnapshot, pool.PoolName, ioByPool)
	return poolSnapshot
}

// defaultRootNode creates a fallback root node when status config tree is absent.
func defaultRootNode(poolName string) metrics.ZFSVdevSnapshot {
	return metrics.ZFSVdevSnapshot{Type: metrics.ZFSVdevKindPool, Name: poolName}
}

// applyRootNode maps parsed config-tree root into the snapshot root node.
func applyRootNode(poolSnapshot *metrics.ZFSPoolSnapshot, pool statusparser.PoolBlock) {
	if len(pool.Config.Roots) == 0 {
		return
	}
	poolSnapshot.Root = mapNode(pool.Config.Roots[0], pool.PoolName)
}

// applyUsage attaches usage statistics when list output has data for the pool.
func applyUsage(
	poolSnapshot *metrics.ZFSPoolSnapshot,
	poolName string,
	usageByPool map[string]listparser.PoolUsage,
) {
	usage, ok := usageByPool[poolName]
	if !ok {
		return
	}

	poolSnapshot.Usage = &metrics.ZFSUsage{
		SizeBytes:      usage.SizeBytes,
		AllocatedBytes: usage.AllocatedBytes,
		FreeBytes:      usage.FreeBytes,
		CapacityPct:    usage.CapacityPct,
	}
}

// applyIOStat attaches I/O statistics when iostat output has data for the pool.
func applyIOStat(
	poolSnapshot *metrics.ZFSPoolSnapshot,
	poolName string,
	ioByPool map[string]iostatparser.PoolIOStat,
) {
	ioStat, ok := ioByPool[poolName]
	if !ok {
		return
	}

	poolSnapshot.IOStat = &metrics.ZFSIOStat{
		Operations: metrics.ZFSIOStatValues{
			Read:  ioStat.Operations.Read,
			Write: ioStat.Operations.Write,
		},
		Bandwidth: metrics.ZFSIOStatValues{
			Read:  ioStat.Bandwidth.Read,
			Write: ioStat.Bandwidth.Write,
		},
	}
}

// normalizeHealth maps raw status health values to contract enum values.
func normalizeHealth(value string) metrics.ZFSPoolHealth {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case string(metrics.ZFSPoolHealthOnline):
		return metrics.ZFSPoolHealthOnline
	case string(metrics.ZFSPoolHealthDegraded):
		return metrics.ZFSPoolHealthDegraded
	case string(metrics.ZFSPoolHealthFaulted):
		return metrics.ZFSPoolHealthFaulted
	case string(metrics.ZFSPoolHealthOffline):
		return metrics.ZFSPoolHealthOffline
	case string(metrics.ZFSPoolHealthRemoved):
		return metrics.ZFSPoolHealthRemoved
	default:
		return metrics.ZFSPoolHealthUnknown
	}
}

// mapNode recursively maps one parsed config node into snapshot vdev shape.
func mapNode(node statusparser.ConfigNode, poolName string) metrics.ZFSVdevSnapshot {
	mapped := metrics.ZFSVdevSnapshot{
		Type:     inferKind(node.Name, poolName),
		Name:     node.Name,
		Path:     inferPath(node.Name),
		Errors:   parseErrors(node.Columns),
		Children: make([]metrics.ZFSVdevSnapshot, 0, len(node.Children)),
	}
	for _, child := range node.Children {
		mapped.Children = append(mapped.Children, mapNode(child, poolName))
	}
	return mapped
}

// inferKind infers vdev kind from node naming conventions.
func inferKind(name string, poolName string) metrics.ZFSVdevKind {
	if name == poolName {
		return metrics.ZFSVdevKindPool
	}
	if strings.HasPrefix(name, "/") {
		return metrics.ZFSVdevKindDevice
	}

	if kind, ok := kindByName(name); ok {
		return kind
	}
	if kind, ok := kindByPrefix(name); ok {
		return kind
	}
	return metrics.ZFSVdevKindUnknown
}

// kindByName matches exact logical vdev group names.
func kindByName(name string) (metrics.ZFSVdevKind, bool) {
	exactNames := map[string]metrics.ZFSVdevKind{
		"spares":  metrics.ZFSVdevKindSpare,
		"logs":    metrics.ZFSVdevKindLog,
		"cache":   metrics.ZFSVdevKindCache,
		"special": metrics.ZFSVdevKindSpecial,
	}
	kind, ok := exactNames[strings.ToLower(name)]
	return kind, ok
}

// kindByPrefix matches prefixed vdev names such as mirror-* and raidz*.
func kindByPrefix(name string) (metrics.ZFSVdevKind, bool) {
	switch {
	case strings.HasPrefix(name, "mirror-"):
		return metrics.ZFSVdevKindMirror, true
	case strings.HasPrefix(name, "raidz"):
		return metrics.ZFSVdevKindRaidz, true
	default:
		return metrics.ZFSVdevKindUnknown, false
	}
}

// inferPath returns node path only for device entries.
func inferPath(name string) string {
	if strings.HasPrefix(name, "/") {
		return name
	}
	return ""
}

// parseErrors maps READ/WRITE/CKSUM columns into typed error counters.
func parseErrors(columns map[string]string) metrics.ZFSIOErrors {
	return metrics.ZFSIOErrors{
		Read:     parseUint(columns["READ"]),
		Write:    parseUint(columns["WRITE"]),
		Checksum: parseUint(columns["CKSUM"]),
	}
}

// parseUint parses uint64 values and falls back to zero on malformed input.
func parseUint(value string) uint64 {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
