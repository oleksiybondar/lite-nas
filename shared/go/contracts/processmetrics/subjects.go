package processmetrics

const (
	// SnapshotEventSubject publishes the latest collected process snapshot.
	SnapshotEventSubject = "process.metrics.events.snapshot"

	// SnapshotRPCSubject serves the latest collected process snapshot via
	// request/reply messaging.
	SnapshotRPCSubject = "process.metrics.rpc.snapshot.get"
)
