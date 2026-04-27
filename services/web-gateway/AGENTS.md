# Web Gateway Agent Instructions

These instructions apply to work under `services/web-gateway/`.

They supplement the repository-level `AGENTS.md`. When these instructions
conflict with the repository-level file, these local instructions take
precedence.

## Architecture Boundaries

The gateway follows this dependency direction:

- `routes -> controllers -> services`

Responsibilities are intentionally split as follows:

- `routes/` own endpoint composition and router mounting only
- `controllers/` own endpoint flow, service orchestration, error mapping, and
  HTTP response shaping
- `controllers/` are the only layer that may be aware of browser-facing HTTP
  transport details such as headers, cookies, and payload transport precedence
- `services/` own backend-facing capabilities such as NATS request/reply calls
  and must hide transport details from controllers
- `dto/` owns browser-facing HTTP/OpenAPI request and response objects only
- `modules/` own runtime composition and dependency assembly
- `middlewares/` own reusable transport-level middleware

## DTO and Service Rules

- Services must work with shared structs or service-owned structs, not HTTP DTOs.
- Services must not accept HTTP transport-shaped request objects when plain
  extracted values are sufficient.
- Controllers must map service results into `dto/` types for browser-facing
  responses.
- Controllers must extract the final service inputs from validated DTOs before
  invoking services.
- Do not treat `dto/` as a domain-model layer.
- Do not move raw NATS request logic into controllers.
- Within `dto/`, prefer endpoint-oriented files for endpoint-specific DTOs.
- When several endpoints intentionally share one DTO shape, prefer a narrowly
  named concept file such as `session_output.go` instead of a generic
  `shared.go`.
- For auth transports, protected endpoints should use header-or-cookie access
  token policies through middleware, while selected public auth endpoints may
  define explicit token payload fields in their DTOs.

## Routing Rules

- Prefer a root router that mounts smaller route slices instead of collecting
  all route registration in one large file.
- Keep static resource routing and documented API routing clearly separated.
- It is acceptable and preferred to register controller methods directly from
  the router when the framework signature allows it.

## Middleware Rules

- Prefer a split between authentication extraction middleware and
  authentication enforcement middleware.
- Extraction middleware may run broadly on documented API endpoints and should
  normalize supported token transport policies into request context.
- Enforcement middleware should run only on protected endpoints.

## Static Asset Rules

- The gateway serves packaged static assets from `/usr/share/lite-nas/web-gateway`.
- Treat packaged frontend files as service-owned read-only resources rather than
  configurable runtime content unless the design is explicitly changed.
- Prefer a dedicated files module for wiring packaged file readers from the
  installed `/usr/share/lite-nas/web-gateway/assets` paths.
- Prefer explicit injected file readers for packaged frontend files over
  directory-backed static file serving.
- Prefer explicit static `GET` routes for known packaged files instead of a
  wildcard assets mount.
