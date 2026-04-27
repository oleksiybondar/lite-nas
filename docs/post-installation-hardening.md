# Post-Installation Hardening

## Purpose

This document records hardening work that remains after the LiteNAS Debian
package has been installed. These actions exist because some controls cannot be
safely decided in a generic package without understanding the target host,
attached devices, network exposure, and operational workflow.

## Operator Responsibility

Operators should treat the package baseline as the starting point, not the end
state.

Post-installation hardening is expected to include:

- reviewing the deployed baseline on the target host
- validating that enabled controls fit the real hardware and service topology
- running assessment tooling when available
- reviewing findings and selecting remediation actions
- applying safe automatic remediation where available
- completing manual remediation where environment-specific judgment is required

## Planned Assessment Flow

The intended LiteNAS hardening workflow is:

1. Install the LiteNAS package baseline.
2. Run LiteNAS hardening assessment or linting tooling when available.
3. Review findings and distinguish expected deviations, true hardening gaps,
   and environment-specific accepted risk.
4. Remediate issues automatically where LiteNAS can do so safely.
5. Complete any remaining remediation manually.

## Typical Runtime Topics

This document should later cover topics such as:

- host-specific network exposure review
- administrator account and access policy review
- review of integrity-monitoring baselines after installation
- review and tuning of malware scanning scope and schedules
- review and enrollment of `usbguard` policy for the actual hardware set
- validation of backup, recovery, and audit expectations

## Relationship To Other Security Docs

- Use [Preinstalled Hardened Configs](./preinstalled-hardened-configs.md) for
  controls that LiteNAS already installs and configures by default.
- Use [Security Tooling](./security-tooling.md) for a tool inventory and
  responsibilities.
- Use [Security Policy Exceptions](./security-policy-exceptions.md) when a
  benchmark-style recommendation is intentionally not followed.
