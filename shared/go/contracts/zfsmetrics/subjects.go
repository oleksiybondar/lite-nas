package zfsmetrics

const (
	// SnapshotEventSubject publishes the latest processed ZFS snapshot.
	SnapshotEventSubject = "zfs.metrics.events.snapshot"

	// SnapshotRPCSubject serves the latest processed ZFS snapshot via request/reply messaging.
	SnapshotRPCSubject = "zfs.metrics.rpc.snapshot.get"

	// HistoryRPCSubject serves retained ZFS snapshots via request/reply messaging.
	HistoryRPCSubject = "zfs.metrics.rpc.history.get"
)
