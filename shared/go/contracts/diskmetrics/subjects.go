package diskmetrics

const (
	// SnapshotEventSubject publishes the latest processed disk snapshot.
	SnapshotEventSubject = "disk.metrics.events.snapshot"

	// SnapshotRPCSubject serves the latest processed disk snapshot via
	// request/reply messaging.
	SnapshotRPCSubject = "disk.metrics.rpc.snapshot.get"

	// HistoryRPCSubject serves retained disk snapshots via request/reply
	// messaging.
	HistoryRPCSubject = "disk.metrics.rpc.history.get"
)
