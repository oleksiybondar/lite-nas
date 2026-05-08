# Packaging Runtime Flow

## Purpose

Define a single packaging/deployment flow that avoids "works on my machine"
drift between local scripts and CI package installation.

## Decision

LiteNAS keeps a strict separation between:

- runtime deployment logic (required on installed hosts)
- development workflows (convenience for local iteration)

The Debian package may include runtime scripts, but must not depend on dev-only
entrypoints or dev-only side effects.

## Runtime vs Dev Script Policy

1. Runtime scripts are allowed in package payload when they represent real
   business/runtime installation behavior.
2. Dev-only scripts are not part of the hardened deployment contract and should
   not be required by package `postinst`.
3. `postinst` remains a small orchestrator that invokes runtime deployment
   entrypoints in deterministic order.

## Required Post-Install Sequence

The runtime installation flow should execute, in order:

1. Deploy baseline configs.
2. Ensure runtime dependencies are present and installed by package manager.
3. Generate certificates when missing.
4. Deploy/enable service units and runtime assets.
5. Reload/restart affected services so newly deployed config becomes effective.
6. Run readiness checks against effective service behavior (not only process
   "active" state).

## CI/CD Parity Rule

System tests must validate the same runtime behavior expected from a fresh host
package install. Any install mode that intentionally skips service start/reload
is a packaging-structure check, not a runtime-readiness check.

## Implementation Guidance

- Keep reusable shell functions in helper libraries.
- Keep runtime step scripts explicit and ordered.
- Keep dev orchestration separate and composed from shared helpers where needed.
- Avoid dynamic "concatenate source files into postinst" generation.
- Prefer deterministic entrypoint invocation from `postinst` for traceable logs
  and failure diagnosis.
