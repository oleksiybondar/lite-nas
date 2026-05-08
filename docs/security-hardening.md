# Security Hardening

## Purpose

LiteNAS is intended to deploy a security-focused host baseline for a dedicated
NAS environment. That baseline reduces common risk by installing selected
security packages, deploying repository-managed configuration, and enabling
default controls that fit the LiteNAS operating model.

This document explains the overall hardening model and how the related
documentation is organized.

## Scope

LiteNAS hardening is split into three layers:

1. Preinstalled and preconfigured hardening shipped by the LiteNAS Debian
   package.
2. Post-installation hardening work that depends on the target host,
   environment, and operational requirements.
3. Intentional policy exceptions where a generic benchmark recommendation does
   not fit the LiteNAS Raspberry Pi and NAS deployment model.

LiteNAS does not claim that package installation alone produces a fully
hardened host for every environment. Real deployments still require operator
review, host-specific decisions, and ongoing assessment.

## Baseline Model

The LiteNAS Debian package is expected to apply a host profile that includes:

- installation of selected security-related packages
- deployment of pre-hardened LiteNAS-managed configs
- deployment of service and system integration files
- enablement of baseline controls that are safe for the intended platform

This baseline is designed to move the host closer to a secure dedicated NAS
profile immediately after installation, while still allowing later
environment-specific refinement.

## Runtime Hardening Model

After installation, operators are expected to:

- review the resulting security posture on the target host
- run LiteNAS-provided assessment or linting tools when available
- review findings and classify them
- remediate findings automatically where supported
- complete manual remediation where automatic changes are not safe or are too
  environment-specific

## Policy Exception Model

LiteNAS may intentionally differ from generic CIS, NIST, or similar benchmark
recommendations when those recommendations conflict with the practical needs of
the target platform.

These exceptions must be documented explicitly rather than left implicit. Each
exception should capture:

- the upstream recommendation or control intent
- the LiteNAS decision
- the operational reason for that decision
- the risk introduced by the deviation
- the compensating controls or operator expectations

## Related Documents

- [Preinstalled Hardened Configs](./preinstalled-hardened-configs.md)
- [Post-Installation Hardening](./post-installation-hardening.md)
- [Security Tooling](./security-tooling.md)
- [Security Policy Exceptions](./security-policy-exceptions.md)
