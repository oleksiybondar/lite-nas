# Web Gateway — Shared Browser Session and Application Routing Task

## Goal

Keep Web Gateway as the owner of the existing browser authentication
endpoints, use the shared browser-auth Go module, and support shared secure
session cookies for the admin-panel host and the Moments subdomain.

## Required behavior

- Preserve the existing login, refresh, logout, and current-user HTTP
  contracts.
- Configure an optional cookie parent domain through the established service
  configuration. When configured, issue and clear `lite-nas-at` and
  `lite-nas-rt` consistently for the admin and `moments.<host>` hosts.
- Keep cookies `Secure`, `HttpOnly`, `SameSite=Lax`, and `Path=/`. Do not
  weaken TLS or expose tokens to frontend JavaScript.
- Continue serving the admin panel and existing APIs. Nginx owns the
  `/apps/file-explorer/`, `/apps/moments/`, and standalone Moments hostname
  routing; do not turn Web Gateway into the owner of Moments domain routes.
- Use the shared auth implementation at
  [shared browser auth](../../shared/go/browserauth/task.md), while retaining
  gateway-owned HTTP DTOs, Huma route registration, and response shaping.

## Implementation actions

1. Review the current auth route/controller/config and test contracts.
2. Add optional cookie-domain configuration and use the shared auth module.
3. Preserve logout cookie scope exactly as issued by login and refresh.
4. Update Nginx and deployment configuration only for gateway-owned routes;
   keep Moments proxy and app asset routing in the appropriate deployment
   configuration.
5. Add focused tests for host-only and shared-domain cookie behavior, refresh,
   logout, and existing gateway contract compatibility.

## Acceptance criteria

- Existing Web Gateway auth routes retain their current contracts.
- Cookie domain can be configured for the production parent domain without
  hard-coding it into source.
- Logout clears cookies using the same scope as issuance.
- Web Gateway remains a shared auth/routing service and contains no
  Moments-specific album or media business logic.
