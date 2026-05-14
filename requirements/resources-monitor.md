# Resources Monitor — Requirements

## Overview

The Resources Monitor service is a worker that consumes metrics snapshot events,
evaluates rule-based alert conditions, and emits alert lifecycle updates.

The service is stateful for active alert tracking and integrates with
logging-manager alert contracts while keeping event-consumption flow resilient to
downstream failures.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Load and validate rule specifications from JSON

#### FR-001 Description

The service MUST load monitoring rules from a JSON-formatted list and reject
invalid rule entries before they are used for runtime evaluation.

#### FR-001 Input

- Rule specification JSON list
- Rule fields: `event`, `event_prefix`, `field`, `condition`, `values`,
  `message`, optional `description`, optional `normal_message`, and event
  metadata fields required by logging-manager event DTOs except date fields

#### FR-001 Output

- Validated in-memory rule set
- Structured validation errors for invalid rule files or entries

#### FR-001 Acceptance Criteria

- Only supported conditions are accepted: `>`, `>=`, `==`, `<=`, `<`, `in`,
  `!=`
- `in` accepts only array-typed `values`
- Numeric comparisons require numeric runtime values and numeric rule values
- String-encoded numeric values (for example `"80"` for a numeric threshold) are
  rejected
- `event_prefix` length is validated to be at most 8 characters

---

### FR-002 Subscribe and process metrics snapshot events

#### FR-002 Description

The service MUST subscribe to the system metrics snapshot event stream and
evaluate matching rules for each received snapshot.

#### FR-002 Input

- Messaging events on `system.metrics.events.stats`
- Validated rule set from FR-001

#### FR-002 Output

- Rule evaluation outcomes for each applicable rule per incoming snapshot

#### FR-002 Acceptance Criteria

- The service subscribes to `system.metrics.events.stats`
- For each incoming snapshot, the service evaluates all rules where
  `rule.event` matches the incoming event identity
- Rule evaluation reads monitored values from nested paths under `data`, where
  `field` uses dot notation (`a.b.c` => `data.a.b.c`)
- Unknown or missing monitored fields do not crash the worker loop

---

### FR-003 Evaluate rule conditions and manage alert state transitions

#### FR-003 Description

The service MUST evaluate each matching rule and apply a three-state lifecycle:
`new -> active`, `active -> active`, `active -> normal`.

#### FR-003 Input

- Rule evaluation outcomes from FR-002
- In-memory active alert cache

#### FR-003 Output

- Lifecycle actions: create event, create occurrence, normalize event
- Updated in-memory active alert cache

#### FR-003 Acceptance Criteria

- When condition transitions from false to true, the service creates a new alert
  event and caches it as active
- While condition remains true, the service creates occurrences for the cached
  active event
- When condition transitions from true to false, the service normalizes the
  event and removes it from cache
- After normalization and cache deletion, a future true condition creates a new
  event (not reuse of previous normalized event)

---

### FR-004 Generate event IDs with prefix and bounded counter format

#### FR-004 Description

The service MUST generate event IDs in the format `<prefix><num>` where `prefix`
comes from rule configuration and `num` comes from a monotonic in-memory
counter.

#### FR-004 Input

- Rule `event_prefix`
- Counter source from logging manager on startup and periodic refresh
- Local fallback seed when remote counter retrieval fails

#### FR-004 Output

- Generated event IDs for new alert creation

#### FR-004 Acceptance Criteria

- Event ID length is exactly 20 characters
- Prefix is taken from the rule and is at most 8 characters
- Numeric part is in range `1..99999999` and advances on each new event creation
- The service attempts to initialize the counter from logging manager at startup
- If counter initialization fails, the service uses a deterministic local
  fallback seed strategy

---

### FR-005 Use flattened alert payload shape for downstream event contracts

#### FR-005 Description

The service MUST emit flattened alert payloads compatible with logging-manager
CLI/event DTO expectations.

#### FR-005 Input

- Active rule, monitored field value, event metadata fields from rule
- Current lifecycle transition from FR-003

#### FR-005 Output

- Flattened create-event and occurrence payloads
- Normalization payload with target event identity and state

#### FR-005 Acceptance Criteria

- Emitted create-event payloads include required logging-manager DTO fields
- Emitted occurrence payloads reference the active cached event ID
- Normalization payloads include the target event ID and normalized state
- Date/time fields are produced by runtime event processing, not static rule
  entries

---

## Interface Requirements

### IR-001 Consume system metrics snapshot subject

#### IR-001 Description

The service MUST consume snapshot events from the shared system metrics contract
subject.

#### IR-001 Input

- NATS events on `system.metrics.events.stats`

#### IR-001 Output

- Decoded snapshot payloads for rule evaluation

#### IR-001 Acceptance Criteria

- Subject name matches the system metrics contract constant for snapshot events
- Invalid payloads are rejected with structured logs and do not terminate the
  service loop

---

### IR-002 Publish generic alert and occurrence events

#### IR-002 Description

The service MUST publish alert lifecycle outputs as generic alert contracts for
interested subscribers.

#### IR-002 Input

- Lifecycle actions from FR-003

#### IR-002 Output

- Alert creation messages
- Alert occurrence messages

#### IR-002 Acceptance Criteria

- New active transitions publish alert creation messages
- Active-state updates publish occurrence messages
- Publish failures are isolated and do not stop metrics event consumption

---

### IR-003 Integrate with logging-manager lifecycle interfaces

#### IR-003 Description

The service MUST support logging-manager-specific lifecycle flows for event
creation, occurrence recording, and normalization/state updates.

#### IR-003 Input

- Alert lifecycle transitions and payloads from FR-003 and FR-005

#### IR-003 Output

- Logging-manager compatible create event calls/messages
- Logging-manager compatible create occurrence calls/messages
- Logging-manager compatible normalize/state-update calls/messages

#### IR-003 Acceptance Criteria

- Logging-manager integration uses the repository logging-manager DTO contracts
- Logging-manager delivery failures are handled independently from generic alert
  publication
- Logging-manager unavailability does not terminate the monitor service loop

---

## Reliability Requirements

### RR-001 Keep monitoring loop alive under downstream failures

#### RR-001 Description

The service MUST continue consuming and evaluating metrics events when
downstream publication or RPC dependencies are unavailable.

#### RR-001 Acceptance Criteria

- Failures in logging-manager integration do not stop event consumption
- Failures in generic alert publication do not stop event consumption
- Failure in one lifecycle action for one event does not crash the process

---

### RR-002 Recover from messaging outages

#### RR-002 Description

The service MUST recover automatically after NATS disconnections without manual
process restart.

#### RR-002 Acceptance Criteria

- Client connection loss triggers automatic reconnect behavior
- After reconnect, the service resumes subscription consumption
- Temporary broker outages do not require process restart for normal operation

---

## Operational Requirements

### OR-001 Run as a simple worker service with bounded in-memory state

#### OR-001 Description

The service MUST run as a worker process and keep only active alert state in
memory.

#### OR-001 Acceptance Criteria

- In-memory cache contains only active (non-normalized) alerts
- Normalized alerts are removed from in-memory cache
- The cache structure is flat and maps to downstream flattened payload needs

---

### OR-002 Support future multi-domain metric event expansion

#### OR-002 Description

The service MUST be structured so additional metric event domains (for example
ZFS, network, and sensor metrics) can be added through rule and subscription
configuration.

#### OR-002 Acceptance Criteria

- Current behavior supports `system.metrics.events.stats`
- Rule schema includes explicit event matching (`event`) and event ID prefix
  (`event_prefix`)
- Interface contracts permit additional subscription subjects in future without
  changing the lifecycle model

---

## Security Requirements

### SR-001 Enforce strict input-type validation at adapter boundaries

#### SR-001 Description

The service MUST validate rule-file inputs and decoded event payload inputs
before condition evaluation and downstream lifecycle emission.

#### SR-001 Acceptance Criteria

- Invalid rule schema entries are rejected before runtime monitoring starts
- Unsupported value types and condition/type combinations are rejected
- Invalid incoming event payload shapes are rejected per-message and logged
- Rejected inputs do not cause execution of business lifecycle actions

---

## Testability Requirements

### TR-001 Keep lifecycle behavior deterministically testable

#### TR-001 Description

Rule matching, condition evaluation, state transitions, and event ID generation
MUST be testable with deterministic inputs and controlled dependencies.

#### TR-001 Acceptance Criteria

- Unit tests can cover `new -> active`, `active -> active`, and
  `active -> normal` transitions
- Unit tests can verify strict type-validation behavior for rule/value
  combinations
- Unit tests can verify event ID generation with startup counter load and
  fallback seed
- Integration tests can verify that logging-manager delivery failure does not
  stop subscription consumption

---

## Notes

- This requirement set defines the `resources-monitor` service behavior for the
  initial system metrics scope
- The exact logging-manager transport mechanics (subject publish vs RPC for
  create/occurrence/state update) are implementation-level integration details
  and must map to existing repository contracts
