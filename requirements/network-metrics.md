# Network Metrics — Requirements

## Overview

The Network Metrics service is responsible for collecting factual network-state
data from the host system, composing one atomic snapshot per polling cycle,
maintaining a short in-memory history, and providing access to current and
historical snapshots.

The service follows the same general polling and retention model as the other
metrics services, but its snapshot contract is domain-specific and includes
interface, protocol, socket, and kernel-pressure sections.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Poll network metrics and compose one atomic snapshot

#### FR-001 Description

The service MUST poll host network metrics on a configurable interval and
compose one atomic network snapshot for each successful polling cycle.

#### FR-001 Input

- Configured poll interval
- Host network metric sources

#### FR-001 Output

- One `NetworkMetricsSnapshot` value per successful polling cycle

#### FR-001 Acceptance Criteria

- Polling runs continuously at the configured interval
- The default poll interval is 1 second
- Each successful cycle produces exactly one snapshot
- All snapshot sections in one snapshot represent the same polling cycle

---

### FR-002 Collect interface metrics from host network interfaces

#### FR-002 Description

The service MUST collect per-interface factual metrics and attributes from the
host network interfaces.

#### FR-002 Input

- `/proc/net/dev`
- `/sys/class/net/<iface>/*`
- `/sys/class/net/<iface>/statistics/*`

#### FR-002 Output

- Snapshot `interfaces` collection containing one entry per discovered
  interface

#### FR-002 Acceptance Criteria

- Each interface entry includes stable interface identity such as interface name
- Each interface entry may include factual adapter identity metadata such as
  vendor ID, device ID, vendor name, device name, and human-readable adapter
  description when locally resolvable
- Each interface entry may include factual link attributes such as MTU, MAC
  address, operational state, carrier state, speed, duplex, and queue length
  when available from the host
- Interface counters remain factual/raw and are not converted into derived
  security judgments
- The service does not require interfaces to be in a specific state to include
  them in the snapshot

---

### FR-003 Collect protocol counters from kernel network statistics

#### FR-003 Description

The service MUST collect protocol-level counters from the Linux kernel network
statistics interfaces.

#### FR-003 Input

- `/proc/net/snmp`
- `/proc/net/netstat`

#### FR-003 Output

- Snapshot `protocols` section containing protocol counter groups

#### FR-003 Acceptance Criteria

- The snapshot includes protocol groups derived from available kernel sources,
  such as IP, ICMP, TCP, UDP, and kernel extension groups
- Protocol values are represented primarily as raw counters
- Counter naming remains close to the kernel-provided names rather than being
  rewritten into higher-level judgments
- Missing protocol groups from the host are represented without crashing the
  polling cycle

---

### FR-004 Collect summarized socket metrics

#### FR-004 Description

The service MUST collect summarized socket-state metrics from the host rather
than emitting full per-socket dumps in the snapshot contract.

#### FR-004 Input

- `/proc/net/tcp`
- `/proc/net/tcp6`
- `/proc/net/udp`
- `/proc/net/udp6`
- `/proc/net/sockstat`

#### FR-004 Output

- Snapshot `sockets` section containing summarized socket metrics

#### FR-004 Acceptance Criteria

- The `sockets` section includes aggregate totals
- The `sockets` section includes counts grouped by socket state where
  applicable
- The `sockets` section includes summarized rankings such as top local ports
  and top remote IPs
- The snapshot does not expose an unbounded list of raw socket rows

---

### FR-005 Collect kernel network pressure indicators

#### FR-005 Description

The service MUST collect factual kernel-side network pressure indicators related
to network soft interrupt activity.

#### FR-005 Input

- `/proc/softirqs`

#### FR-005 Output

- Snapshot `kernel_pressure` section

#### FR-005 Acceptance Criteria

- The snapshot includes network-related softirq counters available from the
  host
- Kernel pressure values remain factual counters rather than interpreted health
  or anomaly conclusions
- The snapshot may include both total and per-CPU values when available

---

### FR-006 Keep the snapshot contract factual and judgment-free

#### FR-006 Description

The service MUST keep the snapshot contract factual/raw and MUST NOT embed
collector metadata or derived security conclusions.

#### FR-006 Input

- Collected network metrics from FR-002 through FR-005

#### FR-006 Output

- Normalized snapshot contract suitable for downstream consumers

#### FR-006 Acceptance Criteria

- The snapshot root shape is:
  - `timestamp`
  - `interfaces`
  - `protocols`
  - `sockets`
  - `kernel_pressure`
- The snapshot does not include a `collector` section
- The snapshot does not include derived security fields such as `ddos`,
  `port_scan`, `conn_storm`, `link_anomaly`, or `suspicious_score`
- Derived judgments are left to downstream consumers rather than emitted by this
  service

---

### FR-007 Maintain bounded snapshot history

#### FR-007 Description

The service MUST maintain a bounded in-memory history of collected network
snapshots.

#### FR-007 Input

- Stream of successfully collected snapshots from FR-001

#### FR-007 Output

- Chronological retained snapshot history

#### FR-007 Acceptance Criteria

- The default retained history contains at most 120 snapshots
- Older snapshots are discarded automatically when the bound is reached
- History retrieval returns snapshots in chronological order
- Runtime memory usage for history remains bounded by configuration

---

### FR-008 Provide the latest network snapshot

#### FR-008 Description

The service MUST provide access to the latest successfully collected network
snapshot.

#### FR-008 Input

- Latest snapshot retained from FR-001

#### FR-008 Output

- Single latest network snapshot

#### FR-008 Acceptance Criteria

- The latest snapshot always reflects the most recently completed successful
  polling cycle
- The latest snapshot can be returned independently of history retrieval
- When no snapshot has been collected yet, the service can report that no
  snapshot is available

---

### FR-009 Provide retained network snapshot history

#### FR-009 Description

The service MUST provide access to retained network snapshot history within the
configured retention bound.

#### FR-009 Input

- Request for snapshot history

#### FR-009 Output

- Chronological collection of retained network snapshots

#### FR-009 Acceptance Criteria

- Returned history does not exceed the retained in-memory bound
- Empty history results are allowed when no snapshots have been collected
- History output remains JSON-serializable

---

### FR-010 Publish snapshot update events

#### FR-010 Description

The service MUST publish a snapshot update event for each successfully composed
network snapshot.

#### FR-010 Input

- Successfully composed snapshot from FR-001

#### FR-010 Output

- Snapshot update event containing the current network snapshot

#### FR-010 Acceptance Criteria

- One event is published for each successful polling cycle
- Event payload contains the composed network snapshot
- Event production frequency follows the successful polling cadence
- Event publication does not depend on change detection between snapshots

---

## Interface Requirements

### IR-001 Expose messaging contracts for latest snapshot and history

#### IR-001 Description

The service MUST expose messaging contracts for retrieving the latest network
snapshot and retained network snapshot history.

#### IR-001 Input

- Messaging requests for latest snapshot
- Messaging requests for retained history

#### IR-001 Output

- Request/reply responses containing latest snapshot or retained history

#### IR-001 Acceptance Criteria

- Latest snapshot access is exposed through a dedicated request/reply contract
- History access is exposed through a dedicated request/reply contract
- When no latest snapshot exists, the latest-snapshot response reports that no
  snapshot is available rather than returning invented data
- Response payloads conform to the shared network-metrics contract package

---

### IR-002 Publish shared snapshot update events

#### IR-002 Description

The service MUST publish network snapshot update events through the shared
network-metrics event contract.

#### IR-002 Input

- Snapshot update events from FR-010

#### IR-002 Output

- Messaging events containing the current network snapshot

#### IR-002 Acceptance Criteria

- Published event payloads conform to the shared network-metrics contract
  package
- Event payloads are JSON-serializable
- Consumers can treat each event as one complete network snapshot

---

### IR-003 Use stable snapshot field naming and timestamp formatting

#### IR-003 Description

The service MUST expose a stable network snapshot JSON contract with explicit
field naming and timestamp formatting.

#### IR-003 Input

- Composed network snapshot from FR-006

#### IR-003 Output

- Stable JSON-serializable snapshot payload

#### IR-003 Acceptance Criteria

- The snapshot uses the field name `timestamp` at the root
- The snapshot timestamp format matches project timestamps such as
  `2026-06-09T10:34:28.461988712+02:00`
- The snapshot uses the section names `interfaces`, `protocols`, `sockets`, and
  `kernel_pressure`
- Repeated entities are represented by arrays, and named snapshot sections are
  represented by objects

---

## Reliability Requirements

### RR-001 Continue polling after collection failures

#### RR-001 Description

The service MUST continue future polling cycles when one polling cycle fails.

#### RR-001 Acceptance Criteria

- Failure to collect one or more source files for one cycle does not terminate
  the service process
- Polling failures are logged with enough detail for diagnosis
- A later successful cycle still produces a snapshot and publication event

---

### RR-002 Isolate publication and request handling from snapshot retention

#### RR-002 Description

The service MUST keep latest-snapshot retention and history retention resilient
to downstream publication failures.

#### RR-002 Acceptance Criteria

- A snapshot that is successfully composed is still retained even if event
  publication fails
- Failure to publish one snapshot event does not block future polling cycles
- Request/reply access to retained snapshots does not depend on successful
  publication of prior events

---

## Operational Requirements

### OR-001 Support configurable polling and history bounds

#### OR-001 Description

The service MUST provide configuration for polling cadence and retained history
capacity.

#### OR-001 Acceptance Criteria

- Poll interval is configurable
- Retained history capacity is configurable
- Default configuration supports 1-second polling and 120 retained snapshots

---

### OR-002 Operate with bounded in-memory retention

#### OR-002 Description

The service MUST operate with bounded in-memory retention rather than
unbounded accumulation of historical network data.

#### OR-002 Acceptance Criteria

- Runtime memory use for retained snapshot history is bounded by configuration
- The service does not retain an unbounded list of prior snapshots in memory
- Socket summarization avoids unbounded snapshot growth from raw socket dumps

---

## Testability Requirements

### TR-001 Keep collection and composition testable without live host state

#### TR-001 Description

The service MUST keep network source reading, snapshot composition, and
messaging publication testable without requiring a live host network state in
unit tests.

#### TR-001 Acceptance Criteria

- Source readers can be substituted with controlled test inputs
- Snapshot composition can be tested from deterministic source fixtures
- Messaging publication and request/reply flows can be tested with injected test
  doubles
- Runtime orchestration can be tested independently from live host files where
  practical
