# Logging Managers — Requirements

## Overview

The Logging Managers are host-local services responsible for ingesting,
persisting, and serving operational event data for LiteNAS.

LiteNAS includes two managers that share the same core behavior:

- `security-logging-manager`
- `system-logging-manager`

They use the same persistence and resource-management model, while input
subjects and domain-specific payload contracts may differ by service.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Ingest event messages from messaging subjects

#### FR-001 Description

Each logging manager MUST consume event-related messages from its configured
messaging subscriptions and process them through a shared eventstore flow.

#### FR-001 Input

- Incoming messaging traffic on manager-specific subscribed subjects
- Validated event-related payloads

#### FR-001 Output

- Internal write operations to eventstore state and history records

#### FR-001 Acceptance Criteria

- Each manager consumes only its configured subjects
- Subject configuration supports separate source streams per manager
- Ingestion supports continuous processing under normal operating load
- Message handling failures for one message do not stop the service loop

---

### FR-002 Maintain bounded current event state

#### FR-002 Description

The system MUST maintain bounded current event state for tracked entities.

#### FR-002 Input

- Event identity and state updates from FR-001

#### FR-002 Output

- Current event records with updated state and timestamps

#### FR-002 Acceptance Criteria

- Current state storage is bounded by configured `max_events`
- Event updates target existing logical slots or assigned slots within bounds
- Event state writes do not cause unbounded growth in current-state storage

---

### FR-003 Maintain lifecycle state per event

#### FR-003 Description

The system MUST maintain lifecycle state associated with current events.

#### FR-003 Input

- Lifecycle updates such as acknowledge and mute operations

#### FR-003 Output

- Persisted lifecycle records associated with current event records

#### FR-003 Acceptance Criteria

- Lifecycle state is maintained as a one-to-one mapping with current event slots
- Lifecycle updates are persisted durably in the same SQLite database
- Lifecycle fields support operator attribution and timestamps

---

### FR-004 Persist historical occurrences with bounded retention

#### FR-004 Description

The system MUST persist event occurrence history for diagnostics and audit
scenarios with bounded retention.

#### FR-004 Input

- Occurrence records generated from ingested event traffic

#### FR-004 Output

- Time-ordered persisted occurrence records

#### FR-004 Acceptance Criteria

- Occurrences are stored in append-oriented form
- Total occurrence count is bounded by configured `max_occurrences`
- Occurrence cleanup removes oldest records in configured cleanup batches
- Occurrence retention remains bounded without manual intervention

---

### FR-005 Support eventstore query access

#### FR-005 Description

The system MUST expose read access to current event state, lifecycle state, and
recent occurrence history for manager-owned data.

#### FR-005 Input

- Read/query requests from manager consumers

#### FR-005 Output

- Query results containing current state, lifecycle state, and/or occurrences

#### FR-005 Acceptance Criteria

- Query access supports current event retrieval
- Query access supports lifecycle retrieval
- Query access supports recent occurrence retrieval
- Query behavior does not require loading the full occurrence history into RAM

---

## Interface Requirements

### IR-001 Keep manager messaging contracts independently configurable

#### IR-001 Description

The two logging managers MUST support independent messaging subject contracts
while reusing the same core behavior.

#### IR-001 Input

- Per-manager messaging subject configuration

#### IR-001 Output

- Manager-specific messaging subscription wiring

#### IR-001 Acceptance Criteria

- Security and system managers can subscribe to different subject sets
- Shared core logic does not require per-manager forks for subscription wiring
- Subject and payload contract details are defined in manager-specific interface
  requirements when needed

---

## Reliability Requirements

### RR-001 Use single-writer batched SQLite persistence

#### RR-001 Description

The system MUST serialize SQLite writes through a single-writer flow and use
batch flushing to reduce write amplification while preserving high durability.

#### RR-001 Acceptance Criteria

- Write execution uses a single DB-writer path for concurrent producers
- Flush triggers include configured batch size and flush interval
- Shutdown triggers a final flush attempt
- Critical lifecycle writes support expedited flushing behavior
- Successfully flushed batches are durably committed before acknowledgment of
  batch completion

---

### RR-002 Provide high durability with a bounded crash-loss window

#### RR-002 Description

The system MUST operate as effectively lossless during normal service operation
and MUST bound potential loss during abrupt process or host failure to the
pre-flush in-memory batch window.

#### RR-002 Acceptance Criteria

- During normal operation, accepted messages are persisted through periodic or
  threshold-based batch flushing without intentional dropping
- Unflushed data-loss risk on crash is limited to queued/unflushed batch
  contents
- The service behavior and operator documentation describe this durability
  tradeoff explicitly
- Batch and flush settings are configurable to tune the loss window

---

## Operational Requirements

### OR-001 Enforce explicit storage and memory bounds

#### OR-001 Description

The system MUST provide explicit, configurable bounds for eventstore capacity
to support predictable operation on constrained hardware.

#### OR-001 Acceptance Criteria

- Event capacity is controlled by `max_events`
- Occurrence retention capacity is controlled by `max_occurrences`
- Runtime behavior does not depend on loading full historical data into RAM
- Resource bounds are evaluable from configuration values

---

### OR-002 Provide shared reusable logging-manager configuration

#### OR-002 Description

The system MUST provide a shared configuration contract reusable by both
logging-manager services.

#### OR-002 Acceptance Criteria

- Shared config includes messaging and eventstore sections
- Eventstore config sections include storage, writer, and cleanup controls
- Both managers can use the same config schema with different values

---

### OR-003 Prioritize SD-card-friendly persistence behavior

#### OR-003 Description

The persistence strategy MUST minimize unnecessary write amplification for
SD-card-backed deployment environments.

#### OR-003 Acceptance Criteria

- Batched writes are used instead of per-message forced writes
- Cleanup executes in batches rather than continuous per-record deletion
- Current event and lifecycle storage remains bounded without unbounded growth

---

## Security Requirements

### SR-001 Keep manager data domains isolated by contract

#### SR-001 Description

Security and system logging domains MUST remain logically separated by service
contract and data ownership.

#### SR-001 Acceptance Criteria

- Each manager handles only its configured subscription domain
- Cross-domain data mixing is prevented by service-level contract boundaries
- Separation is preserved even when both services reuse the same core module

---

## Testability Requirements

### TR-001 Keep core eventstore behavior verifiable with automated tests

#### TR-001 Description

The shared eventstore and logging-manager configuration behavior MUST be
testable with automated tests without requiring full deployed infrastructure.

#### TR-001 Acceptance Criteria

- Configuration parsing and validation are covered by unit tests
- Bounded retention behavior is testable with deterministic fixtures
- Writer batching and flush trigger behavior is testable with controlled inputs
- Manager-specific subject wiring can be validated independently from shared
  core logic

---

## Notes

- This document defines common requirements shared by both logging-manager
  services
- Service-specific messaging subjects and payload contracts are expected to be
  defined in follow-up manager-specific interface requirements
