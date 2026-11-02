# Logging Managers Context

## Purpose and Scope

LiteNAS includes two logging-manager services:

- `security-logging-manager`
- `system-logging-manager`

Both services use the same shared logging/loggingmanager core, but they serve
different operational domains and own separate data flows.

In the current implementation, logging managers are primarily alert consumers
that keep a more user-friendly, state-based view of alert data.

Dedicated monitoring services act as alert sources. `resources-monitor` consumes
metrics events for system alerts. Log Analysis, Network Security, Antivirus,
Service Watchdog, and LiteNAS Integrity feed security alerts to the security
manager. Traditional operational logs remain a separate source of detail.

## Shared Core and Service Separation

The two managers are intentionally separated at the service level even when they
reuse the same internal implementation patterns.

At minimum, they differ by subscribed NATS subjects, authorization policy, and
the source streams they consume. Their external NATS contracts may evolve
independently, but service isolation is a baseline design constraint.

Both managers now validate access tokens at the messaging boundary before
business handlers run. Role enforcement is then applied per subject so state
changes are accepted only from appropriately authorized callers.

## Platform and Runtime Constraints

The target deployment environment is Raspberry Pi hardware with SD-card-backed
storage in a hardened runtime profile.

Design decisions for logging managers are driven by:

- low and predictable RAM usage
- reduced SD-card write amplification
- bounded and inspectable storage usage
- straightforward operational tuning of resource limits

## Persistence and Write-Behavior Rationale

The logging stack uses SQLite-backed persistence with batching and controlled
cleanup windows to reduce write pressure on SD storage.

This approach prioritizes:

- implementation simplicity
- practical reliability for embedded-style deployments
- bounded growth and manageable retention

The selected model accepts a limited in-memory buffering window before flush as
a tradeoff for better write efficiency.

## Data-Model Rationale

Event, lifecycle, and occurrence information is split into separate storage
areas to reduce write pressure and allow smaller, targeted write operations.

This decomposition is primarily an operational write-behavior decision, not a
database-normalization goal.

## Boundaries of This Document

The admin panel and CLI consume the stateful alert data through their existing
service boundaries.

This document captures architecture context and design rationale.

It does not define normative software requirements, API contracts, or complete
acceptance criteria. Those are defined in dedicated requirement documents.
