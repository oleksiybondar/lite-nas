# Release Notes

This file tracks release-level changes for LiteNAS.

The working format and guidance are documented in
[docs/release-notes.md](docs/release-notes.md).

## 0.2.0 - Authenticated monitoring and alerting

### RL-0.2.0 Summary

- Extends the platform skeleton into a fuller monitoring and alerting slice
  with RBAC-backed roles, authenticated internal service calls, stateful alert
  consumption, and end-to-end resource alert generation from system and ZFS
  metrics.

### RL-0.2.0 Added

- Added `rbac-service` as the internal authorization decision point for role
  lookup and capability checks.
- Added service-to-service token login and refresh flows in `auth-service`.
- Added `zfs-metrics` as a ZFS snapshot producer with event and RPC contracts.
- Added `resources-monitor` as a rule-based alert source for system and ZFS
  metric events.
- Added `system-logging-manager`, `security-logging-manager`, and their CLI
  surfaces to the packaged runtime path.

### RL-0.2.0 Changed

- Changed logging-manager runtime boundaries to validate access tokens and
  enforce subject-level role policy before applying writes or state changes.
- Changed auth token issuance to resolve role context through `rbac-service`
  for both user and service principals.
- Changed the effective monitoring path from isolated metrics reporting into a
  complete alert lifecycle flow ending in stateful logging-manager consumers.
- Clarified the split between system and security logging-manager domains as an
  intentional architecture boundary rather than an accidental duplication.

### RL-0.2.0 Platform

- Expanded package build and runtime deployment to include the new RBAC,
  logging-manager, and resources-monitor components.
- Expanded managed transport-certificate coverage and deploy flows for the new
  service and CLI identities.
- Increased the value of existing system-level CLI checks because the same
  simple assertions now transit deeper integration paths across permissions,
  service authentication, RBAC, and manager authorization.

### RL-0.2.0 Notes

- The current operational priority is CLI-first resource monitoring and
  condition-change alerting for a home-lab deployment.
- Security-monitoring producers and richer UI alert consumers remain follow-up
  work on top of the current infrastructure.

## 0.1.1 - Platform web skeleton

### RL-0.1.1 Summary

- Extends the platform skeleton with a web-facing slice, packaged admin-panel
  assets, auth gateway integration, and CI/CD routines that make the repository
  ready for product-focused work in later branches.

### RL-0.1.1 Added

- Added `web-gateway` as the browser-facing HTTP boundary for static frontend
  assets and browser API adaptation.
- Added `auth-service` as the PAM-backed authentication authority behind the
  internal NATS boundary.
- Added `admin-panel` as a Vite + React + TypeScript browser application with
  provider, context, hook, route, alias, and theme structure.
- Added `scripts/build-admin-panel.sh` and shared admin-panel asset helpers for
  reproducible frontend build and deployment handoff.

### RL-0.1.1 Changed

- Changed web-gateway deployment to install built admin-panel assets from
  `.build/admin-panel` into `/usr/share/lite-nas/web-gateway/assets`.
- Changed Debian packaging to build, validate, include, and install admin-panel
  artifacts as the web-gateway static asset source.
- Changed CI/developer Node setup and build flows so the app-local frontend
  dependencies and admin-panel build participate in the repository workflow.
- Changed PR validation and the main package gate to run real JS/TS build and
  unit test jobs for the admin-panel application instead of placeholder lanes.
- Changed PR and main package assembly to consume an explicit admin-panel asset
  artifact so Debian packages use the same built web assets produced by CI.
- Changed browser-facing auth refresh and logout behavior to use the
  HTTP-only refresh-token cookie instead of request-body refresh-token fields,
  matching the gateway's BFF session boundary.
- Changed web-gateway routing so JSON API endpoints live under `/api` while
  non-API browser navigation paths serve the packaged SPA `index.html`.
- Changed the admin-panel Vite development proxy to forward `/api` to the
  gateway target while leaving other paths to the SPA fallback.
- Added the admin-panel `useAuth` context/provider for BFF startup session
  detection, themed authentication loading state, a login route target, and
  sequential `/api/auth/me` then `/api/auth/refresh` handling for `401`
  responses.
- Added admin-panel requirements for BFF cookie auth, auth-state detection, and
  token-free browser JavaScript behavior.
- Clarified that cookie-based BFF auth remains usable by non-browser clients
  through normal cookie-jar handling.
- Changed JS/TS analysis and formatting exclusions to ignore generated output
  and nested dependency directories while still checking source files.

### RL-0.1.1 Platform

- The skeleton now includes multiple microservices communicating over NATS,
  a browser-facing gateway, a browser app shell, deployment scripts, Debian
  packaging, and CI/CD validation for the current platform slice.
- Admin-panel Vite output is built into `.build/admin-panel` and normalized
  into the gateway-owned static asset layout during deploy and package assembly.
- Added repo-wide duplication enforcement for Go and shell code, including
  cross-module detection instead of relying only on per-module checks.
- Added JS/TS duplication enforcement for frontend source.
- Established a repository testing discipline that moves repeated setup into
  shared `testutil`, subproject-local `testutil`, and package-local fixtures
  rather than allowing test-only duplication to accumulate.

### RL-0.1.1 Notes

- This release is still expected to be low direct business value and high
  platform-completion value.
- The practical outcome is that future branches can start actual LiteNAS
  product development on top of an already wired service, gateway, frontend,
  packaging, deployment, and CI/CD skeleton.

## 0.1.0 - Initial platform skeleton

### RL-0.1.0 Summary

- Establish the first reproducible LiteNAS platform slice with a minimal
  service, a minimal CLI app, shared communication/runtime foundations, and
  packageable installation flow.

### RL-0.1.0 Added

- Added the `system-metrics` service as the first platform-seeding backend
  service.
- Added the `system-metrics-cli` app as the first platform-seeding consumer
  application.
- Added shared Go modules for messaging, logging, configuration, metrics, and
  related runtime support.

### RL-0.1.0 Changed

- Wired service-to-app communication over NATS as the first internal platform
  communication path.
- Established repository structure and construction patterns for services,
  apps, and shared modules.

### RL-0.1.0 Platform

- Added reproducible Debian packaging, deployment, runtime configuration, and
  install validation so setup no longer depends on local scripts, ad hoc
  commands, or manual host tweaks.
- Added CI/CD validation for analysis, build, test, packaging, and package
  installability.

### RL-0.1.0 Notes

- This release has intentionally modest direct business value.
- Its primary value is to establish the initial LiteNAS platform skeleton so
  later work can focus more on product behavior than infrastructure bootstrapping.
