# Security Monitoring Requirements

This document records the implemented security-monitoring contracts. Service
implementations and tests are the evidence for the current behavior; the old
implementation task is not a source of outstanding work.

- **security-monitoring/FR-001 — Log analysis:** Receive rsyslog records on a
  socket-activated Unix datagram socket, accept only configured rsyslog sender
  identities, validate each record, select it by configured source, apply
  compiled detection instances and thresholds, and publish security alert
  creation, occurrences, and resolution. The Auth service enforces authorized
  lockdown requests.
- **security-monitoring/FR-002 — Mount lifecycle:** Detect mount changes from
  mountinfo snapshots, publish typed lifecycle events, and create or normalize
  informational system alerts using readable `sysmnt` counter IDs.
- **security-monitoring/FR-003 — Antivirus:** Consume validated mount events,
  scan eligible mounts with ClamAV, and publish security alerts for detections
  using readable `secav` counter IDs.
- **security-monitoring/FR-004 — Service monitoring:** Observe configured
  systemd units, publish typed state changes, and report unavailable units to
  System Logging Manager using readable `syssvc` counter IDs.
- **security-monitoring/FR-005 — Service watchdog:** Consume validated state
  events, evaluate critical-unit policy, and report security incidents for
  recovery failures or unexpected disablement.
- **security-monitoring/FR-006 — USB device connection:** Receive kernel USB
  enumeration records forwarded by rsyslog, ignore USB root hubs, and record a
  critical security alert with the USB port and vendor/product IDs for every
  newly connected device using readable `secusb` counter IDs.
- **security-monitoring/FR-007 — Privilege and account changes:** Count sudo
  password failures with the existing PAM rules, and record critical alerts for
  sudo policy denials and privileged group membership (`secpriv`), new accounts
  with a login shell and root password changes (`secacct`), and repeated SSH
  pre-authentication failures from one source address (`secssh`).
- **security-monitoring/FR-008 — On-demand antivirus:** Accept an authorized
  administrator scan request for any existing folder on a currently mounted
  filesystem, expose task status and retry, and report completed results
  through the same security-alert path as scheduled mount scans. Scheduled
  scans remain limited to configured eligible mounts.
- **security-monitoring/IR-001 — Detection state machines:** Use the shared
  table-driven FSM implementation for configured detection lifecycles.
- **security-monitoring/IR-002 — Lifecycle integration:** Producers retain an
  allocated alert ID through retries, occurrences, and resolution within the
  active process. Volatile state does not guarantee cross-restart correlation.
- **security-monitoring/IR-003 — On-demand scan API:** Expose authenticated
  start, status, and retry gateway endpoints backed by antivirus-specific
  request/reply subjects and the shared background-task contracts.

The producer-owned event ID format is `<domain>_<eight decimal digits>`;
`CONTRIBUTING.md` defines the repository-wide rule. These requirements describe
implemented behavior, not a roadmap or an assertion that every integration has
been exercised on an installed system.
