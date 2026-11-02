# Security Tooling

LiteNAS ships host controls and runtime security monitoring. Compliance review
uses OpenSCAP with a profile selected by the operator. OpenSCAP is not a LiteNAS
package dependency and LiteNAS does not ship its own assessment or remediation
engine.

| Tool or service | Current role | Operator action |
| --- | --- | --- |
| AIDE and LiteNAS Integrity | Package dependency, LiteNAS rules, scheduled integrity checks, and authorized baseline regeneration | Initialize and review the baseline after installation and planned changes |
| ClamAV and Antivirus | Package dependency and scan service for eligible mounts and authorized on-demand scans | Review findings and scan scope |
| USBGuard and USB Manager | Package dependency, device policy enforcement, and managed operations | Review and enroll the host's required devices |
| `logrotate` | Packaged rotation policy for LiteNAS logs | Review retention against available storage |
| rsyslog and Log Analysis | Selected host records forwarded to a validated socket and evaluated by configured detection plugins | Review active rules after package updates |
| Network Security | Correlates validated blocked TCP attempts from Log Analysis | Keep UFW blocked-packet logging enabled for scan detection |
| Security Logging Manager and Security Email Notifier | Persist security alert lifecycles and deliver selected alerts | Review incidents and configure mail delivery |
| OpenSCAP | External compliance assessment using selected benchmark content | Install separately, choose a suitable profile, review results, and apply host-specific changes |

AIDE checks file integrity against its database; it does not determine whether a
host passes a compliance profile. LiteNAS's AIDE rules cover system and service
paths while excluding mutable and virtual data. Regenerate the baseline after
changing the rules or installing software. See [Preinstalled Hardened
Configs](./preinstalled-hardened-configs.md) and [Security
Monitoring](./security-monitoring.md) for deployed behavior.

The active Log Analysis rules are under `/etc/lite-nas/log-analysis/rules.json`.
The package ships a default file but preserves local changes on upgrade. Compare
the packaged and active rules before adopting new detections. More rules and
compiled plugin types can be added in later releases; only the documented
current set is deployed today.
