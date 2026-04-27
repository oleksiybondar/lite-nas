# Web Gateway — Requirements

## Overview

The Web Gateway is the browser-facing HTTP gateway for LiteNAS.

It serves packaged frontend static assets, exposes the HTTP API used by
the web application, and adapts browser-facing requests to internal
service interfaces behind the messaging layer.

The Web Gateway is intentionally a thin layer. It does not own core
domain logic, authorization policy, or file-transfer logic.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Serve packaged frontend static assets

#### FR-001 Description

The Web Gateway MUST serve the packaged static assets of the LiteNAS web
application.

#### FR-001 Input

- HTTP requests for frontend application assets
- Packaged frontend build output

#### FR-001 Output

- Static asset responses for HTML, CSS, JavaScript, and related frontend files

#### FR-001 Acceptance Criteria

- The gateway serves packaged frontend assets from its owned packaged asset area
- Static asset serving does not require direct access to domain services
- The gateway serves only its owned static application assets

---

### FR-002 Expose a browser-facing HTTP API

#### FR-002 Description

The Web Gateway MUST expose an HTTP API used by the LiteNAS web
application.

#### FR-002 Input

- Browser-originated HTTP API requests

#### FR-002 Output

- HTTP API responses suitable for frontend consumption

#### FR-002 Acceptance Criteria

- The HTTP API is routable through the gateway
- API responses use formats suitable for browser application use
- The API remains separate from static asset serving concerns

---

### FR-003 Adapt frontend API requests to internal service contracts

#### FR-003 Description

The Web Gateway MUST translate supported frontend-facing API calls into
requests to internal services over the messaging layer.

#### FR-003 Input

- Browser-facing API requests
- Internal service contracts exposed through messaging

#### FR-003 Output

- Messaging requests to internal services
- Adapted HTTP responses derived from internal service responses

#### FR-003 Acceptance Criteria

- Supported browser-facing API calls are mapped to internal service interactions
- The gateway does not embed the underlying domain logic it forwards to
- Transport and payload adaptation are encapsulated within the gateway layer

---

### FR-004 Keep browser token transport concerns isolated

#### FR-004 Description

The Web Gateway MUST handle browser-facing secure token transport
concerns without owning authorization policy.

#### FR-004 Input

- Access token and refresh token values issued by external authorization services
- Browser requests carrying or receiving secure session state

#### FR-004 Output

- HTTP-only secure token transport behavior at the browser-facing boundary

#### FR-004 Acceptance Criteria

- The gateway can set and forward HTTP-only secure token material as required
- The gateway does not issue authorization policy decisions itself
- Token transport handling remains separate from domain service logic

---

## Interface Requirements

### IR-001 Provide documented HTTP routing for development and debugging

#### IR-001 Description

The Web Gateway MUST provide HTTP routing and automatically generated
OpenAPI documentation for its browser-facing API.

#### IR-001 Input

- Declared HTTP routes and schemas

#### IR-001 Output

- Routable HTTP endpoints
- Generated OpenAPI documentation

#### IR-001 Acceptance Criteria

- The gateway exposes its supported HTTP API through a documented router
- OpenAPI output is generated from the gateway API definitions
- The documentation is suitable for development and debugging use

---

### IR-002 Depend on the messaging layer for backend communication

#### IR-002 Description

The Web Gateway MUST use the messaging layer for communication with
internal services.

#### IR-002 Input

- Configured messaging client
- Browser-facing API requests requiring backend service interaction

#### IR-002 Output

- Messaging traffic between the gateway and internal services

#### IR-002 Acceptance Criteria

- Internal service communication is performed through the messaging client
- Messaging client construction is injectable into gateway runtime wiring
- Browser-facing HTTP handling remains decoupled from direct domain implementation

---

## Operational Requirements

### OR-001 Operate behind a reverse proxy

#### OR-001 Description

The Web Gateway MUST be deployable behind an external reverse proxy.

#### OR-001 Acceptance Criteria

- The gateway can run as an internal HTTP service behind a reverse proxy
- TLS termination is not required to be implemented in the gateway itself
- The reverse proxy and gateway responsibilities remain separable

---

### OR-002 Support packaged frontend asset inclusion

#### OR-002 Description

The Web Gateway MUST support packaged inclusion of frontend build
artifacts during assembly.

#### OR-002 Acceptance Criteria

- Frontend build output can be copied into the gateway-owned static asset area
- Package assembly can include frontend assets without manual host edits
- Installation remains reproducible from packaged artifacts

---

## Security Requirements

### SR-001 Exclude file-transfer responsibility from the gateway

#### SR-001 Description

The Web Gateway MUST NOT own general file upload or file download
responsibilities.

#### SR-001 Acceptance Criteria

- File-transfer behavior is not treated as a gateway-owned capability
- File serving for downloaded datasets remains assignable to a separate least-privilege service
- Gateway requirements stay limited to static frontend assets and API transport

---

### SR-002 Exclude authorization policy ownership from the gateway

#### SR-002 Description

The Web Gateway MUST NOT own authorization policy decisions.

#### SR-002 Acceptance Criteria

- Authorization policy decisions are delegated to dedicated backend services
- The gateway is limited to transport, adaptation, and browser boundary concerns
- Gateway behavior remains valid even when authorization service implementation details evolve

---

## Testability Requirements

### TR-001 Keep HTTP adaptation logic testable without live infrastructure

#### TR-001 Description

The Web Gateway MUST keep routing, request adaptation, and response
mapping testable without requiring a full deployed environment.

#### TR-001 Acceptance Criteria

- HTTP handlers can be tested without a live reverse proxy
- Messaging interactions can be replaced by test doubles
- Static asset serving behavior can be tested against controlled test assets
- OpenAPI generation can be validated without full deployment

---

## Notes

- The initial implementation is expected to use `chi` for routing and middleware
- The initial implementation is expected to use `huma` for OpenAPI generation
- Authorization service token issuance and file-service token mechanics are out of scope for this document
