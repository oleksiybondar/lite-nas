# Auth Service

## Purpose

`auth-service` is the host-local authentication authority for LiteNAS.

It verifies real login-capable users of the managed Linux machine through
PAM-backed flows, issues session tokens for the browser-facing management
interface, and exposes auth-related request/reply and event contracts over
NATS.

## Architectural Role

The auth service owns:

- host authentication and account-state evaluation through PAM
- password-change-required flows required by host auth policy
- access-token issuance and refresh-token rotation
- refresh-token revocation and volatile in-memory refresh-session state
- lockdown state and lockdown event publication
- online token validation for critical backend flows

The auth service does not own:

- browser cookie transport policy
- reverse proxy concerns
- general domain authorization decisions for unrelated services

## Deployment Context

The expected deployment model is:

- `nginx` as the public reverse proxy
- `web-gateway` as the browser-facing HTTP application gateway
- `auth-service` as the PAM-backed auth authority for this machine
- NATS as the internal service-to-service transport

The browser is an HMI for managing the same host on which LiteNAS runs. The
gateway relays auth requests to the auth service rather than embedding PAM
logic directly.

## Build Requirements

`auth-service` is a PAM-only component.

It is expected to build only in environments that provide PAM development
headers and a working CGO toolchain. A non-PAM fallback build is not supported
by the LiteNAS product model.

## Identity Model

The auth service is intended to authenticate only real users who can actually
log in to the managed machine under the active host account model.

This means:

- local and external identities may both be valid when the host PAM and NSS
  configuration permits them
- existence in `/etc/passwd` alone is not sufficient
- service users and non-interactive system accounts must be rejected by policy

## Session Model

The current intended session model is:

- access tokens are signed JWTs with a short lifetime
- default access-token lifetime is 15 minutes
- refresh tokens are stored as server-managed in-memory session state
- a successful login issues a refresh token with a 30 day lifetime
- refresh rotates the token pair without extending the original refresh expiry
- service restart invalidates all refresh sessions

This is intentionally a volatile single-node session model.

## Trust Model

The trust model is intentionally split:

- services may validate JWT access tokens locally for normal protected flows
- services may call the auth service over NATS for live token validation during
  critical or dangerous flows

This allows bounded local trust for routine requests while keeping an online
credibility check available when revocation-sensitive behavior matters.

## Lockdown Model

The auth service is expected to support a lockdown mode.

When lockdown is enabled:

- all auth and refresh requests are rejected
- all refresh-session state is discarded immediately
- a lockdown-enabled event is published over NATS

When lockdown is disabled:

- new auth flows may resume
- previously discarded sessions are not restored
- a lockdown-disabled event is published over NATS

The detailed lockdown behavior and downstream service reactions are expected to
evolve further as the surrounding platform design grows.

## Messaging Role

The auth service is expected to expose explicit NATS contracts for:

- login
- password-change-driven auth continuation
- refresh
- logout
- access-token validation
- lockdown state transitions

It should also publish explicit lockdown state events so other services can
react without needing to poll.

## Boundary With Web Gateway

`web-gateway` remains the browser transport boundary.

The gateway should own:

- HTTP routes and OpenAPI surface
- secure cookie transport concerns
- extraction of access and refresh tokens from supported browser transports
- adaptation of browser-facing auth flows into backend messaging calls

The auth service should own:

- interpretation of auth outcomes
- token issuance semantics
- host auth integration
- live auth credibility decisions

## Non-Goals

The auth service should not become a general policy engine for all service
authorization rules.

It establishes authenticated identity and session credibility for LiteNAS. Each
domain service remains responsible for its own operation-level authorization
decisions on top of that identity context.
