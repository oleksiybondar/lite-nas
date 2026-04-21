# Requirement Types Standard

## Introduction

This document defines the standard requirement classification used across
the project.

Its purpose is to ensure that requirements are written in a consistent,
readable, and testable way across all services and components. A shared
classification also makes it easier to review requirements, map them to
tests, and distinguish business behavior from operational, security, or
implementation concerns.

This standard is intended to be lightweight. It should help structure
requirements without turning documentation into bureaucracy.

---

## Core Principles

Requirements written under this standard SHOULD follow these principles:

- Requirements describe **what** the system must do, not **how** it is
  implemented
- Requirements SHOULD be clear, specific, and testable
- Requirements SHOULD avoid unnecessary implementation details
- Each requirement MUST belong to exactly one requirement type
- Each requirement MUST have a unique identifier within its document

---

## Requirement Types

### FR — Functional Requirements

Describe what the service or component does.

These requirements define the expected behavior, capabilities, and
business or technical functions of the system.

Examples:

- Collect system metrics
- Maintain a bounded metrics history
- Publish change events
- Respond to stat requests

---

### IR — Interface Requirements

Describe how the service or component communicates with external systems
or other internal components.

These requirements cover interaction contracts rather than internal
behavior.

Includes:

- Messaging subjects and request/reply patterns
- API endpoints
- Input/output contracts
- Serialization expectations

Examples:

- The service MUST respond to `getStats` requests over the messaging system
- The CLI MUST output JSON-parseable records

---

### RR — Reliability Requirements

Describe how the service behaves under failure, degradation, or recovery
conditions.

These requirements define resilience expectations and fault-handling
behavior.

Includes:

- Retry behavior
- Reconnection logic
- Graceful degradation
- Startup behavior during dependency failure
- Bounded failure handling

Examples:

- The service MUST start even if NATS is unavailable
- Connection retries MUST use bounded backoff

---

### OR — Operational Requirements

Describe runtime and deployment expectations.

These requirements define how the service should behave when operated in
real environments.

Includes:

- Configuration
- Logging
- Resource usage expectations
- Startup and shutdown behavior
- Runtime limits

Examples:

- Poll interval MUST be configurable
- Memory usage for history MUST remain bounded

---

### SR — Security Requirements

Describe authentication, authorization, and protection of sensitive data.

These requirements define the security expectations for service operation
and integration.

Includes:

- Credentials handling
- Authentication methods
- Authorization boundaries
- Secret management
- Secure defaults

Examples:

- NATS authentication MUST be configurable
- Secrets MUST NOT be hardcoded in source code

---

### TR — Testability Requirements

Describe architectural or design constraints needed to support automated
testing.

These requirements exist to ensure the codebase remains testable without
unnecessary coupling to live infrastructure.

Includes:

- Dependency injection
- Interface-based design
- Use of test doubles
- Separation of orchestration from side effects

Examples:

- Messaging dependencies MUST be injectable
- Business logic MUST be testable without a live NATS server

---

## Requirement Identifier Format

Each requirement MUST use the following identifier format:

```text
<TYPE>-<NUMBER>
```

Examples:

- `FR-001`
- `IR-002`
- `RR-001`
- `SR-003`

Rules:

- `<TYPE>` MUST be one of: `FR`, `IR`, `RR`, `OR`, `SR`, `TR`
- `<NUMBER>` SHOULD use three digits
- Identifiers MUST be unique within a document

---

## Requirement Writing Rules

Each requirement SHOULD include:

- Title
- Description
- Input
- Output
- Acceptance Criteria

Optional sections may include:

- Notes
- Assumptions
- Constraints
- Open Questions

Recommended structure:

```text
### FR-001 Requirement title

#### Description

Describe what the system must do.

#### Input

Describe what triggers or influences the behavior.

#### Output

Describe the externally meaningful result.

#### Acceptance Criteria

- Testable condition
- Testable condition

#### Notes

Optional clarifications.
```

---

## Scope Boundaries

This standard classifies requirements. It does not define:

- implementation design
- source code structure
- deployment topology
- technology choices

These belong in separate documents such as:

- architecture documents
- design notes
- ADRs
- deployment documentation

---

## Guidance on Implementation Details

Requirements SHOULD avoid implementation-specific details unless those
details are themselves required by the system.

Prefer:

- The service MUST collect CPU metrics from the host system

Avoid:

- The service MUST read `/proc/stat`

The first describes required behavior.
The second describes a specific implementation.
