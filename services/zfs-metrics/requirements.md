# zfs-metrics Service Requirements

This document defines functional and integration requirements for the
`zfs-metrics` service.

Requirement IDs use the `zfs-metrics-svc/<TYPE>-<NNN>` format.

## Functional Requirements

### zfs-metrics-svc/FR-001 Snapshot Polling

The service must poll ZFS data on a configurable interval and trigger collection
immediately on startup.

### zfs-metrics-svc/FR-002 Zpool Command Collection

For each polling cycle, the service must collect raw outputs from:

- `zpool status -P -L`
- `zpool list -H -p -o name,size,alloc,free,cap,health`
- `zpool iostat -H -p`

### zfs-metrics-svc/FR-003 Structured Parsing

The service must parse collected command outputs into structured intermediate
models using shared parser packages under `shared/go/parsers/zfs`.

### zfs-metrics-svc/FR-004 Snapshot Composition

The service must compose one normalized `metrics.ZFSSnapshot` entity from the
parsed `status`, `list`, and `iostat` data for each polling cycle.

### zfs-metrics-svc/FR-005 Snapshot Event Publication

After successful snapshot composition, the service must publish
`zfsmetrics.SnapshotUpdatedEvent` to
`zfs.metrics.events.snapshot`.

### zfs-metrics-svc/FR-006 Latest Snapshot RPC

The service must expose request/reply access to latest snapshot via
`zfs.metrics.rpc.snapshot`.

When no snapshot has been collected yet, the service must return
`Available=false`.

### zfs-metrics-svc/FR-007 Poll Failure Handling

If one polling cycle fails, the service must log the failure and continue
subsequent polling cycles.

### zfs-metrics-svc/FR-008 Zpool Path Validation

Before command execution, the service must validate configured `zpool_path`:

- non-empty
- absolute path
- basename equals `zpool`
- exists and is not a directory
- matches an allowed path from service allowlist

## Integration Requirements

### zfs-metrics-svc/IR-001 Runtime Lifecycle

The service must support graceful shutdown on process context cancellation and
drain runtime resources through shared infra close routines.

### zfs-metrics-svc/IR-002 Shared Infrastructure

The service must use shared infrastructure modules for:

- configuration loading
- logging
- NATS messaging client/server lifecycle
- timer worker orchestration

### zfs-metrics-svc/IR-003 Contract Compatibility

Published events and RPC payloads must conform to
`shared/go/contracts/zfsmetrics`.
