# Auth Service — Requirements

## Overview

The Auth Service is the host-local authentication authority for LiteNAS.

It verifies real login-capable users of the managed Linux system through
PAM-backed authentication flows, issues session tokens for the browser-facing
management interface, and exposes auth-related request/reply and event
contracts over NATS.

The service is responsible for:

- PAM-backed authentication and account-state handling
- password-change flows required by PAM account policy
- access-token issuance and refresh-token rotation
- refresh-token revocation and volatile in-memory session tracking
- lockdown state management and auth-related event publication
- online token validation for critical backend flows

The service is not a general authorization engine for all domain decisions.
It establishes identity and session credibility for the current machine and
publishes enough auth state for other components to enforce their own
boundaries.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Authenticate real login-capable system users

#### FR-001 Description

The Auth Service MUST authenticate only real users who are allowed to log in to
the managed system through the configured PAM and NSS account model.

#### FR-001 Input

- Auth requests containing a submitted login identifier
- Submitted authentication material supported by the PAM stack
- Host account data resolved through the system PAM and NSS configuration

#### FR-001 Output

- Auth success or auth failure responses tied to a host identity

#### FR-001 Acceptance Criteria

- Authentication checks are delegated to PAM-backed host authentication flows
- Users that cannot log in to the system are not treated as valid interactive
  LiteNAS principals
- Service accounts and non-interactive accounts are rejected by policy
- Account eligibility is based on actual login capability rather than only
  local `/etc/passwd` presence

---

### FR-002 Return structured authentication outcomes

#### FR-002 Description

The Auth Service MUST return structured authentication outcomes rather than a
simple boolean result.

#### FR-002 Input

- Authentication attempts
- PAM account-management and credential-management results

#### FR-002 Output

- Auth outcome objects containing machine-readable status and user-facing
  messages

#### FR-002 Acceptance Criteria

- Responses distinguish at least successful authentication, invalid
  credentials, password-change-required, account-locked, temporarily blocked,
  denied, lockdown-active, and service-unavailable outcomes
- Responses include a stable machine-readable outcome code
- Responses include one or more human-readable messages suitable for the HMI
- Responses can carry structured timing or retry information when PAM exposes
  it or when the service can derive it safely

---

### FR-003 Support PAM-driven password change flows

#### FR-003 Description

The Auth Service MUST support password update flows when PAM requires a user to
change their password before normal login can complete.

#### FR-003 Input

- Authentication results indicating password expiry or forced password change
- Password change requests containing the required current and replacement
  credentials

#### FR-003 Output

- Password change success or failure responses
- Follow-up auth state indicating whether a normal session can now be issued

#### FR-003 Acceptance Criteria

- The service can surface a password-change-required state without issuing a
  normal session
- The service can process a PAM-backed password update flow for eligible users
- Password update failures are returned as structured outcomes with messages
- Successful password update allows a follow-up authentication flow to complete

---

### FR-004 Issue short-lived access tokens and long-lived refresh tokens

#### FR-004 Description

The Auth Service MUST issue a short-lived access token and a refresh token for
successful login flows.

#### FR-004 Input

- Successful authentication result for an eligible user
- Current service time

#### FR-004 Output

- Signed JWT access token
- Refresh token bound to server-side refresh state

#### FR-004 Acceptance Criteria

- Access tokens are signed JWTs suitable for local validation by other
  services
- Access-token lifetime defaults to 15 minutes
- A new login issues a refresh token with a 30 day lifetime
- Refresh tokens are represented in server-managed session state rather than
  treated as self-sufficient bearer truth

---

### FR-005 Rotate refresh tokens without extending absolute session lifetime

#### FR-005 Description

The Auth Service MUST rotate refresh tokens on refresh while preserving the
absolute refresh-session expiry established at login.

#### FR-005 Input

- Refresh requests containing a currently valid refresh token
- Existing refresh-session state

#### FR-005 Output

- New access and refresh token pair
- Revoked prior refresh token state

#### FR-005 Acceptance Criteria

- Refresh-token rotation invalidates the submitted refresh token after
  successful use
- Refresh requests issue a new access token and a new refresh token
- The rotated refresh token inherits the original refresh-session expiry rather
  than extending it
- Replay of an already rotated or revoked refresh token is rejected

---

### FR-006 Revoke sessions on logout and service restart

#### FR-006 Description

The Auth Service MUST revoke refresh-session state on logout and MUST treat
service restart as full refresh-session loss.

#### FR-006 Input

- Logout requests
- Service shutdown and restart lifecycle events

#### FR-006 Output

- Revoked session state
- Logout confirmation responses

#### FR-006 Acceptance Criteria

- Logout invalidates the corresponding refresh token state
- Restart discards all in-memory refresh-token state
- Refresh requests for discarded or logged-out sessions are rejected
- The volatile session model is treated as intentional behavior rather than an
  implementation accident

---

### FR-007 Enforce lockdown mode

#### FR-007 Description

The Auth Service MUST support a lockdown mode that immediately blocks
authentication and discards token-based session continuity.

#### FR-007 Input

- Lockdown enable and disable commands
- Auth and refresh requests while lockdown is active

#### FR-007 Output

- Lockdown state transitions
- Rejected auth responses while lockdown is active

#### FR-007 Acceptance Criteria

- Enabling lockdown causes all refresh-token state to be discarded immediately
- Login, refresh, and password-change-dependent session issuance requests are
  rejected while lockdown is active
- Responses during lockdown return a distinct lockdown outcome
- Disabling lockdown permits normal auth handling to resume without restoring
  discarded sessions

---

## Interface Requirements

### IR-001 Expose auth request/reply contracts over NATS

#### IR-001 Description

The Auth Service MUST expose its primary service interface over NATS
request/reply contracts.

#### IR-001 Input

- NATS requests from the Web Gateway and other internal services

#### IR-001 Output

- NATS replies containing auth results, token payloads, or validation results

#### IR-001 Acceptance Criteria

- Login, refresh, logout, and token-validation capabilities are exposed through
  explicit NATS request/reply subjects
- Request and response contracts distinguish successful sessions from
  structured auth outcomes
- The messaging contract is usable by the Web Gateway without embedding PAM
  logic in the gateway

---

### IR-002 Publish lockdown state events

#### IR-002 Description

The Auth Service MUST publish lockdown state changes to the messaging layer so
other services can react to emergency auth-state changes.

#### IR-002 Input

- Lockdown enable and disable transitions

#### IR-002 Output

- Lockdown event messages over NATS

#### IR-002 Acceptance Criteria

- A lockdown-enabled event is published when lockdown becomes active
- A lockdown-disabled event is published when lockdown is cleared
- Events carry an explicit state value rather than relying on message absence
- Event publication is separate from request/reply auth flows

---

### IR-003 Support local JWT validation with online validation fallback

#### IR-003 Description

The Auth Service MUST support a trust model where other services can validate
JWT access tokens locally while also being able to request live validation for
critical operations.

#### IR-003 Input

- Access tokens presented to internal services
- NATS validation requests for critical flows

#### IR-003 Output

- Locally verifiable JWT claims
- Live validation responses for critical flows

#### IR-003 Acceptance Criteria

- Access tokens contain claims sufficient for local signature and expiry checks
- Services are not required to call the Auth Service for every protected action
- A dedicated validation contract exists for flows that require live credibility
  confirmation
- The validation path can reject tokens that remain cryptographically valid but
  are no longer considered credible by current auth state

---

## Reliability Requirements

### RR-001 Fail closed for auth decisions

#### RR-001 Description

The Auth Service MUST fail closed when it cannot safely determine auth
credibility.

#### RR-001 Acceptance Criteria

- Authentication requests are rejected when required PAM processing cannot be
  completed safely
- Refresh requests are rejected when refresh-session state is unavailable
- Token-validation responses do not report success when current credibility
  cannot be established

---

### RR-002 Keep lockdown state authoritative during runtime

#### RR-002 Description

The Auth Service MUST apply lockdown state consistently to all runtime auth
flows while the process remains active.

#### RR-002 Acceptance Criteria

- Lockdown checks occur before normal auth success responses are issued
- Refresh requests do not bypass lockdown state
- Runtime state transitions do not leave partially active sessions behind after
  lockdown enablement

---

## Operational Requirements

### OR-001 Use volatile in-memory refresh-session storage

#### OR-001 Description

The Auth Service MUST use in-memory storage for refresh-session state in the
initial implementation.

#### OR-001 Acceptance Criteria

- Refresh tokens are tracked in process memory
- The service does not require a database to issue and rotate refresh tokens
- Restart-driven session invalidation is documented as intentional behavior

---

### OR-002 Expose configurable token and policy settings

#### OR-002 Description

The Auth Service MUST expose configuration for token, PAM-policy, and messaging
settings needed for operation.

#### OR-002 Acceptance Criteria

- Access-token signing configuration is externalized
- Access-token and refresh-token lifetimes are configurable
- PAM service selection or equivalent auth-policy configuration is
  externalized
- NATS connectivity and subject configuration are externalized

---

## Security Requirements

### SR-001 Treat PAM as the host authentication authority

#### SR-001 Description

The Auth Service MUST delegate host credential verification and account-state
evaluation to PAM-backed system flows rather than reimplementing host auth
rules in service code.

#### SR-001 Acceptance Criteria

- Host credential checks are not reimplemented as ad hoc password verification
- Account-state checks follow PAM-backed policy evaluation
- Password-change-required states are surfaced from the host auth model rather
  than replaced with generic auth failure

---

### SR-002 Use signed JWT access tokens for bounded local trust

#### SR-002 Description

The Auth Service MUST issue signed access tokens that support bounded offline
trust within the LiteNAS service environment.

#### SR-002 Acceptance Criteria

- Access tokens are cryptographically signed
- Access tokens include expiry and principal-identifying claims
- Access tokens are short-lived and are not treated as indefinite session
  authority
- Services can combine local validation with live validation for critical
  actions

---

### SR-003 Protect refresh-token continuity as server-managed state

#### SR-003 Description

The Auth Service MUST treat refresh tokens as server-managed session continuity
artifacts and protect them against replay and stale reuse.

#### SR-003 Acceptance Criteria

- Refresh tokens are single-use within a rotation chain
- Refresh-token reuse after rotation is rejected
- Logged-out, discarded, or lockdown-invalidated refresh tokens cannot be used
  to mint new access tokens

---

### SR-004 Apply rate limiting and secure defaults

#### SR-004 Description

The Auth Service MUST apply secure defaults to exposed auth flows.

#### SR-004 Acceptance Criteria

- Authentication attempts are rate-limited or throttled
- Sensitive secrets and signing keys are not hardcoded in source code
- The service design assumes protected transport between the browser boundary
  and backend auth flows

---

## Testability Requirements

### TR-001 Keep PAM, token, and messaging dependencies injectable

#### TR-001 Description

The Auth Service MUST keep its PAM integration, token issuer, refresh store,
clock, and messaging dependencies injectable so the service can be tested
without live host auth or live messaging infrastructure.

#### TR-001 Acceptance Criteria

- Auth flow logic can be tested with PAM test doubles
- Token issuance and validation logic can be tested with deterministic test
  keys and clocks
- Refresh-session behavior can be tested without external persistence
- Messaging contracts can be tested without a live NATS server
