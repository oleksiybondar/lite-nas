# Shared Browser Authentication — Go Module Task

## Goal

Extract the smallest reusable Go module needed for Web Gateway and Moments to
share the browser login/session behavior. Keep route-specific DTOs and HTTP
composition in each service while sharing token/cookie behavior so the two
services do not drift.

## Required behavior

- Preserve the existing login, access-token refresh, logout, and current-user
  behavior backed by the existing Auth Service contracts.
- Use the existing `lite-nas-at` access cookie and `lite-nas-rt` refresh cookie
  with `Secure`, `HttpOnly`, `SameSite=Lax`, and `Path=/`.
- Add configurable cookie domain support so the cookies can be shared by the
  admin host and `moments.<host>`. A host-only cookie must remain possible for
  local/single-host deployments. Do not hard-code a production parent domain.
- Ensure cookie creation and expiration use identical name, domain, path, and
  security attributes, so logout clears cookies issued by login/refresh.
- Preserve token validation, refresh rotation, CSRF defenses, and error
  behavior. A shared package must not give Moments a separate token lifecycle.

## Implementation instructions

1. Inventory existing Web Gateway auth controllers, service helpers,
   `shared/httpcookie`, Auth RPC contracts, and tests.
2. Design the narrow shared package API before moving code. Share stable auth
   and cookie behavior; leave Huma DTOs, routes, and service-specific
   orchestration in their owning services.
3. Add a validated configuration value for the optional cookie domain and
   wire it into Web Gateway and Moments config. Reject malformed cookie-domain
   values rather than emitting unsafe `Set-Cookie` headers.
4. Migrate Web Gateway to the shared behavior without changing its public
   endpoints or response contract. Use the same module from Moments.
5. Verify embedded `/apps/moments` and standalone `moments.<host>` requests
   carry the same session cookies over HTTPS, and that logout expires the same
   cookie scope.

## Acceptance criteria

- Web Gateway and Moments share the same browser auth and cookie behavior.
- A configured common parent domain allows the same secure cookies on the
  admin host and Moments subdomain; unset configuration retains host-only
  behavior.
- Login, refresh, logout, CSRF protection, and cookie clearing retain their
  established contracts.
- No service-specific HTTP DTO or route layer is moved into the shared package.
