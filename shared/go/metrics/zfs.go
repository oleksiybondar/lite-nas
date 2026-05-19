package metrics

import "time"

// ZFSPoolHealth represents normalized pool health states.
type ZFSPoolHealth string

const (
	ZFSPoolHealthOnline   ZFSPoolHealth = "ONLINE"
	ZFSPoolHealthDegraded ZFSPoolHealth = "DEGRADED"
	ZFSPoolHealthFaulted  ZFSPoolHealth = "FAULTED"
	ZFSPoolHealthOffline  ZFSPoolHealth = "OFFLINE"
	ZFSPoolHealthRemoved  ZFSPoolHealth = "REMOVED"
	ZFSPoolHealthUnknown  ZFSPoolHealth = "UNKNOWN"
)

// ZFSVdevKind describes a vdev node kind in the status tree.
type ZFSVdevKind string

const (
	ZFSVdevKindPool    ZFSVdevKind = "pool"
	ZFSVdevKindMirror  ZFSVdevKind = "mirror"
	ZFSVdevKindRaidz   ZFSVdevKind = "raidz"
	ZFSVdevKindDevice  ZFSVdevKind = "device"
	ZFSVdevKindSpare   ZFSVdevKind = "spare"
	ZFSVdevKindLog     ZFSVdevKind = "log"
	ZFSVdevKindCache   ZFSVdevKind = "cache"
	ZFSVdevKindSpecial ZFSVdevKind = "special"
	ZFSVdevKindUnknown ZFSVdevKind = "unknown"
)

// ZFSSnapshot is the top-level service snapshot payload.
type ZFSSnapshot struct {
	Timestamp time.Time
	Pools     []ZFSPoolSnapshot
}

// ZFSPoolSnapshot is a traversal-oriented normalized pool snapshot model.
type ZFSPoolSnapshot struct {
	Name   string
	Health ZFSPoolHealth
	Errors string
	Scan   string
	Root   ZFSVdevSnapshot
	Usage  *ZFSUsage
	IOStat *ZFSIOStat
}

// ZFSVdevSnapshot represents a vdev/device node in the pool tree.
type ZFSVdevSnapshot struct {
	Type     ZFSVdevKind
	Name     string
	Path     string
	Errors   ZFSIOErrors
	Children []ZFSVdevSnapshot
}

// ZFSIOErrors captures status read/write/checksum counters.
type ZFSIOErrors struct {
	Read     uint64
	Write    uint64
	Checksum uint64
}

// ZFSIOStat represents normalized pool-level iostat values.
type ZFSIOStat struct {
	Operations ZFSIOStatValues
	Bandwidth  ZFSIOStatValues
}

// ZFSIOStatValues stores read/write metrics.
type ZFSIOStatValues struct {
	Read  uint64
	Write uint64
}

// ZFSUsage stores normalized capacity values in bytes and percent.
type ZFSUsage struct {
	SizeBytes      uint64
	AllocatedBytes uint64
	FreeBytes      uint64
	CapacityPct    float64
}
