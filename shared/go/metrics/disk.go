package metrics

import "time"

// DiskMetricsSnapshot is the top-level disk metrics snapshot payload.
type DiskMetricsSnapshot struct {
	// Timestamp is the time at which the snapshot was collected.
	Timestamp time.Time `json:"timestamp"`

	// Devices contains flat block-device-like entries discovered on the host.
	Devices []DiskDeviceSnapshot `json:"devices"`

	// Filesystems summarizes observed filesystem types from active mounts.
	Filesystems map[string]DiskFilesystemSummary `json:"filesystems"`

	// Mounts contains active mount entries keyed by mountpoint.
	Mounts map[string]DiskMountSnapshot `json:"mounts"`
}

// DiskDeviceSnapshot represents one block-device-like host entry.
type DiskDeviceSnapshot struct {
	// Node is the runtime kernel node name, for example "sda" or "nvme0n1p1".
	Node string `json:"node"`

	// Kind classifies the device role within the snapshot contract.
	Kind string `json:"kind"`

	// Parent stores the parent kernel node for partitions or derived devices
	// when one parent can be identified.
	Parent *string `json:"parent,omitempty"`

	// Major is the device major number.
	Major uint64 `json:"major"`

	// Minor is the device minor number.
	Minor uint64 `json:"minor"`

	// SizeBytes is the device size in bytes.
	SizeBytes uint64 `json:"size_bytes"`

	// Name stores one optional human-friendly display name.
	Name *string `json:"name,omitempty"`

	// Vendor stores the optional device vendor string.
	Vendor *string `json:"vendor,omitempty"`

	// Model stores the optional device model string.
	Model *string `json:"model,omitempty"`

	// Description stores one optional descriptive device label.
	Description *string `json:"description,omitempty"`

	// Serial stores one optional stable serial identifier.
	Serial *string `json:"serial,omitempty"`

	// WWN stores one optional stable world-wide-name style identifier.
	WWN *string `json:"wwn,omitempty"`

	// ConnectionType classifies the resolved host-side transport or device
	// family.
	ConnectionType string `json:"connection_type"`

	// Rotational reports whether the host marks the device as rotational.
	Rotational *bool `json:"rotational,omitempty"`

	// Removable reports whether the host marks the device as removable.
	Removable *bool `json:"removable,omitempty"`

	// IO stores optional raw disk I/O counters from the kernel.
	IO *DiskDeviceIOCounters `json:"io,omitempty"`

	// Partitions stores child partition node names when discovered.
	Partitions []string `json:"partitions,omitempty"`

	// Mounts stores mountpoints that are currently associated with the device.
	Mounts []string `json:"mounts,omitempty"`

	// ZPoolMemberships stores ZFS pool memberships when external ZFS data is
	// available.
	ZPoolMemberships []string `json:"zpool_memberships,omitempty"`
}

// DiskDeviceIOCounters stores one set of raw Linux disk I/O counters.
type DiskDeviceIOCounters struct {
	// ReadsCompleted is the completed read request count.
	ReadsCompleted uint64 `json:"reads_completed"`

	// ReadsMerged is the merged read request count.
	ReadsMerged uint64 `json:"reads_merged"`

	// SectorsRead is the cumulative read sector count.
	SectorsRead uint64 `json:"sectors_read"`

	// ReadTimeMS is the cumulative read time in milliseconds.
	ReadTimeMS uint64 `json:"read_time_ms"`

	// WritesCompleted is the completed write request count.
	WritesCompleted uint64 `json:"writes_completed"`

	// WritesMerged is the merged write request count.
	WritesMerged uint64 `json:"writes_merged"`

	// SectorsWritten is the cumulative written sector count.
	SectorsWritten uint64 `json:"sectors_written"`

	// WriteTimeMS is the cumulative write time in milliseconds.
	WriteTimeMS uint64 `json:"write_time_ms"`

	// IOInProgress is the current in-progress I/O count.
	IOInProgress uint64 `json:"io_in_progress"`

	// IOTimeMS is the cumulative device busy time in milliseconds.
	IOTimeMS uint64 `json:"io_time_ms"`

	// WeightedIOTimeMS is the cumulative weighted I/O time in milliseconds.
	WeightedIOTimeMS uint64 `json:"weighted_io_time_ms"`

	// DiscardsCompleted is the completed discard request count when available.
	DiscardsCompleted *uint64 `json:"discards_completed,omitempty"`

	// DiscardsMerged is the merged discard request count when available.
	DiscardsMerged *uint64 `json:"discards_merged,omitempty"`

	// SectorsDiscarded is the cumulative discarded sector count when available.
	SectorsDiscarded *uint64 `json:"sectors_discarded,omitempty"`

	// DiscardTimeMS is the cumulative discard time in milliseconds when
	// available.
	DiscardTimeMS *uint64 `json:"discard_time_ms,omitempty"`

	// FlushesCompleted is the completed flush request count when available.
	FlushesCompleted *uint64 `json:"flushes_completed,omitempty"`

	// FlushTimeMS is the cumulative flush time in milliseconds when available.
	FlushTimeMS *uint64 `json:"flush_time_ms,omitempty"`
}

// DiskFilesystemSummary stores one observed filesystem-type summary.
type DiskFilesystemSummary struct {
	// Type classifies the filesystem into one service-level category.
	Type string `json:"type"`

	// MountCount stores the number of currently observed mounts of this type.
	MountCount uint64 `json:"mount_count"`
}

// DiskMountSnapshot stores one active mountpoint snapshot.
type DiskMountSnapshot struct {
	// Mountpoint is the mounted path visible to the process.
	Mountpoint string `json:"mountpoint"`

	// Source is the source string reported by the mount table.
	Source string `json:"source"`

	// Filesystem is the observed filesystem type.
	Filesystem string `json:"filesystem"`

	// Device stores the local kernel node when the source resolves to one.
	Device *string `json:"device,omitempty"`

	// ReadOnly reports whether the mount is read-only.
	ReadOnly bool `json:"readonly"`

	// Remote reports whether the mount is network-backed.
	Remote bool `json:"remote"`

	// MemoryBacked reports whether the mount is memory-backed.
	MemoryBacked bool `json:"memory_backed"`

	// TotalBytes stores total capacity in bytes when usage can be read.
	TotalBytes *uint64 `json:"total_bytes,omitempty"`

	// UsedBytes stores used capacity in bytes when usage can be read.
	UsedBytes *uint64 `json:"used_bytes,omitempty"`

	// FreeBytes stores free capacity in bytes when usage can be read.
	FreeBytes *uint64 `json:"free_bytes,omitempty"`

	// AvailableBytes stores available capacity in bytes when usage can be read.
	AvailableBytes *uint64 `json:"available_bytes,omitempty"`

	// UsedPercent stores used capacity percentage when usage can be read.
	UsedPercent *float64 `json:"used_percent,omitempty"`
}
