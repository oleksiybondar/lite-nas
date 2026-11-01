package litenasintegrity

const (
	// RegenerateBaselineRPCSubject triggers privileged AIDE baseline regeneration.
	RegenerateBaselineRPCSubject = "litenas-integrity.rpc.baseline.regenerate"
	// StartCheckTaskSubject queues an on-demand AIDE check.
	StartCheckTaskSubject = "litenas-integrity.background-tasks.rpc.start"
	// RetryCheckTaskSubject retries a completed AIDE check.
	RetryCheckTaskSubject = "litenas-integrity.background-tasks.rpc.retry"
	// CheckTaskStatusSubject reads one AIDE check task.
	CheckTaskStatusSubject = "litenas-integrity.background-tasks.rpc.status-check"
	// StartBaselineTaskSubject queues an AIDE baseline generation.
	StartBaselineTaskSubject = "litenas-integrity.background-tasks.rpc.baseline.start"
	// InfoSubject reads the active AIDE baseline state.
	InfoSubject = "litenas-integrity.rpc.info"
)
