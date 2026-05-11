# Logging Managers Context

## Purpose and Scope

LiteNAS will include two logging-manager services:

- `security-logging-manager`
- `system-logging-manager`

Both services use the same shared logging/loggingmanager core, but they serve
different operational domains and own separate data flows.

The security manager is responsible for security-oriented alert and metric
tracking. The system manager is responsible for system health alerts and
performance metric tracking.

## Shared Core and Service Separation

The two managers are intentionally separated at the service level even when they
reuse the same internal implementation patterns.

At minimum, they differ by subscribed NATS subjects and the source streams they
consume. Their external NATS contracts may evolve independently, but service
isolation is a baseline design constraint.

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

This document captures architecture context and design rationale.

It does not define normative software requirements, API contracts, or complete
acceptance criteria. Those are defined in dedicated requirement documents.
