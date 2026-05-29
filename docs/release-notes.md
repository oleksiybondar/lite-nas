# Release Notes

## 0.2.0 - 2026-05-29

### RL-0.2.0 Summary

- Extended the platform skeleton into an authenticated monitoring and alerting
  slice with RBAC-backed service roles, stateful logging-manager enforcement,
  and rule-based alert generation from system and ZFS metrics. The release
  stays operationally CLI-first, but it now exercises a much more complete
  internal runtime path.

### RL-0.2.0 Added

- Added `rbac-service` as the internal authorization decision service for host
  role/group resolution and capability checks.
- Added service-to-service token login and refresh flows so internal LiteNAS
  services can authenticate across the NATS boundary.
- Added `zfs-metrics` as a ZFS snapshot producer with ZFS event/RPC contracts.
- Added `resources-monitor` as a rule-based alert source consuming system and
  ZFS metrics events and publishing alert lifecycle updates.
- Added packaged and deployable system/security logging-manager CLI binaries and
  the shared manager runtime wiring they exercise.

### RL-0.2.0 Changed

- Changed logging managers from unauthenticated internal message consumers to
  authenticated boundaries that validate access tokens and enforce role policy
  before applying writes or state mutations.
- Changed auth-service token issuance so issued user and service access tokens
  resolve role context through `rbac-service`.
- Clarified the system/security logging-manager split as intentional
  architecture, even though only the system side currently has richer alert
  sources behind it.
- Changed the monitoring path from a simple metrics seed slice into an
  end-to-end alert flow: metrics producers -> resources-monitor ->
  logging-manager state -> CLI consumers.

### RL-0.2.0 Platform

- Expanded the packaged LiteNAS runtime to include `rbac-service`,
  `system-logging-manager`, `security-logging-manager`,
  `system-logging-manager-cli`, `security-logging-manager-cli`, and
  `resources-monitor` beside the previously packaged auth, metrics, and web
  components.
- Extended repository build/deploy flows and managed NATS certificate coverage
  for the new service and CLI identities participating in authenticated
  messaging.
- Increased the integration value of existing system tests because simple CLI
  checks now also transit permission boundaries, service authentication, RBAC
  role retrieval, and manager-side authorization logic.

### RL-0.2.0 Notes

- The current monitoring focus is resource and condition-change alerting for a
  CLI-oriented home-lab workflow.
- Security and UI-facing alert consumers are intentionally prepared but not yet
  expanded into full security-monitoring or browser-native alert-management
  products.

## 0.1.1 - 2026-05-06

### RL-0.1.1 Summary

- Expanded the LiteNAS platform skeleton beyond the initial monitoring slice
  with authentication, web gateway, admin-panel, and system-level validation
  wiring. This release is still primarily scaffolding: it establishes runtime,
  packaging, and CI/CD shape for future product work rather than delivering a
  complete user-facing NAS administration experience.

### RL-0.1.1 Added

- Added an `auth-service` skeleton that owns PAM-backed host authentication,
  auth token issuance, and internal auth messaging behind the LiteNAS service
  boundary.
- Added a `web-gateway` skeleton that serves packaged admin-panel assets and
  exposes browser-facing API routes backed by internal services.
- Added the initial `admin-panel` web app skeleton and its build handoff into
  Debian package assembly.
- Added Python HyperionTF system-level tests for installed LiteNAS behavior,
  grouped by infra, CLI, API, and UI categories.
- Added system-test documentation covering category markers, service markers,
  HTML-report docstrings, focused verification points, and shared fixtures.

### RL-0.1.1 Changed

- Extended local and CI validation so Python system tests run after Python
  static analysis and before later duplication checks.
- Ordered system-test execution by credibility gate: infrastructure first,
  then CLI, API, and UI.
- Clarified package analysis so static config checks validate repository files
  while installed-path behavior remains covered by package and system tests.
- Added PAM development headers to Go analysis setup so auth-service cgo
  typechecking matches the build environment.
- Made PR package validation depend on the admin-panel production build as well
  as JS/TS tests before consuming frontend assets.
- Cleaned up CI package validation warnings by declaring the cgo libc runtime
  dependency, building cgo binaries as PIE, fixing multi-module Go cache setup,
  and always uploading a system-test log zip.
- Makes the installed system-metrics CLI usable from normal local user shells
  by installing readable CLI config and client credentials, keeping the shared
  LiteNAS config directory traversable, and initializing its log file during
  deployment.
- Promoted the AIDE binary to a package dependency while keeping AIDE database
  initialization out of Debian installation and CI package validation.
- Standardized HyperionTF log output under `tests/logs` so CI can publish
  predictable test artifacts.

### RL-0.1.1 Platform

- Release package validation now installs the built LiteNAS Debian package and
  runs the Python system test suite against the installed services.
- CI uploads Python system-test HTML logs as `system_tests_logs.zip` artifacts
  with a 15-day retention window.
- Python developer and CI dependency installation now creates the repository
  virtual environment, installs the HyperionTF/pytest stack, and prepares the
  Playwright runtime needed for future UI and visual tests.

### RL-0.1.1 Notes

- Authentication, web gateway, admin-panel, and system tests are intentionally
  early platform scaffolding. The current work verifies wiring, packaging, and
  runtime behavior needed for later feature development.

## Why release notes are needed early

LiteNAS is still in the platform-skeleton stage, but release notes are useful
already.

At this stage, many changes do not add much immediate business value. They
still change the shape of the platform in important ways:

- service and app skeletons
- runtime and messaging wiring
- packaging and installation behavior
- deployment expectations
- CI/CD and release reproducibility

Release notes should make those changes visible so platform progress can be
tracked intentionally rather than inferred from commit history.

## Format

Use one section per release.

Recommended structure:

```md
## X.Y.Z - YYYY-MM-DD

### RL-X.Y.Z Summary

- One short paragraph or 2-4 bullets describing the release intent.

### RL-X.Y.Z Added

- New service, app, module, script, or packaging capability.

### RL-X.Y.Z Changed

- Important behavior or structure changes.

### RL-X.Y.Z Fixed

- Important defects resolved.

### RL-X.Y.Z Platform

- CI/CD, packaging, deployment, reproducibility, or developer-workflow changes.

### RL-X.Y.Z Notes

- Optional limitations, follow-up work, or intentionally incomplete areas.
```

## Guidance

- Prefer business-meaningful summaries over commit-by-commit narration.
- It is acceptable for early releases to emphasize platform and infrastructure
  value over direct end-user value.
- Mention intentionally incomplete slices when a release mainly prepares later
  product work.
- When a release introduces or changes installation, packaging, runtime
  dependencies, or deployment behavior, record that explicitly.
- When a release materially changes repository-wide quality enforcement or
  contributor workflow, such as duplication gates, shared test-helper
  conventions, or CI validation policy, record that under the release
  `Platform` section.
- Use release-qualified subsection headings such as `RL-0.1.0 Summary` and
  `RL-0.1.0 Platform` so markdown headings stay globally unique inside the
  document.
- Keep wording factual and concise.
