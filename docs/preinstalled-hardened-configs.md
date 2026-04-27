# Preinstalled Hardened Configs

## Purpose

This document records the security baseline that LiteNAS installs and configures
as part of the Debian package. It should describe controls that are expected to
be present immediately after installation without requiring separate manual
setup by the operator.

## Installed Security Packages

The LiteNAS package baseline is expected to include selected security tools and
supporting runtime dependencies.

Current planned baseline areas:

- integrity monitoring
- malware scanning
- USB device control
- log retention management
- firewall baseline

Planned preinstalled tools include:

- `aide`
- `clamav`
- `usbguard`
- `logrotate`

This section should later record:

- whether the tool is installed as a hard dependency or recommended dependency
- what LiteNAS config files are deployed for it
- whether the service or timer is enabled by default

## Preconfigured Service Baseline

LiteNAS-managed services and apps should ship with repository-managed runtime
configuration and service integration defaults that support a safer dedicated
host profile.

Examples of baseline areas to track here:

- service users and groups
- file ownership and permissions
- systemd unit hardening directives
- runtime directory ownership
- default network exposure

## Logging And Retention Baseline

LiteNAS should ship managed log rotation defaults for LiteNAS-owned services
and apps.

Current intent:

- ship preconfigured `logrotate` entries with the package
- keep retention intentionally short to reduce write pressure on Raspberry Pi
  SD-card-based environments
- prefer a conservative default such as 7 days of rotation for LiteNAS-managed
  logs unless a specific service requires a different policy

This section should later list each shipped log target and its retention rule.

## Integrity Monitoring Baseline

LiteNAS is expected to preinstall and preconfigure filesystem integrity
monitoring.

This section should later define:

- what `aide` configuration LiteNAS provides
- what paths are included or excluded by default
- when database initialization is expected to happen
- what operator review remains necessary

## Malware Scanning Baseline

LiteNAS is expected to preinstall and preconfigure malware scanning components.

This section should later define:

- what `clamav` packages and services are installed
- whether signatures are updated automatically
- what scan defaults are enabled
- what scan scheduling or review actions remain with the operator

## USB Device Control Baseline

LiteNAS is expected to preinstall and preconfigure USB device control suitable
for a NAS appliance profile.

This section should later define:

- what `usbguard` baseline policy is shipped
- what LiteNAS enables automatically
- what operator review is required for local hardware and peripherals
- how LiteNAS handles deployment scenarios that depend on USB-attached storage

## Deferred Items

Some hardening controls may be intentionally deferred to post-installation work
when they depend on the physical environment, network model, administrator
workflow, or attached hardware.

Those items should be tracked in
[Post-Installation Hardening](./post-installation-hardening.md) instead of
being described here as if they were fully automated.
