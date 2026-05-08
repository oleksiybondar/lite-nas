# Security Tooling

## Purpose

This document tracks the security-related tools that LiteNAS installs, configures,
or plans to provide as part of its hardening model.

It should answer:

- why the tool exists in the LiteNAS baseline
- whether it is installed by default
- what LiteNAS configures automatically
- what operator action is still required
- whether remediation is automatic, manual, or mixed

## Tool Inventory

| Tool | Purpose | Installed By Default | Preconfigured By LiteNAS | Primary Phase | Operator Follow-Up |
| --- | --- | --- | --- | --- | --- |
| `aide` | Filesystem integrity monitoring | Planned | Planned | Baseline and runtime review | Review reports and handle findings manually |
| `clamav` | Malware scanning | Planned | Planned | Baseline and runtime review | Review findings and tune scans as needed |
| `usbguard` | USB device control baseline | Planned | Planned | Baseline and runtime review | Review and adapt policy for attached hardware |
| `logrotate` | Managed log retention for LiteNAS services and apps | Planned | Planned | Baseline | Validate retention policy fits storage constraints |
| LiteNAS hardening linter | Host hardening assessment | Planned | N/A | Post-installation | Review findings and decide remediation |
| LiteNAS remediation tooling | Assisted or automatic remediation | Planned | N/A | Post-installation | Confirm proposed changes before or after execution |

## Tool Roles

### `aide`

`aide` is intended to provide filesystem integrity monitoring for important
system and LiteNAS-managed paths. LiteNAS should define a baseline config that
fits the appliance profile without pretending that all deployments share the
same acceptable file-change patterns.

### `clamav`

`clamav` is intended to provide a malware scanning capability within the
preinstalled baseline. LiteNAS should document whether signature updates, scan
scheduling, and scan targets are fully configured by default or require
additional operator decisions.

### `usbguard`

`usbguard` is intended to provide a USB device control baseline. Because
LiteNAS may run on Raspberry Pi systems with USB-attached storage, the baseline
must distinguish between generic removable-device restrictions and the actual
device model required by a NAS host.

### `logrotate`

`logrotate` is intended to bound retained log volume and reduce uncontrolled
growth of LiteNAS-managed logs. The default policy should reflect storage
constraints typical of Raspberry Pi deployments, including SD-card wear
considerations.

## Future LiteNAS Tooling

LiteNAS is expected to add its own host assessment and remediation tooling.

That future tooling should:

- check the applied baseline against LiteNAS policy
- identify missing or drifted controls
- separate findings from intentional policy exceptions
- support safe automatic remediation where the change is deterministic
- leave environment-sensitive decisions to explicit operator review
