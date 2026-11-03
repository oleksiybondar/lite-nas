# Service Metrics — Requirements

## Overview

The Service Metrics service is responsible for collecting a factual
service-manager view of host services from systemd, composing one atomic
snapshot per polling cycle, maintaining a short in-memory history, and
providing access to current and historical snapshots.

The service follows the same snapshot-first runtime model as the Network
Metrics and Disk Metrics services. Its contract is intentionally limited to
systemd-managed service state, unit ownership, process relationship, and
cgroup-scoped service metadata. It MUST NOT become a general-purpose process
monitor or task-manager surface.

Primary target platform is Debian and Ubuntu hosts that use systemd as the
service manager.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Poll service-manager state and compose one atomic snapshot

#### FR-001 Description

The service MUST poll systemd-managed service state on a configurable interval
and compose one atomic service snapshot for each successful polling cycle.

#### FR-001 Input

- Configured poll interval
- systemd runtime metadata
- Related cgroup runtime metadata

#### FR-001 Output

- One `ServiceMetricsSnapshot` value per successful polling cycle

#### FR-001 Acceptance Criteria

- Polling runs continuously at the configured interval
- The default poll interval is 1 second
- Each successful cycle produces exactly one snapshot
- All service entries within one snapshot represent the same polling cycle

---

### FR-002 Inventory managed systemd service units

#### FR-002 Description

The service MUST report managed systemd service units as the primary inventory
surface for the snapshot contract.

#### FR-002 Input

- systemd unit inventory
- systemd unit-file inventory

#### FR-002 Output

- Snapshot `services` collection containing one entry per discovered service
  unit

#### FR-002 Acceptance Criteria

- The snapshot reports systemd `*.service` units only
- The inventory is derived from systemd-managed unit data rather than a raw
  process list
- The service MAY merge runtime unit inventory and unit-file inventory so
  enablement state remains visible even when a service is inactive
- The service does not require a unit to be active in order to include it in
  the snapshot when systemd can still identify it as a managed service unit

---

### FR-003 Collect factual service state and enablement fields

#### FR-003 Description

The service MUST collect factual runtime and enablement fields for each
reported service unit without converting them into higher-level health
judgments.

#### FR-003 Input

- systemd unit properties
- systemd unit-file state metadata

#### FR-003 Output

- Service entry fields describing runtime and enablement state

#### FR-003 Acceptance Criteria

- Each service entry includes `name`
- Each service entry includes `manager`
- Each service entry includes `unit_type`
- Each service entry includes `load_state` when runtime state is available
- Each service entry includes `active_state` when runtime state is available
- Each service entry includes `sub_state` when runtime state is available
- Each service entry includes `enabled_state` when unit-file state is available
- Each service entry may include `result`
- The contract preserves systemd state strings rather than rewriting them into
  custom enums

---

### FR-004 Report service failures and failure-adjacent outcome fields

#### FR-004 Description

The service MUST expose factual failure-relevant unit outcome fields from
systemd so downstream consumers can reason about failed or degraded services.

#### FR-004 Input

- systemd unit result and failure metadata

#### FR-004 Output

- Service entry failure-relevant state fields

#### FR-004 Acceptance Criteria

- A failed service remains visible through factual state fields such as
  `active_state`, `sub_state`, and `result`
- The snapshot does not add derived booleans such as `is_healthy`,
  `needs_restart`, or `severity`
- Failure reporting is unit-scoped and does not require per-process metrics

---

### FR-005 Report service ownership and unit-to-process relationships

#### FR-005 Description

The service MUST report factual ownership and unit-to-process relationship data
for each service unit without exposing unrelated global process inventory.

#### FR-005 Input

- systemd unit properties
- Unit cgroup membership data

#### FR-005 Output

- Service entry ownership and process relationship fields

#### FR-005 Acceptance Criteria

- Each service entry may include `main_pid`
- Each service entry may include `control_pid`
- Each service entry may include `pids`
- Each service entry may include `cgroup`
- Each service entry may include `slice`
- `pids` contains only process IDs that belong to the reported unit
- The snapshot does not expose a host-wide raw process list
- The snapshot does not expose per-process CPU, memory, open-file, thread, or
  task-manager style fields

---

### FR-006 Report service resource metadata scoped to the unit cgroup

#### FR-006 Description

The service MUST report service resource metadata only at the unit or cgroup
scope and MUST NOT devolve into per-process resource accounting.

#### FR-006 Input

- systemd unit accounting properties
- Related cgroup resource files when needed

#### FR-006 Output

- Optional service entry `resources` object

#### FR-006 Acceptance Criteria

- `resources` may include `memory_bytes`
- `resources` may include `cpu_usage_usec`
- Resource values represent the service unit or service cgroup as a whole
- The snapshot does not include per-process CPU or memory values
- Missing optional resource-accounting data does not fail the whole snapshot

---

### FR-007 Report supporting unit metadata relevant to service management

#### FR-007 Description

The service MUST include supporting factual unit metadata that helps consumers
identify where a service definition and runtime ownership originate.

#### FR-007 Input

- systemd unit properties

#### FR-007 Output

- Supporting service entry metadata

#### FR-007 Acceptance Criteria

- Service entries may include `description`
- Service entries may include `fragment_path`
- Service entries may include `started_at`
- Optional metadata is omitted when the host or unit does not expose it
- The service does not invent placeholder values for unavailable metadata

---

### FR-008 Keep the snapshot contract factual and service-scoped

#### FR-008 Description

The service MUST keep the snapshot contract factual, service-scoped, and free
from collector internals or process-monitor features.

#### FR-008 Input

- Collected service metrics from FR-002 through FR-007

#### FR-008 Output

- Normalized snapshot contract suitable for downstream consumers

#### FR-008 Acceptance Criteria

- The snapshot root shape is:
  - `timestamp`
  - `services`
- Each service entry uses one normalized DTO shape
- The snapshot does not include a `collector` section
- The snapshot does not include raw command output or `systemctl` text dumps
- The snapshot does not include host-wide process-inspection sections

---

### FR-009 Prefer systemd D-Bus collection over systemctl CLI execution

#### FR-009 Description

The service MUST collect systemd service-manager data from systemd APIs and
MUST prefer D-Bus-backed property access over spawning `systemctl`.

#### FR-009 Input

- systemd D-Bus APIs
- Related cgroup runtime metadata

#### FR-009 Output

- Service snapshot populated from supported host interfaces

#### FR-009 Acceptance Criteria

- Runtime service-manager fields are sourced from systemd APIs
- `systemctl` CLI execution is not required for normal polling
- Related cgroup files may be read for fields that are cgroup-owned rather than
  native unit properties
- Polling continues across partial field-read failures when a snapshot can
  still be composed

---

### FR-010 Maintain bounded snapshot history

#### FR-010 Description

The service MUST maintain a bounded in-memory history of collected service
snapshots.

#### FR-010 Input

- Stream of successfully collected snapshots from FR-001

#### FR-010 Output

- Chronological retained snapshot history

#### FR-010 Acceptance Criteria

- The default retained history contains at most 120 snapshots
- Older snapshots are discarded automatically when the bound is reached
- History retrieval returns snapshots in chronological order
- Runtime memory usage for history remains bounded by configuration

---

### FR-011 Provide the latest service snapshot

#### FR-011 Description

The service MUST provide access to the latest successfully collected service
snapshot.

#### FR-011 Input

- Latest snapshot retained from FR-001

#### FR-011 Output

- Single latest service snapshot

#### FR-011 Acceptance Criteria

- The latest snapshot always reflects the most recently completed successful
  polling cycle
- The latest snapshot can be returned independently of history retrieval
- When no snapshot has been collected yet, the service can report that no
  snapshot is available

---

### FR-012 Provide retained service snapshot history

#### FR-012 Description

The service MUST provide access to retained service snapshot history within the
configured retention bound.

#### FR-012 Input

- Request for snapshot history

#### FR-012 Output

- Chronological collection of retained service snapshots

#### FR-012 Acceptance Criteria

- Returned history does not exceed the retained in-memory bound
- Returned snapshots are ordered chronologically
- Empty history is allowed when no snapshot has been collected yet

---

### FR-013 Produce service snapshot update events

#### FR-013 Description

The service MUST produce a service snapshot update event for each completed
successful polling cycle.

#### FR-013 Input

- Collected service snapshot from FR-001

#### FR-013 Output

- Service snapshot update event containing the current snapshot

#### FR-013 Acceptance Criteria

- An event is produced after each successful snapshot collection
- Event production frequency matches the configured poll interval
- Event payloads are based on the normalized snapshot contract
- Event publication MUST NOT depend on change detection between snapshots

---

## Interface Requirements

### IR-001 Respond to latest snapshot requests

#### IR-001 Description

The service MUST respond to messaging requests for the latest service snapshot.

#### IR-001 Input

- Request message for latest service snapshot

#### IR-001 Output

- Response containing:
  - `available`
  - `snapshot`

#### IR-001 Acceptance Criteria

- The response shape is JSON-serializable
- `available=false` is returned when no snapshot has been collected yet
- `snapshot` is omitted when `available=false`

---

### IR-002 Respond to retained history requests

#### IR-002 Description

The service MUST respond to messaging requests for retained service snapshot
history.

#### IR-002 Input

- Request message for retained service snapshot history

#### IR-002 Output

- Response containing:
  - `items`

#### IR-002 Acceptance Criteria

- The response shape is JSON-serializable
- `items` is ordered chronologically
- An empty `items` collection is valid when no history is available

---

### IR-003 Publish service snapshot updates

#### IR-003 Description

The service MUST publish service snapshot update events via the messaging
system.

#### IR-003 Input

- Service snapshot update events from FR-013

#### IR-003 Output

- Event message containing the current service snapshot

#### IR-003 Acceptance Criteria

- An event is published for each service snapshot update event
- Event payloads are JSON-serializable
- Event publication MUST NOT block snapshot collection

---

### IR-004 Follow the shared metrics-service implementation analogy

#### IR-004 Description

The service MUST follow the same implementation analogy already used by the
existing metrics services unless a later requirement explicitly diverges.

#### IR-004 Input

- Service configuration
- Shared runtime infrastructure

#### IR-004 Output

- Service runtime assembled with the same architectural pattern used by peer
  metrics services

#### IR-004 Acceptance Criteria

- The service uses a polling worker that composes one snapshot per cycle
- The service stores latest snapshot and bounded history in a dedicated state
  component
- The service exposes latest snapshot and history through dedicated contract
  packages under `shared/go/contracts/servicemetrics`
- The service keeps its snapshot DTOs under `shared/go/metrics`
- The service uses shared infrastructure patterns for configuration loading,
  logging, messaging, runtime lifecycle, and worker orchestration

---

## Field Source Mapping

The following mapping defines the expected ownership of snapshot fields by
source. The service should prefer systemd D-Bus properties for manager-owned
state and use cgroup data only for fields that are inherently cgroup-scoped or
not consistently surfaced as unit properties.

| Field | Expected primary source | Notes |
| --- | --- | --- |
| `timestamp` | service runtime clock | Same polling-cycle timestamp model as other metrics services |
| `services[].name` | systemd unit identity | Unit name such as `smbd.service` |
| `services[].manager` | constant | Expected value: `systemd` |
| `services[].unit_type` | systemd unit identity | Expected value for initial scope: `service` |
| `services[].description` | systemd unit property | Factual unit description when available |
| `services[].load_state` | systemd unit property | Runtime property |
| `services[].active_state` | systemd unit property | Runtime property |
| `services[].sub_state` | systemd unit property | Runtime property |
| `services[].enabled_state` | systemd unit-file state | Enablement state is not the same as runtime activity |
| `services[].result` | systemd unit property | Runtime outcome/failure-adjacent property |
| `services[].main_pid` | systemd unit property | Unit-scoped PID relationship |
| `services[].control_pid` | systemd unit property | Unit-scoped PID relationship |
| `services[].pids` | unit cgroup membership | Unit-owned PID list only, not global process inventory |
| `services[].cgroup` | systemd unit property | Unit cgroup path when available |
| `services[].slice` | systemd unit property | Ownership context such as `system.slice` |
| `services[].fragment_path` | systemd unit property | Unit definition path when available |
| `services[].started_at` | systemd unit property | Normalized timestamp derived from systemd runtime timestamps when available |
| `services[].resources.memory_bytes` | systemd unit property or cgroup memory file | Prefer systemd-exposed unit accounting; otherwise read unit cgroup value |
| `services[].resources.cpu_usage_usec` | systemd unit property or cgroup CPU stat | Prefer systemd-exposed unit accounting; otherwise read unit cgroup value |

---

## Proposed Snapshot Schema

```json
{
  "timestamp": "RFC3339 timestamp",
  "services": [
    {
      "name": "string",
      "manager": "systemd",
      "unit_type": "service",
      "description": "string",
      "load_state": "string",
      "active_state": "string",
      "sub_state": "string",
      "enabled_state": "string",
      "result": "string",
      "main_pid": 0,
      "control_pid": 0,
      "pids": [0],
      "cgroup": "string",
      "slice": "string",
      "fragment_path": "string",
      "started_at": "RFC3339 timestamp",
      "resources": {
        "memory_bytes": 0,
        "cpu_usage_usec": 0
      }
    }
  ]
}
```

### JSON Example

```json
{
  "timestamp": "2026-06-13T09:14:22Z",
  "services": [
    {
      "name": "smbd.service",
      "manager": "systemd",
      "unit_type": "service",
      "description": "Samba SMB Daemon",
      "load_state": "loaded",
      "active_state": "active",
      "sub_state": "running",
      "enabled_state": "enabled",
      "result": "success",
      "main_pid": 1234,
      "control_pid": 0,
      "pids": [1234, 1235],
      "cgroup": "/system.slice/smbd.service",
      "slice": "system.slice",
      "fragment_path": "/usr/lib/systemd/system/smbd.service",
      "started_at": "2026-06-13T08:50:10Z",
      "resources": {
        "memory_bytes": 43896832,
        "cpu_usage_usec": 1812500
      }
    },
    {
      "name": "nmbd.service",
      "manager": "systemd",
      "unit_type": "service",
      "description": "Samba NMB Daemon",
      "load_state": "loaded",
      "active_state": "failed",
      "sub_state": "failed",
      "enabled_state": "enabled",
      "result": "exit-code",
      "main_pid": 0,
      "control_pid": 0,
      "pids": [],
      "cgroup": "/system.slice/nmbd.service",
      "slice": "system.slice",
      "fragment_path": "/usr/lib/systemd/system/nmbd.service",
      "started_at": "2026-06-13T08:49:58Z",
      "resources": {
        "memory_bytes": 0,
        "cpu_usage_usec": 0
      }
    }
  ]
}
```

---

## Proposed Go DTO Definitions

```go
package metrics

import "time"

// ServiceMetricsSnapshot is the top-level service metrics snapshot payload.
type ServiceMetricsSnapshot struct {
 Timestamp time.Time               `json:"timestamp"`
 Services  []ServiceUnitSnapshot   `json:"services"`
}

// ServiceUnitSnapshot represents one systemd-managed service unit snapshot.
type ServiceUnitSnapshot struct {
 Name         string                `json:"name"`
 Manager      string                `json:"manager"`
 UnitType     string                `json:"unit_type"`
 Description  *string               `json:"description,omitempty"`
 LoadState    *string               `json:"load_state,omitempty"`
 ActiveState  *string               `json:"active_state,omitempty"`
 SubState     *string               `json:"sub_state,omitempty"`
 EnabledState *string               `json:"enabled_state,omitempty"`
 Result       *string               `json:"result,omitempty"`
 MainPID      *uint32               `json:"main_pid,omitempty"`
 ControlPID   *uint32               `json:"control_pid,omitempty"`
 PIDs         []uint32              `json:"pids,omitempty"`
 CGroup       *string               `json:"cgroup,omitempty"`
 Slice        *string               `json:"slice,omitempty"`
 FragmentPath *string               `json:"fragment_path,omitempty"`
 StartedAt    *time.Time            `json:"started_at,omitempty"`
 Resources    *ServiceUnitResources `json:"resources,omitempty"`
}

// ServiceUnitResources stores cgroup-scoped resource values for one service.
type ServiceUnitResources struct {
 MemoryBytes  *uint64 `json:"memory_bytes,omitempty"`
 CPUUsageUsec *uint64 `json:"cpu_usage_usec,omitempty"`
}
```

The matching shared contract packages should follow the same pattern used by
the existing metrics services:

```go
package servicemetrics

import "lite-nas/shared/metrics"

type GetSnapshotRequest struct{}

type GetSnapshotResponse struct {
 Available bool                           `json:"available"`
 Snapshot  metrics.ServiceMetricsSnapshot `json:"snapshot,omitempty"`
}

type SnapshotUpdatedEvent struct {
 Snapshot metrics.ServiceMetricsSnapshot `json:"snapshot"`
}

type GetHistoryRequest struct{}

type GetHistoryResponse struct {
 Items []metrics.ServiceMetricsSnapshot `json:"items"`
}
```

---

## Notes

- The service is intentionally a service-manager metrics surface, not a process
  inspector.
- The contract preserves factual systemd states rather than normalizing them
  into LiteNAS-specific health enums.
- Unit-scoped PID relationships are allowed because they express service
  ownership boundaries, but they do not authorize per-process monitoring
  features.
- Resource fields should remain optional because unit accounting availability
  depends on host configuration and systemd exposure.
