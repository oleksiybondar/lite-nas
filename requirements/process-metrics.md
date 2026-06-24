# Process Metrics — Requirements

## Overview

The Process Metrics service is responsible for collecting a live host process
inventory directly from Linux procfs, composing one atomic snapshot per polling
cycle, and providing access to the latest collected snapshot.

The service is intentionally limited to process inventory and per-process
resource usage. It MUST NOT become a service-manager surface and MUST NOT
contain systemd or service lifecycle logic.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Poll procfs and compose one atomic process snapshot

#### FR-001 Description

The service MUST poll procfs on a configurable interval and compose one atomic
process snapshot for each successful polling cycle.

#### FR-001 Acceptance Criteria

- Polling runs continuously at the configured interval
- The default poll interval is 5 seconds
- Each successful cycle produces exactly one snapshot
- All process entries within one snapshot represent the same polling cycle

---

### FR-002 Report live process inventory and factual process identity fields

#### FR-002 Description

The service MUST report a live process inventory derived directly from Linux
procfs and expose factual identity and hierarchy fields for each collected
process.

#### FR-002 Acceptance Criteria

- Each process entry includes `pid`
- Each process entry includes `ppid`
- Each process entry includes `name`
- Each process entry includes `state`
- Each process entry includes `uid`
- Each process entry includes `gid`
- The service may include `username` when it can be resolved
- Process names containing spaces or parentheses are parsed correctly

---

### FR-003 Tolerate partial procfs visibility and disappearing processes

#### FR-003 Description

The service MUST tolerate partial per-process read failures without failing the
whole snapshot.

#### FR-003 Acceptance Criteria

- A single PID read failure does not fail the whole snapshot
- Disappearing processes are skipped safely
- Permission-denied reads do not terminate the polling loop
- Missing `cmdline`, `exe`, or `cwd` data is tolerated
- Kernel threads with empty command lines remain representable

---

### FR-004 Report raw process CPU and memory counters in UNIX-friendly units

#### FR-004 Description

The service MUST expose raw process CPU and memory counters in stable,
UNIX-friendly units rather than derived percentages.

#### FR-004 Acceptance Criteria

- CPU counters are reported as raw clock ticks
- Resident memory is reported in bytes
- Virtual memory size is reported in bytes
- Process start time is derived from boot time and process start ticks when
  that information is available

---

## Interface Requirements

### IR-001 Provide latest process snapshot over messaging

#### IR-001 Description

The service MUST provide access to the latest process snapshot via the shared
messaging transport.

#### IR-001 Acceptance Criteria

- The service responds to a latest-snapshot request subject
- The response indicates whether a snapshot is available
- The response payload is JSON-serializable
- The service does not expose a history RPC for process snapshots

---

### IR-002 Publish process snapshot update events

#### IR-002 Description

The service MUST publish the latest process snapshot after each successful
polling cycle.

#### IR-002 Acceptance Criteria

- Each successful polling cycle publishes one snapshot event
- The event payload is JSON-serializable
- Event publishing failures are isolated from the polling loop
