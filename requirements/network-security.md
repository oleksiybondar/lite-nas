# Network Security Requirements

- **network-security/FR-001 — Firewall observations:** Log Analysis forwards
  validated `imklog` UFW `BLOCK` TCP records as typed NATS events containing
  timestamp, canonical source and destination IPs, and destination port.
  Untrusted kernel claims, malformed addresses, and invalid ports are ignored.
- **network-security/FR-002 — Port scan detection:** Network Security detects
  one source probing at least the configured number of distinct blocked TCP
  destination ports on one destination within the configured window. Repeated
  observations of the same port do not advance the threshold. Network metrics
  snapshots provide timestamp context but do not drive scan verdicts.
- **network-security/FR-003 — Cooldown:** One finding is emitted per source and
  destination pair during cooldown. A later burst can create a new finding.
  Detector state is bounded and process local.
- **network-security/FR-004 — Security alert:** A detected scan publishes one
  warning alert through Security Logging Manager with a stable event ID and
  source, destination, distinct port count, and window in the message.
- **network-security/FR-005 — Horizontal scan:** Count distinct blocked TCP
  destinations per source and destination port, then alert at the configured
  threshold with a separate cooldown.
- **network-security/FR-006 — Repeated attempts:** Count blocked TCP attempts
  per source, destination, and port, including repeated attempts to one port.
  Alert at the configured threshold with a separate cooldown. This is a
  connection pressure signal, not proof of an application login attack.
- **network-security/IR-001 — Runtime integration:** Deploy Network Security as
  a dedicated unprivileged systemd user with its AppArmor profile enforced.
  Give it only the NATS subscriptions for firewall attempts and network
  snapshots, and the publish and reply subjects needed for authentication and
  security alerts. Service Watchdog protects and may restart its unit.
- **network-security/IR-002 — Configuration recovery:** Include the service
  configuration, systemd unit, AppArmor profile, NATS authorization, and
  transport and identity certificates in the existing backup specifications.

## Software requirements

The installed system needs systemd, AppArmor, NATS Server with the LiteNAS TLS
and authorization configuration, rsyslog with `imklog` and the LiteNAS forwarding
drop-in, UFW with blocked TCP packet logging, Auth, Security Logging Manager,
Network Metrics, and Log Analysis. The single LiteNAS Debian package installs
the new service and its configuration; its existing package dependencies cover
the external runtime programs. Network Metrics supplies optional timestamp
context; blocked attempts from Log Analysis are required for detection.

The local system test additionally needs Python, HyperionTF, `ip` from
`iproute2`, UFW, and the installed test sudo helper. It creates a temporary
network namespace, scans blocked local ports, and waits for a security alert.

Detection depends on UFW logging blocked probes. Permitted connections and
packets omitted or rate limited by UFW cannot contribute to the threshold.
