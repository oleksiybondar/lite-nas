# Release Notes

This file tracks release-level changes for LiteNAS.

The working format and guidance are documented in
[docs/release-notes.md](docs/release-notes.md).

## 0.1.1 - Planned

### RL-0.1.1 Summary

- Extend the current platform skeleton with a minimal web-facing slice so the
  browser delivery path, frontend build flow, and packaging handoff become
  reproducible parts of the platform.

### RL-0.1.1 Added

- Planned introduction of a simple Go web service.
- Planned introduction of a simple Vite + React web app.

### RL-0.1.1 Changed

- Planned extension of the runtime wiring to include web-facing integration.
- Planned extension of the assembly flow so JS/TS build artifacts are consumed
  by backend packaging.

### RL-0.1.1 Platform

- Planned CI/CD updates so JS/TS build jobs produce real artifacts for later
  packaging stages.
- Planned NATS configuration updates required for the web-facing slice.
- Planned service hardening updates needed to support the next platform stage.

### RL-0.1.1 Notes

- Intended to expose a basic public page that shows CPU and RAM trend data.
- This release is still expected to be low direct business value and high
  platform-completion value.

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
