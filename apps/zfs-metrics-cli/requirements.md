# zfs-metrics-cli Application Requirements

This document defines functional and integration requirements for the
`zfs-metrics-cli` application.

Requirement IDs use the `zfs-metrics-cli/<TYPE>-<NNN>` format.

## Functional Requirements

### zfs-metrics-cli/FR-001 Snapshot Retrieval Command

The CLI must provide a command path to request the latest ZFS snapshot from
`zfs-metrics` over request/reply messaging.

### zfs-metrics-cli/FR-002 Availability-Aware Output

When service response reports snapshot unavailable (`Available=false`), the CLI
must render a clear non-error message and exit successfully.

### zfs-metrics-cli/FR-003 Structured Output Modes

The CLI must support machine-readable output for automation use-cases.

### zfs-metrics-cli/FR-004 Human Output Mode

The CLI must support a human-readable output mode optimized for terminal usage.

### zfs-metrics-cli/FR-005 Input Validation

The CLI must validate user-provided arguments/options at the CLI boundary
before invoking downstream processing.

### zfs-metrics-cli/FR-006 Error Reporting

The CLI must print actionable error messages to stderr and return non-zero exit
code for request/config/runtime failures.

## Integration Requirements

### zfs-metrics-cli/IR-001 Shared Config Compatibility

The CLI must support shared LiteNAS configuration loading patterns for
messaging and logging settings.

### zfs-metrics-cli/IR-002 Contract Compatibility

CLI request and response handling must conform to
`shared/go/contracts/zfsmetrics`.

### zfs-metrics-cli/IR-003 Service Interoperability

The CLI must interoperate with a deployed `zfs-metrics` service using the same
TLS/NATS trust model as other LiteNAS runtime components.

## Current Implementation Status

Current code is a runtime stub. Requirements above describe the target
behavior for full implementation.
