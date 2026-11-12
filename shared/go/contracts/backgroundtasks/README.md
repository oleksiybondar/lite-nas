# Background task RPC contract

The contract defines three request/reply operations: `StartSubject`,
`RetrySubject`, and `StatusCheckSubject`. A start request names a task type
implemented by that service and carries its task-specific JSON parameters. The
service validates those parameters and maps them to a trusted command
configuration; clients never submit executable paths or raw command arguments.
When several services expose tasks on the same NATS server, each service must
use service-qualified subjects with these request and response types so one RPC
cannot be handled by multiple task owners. Antivirus and integrity currently
use `antivirus.background-tasks.rpc.*` and
`litenas-integrity.background-tasks.rpc.*` respectively.

Retry is a new attempt, not process continuation. It is allowed only after the
referenced attempt is terminal, uses the saved validated request, and returns a
new `task_id`. The old attempt remains available for status polling. The generic
runner stores task state in memory, so status history disappears when the owning
service restarts.

Status can be polled while a command runs. The shared manager accepts tasks into
a bounded queue, runs them through one command worker, and sends state,
progress, and completion events through a bounded output channel to its state
consumer. The state consumer publishes terminal snapshots on a separate bounded
`Results` channel; exactly one owning service dispatcher consumes it and maps
snapshots into that service's typed reporting channel. Queue saturation is
returned as an error. Output is a bounded tail
and `output_truncated` reports discarded older bytes. `progress` is optional
because commands such as `aide --check` do not expose a reliable percentage by
default; a service may add a progress parser for commands with a documented
progress format. Shutdown stops task admission, cancels active work through the
service context, accounts for queued tasks as cancelled, and drains result
events before the manager returns. Services keep the result dispatcher and
reporting worker alive until the manager closes `Results`, then close their
downstream result channel from its owner.

JSON Schemas for the request and response messages are in `schemas/`.
