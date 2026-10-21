# LiteNAS Integrity Service — Requirements

## Overview

The LiteNAS Integrity Service is a host-integrity worker that runs AIDE checks,
detects filesystem-integrity failures, and emits security alert lifecycle
updates.

The service performs an integrity check at startup and then continues as a
scheduled worker using a configured hour-based interval. It also exposes a
privileged baseline-regeneration RPC for authorized human operators.

The service is responsible for:

- running scheduled `aide --check` executions
- running `aide --init` during authorized baseline regeneration
- replacing the active AIDE database with the generated `.new` baseline after a
  successful initialization
- tracking one active in-memory integrity incident across repeated failed
  checks
- publishing security alert creation and occurrence updates while a failure
  remains active
- normalizing the active alert when integrity returns to normal or after a
  successful baseline regeneration
- validating privileged RPC callers through Auth and RBAC services
- operating under a dedicated service account with scoped passwordless `sudo`
  rights for required AIDE commands

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Run an integrity check at startup and on a fixed hourly schedule

#### FR-001 Description

The service MUST run an AIDE integrity check once at startup and then continue
running checks every configured number of hours after launch.

#### FR-001 Input

- Service startup lifecycle
- `[AIDE]` configuration containing `schedule`
- Host-local AIDE installation and configuration

#### FR-001 Output

- One `aide --check` execution attempt at startup
- Repeated `aide --check` execution attempts on the configured interval

#### FR-001 Acceptance Criteria

- The service runs an integrity check immediately after runtime startup
- The service waits the configured number of hours between later checks
- The schedule is interval-based and not tied to wall-clock calendar times
- Invalid or non-positive schedule values are rejected at configuration-load
  time

---

### FR-002 Detect integrity-failure state from AIDE check results

#### FR-002 Description

The service MUST interpret AIDE check results to distinguish normal integrity
state from integrity-failure state.

#### FR-002 Input

- Exit status and process output from `aide --check`

#### FR-002 Output

- Normal integrity result
- Active integrity-failure result
- Structured execution failure result when the check itself cannot be completed

#### FR-002 Acceptance Criteria

- A successful AIDE check is treated as normal integrity state
- An AIDE-reported mismatch or integrity violation is treated as an active
  integrity-failure state
- Failure to execute or complete the AIDE command is surfaced distinctly from a
  clean integrity result
- Check-result interpretation does not crash the runtime loop

---

### FR-003 Track one active integrity incident until normalized

#### FR-003 Description

The service MUST keep one in-memory cached active integrity incident and reuse
it across repeated failed checks until the incident is normalized.

#### FR-003 Input

- Current AIDE check result from FR-002
- In-memory cached active incident state

#### FR-003 Output

- Create-alert action for first active failure
- Add-occurrence action for later repeated failures
- Normalize-and-clear action when the incident ends

#### FR-003 Acceptance Criteria

- When integrity transitions from normal to failed, the service creates a new
  active security alert and caches its event identity
- While integrity remains failed, later checks add occurrences to the cached
  alert instead of creating new alerts
- When integrity returns to normal, the service normalizes the cached alert and
  removes it from memory
- After service restart, cached incident state is lost by design
- If the service restarts while integrity is still failing, a later failed
  check can create a new alert entry

---

### FR-004 Regenerate the AIDE baseline through an authorized RPC

#### FR-004 Description

The service MUST expose an internal RPC that regenerates the AIDE baseline only
for currently authorized callers.

#### FR-004 Input

- RPC request containing only `access_token`
- Current auth-service validation result
- Current RBAC role-resolution result

#### FR-004 Output

- Authorized baseline regeneration attempt
- Denied response for unauthorized or invalid callers
- Structured success or failure response

#### FR-004 Acceptance Criteria

- The request contract accepts only `access_token`
- The service validates the submitted token through the auth-service live
  validation RPC before any baseline action
- The service uses the validated username from auth-service to resolve current
  role context through RBAC
- Only callers holding security-role or administrator-equivalent privileges are
  allowed to regenerate the baseline
- Authorization fails closed when token validation or RBAC resolution cannot be
  completed safely

---

### FR-005 Replace the active AIDE database after successful initialization

#### FR-005 Description

After a successful baseline-regeneration run, the service MUST replace the
active AIDE database with the newly generated baseline database.

#### FR-005 Input

- Authorized baseline-regeneration action from FR-004
- `aide --init` command result
- Generated AIDE database `.new` file

#### FR-005 Output

- Updated active AIDE database
- Failure response if replacement cannot be completed safely

#### FR-005 Acceptance Criteria

- Baseline regeneration runs `aide --init`
- The service expects a newly generated `.new` database artifact after
  successful initialization
- The service replaces the active database with the new baseline only after a
  successful initialization run
- The replacement step is treated as part of the baseline-regeneration
  operation, not as a manual follow-up task
- Failure in the replacement step causes the RPC to report failure

---

### FR-006 Publish critical security alert creation and occurrence updates

#### FR-006 Description

The service MUST publish security alert lifecycle outputs for detected
integrity incidents using critical severity.

#### FR-006 Input

- First active integrity-failure transition from FR-003
- Repeated active integrity-failure results from FR-003
- Service-to-service access token held by the runtime

#### FR-006 Output

- Security alert creation message for the first active failure
- Security alert occurrence message for repeated active failures

#### FR-006 Acceptance Criteria

- The first active integrity failure publishes one security alert creation
  payload
- Repeated failed checks publish occurrences for the cached active alert
- Created integrity alerts use the highest configured severity level
- Published alert payloads include a service-authenticated access token
  belonging to the integrity service runtime

---

### FR-007 Normalize the active alert after recovery or baseline regeneration

#### FR-007 Description

The service MUST normalize the cached active alert after a clean integrity
check or after a successful baseline-regeneration flow.

#### FR-007 Input

- Transition from failed integrity state to clean integrity state
- Successful baseline-regeneration result from FR-005
- Cached active alert identity

#### FR-007 Output

- Logging-manager state-update RPC setting the active alert state to `normal`
- Cleared in-memory cached alert identity

#### FR-007 Acceptance Criteria

- A clean integrity result after an active failure triggers a state-update RPC
  that sets the alert state to `normal`
- A successful baseline-regeneration flow also triggers the same normalization
  behavior when an active alert exists
- After successful normalization, the cached event identity is removed from
  memory
- A future integrity failure after normalization creates a new alert

---

## Interface Requirements

### IR-001 Expose a baseline-regeneration RPC contract

#### IR-001 Description

The service MUST expose one internal RPC contract for authorized baseline
regeneration.

#### IR-001 Input

- RPC request containing `access_token`

#### IR-001 Output

- Structured success or failure response for the baseline-regeneration
  operation

#### IR-001 Acceptance Criteria

- The RPC contract is request/reply based
- The request shape contains only the caller access token
- The response indicates whether the baseline regeneration completed
  successfully
- Invalid request payloads are rejected at the RPC boundary

---

### IR-002 Integrate with auth-service live token validation

#### IR-002 Description

The service MUST use auth-service live validation to confirm that a submitted
access token is still credible at the time of a privileged baseline request.

#### IR-002 Input

- Caller access token from the regeneration RPC

#### IR-002 Output

- Current token-validity decision
- Validated username for downstream authorization checks

#### IR-002 Acceptance Criteria

- Token validation uses the repository auth live-validation RPC contract
- The integrity service does not trust stale local authorization state instead
  of live validation for this privileged flow
- Token validation can reject tokens that are expired, revoked, or otherwise no
  longer credible

---

### IR-003 Integrate with RBAC for caller role resolution

#### IR-003 Description

The service MUST use RBAC to resolve current role context for a validated
baseline-regeneration caller.

#### IR-003 Input

- Validated username from auth-service

#### IR-003 Output

- Current RBAC role/group context for authorization evaluation

#### IR-003 Acceptance Criteria

- RBAC lookup uses the repository subject-role request/reply contract
- Authorization accepts security-role and administrator-equivalent callers
- The integrity service does not accept caller-provided usernames outside the
  validated auth result

---

### IR-004 Integrate with security logging-manager alert contracts

#### IR-004 Description

The service MUST integrate with the security logging-manager contracts for
alert creation, occurrence recording, and normalization.

#### IR-004 Input

- Alert lifecycle actions from FR-003, FR-006, and FR-007

#### IR-004 Output

- Security alert creation messages
- Security alert occurrence messages
- Security alert state-normalization RPCs

#### IR-004 Acceptance Criteria

- The service publishes new alerts through the security alert subject
- The service publishes repeated-failure occurrences through the security alert
  occurrence subject
- The service normalizes alerts through the logging-manager state-update RPC
  interface
- The service authenticates alert-publishing actions as its own service
  identity

---

## Reliability Requirements

### RR-001 Keep the scheduled runtime loop alive under check and downstream failures

#### RR-001 Description

The service MUST keep its scheduling loop alive when AIDE checks, alert
publishing, or normalization dependencies fail.

#### RR-001 Acceptance Criteria

- A failed scheduled check does not terminate the service process by itself
- Alert publication failures do not stop later scheduled checks
- Logging-manager normalization failures do not stop later scheduled checks
- One failed loop iteration does not prevent later iterations from running

---

### RR-002 Fail closed for privileged baseline regeneration

#### RR-002 Description

The service MUST fail closed when it cannot safely authorize or complete a
baseline-regeneration request.

#### RR-002 Acceptance Criteria

- Missing or invalid caller tokens are rejected
- Auth-service validation failure rejects the request
- RBAC lookup failure rejects the request
- AIDE initialization failure rejects the request
- Active-database replacement failure rejects the request

---

## Operational Requirements

### OR-001 Use a dedicated service account with scoped AIDE sudo rights

#### OR-001 Description

The service MUST run under its own dedicated operating-system account and use a
repo-managed sudoers policy for required AIDE commands.

#### OR-001 Acceptance Criteria

- The systemd unit runs the service under a dedicated user and group
- Required `sudo` execution is passwordless for the service account
- The sudoers policy is scoped to the required AIDE commands and not broad
  interactive root access
- Baseline regeneration and integrity checks do not depend on manual operator
  password entry

---

### OR-002 Externalize AIDE scheduling in service configuration

#### OR-002 Description

The service MUST load its AIDE runtime schedule from repository-managed service
configuration.

#### OR-002 Acceptance Criteria

- The service has a dedicated config file under `/etc/lite-nas/`
- The config includes an `[AIDE]` section
- The `[AIDE]` section defines `schedule` in hours
- Missing or invalid required AIDE configuration prevents startup

---

### OR-003 Treat AIDE as a required runtime dependency

#### OR-003 Description

The service MUST treat AIDE as a mandatory runtime dependency rather than an
optional integration.

#### OR-003 Acceptance Criteria

- Startup or runtime execution surfaces a clear failure when AIDE is not
  available
- The service does not silently degrade into a no-op mode when AIDE is absent
- Packaging and deployment for the service assume AIDE is installed

---

### OR-004 Use a dedicated service identity in transport security and authorization

#### OR-004 Description

The service MUST have its own transport identity and messaging authorization
entries for internal communication.

#### OR-004 Acceptance Criteria

- The service has its own client certificate material path in configuration
- The messaging layer uses a dedicated service client name
- NATS authorization includes subjects needed for token login, token refresh,
  auth validation, RBAC lookup, alert publication, and alert normalization

---

## Security Requirements

### SR-001 Restrict baseline regeneration to security-authorized operators

#### SR-001 Description

Baseline regeneration MUST be limited to callers that are currently both
authenticated and authorized for security-sensitive operations.

#### SR-001 Acceptance Criteria

- Authorization requires successful token validation plus current RBAC role
  resolution
- Security-role callers are accepted
- Administrator-equivalent callers are accepted
- Other authenticated callers are rejected

---

### SR-002 Separate human authorization from service-to-service alert identity

#### SR-002 Description

The service MUST distinguish between human authorization for baseline
regeneration and service identity used for background alert publication.

#### SR-002 Acceptance Criteria

- Regeneration RPC authorization is based on the submitted human access token
- Scheduled checks do not require a human token to publish alerts
- Alert publication uses the integrity service's own service token

---

### SR-003 Minimize host privilege exposure

#### SR-003 Description

The service MUST minimize host-level privilege exposure while still being able
to execute required AIDE operations.

#### SR-003 Acceptance Criteria

- The runtime executes under a non-root dedicated user
- Elevated host access is delegated through narrowly scoped `sudo`
- The service is not granted unrestricted root-shell execution capability

---

## Testability Requirements

### TR-001 Keep AIDE execution, authorization, and alerting dependencies injectable

#### TR-001 Description

The service MUST keep AIDE command execution, authorization clients, and alert
delivery dependencies injectable for deterministic testing.

#### TR-001 Acceptance Criteria

- Tests can simulate clean checks, integrity failures, and command-execution
  failures without requiring live AIDE execution
- Tests can simulate auth validation allow and deny results
- Tests can simulate RBAC allow and deny results
- Tests can verify first-failure alert creation, repeated-failure occurrence
  publication, and normalization behavior
