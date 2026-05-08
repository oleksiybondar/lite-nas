# Admin Panel — Requirements

## Overview

The Admin Panel is the browser application for LiteNAS.

It is served as packaged static assets by `web-gateway` and uses the gateway's
browser-facing HTTP API. It does not own authentication token storage,
authorization policy, host-authentication behavior, or backend domain logic.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Use BFF cookie-based session state

#### FR-001 Description

The Admin Panel MUST use the Web Gateway BFF session contract for
authentication.

#### FR-001 Input

- Browser-visible authentication state from gateway API responses
- HTTP-only access-token and refresh-token cookies managed by the browser

#### FR-001 Output

- Browser API requests that include gateway session cookies
- UI authentication state derived from gateway responses

#### FR-001 Acceptance Criteria

- The app does not read, store, or parse access-token or refresh-token values
- Gateway requests that need cookies use credential-including request behavior
- Gateway API calls use `/api`-prefixed paths
- Initial auth detection is based on gateway session endpoints rather than
  direct JavaScript cookie inspection
- A `401` from the access-token-backed user lookup is treated as a possible
  refresh path before the app decides the user is not authenticated

---

### FR-002 Keep authentication state explicit in the UI shell

#### FR-002 Description

The Admin Panel MUST represent authentication state explicitly so the shell can
distinguish startup checks and authenticated sessions.

#### FR-002 Input

- Gateway user/session responses
- Gateway `401` responses from protected endpoints
- Gateway refresh success or failure responses

#### FR-002 Output

- UI state suitable for routing, loading, and login decisions

#### FR-002 Acceptance Criteria

- The app can represent at least initialized, not initialized, authenticated,
  and not authenticated states
- The auth provider exposes auth-initialized, authenticated, and auth-status
  handling operations to the rest of the app
- Protected UI flows are not rendered as authenticated until the gateway confirms
  the session
- Failed refresh causes the app to clear authenticated UI state

---

## Interface Requirements

### IR-001 Use gateway-relative API calls

#### IR-001 Description

The Admin Panel MUST call the Web Gateway API as its BFF instead of calling
internal services directly.

#### IR-001 Acceptance Criteria

- The app does not call NATS-backed services directly
- Auth, system metrics, and future NAS features go through gateway-owned
  `/api` HTTP endpoints
- API client code remains testable without a deployed backend

---

## Security Requirements

### SR-001 Keep token material out of browser JavaScript

#### SR-001 Description

The Admin Panel MUST treat access-token and refresh-token values as unavailable
to browser JavaScript.

#### SR-001 Acceptance Criteria

- Token values are not stored in local storage, session storage, React state, or
  query caches
- Token values are not expected in login or refresh JSON response bodies
- Logout and refresh behavior depends on gateway-managed HTTP-only cookies

---

## Testability Requirements

### TR-001 Test auth flow through observable gateway behavior

#### TR-001 Description

The Admin Panel MUST test auth behavior through request/response outcomes rather
than direct token inspection.

#### TR-001 Acceptance Criteria

- Unit tests can mock gateway responses for `/api/auth/me` and
  `/api/auth/refresh`
- Tests verify credential-including request configuration where auth cookies are
  required
- Tests do not require JavaScript access to HTTP-only cookie values

---
