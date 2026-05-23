# RBAC Service — Strict Boundaries

## Purpose

This document defines strict ownership boundaries between Auth Service, RBAC
Service, and target LiteNAS services for authorization-related flows.

The goal is to keep contracts explicit and prevent responsibility drift.

---

## Core Boundary Principle

- Auth proves identity.
- RBAC answers capability decisions.
- Target services enforce workflow and business action rules.

No service should absorb the primary responsibility of another service in this
chain.

---

## Auth Service Boundaries

### Auth Service owns

- Credential authentication and session credibility
- Session/token issuance and refresh lifecycle
- Subject identity claims presented to other services
- Username-based subject context used to request RBAC role resolution

### Auth Service does not own

- Final resource-specific capability decisions (file, app, sudo)
- Host permission probing for authorization decisions
- Domain-specific authorization logic of target services

---

## RBAC Service Boundaries

### RBAC Service owns

- Role/group resolution for authenticated subjects
- UID return as canonical key for downstream capability checks
- Boolean decision answers for:
  - path access (`read`, `write`, `modify`)
  - non-sudo app execution
  - sudo capability for requested command identity
- Permission evaluation based on host identity and permission sources
- Symlink resolution and target-path permission evaluation
- POSIX ACL-aware path decision logic
- Decision caching with configurable TTL and explicit invalidation support
- Fail-closed behavior on inability to evaluate safely
- Service-owned audit logs under app-level control

### RBAC Service does not own

- Authentication or credential verification
- Browser-facing access interfaces
- JWT lifecycle management
- Execution of target commands as a side effect of decision checks
- Target-service business workflows after decision results are returned

---

## Target Service Boundaries

### Target Services own

- JWT validation and request authentication at their own boundaries
- Coarse pre-check authorization decisions from token role context
- Invocation of RBAC for resource-specific or elevated checks
- Final action enforcement in service business workflows
- Service-local logging, error handling, and user-facing response semantics

### Target Services do not own

- Reimplementation of host group/ACL/sudo decision logic already owned by RBAC
- Ad hoc host permission probing bypassing RBAC decision contracts

---

## Data and Subject Model Boundaries

- Username is the canonical lookup key for role-resolution RPC.
- UID is the canonical subject key for capability-check RPCs.
- Non-interactive system users follow the same RBAC model as interactive users.
- Group membership is treated as coarse role context, not as a complete
  replacement for resource-specific checks.

---

## Decision Model Boundaries

- Capability responses are internal boolean decisions (`allow` or `deny`).
- Calling services must not reinterpret transport or evaluation failure as
  implicit allow.
- RBAC must return deny when safe evaluation is not possible.

---

## Audit and Noise Boundaries

- RBAC must produce service-owned audit logs for traceability.
- RBAC decision mechanisms must minimize unnecessary host-system audit noise.
- The selected sudo-evaluation mechanism is an implementation detail, but it
  must satisfy the noise constraint and the no-command-execution constraint.

---

## Open Implementation Point (Explicitly Deferred)

The exact sudo-capability evaluation mechanism is intentionally deferred.

Current mandatory constraints for the final mechanism:

- Must return reliable boolean sudo-capability decisions for requested command
  identity
- Must avoid executing target commands
- Must minimize unnecessary host audit entries
- Must support cache-backed evaluation and explicit invalidation
