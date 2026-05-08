# Development Notes

## Why system-metrics came first

The `system-metrics` service is not the highest business-value capability in the
intended LiteNAS scope. By itself, it delivers only a small part of the overall
platform vision.

It was still chosen as the first implemented service on purpose.

`system-metrics` is simple enough to let the project establish the foundational
building blocks first:

- repository structure
- shared Go modules
- service runtime patterns
- configuration loading
- logging
- messaging integration
- test conventions
- build scripts
- deployment scripts
- packaging
- CI/CD workflow structure

This follows the intended LiteNAS development approach: build the platform from
small, understandable bricks first, and only then expand into higher-value
business logic and broader product capabilities.

The goal is to avoid jumping directly into complex domain behavior before the
base infrastructure is reproducible, testable, and maintainable.

In practice, `system-metrics` acts as an infrastructure-seeding service rather
than a statement of final product priority.

## Why the web skeleton is still infrastructure-first

The current web skeleton adds:

- `web-gateway` as the browser-facing Go HTTP boundary
- `auth-service` as a dedicated PAM-backed auth authority
- `admin-panel` as a Vite + React browser app
- packaging and assembly wiring so web assets are consumed by backend packaging
- CI/CD updates so JS/TS build artifacts are part of the reproducible build flow

This is also intentionally low immediate business value.

The purpose is not to ship a feature-rich user-facing product yet. The purpose
is to complete the next missing infrastructure boundary:

- browser-facing app wiring
- service-to-web integration points
- frontend build integration in CI/CD
- asset handoff into packaging
- reproducible installation and release flow beyond local scripts or manual tweaks

This stage should make the platform feel operationally complete enough that the
remaining effort is mostly product and domain implementation rather than
continued build and deployment bootstrapping.

## Platform-web-skeleton branch outcome

- The platform now has several services and apps that build and install through
  the same repository-owned flows.
- Internal service communication is wired over NATS rather than direct browser
  access to backend internals.
- The browser gateway serves packaged admin-panel assets and exposes documented
  browser-facing auth and system metrics surfaces.
- The frontend build output is produced under `.build/admin-panel`, normalized
  for gateway serving, and included in Debian package assembly.
- CI/CD scripts cover frontend dependencies, frontend build, JS/TS analysis,
  duplication checks, package analysis, and package install validation.
- The branch remains intentionally small in end-user behavior, but it completes
  enough of the skeleton that future branches can focus on actual NAS product
  development rather than platform bootstrapping.

## Boundary Validation Direction

LiteNAS treats validation as a boundary concern. User-facing and
service-facing inputs should be validated before any meaningful processing
happens inside the application.

For HTTP and HTTPS boundaries, `web-gateway` should keep using Huma DTO
validation because those DTOs already describe browser-facing transport
contracts and OpenAPI metadata.

For non-HTTP Go boundaries, including CLI tools, NATS request/reply handlers,
NATS subscription handlers, and other message handlers, the default validation
library should be `go-playground/validator/v10`. These handlers should validate
incoming command or message structs before mapping them into service calls.

For frontend runtime validation, the admin panel should use Zod by default.
TypeScript DTOs document compile-time shapes, but Zod schemas should enforce
runtime input constraints before requests are sent. Form components may own
user-facing field state and messages, but providers and API actions should still
reuse the same schemas before accepting submitted payloads.

This split keeps HTTP/OpenAPI concerns in Huma, gives internal service messages
and CLI inputs a transport-neutral Go validation path, and gives browser code a
shared runtime validation layer.

## Packaging and Runtime Flow

For packaging/runtime orchestration policy and CI parity expectations, see
[Packaging Runtime Flow](./packaging-runtime-flow.md).
