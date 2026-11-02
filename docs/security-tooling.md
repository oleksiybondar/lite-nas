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

## Release OpenSCAP comparison

The main release workflow runs a paired assessment on a fresh GitHub-hosted
Ubuntu 24.04 amd64 VM. It evaluates the ComplianceAsCode Ubuntu 24.04 CIS Level
2 Server profile before and after installing the validated LiteNAS Debian
package. Neither scan applies remediation. Package system tests run in a
separate job after the same package build, so their setup cannot affect the
comparison.

The `openscap-release-delta-amd64` workflow artifact is retained for 30 days.
It contains both OpenSCAP XCCDF results, ARF results, and HTML reports; the
benchmark data stream and SHA-512 download digest; installation and scan logs;
and `delta.json` plus `delta.html`. The delta report groups rules by the
benchmark's top-level categories and counts pass/fail improvements, regressions,
remaining failures, and outcomes that need review. The benchmark SHA-256,
profile, content version, package SHA-256, and runner image version identify the
inputs. Benchmark IEC 62443 references are shown where present, but the CIS
profile and this VM assessment do not establish an IEC 62443 security level or
certify a deployed host. Rules that need hardware, a particular environment,
or manual evidence may remain unassessed.

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
