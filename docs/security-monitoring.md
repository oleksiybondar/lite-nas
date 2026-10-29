# Security monitoring services

The five services run as separate unprivileged Go processes. Within each process, polling, NATS input, scanning
or evaluation, and reporting run as bounded workers. `shared/go/fsm` supplies typed transition tables and a
bounded keyed memory store; Resources Monitor and LiteNAS Integrity use it for their existing incident
lifecycles. Snapshot collection in metrics services is stateless and was not migrated.

| Service | Input and output | Default |
| --- | --- | --- |
| Log Analysis | Receives rsyslog records on a socket-activated Unix datagram socket; publishes security alert lifecycle and optionally calls `auth.rpc.lockdown.set` | One-shot auth-failure alerts; 3 known-account failures or 3 unknown-account failures in 2 minutes with a 3 minute cooldown; one critical, emailed alert per newly connected USB device; source probing and lockdown disabled |
| Mount Monitor | Reads `/proc/self/mountinfo`; publishes `mount.events.lifecycle` and system alerts | 5 second poll; first snapshot is a baseline |
| Antivirus | Subscribes to mount events and reconciles existing mounts; runs `clamscan`; reports clean scans as normal info events, incomplete scans as warnings, and malware as error events in Security Logging Manager; cancels scans when mounts disappear | One scan at a time; `vfat,exfat,ntfs,ntfs3,ext4` under `/media` or `/mnt`; 30 minute timeout |
| System Service Monitor | Probes configured systemd units; publishes `service.monitor.events.state` and system alert lifecycle | 10 second poll; units in `system-service-monitor.conf` |
| Service Watchdog | Subscribes to service events and independently probes its own protected list; publishes security alert lifecycle and starts allowlisted units | 10 second poll, 30 second recovery delay, 1 minute exponential backoff, at most 3 attempts |

The Watchdog's protected list is independent of the Service Monitor list. It includes
`systemd-journald.service`, AppArmor, USBGuard, UFW, and the LiteNAS security and monitoring units in
`service-watchdog.conf`. It attempts recovery after systemd has left the unit inactive, including a manual stop.
It reports unexpected enablement changes. Its systemd unit is ordered after every protected unit, so systemd
stops Watchdog first during normal shutdown and reboot. Watchdog never changes enablement. Sudo allows only
`systemctl start` of the exact configured default units; changing the config alone cannot widen the privilege
list.

## Log Analysis ingress

rsyslog forwards selected records through `/etc/rsyslog.d/30-lite-nas-log-analysis.conf`. Its `omuxsock` action
sends one JSON object per datagram (`timestamp`, `facility`, `input`, `program`, `message`) to
`/run/lite-nas-log-analysis/rsyslog.sock`. `lite-nas-log-analysis.socket` owns that socket (mode `0620`, group
`adm`, credentials passing on), so records queue in the kernel while the service restarts. rsyslog keeps a
disk-assisted action queue in its work directory while the socket is unavailable. If that queue is full, rsyslog
discards only records for this action; it does not stall host logging. The rsyslogd AppArmor include
`/etc/apparmor.d/rsyslog.d/lite-nas-log-analysis` allows the send.

Log Analysis accepts a datagram only when the kernel-attached sender UID belongs to an account in
`sender_users` (default `syslog,root`; missing accounts are skipped). It rejects unknown JSON fields, trailing
data, NUL bytes, malformed timestamps and records longer than 4 KiB, then logs the rejection. Each `sources`
entry in `rules.json` selects records by `facilities`, `inputs` and `programs`. Every non-empty list must match,
and `format` (`auth` or `kernel`) chooses the compiled field extraction. The shipped `kernel` source selects
`inputs: ["imklog"]` rather than the `kern` facility, because local processes can claim a facility but not the
input module. Authentication-facility records can still be written by any local process through `logger`, as
with the previous file source.

LiteNAS ships only this drop-in, not a full `rsyslog.conf`. It relies on the distribution configuration loading
`imklog`, setting `$WorkDirectory`, and including `/etc/rsyslog.d/*.conf`; Debian and Ubuntu defaults do. The
Log Analysis deployment (`deploy-all.sh`, `deploy-log-analysis.sh`, and the Debian package) installs the drop-in
and the rsyslogd AppArmor include as `root:root`. It then recreates the socket, starts the service, validates the
full rsyslog configuration with `rsyslogd -N1`, reloads the rsyslogd profile when it is loaded, and restarts
rsyslog. Active rules that still use the removed `rsyslog_auth` file source are moved to
`rules.json.pre-rsyslog-socket` and replaced by the packaged default.

The shipped authentication rules, all critical and emailed unless noted:

| Rule | Selects | Instance and IDs |
| --- | --- | --- |
| `pam_*` (existing) | PAM `auth` failures for `litenas-auth`, `sshd`, `login`, `su`, `su-l` and `sudo` | Existing thresholds, `secauth` |
| `sudo_policy_violation` | Classic sudo `user NOT in sudoers`, `user NOT authorized on host`, `command not allowed` | Immediate, `secpriv` |
| `privileged_group_membership` | shadow-utils adds to `sudo`, `admin`, `adm`, `root`, `wheel`, `shadow`, `disk`, `docker`, `lxd` | Immediate, `secpriv` |
| `login_account_created` | `useradd` accounts whose shell is not `nologin` or `false` | Immediate, `secacct` |
| `root_password_changed` | PAM `chauthtok` password change for `root` | Immediate, `secacct` |
| `ssh_preauth_probing` | OpenSSH `Invalid user` and `Failed … for <valid user>` records | 10 per source in 5 minutes, `secssh` |

`sudo` is not part of the PAM success pattern: sudo opens a session for the target account on every allowed
service-account command, which would otherwise reset guessing state. The sudo denial reasons match classic sudo
(Debian). Ubuntu 26.04 defaults to sudo-rs, whose syslog denial format has not been confirmed, so that rule may
not trigger there. The SSH rule counts one record per invalid-user connection and one per failed attempt for an
existing account. Detection messages now include the extracted source address when one is present.

AppArmor denials and kernel taint messages are not shipped as rules yet. LiteNAS profiles currently log routine
denials (for example the Go runtime reading cgroup `cpu.max`), so alerts would be constant until those profiles
are tuned. The kernel prints its taint notice only once per boot, and out-of-tree ZFS modules trigger it at
boot, so a later unsigned module would not be reported.

To process more records, extend the selection condition in the rsyslog drop-in, add a matching source and rules
to the active `rules.json`, then restart rsyslog and Log Analysis. Records that no source selects are ignored.

The `usb_device_connected` rule matches kernel `New USB device found` records, excludes Linux root hubs
(`idVendor=1d6b`), and uses the `usb_device_event` instance with its own `secusb` event ID prefix. The alert text
contains only the USB port and hexadecimal vendor/product IDs; device-reported names and serials are not copied
because the device controls them. Devices that are attached at boot are enumerated again on every boot and
raise an alert each time. To allowlist a known device, add an `exclude` pattern such as
`idVendor=0781, idProduct=5583` to the rule. USBGuard still decides whether the device is authorized.

## Message and worker ownership

All transport callbacks validate size and schema before forwarding into a service owned channel. The Mount
Monitor and Service Monitor own bounded observation and outbound queues; a single outbound worker publishes to
NATS. Antivirus owns a 64 entry mount input, scan job, result, and report queue. Watchdog owns a 32 entry
notification and effect queue plus 16 entry recovery queues. Log Analysis owns a 64 entry record queue, one
bounded input per constructed plugin, and a 64 entry detection queue. Its single rsyslog receiver produces
the record queue, and the runtime closes it after the receiver stops. The dispatch worker sends separate copies
to every matching plugin. Producers stop before their channel is closed, and NATS subscriptions drain before
input closure. Cancellation prevents blocked sends during shutdown. Reporting retries while running, then logs
any pending effect that cannot be delivered at stop.

Both new lifecycle subjects use versioned JSON and have explicit NATS publisher and subscriber ACLs. Antivirus
treats mount messages as hints and verifies the live mount identity before scanning; it also scans eligible
mounts present on startup. Watchdog verifies systemd state and its configured allowlist before every recovery
decision, so missed, duplicate, or stale NATS events do not authorize an action. The current NATS integration
uses core NATS, so mount notifications and queued effects are not durable across a crash. Watchdog's periodic
reconciliation and Antivirus startup reconciliation recover current host state; historical transitions and an
unfinished alert publication may be lost. Correlation keys are stable per rule, mount, or unit to reduce
duplicate manager incidents.

`mount.events.lifecycle` carries version, event ID, action (`mounted`, `unmounted`, `changed`), kernel mount ID,
mount point, source, filesystem type, and observation time. `service.monitor.events.state` carries version,
event ID, unit, condition, systemd `ActiveState`, `SubState`, `Result`, restart count, and observation time.
Neither contract is an authorization signal. Producers use stable identities for a given transition; consumers
validate payloads and read current host state where an action depends on it.

## Auth and email behavior

Auth lockdown state is persisted in `/var/lib/lite-nas-auth/lockdown.json`. While active, new user logins and
refreshes are denied. Service token login and refresh remain available. Log Analysis has no enabled lockdown
rule in the sample configuration. To enable one, an administrator must set `lockdown: true` on a critical rule
and review the impact before restarting Log Analysis. The RPC has a NATS ACL limited to Log Analysis and Auth;
Auth owns enforcement and its state change event.

Security alerts follow the existing Security Logging Manager and Security Email Notifier path; no new SMTP
client is used. A single failed login remains a warning without email. Repeated known-account or unknown-account
failures are critical and set `email_on_critical`, so they send email when the configured cooldown creates a new
incident. Malware detections use error severity and explicitly request email. The disabled critical source-probing
example also sends email if enabled. Publishers that omit the alert contract's optional `email` flag retain the
notifier's previous behavior. Transport or SMTP failure must be checked in the corresponding service logs.
Monitor event IDs follow the resources monitor's readable prefix and eight-digit counter format:
`sysmnt_00000001` for mounts, `syssvc_00000001` for service availability,
`secauth_00000001` for authentication, `secusb_00000001` for USB devices, and `secav_00000001` for antivirus.
Log Analysis uses the configurable `event_id_prefix` in `rules.json` (shipped as `secauth`); an instance may
override it with its own `event_id_prefix` (shipped as `secusb` for USB devices). One counter is shared across
detector instances. Counters use the existing startup-time seed
and are volatile; they do not guarantee uniqueness across restarts. Mount correlation is kept
in memory, so a monitor restart loses the IDs needed to normalize pre-restart mount alerts.
Historical event IDs remain unchanged.

## Installation and operations

`deploy-all.sh`, individual `deploy-<service>.sh` entrypoints, and the single `lite-nas` Debian package install
the binaries, NATS and auth identity certificates, dedicated users, systemd units, AppArmor profiles, and the
exact Watchdog sudoers drop-in. ClamAV and USBGuard are package dependencies because the default deployment
enables Antivirus and protects USBGuard. Service `.conf` files and the live Log Analysis rules preserve
administrator edits across local deploys and Debian package upgrades. The packaged default is
`configs/etc/lite-nas/log-analysis/rules.json`; the active rules are
`/etc/lite-nas/log-analysis/rules.json`. Edit the active JSON file to add or change audit detections, then
restart Log Analysis with `sudo systemctl restart lite-nas-log-analysis.service`. On a fresh install, deployment
copies the packaged default into `/etc`; on later deployments, it leaves the active rules alone. To review a
new packaged default against local rules, compare those two files before deciding whether to merge changes.
No embedded passwords or keys are in sample configs. Mount
Monitor and System Service Monitor have the operator group needed for system alerts; Antivirus has operator and
security groups; Watchdog and Log Analysis have the security group. Log Analysis no longer reads log files and
is not added to `adm`; deployments made before the socket ingress keep that existing membership until it is
removed manually. The Antivirus scanner runs without root and never follows symlinks or crosses
filesystems.

Change service settings under `/etc/lite-nas/`, then restart that unit. For Watchdog policy changes, keep
`/etc/sudoers.d/lite-nas-service-watchdog` aligned with the allowed starts and validate with `visudo -c -f`
before restart. Review `journalctl -u lite-nas-<service>` and `/var/log/lite-nas/<service>.log` for queue,
probe, scan, and delivery failures. `systemctl is-active` and `systemctl is-enabled` show current state; service
state events are observations rather than proof of a human action. Rollback uses the previous single package and
preserves `/etc/lite-nas` and `/var/lib/lite-nas-auth`; uninstall follows the existing package removal policy
and does not delete shared logs or identities.

## Known delivery limits

The FSM counters and cooldowns, monitor incident state, and outbound queues are volatile. Process restart
rebuilds current host state where possible, but it cannot guarantee a pending manager effect. Records already
read from the socket but not yet dispatched at shutdown are lost; records still queued in the socket or in
rsyslog's action queue are delivered to the next process. The Auth lockdown state and its pending state-change
event are persisted separately. Full
lossless event delivery would require a durable transport or outbox and is not provided by the current core NATS
infrastructure.
