# RBAC Service — Requirements

## Overview

The RBAC Service is the internal authorization decision service for LiteNAS.

It resolves role context from system identity sources and answers on-demand
capability checks for target resources. The service is a policy decision point
for internal services and does not execute target commands or perform target
service business logic.

The service is responsible for:

- resolving subject role context from host identity state (including groups)
- returning UID as the canonical subject key for downstream checks
- answering boolean capability decisions for app execution, sudo capability,
  and file path access
- evaluating file checks with symlink resolution and POSIX ACL awareness
- supporting cache-backed decision evaluation with explicit invalidation
- failing closed when decisions cannot be determined safely
- producing service-owned audit logs while minimizing system-audit noise

The service is not responsible for:

- primary credential authentication
- JWT issuance or refresh-session management
- executing user commands or performing privileged operations on behalf of
  requesters
- enforcing domain-specific business authorization inside target services

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Resolve subject roles for authenticated principals

#### FR-001 Description

The RBAC Service MUST return role context for a subject identified by username.

#### FR-001 Input

- Username of an already authenticated subject

#### FR-001 Output

- Role context including groups
- UID for the subject

#### FR-001 Acceptance Criteria

- The service accepts username as the input key for role resolution
- The response includes the subject UID as a separate field
- Group membership is included in the response as role context
- Non-interactive system users are resolved using the same role model as
  interactive users

---

### FR-002 Provide boolean path access decisions

#### FR-002 Description

The RBAC Service MUST answer whether a subject can perform a requested access
operation on a specific filesystem path.

#### FR-002 Input

- Subject UID
- Target path
- Requested operation (read, write, modify)

#### FR-002 Output

- Boolean allow or deny decision

#### FR-002 Acceptance Criteria

- The service resolves symlinks and evaluates the resolved target path
- Decision evaluation includes owner/group mode bits and POSIX ACLs
- The response contains a boolean decision suitable for internal service use

---

### FR-003 Provide boolean application execution decisions

#### FR-003 Description

The RBAC Service MUST answer whether a subject can execute a specific
application command without sudo elevation.

#### FR-003 Input

- Subject UID
- Target executable or command identity

#### FR-003 Output

- Boolean allow or deny decision

#### FR-003 Acceptance Criteria

- The service returns allow only when the subject has sufficient effective
  rights for non-sudo execution
- The service does not execute the target application as part of decision
  answering

---

### FR-004 Provide boolean sudo-capability decisions

#### FR-004 Description

The RBAC Service MUST answer whether a subject can run a requested application
command through sudo policy.

#### FR-004 Input

- Subject UID
- Target executable or command identity for sudo evaluation

#### FR-004 Output

- Boolean allow or deny decision

#### FR-004 Acceptance Criteria

- The service returns a boolean decision for sudo eligibility of the requested
  command
- The service does not execute the target application while evaluating sudo
  eligibility
- Sudo-capability evaluation mechanism may evolve, but must satisfy the
  operational and security requirements in this document

---

### FR-005 Support cache-backed decision evaluation

#### FR-005 Description

The RBAC Service MUST support caching for role and capability lookups to reduce
host probing overhead.

#### FR-005 Input

- Repeated role and capability requests for the same subject and resource scope
- Cache invalidation requests triggered by role or permission changes

#### FR-005 Output

- Decisions and role responses that may be served from cache
- Explicit cache invalidation outcomes

#### FR-005 Acceptance Criteria

- Caching behavior is configurable
- Cache TTL is configurable
- A supported invalidation operation exists to force role and permission
  refresh for affected subjects or scopes

---

## Interface Requirements

### IR-001 Expose RBAC contracts over RPC

#### IR-001 Description

The RBAC Service MUST expose its service interface as internal RPC contracts.

#### IR-001 Input

- RPC requests from Auth Service and other LiteNAS internal services

#### IR-001 Output

- RPC replies containing role-resolution data or boolean capability decisions

#### IR-001 Acceptance Criteria

- A role-resolution RPC is available using username input
- Capability-check RPCs are available using UID input
- Capability-check RPCs cover path access, app execution, and sudo capability

---

### IR-002 Keep decision responses machine-simple for internal callers

#### IR-002 Description

The RBAC Service MUST return machine-simple decision payloads for internal
service consumers.

#### IR-002 Input

- Capability-check requests over RPC

#### IR-002 Output

- Boolean decision results

#### IR-002 Acceptance Criteria

- Decision payloads provide boolean allow or deny as the core contract
- Callers can consume decisions without parsing user-facing message text

---

## Reliability Requirements

### RR-001 Fail closed for capability decisions

#### RR-001 Description

The RBAC Service MUST fail closed when it cannot determine capability safely.

#### RR-001 Acceptance Criteria

- Dependency or evaluation failures result in deny decisions
- Timeout or transient lookup failures do not produce allow decisions

---

### RR-002 Keep cache invalidation authoritative

#### RR-002 Description

The RBAC Service MUST ensure invalidation requests can force fresh evaluation
after role or permission changes.

#### RR-002 Acceptance Criteria

- Invalidation removes or bypasses affected cached entries before subsequent
  checks are answered
- Post-invalidation checks reflect current host permission state

---

## Operational Requirements

### OR-001 Use host-system identity and permission sources

#### OR-001 Description

The RBAC Service MUST derive role and capability decisions from host-system
identity and permission state.

#### OR-001 Acceptance Criteria

- Role resolution reflects current host group membership
- File access decisions reflect host filesystem permissions including POSIX ACLs
- Sudo-capability decisions reflect host sudo policy state

---

### OR-002 Minimize unnecessary host audit noise

#### OR-002 Description

The RBAC Service MUST minimize unnecessary host-level audit entries caused by
permission probing.

#### OR-002 Acceptance Criteria

- The service avoids decision mechanisms that create avoidable high-volume host
  audit entries
- Cache usage demonstrably reduces repeated probe noise for repeated checks
- The selected sudo evaluation approach is validated against this noise
  constraint before production rollout

---

### OR-003 Produce service-owned configurable audit logs

#### OR-003 Description

The RBAC Service MUST produce its own audit logs under application control.

#### OR-003 Acceptance Criteria

- Decision requests and outcomes can be logged by the service
- Audit-log detail level is configurable by service policy
- Service audit logging is independent from host system-audit verbosity

---

## Security Requirements

### SR-001 Preserve strict service boundaries

#### SR-001 Description

The RBAC Service MUST remain a decision service and MUST NOT execute requested
commands or perform privileged operations on behalf of subjects.

#### SR-001 Acceptance Criteria

- Execution checks do not run target commands
- Sudo checks do not run target commands
- Target-service business actions remain outside RBAC responsibility

---

### SR-002 Restrict trust to internal service callers

#### SR-002 Description

The RBAC Service MUST be treated as an internal service interface used by
trusted LiteNAS services.

#### SR-002 Acceptance Criteria

- Interfaces are defined for internal service-to-service invocation
- External browser-facing invocation is not treated as the primary access model
- Calling services remain responsible for authenticating end-user sessions
  before invoking RBAC checks

---

## Testability Requirements

### TR-001 Keep identity, filesystem, sudo, clock, and cache adapters injectable

#### TR-001 Description

The RBAC Service MUST keep host-dependent evaluation adapters injectable so
decision logic can be tested without requiring live privileged host state.

#### TR-001 Acceptance Criteria

- Role-resolution logic is testable with identity-source doubles
- Path-access logic is testable with filesystem-permission doubles including ACL
  scenarios
- Sudo-capability logic is testable with sudo-policy doubles
- Cache behavior and invalidation logic are testable with deterministic time
  control
