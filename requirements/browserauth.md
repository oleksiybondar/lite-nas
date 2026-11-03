# Shared Browser Authentication Requirements

- **browserauth/FR-001:** Web Gateway and Moments use one Go implementation for
  Auth Service login, refresh, logout, current-user verification, and session
  claims mapping.
- **browserauth/FR-002:** The configured browser cookie domain is either empty
  or a valid DNS parent domain. Invalid values fail service startup.
- **browserauth/FR-003:** Access and refresh cookies are secure, HTTP-only,
  SameSite Lax, and scoped to `/`. Expiration uses the same name, domain,
  path, and security attributes as issuance.
- **browserauth/FR-004:** Both apps reject unsafe browser requests whose
  Origin host differs from the request host before reaching route logic.
