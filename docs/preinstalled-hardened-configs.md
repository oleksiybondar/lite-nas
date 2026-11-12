# Preinstalled Hardened Configs

LiteNAS's single Debian package installs the service baseline and depends on
AIDE, ClamAV, USBGuard, rsyslog, AppArmor, UFW, and other runtime packages listed
in `packaging/debian/lite-nas/DEBIAN/control.in`. OpenSCAP is not a package
dependency; operators install and use it separately for compliance assessment.

## Service and Host Baseline

The package deploys dedicated service identities, systemd units, AppArmor
profiles, NATS permissions, and repository-managed configuration. Log Analysis
receives selected rsyslog records through a socket-activated Unix datagram
socket. Network Security evaluates UFW blocked TCP attempt events. The current
SIEM path and its limits are described in [Security
Monitoring](./security-monitoring.md).

## Log Retention

The packaged `/etc/logrotate.d/lite-nas` policy rotates direct
`/var/log/lite-nas/*.log` files daily or at 50 MiB, retains seven rotations,
compresses old files, and uses `copytruncate` for loggers that keep their file
open. Operators should check available storage and retention needs.

## Integrity Monitoring

LiteNAS ships AIDE configuration drop-ins and runs LiteNAS Integrity for
scheduled checks and authorized baseline regeneration. Installation does not
initialize the AIDE database automatically. Operators establish and review a
baseline on the target host. See [Integrity
Requirements](../requirements/litenas-integrity.md).

## Malware and USB Controls

ClamAV is a package dependency. Antivirus scans eligible mounts and accepts
authorized on-demand scans; the deployed updater service supports signature
updates. Operators review scan results and scope. USBGuard is also a package
dependency and the platform includes USB Manager. Operators must review device
policy for required local storage and peripherals. These controls require
host-specific decisions after installation.

See [Post-Installation Hardening](./post-installation-hardening.md) for the
operator workflow and [Security Tooling](./security-tooling.md) for tool roles.
