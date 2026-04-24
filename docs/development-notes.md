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

## Why the next web slice is still infrastructure-first

The next planned slice is expected to add:

- a simple Go web service
- a simple Vite + React web app
- packaging and assembly wiring so web assets are consumed by backend packaging
- CI/CD updates so JS/TS build artifacts become part of the reproducible build flow

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

## Planned focus for the platform-web-skeleton branch

- scaffold a minimal Go web service with the same structural conventions used by
  existing services
- scaffold a minimal Vite + React web app with build output suitable for
  packaging
- wire the web app build output into the backend/package assembly flow
- extend CI/CD so JS/TS build jobs produce real artifacts consumed by later
  packaging jobs
- keep the slice intentionally small in end-user behavior while maximizing
  reproducibility and future expansion value
