# System Metrics — Requirements

## Overview

The System Metrics service is responsible for collecting CPU and RAM
metrics from the host system, maintaining a short-term history, and
providing access to current and historical metrics.

The service operates on a polling model and produces metrics updates at
a fixed interval.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Collect system metrics

#### FR-001 Description

The service MUST periodically collect CPU and RAM metrics from the host
system, including total CPU usage and per-core CPU usage.

#### FR-001 Input

- Host system metrics interfaces
- Poll interval (configurable)

#### FR-001 Output

- Metrics data structure containing:
  - timestamp
  - total CPU usage (%)
  - per-core CPU usage (%)
  - total RAM (bytes)
  - used RAM (bytes)
  - used RAM (%)

#### FR-001 Acceptance Criteria

- Metrics are collected at the configured interval (default: 1 second)
- CPU metrics include both total and per-core values
- RAM metrics include total, used, and percentage
- Collection continues even if individual metric sources temporarily fail

---

### FR-002 Maintain metrics history

#### FR-002 Description

The service MUST maintain a time-bounded history of collected metrics.

#### FR-002 Input

- Stream of collected metrics from FR-001

#### FR-002 Output

- Time-ordered collection of metrics entries

#### FR-002 Acceptance Criteria

- History contains at most 120 seconds of data
- Older entries are discarded automatically
- Retrieval returns data in chronological order
- Memory usage remains bounded

---

### FR-003 Provide current metrics snapshot

#### FR-003 Description

The service MUST provide access to the most recent metrics snapshot.

#### FR-003 Input

- Latest collected metrics from FR-001

#### FR-003 Output

- Single metrics data structure representing the latest state

#### FR-003 Acceptance Criteria

- Snapshot always reflects the most recently collected metrics
- Snapshot is available even if history is not yet populated

---

### FR-004 Provide metrics history

#### FR-004 Description

The service MUST provide access to collected metrics history within the
configured retention window.

#### FR-004 Input

- Request for metrics history

#### FR-004 Output

- Collection of metrics entries within the history window

#### FR-004 Acceptance Criteria

- Returned history does not exceed the configured time window
- Entries are ordered chronologically
- Empty result is allowed if no data is available

---

### FR-005 Produce metrics update events

#### FR-005 Description

The service MUST produce a metrics update event for each completed
metrics collection cycle.

#### FR-005 Input

- Collected metrics from FR-001

#### FR-005 Output

- Metrics update event containing the current metrics snapshot

#### FR-005 Acceptance Criteria

- An event is produced after each successful metrics collection
- Event production frequency matches the configured poll interval
- Event production MUST NOT depend on previous metric values
- Event production operates independently of external systems

---

## Interface Requirements

### IR-001 Respond to getStats requests

#### IR-001 Description

The service MUST respond to requests for current metrics and recent
history via the messaging system.

#### IR-001 Input

- Request message for system metrics

#### IR-001 Output

- Response message containing:
  - current metrics snapshot
  - metrics history (up to 120 seconds)

#### IR-001 Acceptance Criteria

- Response includes both current metrics and history
- Response format is JSON-serializable
- Response is returned within 1 second under normal conditions

---

### IR-002 Publish metrics updates

#### IR-002 Description

The service MUST publish metrics update events via the messaging system.

#### IR-002 Input

- Metrics update events from FR-005

#### IR-002 Output

- Event message containing the current metrics snapshot

#### IR-002 Acceptance Criteria

- An event is published for each metrics update event
- Event payload is JSON-serializable
- Event publishing MUST NOT block metric collection
- Events are published at a frequency matching the poll interval

---

## Notes

f

- The service follows a continuous polling and publishing model
- No change-detection or threshold-based filtering is performed
- Consumers are responsible for interpreting and filtering metrics
- Implementation details such as OS-specific data sources are excluded
  from this document
- Messaging system specifics (e.g. subjects, authentication) will be
  defined in additional requirements
