package servicemetrics

const (
	// SnapshotEventSubject publishes the latest processed service snapshot.
	SnapshotEventSubject = "service.metrics.events.snapshot"

	// SnapshotRPCSubject serves the latest processed service snapshot via
	// request/reply messaging.
	SnapshotRPCSubject = "service.metrics.rpc.snapshot.get"

	// HistoryRPCSubject serves retained service snapshots via request/reply
	// messaging.
	HistoryRPCSubject = "service.metrics.rpc.history.get"
)
