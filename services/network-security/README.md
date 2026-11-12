# Network Security

Network Security subscribes to validated blocked TCP firewall attempts from
Log Analysis and network snapshot events from Network Metrics. Its compiled
FSM detectors count distinct ports on one destination, distinct destinations
on one port, and repeated attempts to one destination and port. Each defaults
to 12 blocked attempts in 2 minutes with a 30 minute cooldown. A finding
publishes a warning security alert; the latest snapshot timestamp is included
as context when available.

`/etc/lite-nas/network-security.conf` controls each detector's threshold, window,
cooldown, maximum tracked keys, and operator guidance (`mitigate`). Each finding
publishes that guidance in the security alert. Configuration is
read at startup. The in-memory FSM resets on restart. Network snapshots alone
cannot identify scans because they do not contain source-to-port attempts.

Build with `./scripts/build-network-security-binary.sh`. Local deployment uses
`./scripts/deploy-network-security.sh` after shared LiteNAS infrastructure and
the updated NATS authorization configuration are installed. Log Analysis must
also be redeployed to publish the new firewall event.

The service runs as `lite-nas-network-security` with no Linux capabilities and
an enforced AppArmor profile. Service Watchdog monitors and can restart it.
The single Debian package includes its binary, configuration, unit, and profile;
the normal backup specifications include those files and the NATS credentials.
UFW blocked TCP logging, rsyslog `imklog`, Log Analysis, NATS, Auth, and Security
Logging Manager must be running for alerts to reach Security Logging Manager.

For an installed system, run the HyperionTF CLI test with
`pytest -m cli tests/cli/security_monitoring/test_local_port_scan.py`. The test
preparation script installs its privileged helper; the test also needs
`iproute2`. It sends blocked probes from a temporary network namespace and
expects a `secnet` security alert.
