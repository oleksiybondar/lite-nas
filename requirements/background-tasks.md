# Shared Background Tasks Requirements

This document defines the in-process command-task behavior shared by service
modules. Services own task types, authorization, input validation, and mapping
to trusted command specifications.

## Functional Requirements

### FR-001 Accept service-selected command tasks

The shared background task manager MUST accept only command specifications
constructed by its owning service. It MUST assign each accepted attempt a
unique task ID and expose its lifecycle state for polling.

### FR-002 Capture bounded command results

The manager MUST run configured commands without a shell and retain a bounded
output tail, exit status, timestamps, and terminal result state for each
attempt.

### FR-003 Retry completed task specifications

The manager MUST allow a completed attempt to be retried as a new attempt with
a new ID while preserving the previous attempt for status polling. It MUST
reject retry while the referenced attempt is queued or running.

## Reliability Requirements

### RR-001 Bound task admission and report saturation

The manager MUST use bounded task and result queues. It MUST reject new tasks
when the task queue is full instead of silently discarding them.

The manager MUST expose terminal snapshots on one bounded `Results` channel.
Each owning service MUST keep one result dispatcher consuming that channel and
forward typed outcomes into its existing reporting pipeline. The dispatcher
owns no manager channels; the manager closes `Results` after its event stream
drains, and the service closes its downstream channel after the dispatcher
exits.

### RR-002 Drain task results during shutdown

The manager MUST stop accepting work during shutdown, cancel active commands
through its parent context, mark queued attempts cancelled, and consume all
accepted result events before its run loop exits. Services MUST keep result
dispatchers and downstream reporters running until manager results drain.

## Interface Requirements

### IR-001 Keep task parameters behind the owning service boundary

The shared manager MUST receive trusted command specifications rather than
transport requests. Owning services MUST validate task requests and expose
service-qualified NATS subjects when multiple task owners share one broker.
