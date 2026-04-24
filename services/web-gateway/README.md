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

Authorization policy, token issuance, and token semantics are expected
to be owned by dedicated backend services. File-service short-lived JWT
mechanics are intentionally out of scope for this component.

## HTTP and Messaging Stack

The initial implementation is expected to use:

- `chi` for HTTP routing and middleware
- `huma` for OpenAPI generation and development-facing documentation
- an injected NATS client for communication with internal services

This is an intentional tradeoff: not the most minimal possible stack,
but simple, effective, and useful for documented gateway development and
debugging.

## Frontend Asset Flow

The LiteNAS web application will live under `apps/admin-panel`.

Its built frontend assets are expected to become owned static resources
of the gateway during package assembly. The gateway serves those assets,
while non-static file flows remain outside the gateway boundary.

## Non-Goals

The gateway should not slowly turn into a general backend that owns
business logic.

If a feature requires domain ownership, authorization policy ownership,
or broad file-transfer ownership, that logic should remain in dedicated
backend services rather than being absorbed into the gateway.
