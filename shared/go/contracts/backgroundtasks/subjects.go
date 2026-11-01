// Package backgroundtasks defines the shared RPC contracts for long-running
// command jobs. Owning services select and validate command implementations.
package backgroundtasks

const (
	// StartSubject accepts a service-defined task type and task-specific payload.
	StartSubject = "background-tasks.rpc.start"
	// RetrySubject starts a fresh attempt from a completed task's saved request.
	RetrySubject = "background-tasks.rpc.retry"
	// StatusCheckSubject returns the latest state and bounded output of one attempt.
	StatusCheckSubject = "background-tasks.rpc.status-check"
)
