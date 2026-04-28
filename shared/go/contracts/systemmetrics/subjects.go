package systemmetrics

const (
	// SnapshotEventSubject publishes the latest processed system metrics
	// snapshot to interested consumers.
	SnapshotEventSubject = "system.metrics.events.stats"

	// SnapshotRPCSubject serves the latest processed system metrics snapshot via
	// request/reply messaging.
	SnapshotRPCSubject = "system.metrics.rpc.stats.get"

	// HistoryRPCSubject serves the retained system metrics history via
	// request/reply messaging.
	HistoryRPCSubject = "system.metrics.rpc.history.get"
)
