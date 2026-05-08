# Web Gateway

## Purpose

`web-gateway` is the browser-facing HTTP gateway for LiteNAS.

It has three primary responsibilities:

- serve packaged frontend static assets
- expose the browser-facing HTTP API
- adapt frontend API calls to internal service interfaces behind the messaging layer

## Architectural Role

The gateway is intentionally thin.

It does not own:

- core domain logic
- authorization policy
- host authentication policy
- general file upload or file download behavior

Instead, it acts as the HTTP entrypoint that translates browser-facing
interactions into calls to backend services that own their logic behind
NATS.

## Deployment Context

The expected deployment model is:

- `nginx` as the public reverse proxy
- `web-gateway` as the browser-facing HTTP application gateway
- a separate file-serving service for dataset download and upload flows
- authorization and domain logic behind dedicated backend services

The gateway is expected to run behind the reverse proxy rather than own
public TLS termination directly.

## Tokens and Security Boundary

The gateway is responsible for browser-facing secure token transport
concerns such as handling HTTP-only token delivery at the web boundary.

It is not the authorization authority.

Authorization policy, host authentication, token issuance, and token
semantics are expected to be owned by dedicated backend services.
`auth-service` fills that role. File-service
short-lived JWT mechanics are intentionally out of scope for this
component.

The gateway currently distinguishes two browser-facing token transport
policies:

- access token via `Authorization: Bearer ...`
  intended for explicit REST-style clients
- access token or refresh token via HTTP-only cookies
  intended for browser-facing session transport

Protected endpoints are expected to use access-token transport through
either the `Authorization` header or the access-token cookie. Browser
auth transition endpoints such as refresh and logout use the
HTTP-only refresh-token cookie rather than request-body token fields.

This remains compatible with non-browser HTTP clients that support cookie jars.
Tools such as `curl`, `wget`, Python `requests`, and API test clients can store
`Set-Cookie` headers and replay those cookies without requiring token values in
JSON response bodies.

## HTTP and Messaging Stack

The initial implementation is expected to use:

- `chi` for HTTP routing and middleware
- `huma` for OpenAPI generation and development-facing documentation
- an injected NATS client for communication with internal services

This is an intentional tradeoff: not the most minimal possible stack,
but simple, effective, and useful for documented gateway development and
debugging.

## Frontend Asset Flow

The LiteNAS web application lives under `apps/admin-panel`.

Its Vite build output is written to `.build/admin-panel` by
`scripts/build-admin-panel.sh`. Deployment and packaging normalize that output
into owned static resources of the gateway. The gateway serves those assets,
while non-static file flows remain outside the gateway boundary.

The current gateway-owned packaged asset layout is:

- `/usr/share/lite-nas/web-gateway/assets/index.html`
- `/usr/share/lite-nas/web-gateway/assets/index.css`
- `/usr/share/lite-nas/web-gateway/assets/index.js`
- `/usr/share/lite-nas/web-gateway/assets/favicon.ico` when the frontend build
  provides one

These packaged files are treated as service-owned read-only resources
rather than user-configurable runtime content.

## Package Structure

The current service package layout is:

- `config/`
  service-local runtime config composition for shared `http`,
  `messaging`, and `logging` sections
- `controllers/`
  browser-facing endpoint flow orchestration and HTTP/OpenAPI response
  mapping
- `dto/`
  HTTP request/response DTOs used for browser-facing JSON and OpenAPI
  contracts
- `middlewares/`
  reusable HTTP middleware for authenticated and transport-level
  concerns
- `modules/`
  runtime composition modules for infrastructure, services, and
  controllers
- `routes/`
  root router composition plus mounted route slices such as index,
  assets, auth, and system metrics
- `services/`
  backend-facing gateway capabilities such as NATS request/reply calls
  hidden behind service methods

This split is intended to keep route ownership local, keep controller
logic readable as the HTTP surface grows, and avoid leaking messaging
transport concerns into router code.

## Responsibility Boundaries

The gateway follows this dependency direction:

- `routes -> controllers -> services`

Responsibilities are intentionally separated as follows:

- `routes` compose and mount endpoints, middleware, and route groups
  without owning endpoint behavior
- `controllers` own endpoint flow, call one or more services as needed,
  and map internal/shared data into HTTP DTOs
- `controllers` are the only layer that should be aware of browser-facing
  HTTP transport details such as headers, cookies, and payload transport
  policy precedence
- `services` own reusable backend-facing capabilities such as NATS
  request/reply operations and should not own final HTTP response
  formats or HTTP transport-specific request objects
- `dto` owns browser-facing contract shapes rather than domain models

In practical terms, shared or service-owned structs should remain in
shared packages or service packages, while controllers translate those
structures into browser-facing DTOs and extract the final service inputs
from validated DTOs.

Within `dto/`, prefer endpoint-oriented files such as `login.go`,
`refresh.go`, or `snapshot.go` for endpoint-specific DTOs.

When multiple endpoints intentionally share the same HTTP contract, use
a narrowly named shared file based on the concept being shared, for
example `session_output.go`, instead of collecting unrelated DTOs in a
generic `shared.go`.

## Current HTTP Surface

The current skeleton HTTP surface is:

- `/`
  serves the packaged browser entrypoint HTML resource
- non-API browser navigation paths such as `/dashboard`
  serve the same packaged browser entrypoint HTML resource for SPA routing
- `/favicon.ico`
  serves the packaged browser favicon
- `/assets/index.css`
  serves the packaged browser stylesheet
- `/assets/index.js`
  serves the packaged browser JavaScript bundle
- `/api/auth/...`
  browser-facing auth transport endpoints intended to adapt requests to
  the dedicated `auth-service` over NATS
- `/api/system-metrics/...`
  browser-facing JSON endpoints backed by internal system metrics NATS
  calls
- `/api/docs`
  development-facing OpenAPI documentation generated through `huma`

The gateway currently uses plain `chi` mounting for static resource
routes and a shared `huma` API for documented browser-facing API
endpoints. API endpoints are mounted under `/api`; SPA fallback behavior is
limited to non-API routes.

The static resource layer currently uses a dedicated files module that
wires explicit injected file readers for each packaged frontend file
under `/usr/share/lite-nas/web-gateway/assets`. Static routes should
stay explicit rather than falling back to wildcard directory serving.

Current auth transport middleware behavior is intentionally split into:

- extraction middleware applied broadly to documented API endpoints
- enforcement middleware applied only to protected endpoints

This allows protected endpoints to depend on a normalized
header-or-cookie access-token policy, while public auth endpoints can
depend on the HTTP-only refresh-token cookie for BFF session renewal and
logout flows.

## Non-Goals

The gateway should not slowly turn into a general backend that owns
business logic.

If a feature requires domain ownership, authorization policy ownership,
host authentication ownership, or broad file-transfer ownership, that
logic should remain in dedicated backend services rather than being
absorbed into the gateway.
