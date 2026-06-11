package networkmetrics

const (
	// SnapshotEventSubject publishes the latest processed network snapshot.
	SnapshotEventSubject = "network.metrics.events.snapshot"

	// SnapshotRPCSubject serves the latest processed network snapshot via
	// request/reply messaging.
	SnapshotRPCSubject = "network.metrics.rpc.snapshot.get"

	// HistoryRPCSubject serves retained network snapshots via request/reply
	// messaging.
	HistoryRPCSubject = "network.metrics.rpc.history.get"
)
