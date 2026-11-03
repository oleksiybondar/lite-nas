# Security Automation action and FSM reference

**Status:** Operator-requested and verified Web/API direct HTTPS UFW blocks, automatic 30-minute source-wide blocks
after horizontal then vertical scans, account-and-source web-session revocation, account-and-TTY loginctl termination,
repeated USB connection blocking, and correlated security alerts are implemented. Host effects still need
installed-system verification. SSH and generic network alerts do not prove direct HTTPS ingress.

## Startup rule schema

The service reads `/etc/lite-nas/security-automation.conf`. Its `[automation]` section requires `mode=strict` or
`mode=permissive` and a clean absolute `rules_file` path. The packaged file is
`/etc/lite-nas/security-automation/rules.json`. Sixteen packaged detectors are enabled at startup;
`integrity_service_disruption` remains visible but disabled because AIDE alerts do not identify a service unit. The
horizontal-then-vertical scan rule is the only network rule with an automatic host effect. Deployment merges only the
reviewed new and older action-rule shapes into installed rule files, preserves local settings, and leaves a
`.pre-action-enrollment` backup on the host.

The JSON top level contains `rules`. Each rule has `id`, `name`, `description`, an optional `explanation` (up to 1024
characters describing detection, intended protection, and effect availability), an explicit boolean `enabled`, `fsm`,
`correlate_by`, `steps`, and `action`. Omitting `enabled` fails startup. `fsm` has `type` (`sequence` or
`count_within_window`), `window`, `max_keys`, `max_idle`, and `threshold` for counts. A step has an event `type`
(`security` or `system`), `id_prefix`, optional `category`, `source`, and `severity`, and exactly one `message` matcher:
`equals`, `contains`, or `regex`. A regex must contain a named capture matching `correlate_by` unless that name selects
a validated alert field (`event_id`, `source`, or `category`) or one of the validated combined keys (`account_tty`,
`device_identity`). `action` has `kind` (`alert_only`, `alert`, `builtin`, or `command_profile`). `alert` requires
`severity` and `email` and has no command name or operands.

Unknown JSON fields, duplicate rule IDs, malformed regular expressions, missing correlation captures, and invalid FSM
parameters fail startup. Compiled plugin types are linked into the binary; only rules configured `enabled=true` receive
initialized FSM instances. The `enabled` value is the persistent startup switch: edit the administrator-owned rules file
and restart the service to change it. The UI, CLI, and RPC never write this file; it is covered by AIDE integrity
checks. Their enable/disable override changes only the in-memory effective state and lasts until restart. An RPC
override can only affect an instance that was enabled at startup. A rule with `enabled=false` remains visible with
`configured_enabled=false` and cannot be enabled by RPC. The enrolled effects still need testing against real events on
the installed system.

## Action kinds

An `alert` action publishes a security alert through the existing Logging Manager subject with category
`automation_correlation`, configured severity and email policy, the rule and correlation key, and contributing source
alert IDs. Its event ID is stable for the rule and ordered source IDs across retries. A failed publication is retried
while the worker runs and logged if still undelivered at shutdown. Security Automation ignores its own alerts on input.
The published alert does not claim a host mitigation or offer rollback. `alert_only` remains a service-log outcome with
no published alert.

A `builtin` names a service-owned operation. A `command_profile` names a fixed, reviewed executable, verb, and argument
layout. Rules may select a configured profile but cannot provide a command line, executable path, option, or arbitrary
argument. The runner passes an argument vector directly to the executable, without a shell. A profile is rejected at
startup unless its apply, verification, and restore semantics are defined; explicitly one-way profiles require
`non_restorable` policy.

- **`block_source_ip`**
  - **Operand:** `:requester_ip_address`
  - **Apply and restore:** Insert an identifiable service-owned UFW deny for inbound TCP 443 on verified physical
    interfaces, ahead of any HTTPS allow rule. Expire it after 30 minutes; `restore_source_ip` removes only the recorded
    rule after verification. Cloudflare Tunnel and Tailscale traffic are outside this rule's scope.

- **`block_scan_source_ip`**
  - **Operand:** `:requester_ip_address`
  - **Apply and restore:** After a horizontal scan followed by a vertical port scan from the same source, deny all
    inbound traffic from that IP on verified physical interfaces for 30 minutes. Restore removes only the recorded
    service-owned rules. Tunnel and Tailscale interfaces are outside this scope.

- **`restore_source_ip`**
  - **Operand:** Recorded rule identity
  - **Apply and restore:** Restore operation for `block_source_ip`; it is selected from the persisted incident, never
    from an alert.

- **`revoke_web_sessions`**
  - **Operand:** Verified account and direct HTTPS client IP
  - **Apply and restore:** Ask Auth to revoke only active Admin Panel sessions matching both; this one-way effect has no
    restore command.

The following profiles describe the host commands behind the enrolled Session Manager and USB Manager RPCs. Security
Automation calls those managers with its service identity; it does not execute their host commands itself.

- **`usb_block_device`**
  - **Fixed command form:** `usbguard block-device --permanent :device_id` via USB Manager
  - **Dynamic operand:** One current allowed inventory entry with matching port, vendor/product, and recorded hash
  - **Restore policy:** Keep blocked until manual rollback verifies the same device and restores its recorded allow
    policy.

- **`login_terminate_session`**
  - **Fixed command form:** `loginctl terminate-session -- :session_id` via Session Manager
  - **Dynamic operand:** One active session with exact account and TTY
  - **Restore policy:** One-way; the service records the result and offers no rollback.

`usbguard allow-device` and UFW deletion are inverse operations only. They are selected from a persisted applied-effect
record, not directly from an incoming alert. Broad SSH lockdown is excluded until VPN and interface guards have an exact
contract. There is no account restriction, account/group reversal, or Watchdog retry-suppression command in the current
toolset; rules requiring one remain alert-only until an owning service defines it.

## Named arguments

The colon spelling is a whole argument placeholder, not text interpolation inside another argument. Only a profile may
use a placeholder, and each placeholder is accepted only at a reviewed position for its program. The service validates
the source observation, provenance, syntax, current host identity where applicable, and the program's operand rules
before execution. Missing or ambiguous values cause a skipped effect with an observable reason. Option-shaped values are
rejected even when an end-of-options delimiter is present.

- **`:requester_ip_address`**
  - **Source and validation:** Validated alert field or named regex capture; canonical, unicast IPv4/IPv6 address
  - **Allowed profiles:** `block_source_ip`, `revoke_web_sessions`

- **`:username`**
  - **Source and validation:** Log Analysis account captured from the verified Web/API or sudo alert
  - **Allowed profiles:** `revoke_web_sessions`, `login_terminate_session`

- **`:session_id`**
  - **Source and validation:** Unique Session Manager inventory ID with exact account and TTY
  - **Allowed profiles:** `login_terminate_session`

- **`:device_id`**
  - **Source and validation:** USB Manager inventory ID resolved from the alert port and vendor/product, then tied to a
    recorded hash
  - **Allowed profiles:** `usb_block_device`

No profile accepts an unbounded free-text message, filesystem path, group name, shell fragment, or operator-supplied
extra argument.

## FSM plugin names

These are **Security Automation** plugin names to implement using `shared/go/fsm`. They are distinct from Log Analysis's
existing `immediate`, `threshold`, and `keyed_threshold` plugins, which evaluate raw log observations before publishing
alerts.

- **`sequence`**
  - **Input and completion:** Ordered alert steps; complete only when all steps match within the configured sequence
    deadline
  - **Per-key state:** Step index, first-match time, deadline, contributing alert IDs

- **`count_within_window`**
  - **Input and completion:** Count matching validated alert creations for one key; complete at the configured threshold
    inside the window
  - **Per-key state:** Count, window deadline, contributing alert IDs

Both plugins use a separate `(rule_id, correlation_key)` instance. Unrelated alerts leave an instance unchanged. An
out-of-order alert does not advance `sequence`; after its first accepted step, another first-step match cannot restart
the sequence before its deadline. This also lets two alerts satisfy successive steps when their selectors overlap. A
deadline resets incomplete detection state without applying an effect. Cooldown and mitigation expiry are separate from
the FSM window. Duplicate alert IDs must not advance either plugin twice. No action executes from an intermediate
transition.

## Enrollment

The packaged rules and their action assignments are listed in the [feature
design](../requirements/security-automation.md). An installed host test must confirm each effect and restore path before
production use. Alert-only rules remain observable and cannot claim mitigation success.

## Worker pipeline

The service uses `shared/workers` for independently running loops. NATS callbacks validate and enqueue typed
observations; a dispatcher owns the compiled FSM instances and sends immutable completed matches to an outbound worker.
The outbound worker persists an incident intent before invoking an effect and then publishes the alert lifecycle. A
timer worker sends deadline ticks to the dispatcher and restoration intents to the outbound worker. Operator RPC
requests enter the same instance owner so they serialize with automatic work.

- **Observations**
  - **Producer and owner:** NATS subscription adapter
  - **Consumer:** Dispatcher
  - **Capacity and backpressure:** Bounded; send waits with callback and service cancellation
  - **Shutdown:** Drain NATS, then adapter closes after callbacks stop

- **Completed matches**
  - **Producer and owner:** Dispatcher
  - **Consumer:** Outbound effect worker
  - **Capacity and backpressure:** Bounded; send waits with cancellation and reports any unsent intent
  - **Shutdown:** Dispatcher closes after observations drain

- **Timer ticks**
  - **Producer and owner:** Shared polling timer worker
  - **Consumer:** Dispatcher
  - **Capacity and backpressure:** Bounded as defined by `shared/workers`; ticks carry no security observation
  - **Shutdown:** Timer stops on service cancellation

- **Operator commands**
  - **Producer and owner:** RPC adapter
  - **Consumer:** Instance owner
  - **Capacity and backpressure:** Bounded; request waits or returns a visible timeout
  - **Shutdown:** Drain NATS before closing request queue

- **Restore intents**
  - **Producer and owner:** Timer and instance owner, via the owning coordinator
  - **Consumer:** Outbound effect worker
  - **Capacity and backpressure:** Bounded; no silent drop
  - **Shutdown:** Coordinator stops producers before outbound worker drains

One channel with multiple readers is a work queue. A finding needed by both incident storage and reporting must be
explicitly copied or handled in one owning worker; neither consumer may assume broadcast. The NATS adapter, listener,
dispatcher, and outbound effect worker are implemented. UFW and USB effects have durable incident records and verified
restore paths. Auth and loginctl session actions are one-way, journaled until their manager request and security alert
complete. A failed effect remains visible for review or bounded retry.

## Operator API inventory

The NATS RPCs use the `security-automation.rpc.*` subjects and carry an access token. The gateway exposes them under
`/security/automations` to the security role. The CLI prompts for user credentials through the shared Auth Service flow
and calls the same RPCs. Operator `force-mitigate` accepts the scoped UFW block. USB and session effects run from
verified alerts. Cooldown reset clears only the selected rule's volatile detection state for one correlation key; it
does not change active incidents or restore effects.

- **List all rules and configured/effective enablement**
  - **NATS subject suffix:** `rules.list`
  - **Gateway path:** `GET /rules`
  - **Intended CLI command:** `rules`

- **List active or pending incidents and their mitigation state**
  - **NATS subject suffix:** `active.list`
  - **Gateway path:** `GET /active`
  - **Intended CLI command:** `active`

- **Restore one service-owned effect**
  - **NATS subject suffix:** `rollback`
  - **Gateway path:** `POST /active/{incident_id}/rollback`
  - **Intended CLI command:** `rollback <incident-id>`

- **Reconcile, restore if needed, then normalize only after verification**
  - **NATS subject suffix:** `force-normalize`
  - **Gateway path:** `POST /active/{incident_id}/force-normalize`
  - **Intended CLI command:** `force-normalize <incident-id>`

- **Trigger a configured rule on demand with all required named operands**
  - **NATS subject suffix:** `force-mitigate`
  - **Gateway path:** `POST /rules/{rule_id}/mitigate`
  - **Intended CLI command:** `force-mitigate RULE name=value... --yes`

- **Clear one keyed detector's in-memory state**
  - **NATS subject suffix:** `cooldown.reset`
  - **Gateway path:** `POST /rules/{rule_id}/cooldown/reset`
  - **Intended CLI command:** `cooldown-reset <rule-id> <key>`

- **Temporarily enable or disable one rule**
  - **NATS subject suffix:** `rule.set-enabled`
  - **Gateway path:** `PUT /rules/{rule_id}/enabled`
  - **Intended CLI command:** `enable/disable <rule-id>`

Rule responses expose `state` (`normal`, `pending`, `mitigating`, `restoring`, or `failed`), `enabled` as the effective
in-memory state, and `configured_enabled` as the startup default. Incident responses expose their own mitigation state
and last outcome. For an owned UFW incident, `rollback` and `force-normalize` queue the same verified restore path; the
incident record is removed only after the helper verifies the rule is absent and the security incident alert is
normalized. Operator requests publish separate audit alerts. The CLI uses the same NATS RPCs as the web interface; it
does not write Logging Manager state directly.

`force-mitigate` bypasses only the detection FSM. The operator selects a rule row, supplies every `required_operands`
entry through a dialog, and types `yes` to confirm. For `block_source_ip`, the service validates one canonical source
IP, discovers verified physical interfaces, persists a pending incident, and queues the UFW effect. The response reports
it as pending; the effect worker verifies each installed rule, publishes an incident security alert, and later restores
the block and normalizes that alert. Unknown, missing, or extra operands fail before any effect. An operator-supplied IP
authorizes only a direct HTTPS block; it does not establish provenance for automatic blocking.

The CLI accepts `force-mitigate RULE [name=value ...] --yes`. After authentication it fetches the rule list, prompts on
the terminal for each missing `required_operands` value, and rejects unknown operand names. Operators can supply all
values on the command line for noninteractive use after login; `--yes` remains required. The service independently
validates the canonical source IP and physical-interface set before queuing a UFW effect.

## Trace and control host effects with the system CLI

Use the host commands below to inspect the actual state behind an automation rule. A raw command is a manual host
change, not an automation incident. In particular, `enabled=true` in the rules API means the detector is enabled;
`action_name` alone does not confirm that an effect ran. Check `active` and the host inventory for the effect result. A
`401` from the HTTP API means that request lacked a valid access token; an authenticated request can still fail when the
effect handler rejects it.

Start with the rule ID and time of the request. Check the gateway and automation logs for the exact RPC error, then
inspect the relevant host state. The service's RFC5424 file timestamps use UTC; `journalctl` normally displays local
time.

```bash
sudo journalctl -u lite-nas-web-gateway -u lite-nas-security-automation -u nats-server --since '30 minutes ago' --no-pager
sudo tail -n 100 /var/log/lite-nas/web-gateway.log /var/log/lite-nas/security-automation.log
```

Use `security-automation-cli rules` to see detector enablement and `security-automation-cli active` to see recorded host
effects. `security-automation-cli disable RULE_ID` stops new matches until the service restarts; `enable RULE_ID`
resumes a startup-enabled detector. Disabling an IP-block rule queues restoration of its active UFW incidents. A USB
block stays in place when its detector is disabled, as requested; use `security-automation-cli rollback INCIDENT_ID` to
restore that exact recorded device. Session termination and web-session revocation have no inverse. For a persistent
switch, edit the installed `rules.json` and restart Security Automation.

The following commands apply and reverse **manual** changes. Replace example IP addresses, interface names, device IDs,
session IDs, and unit names with values verified against the current host. Record the original state before changing it.
The reverse command is appropriate only when it restores that original state; do not remove a rule or undo a policy
owned by another service or operator. Use `ip -br link` and `ip -j route show default` to inspect candidate physical
interfaces, then confirm their type and the inbound HTTPS path. Do not use a Tailscale or tunnel interface, or the
default route alone, as proof of a physical interface.

- **`block_source_ip`**
  - **Inspect current state:** `ip -br link` and `sudo ufw status numbered`
  - **Apply manually:** `sudo ufw insert 1 deny in on PHYSICAL_INTERFACE proto tcp from 198.51.100.10 to any port 443
    comment 'manual automation investigation'` (replace `PHYSICAL_INTERFACE` with a verified physical ingress interface)
  - **Restore or disable the manual effect:** Confirm the exact comment and scope, then run `sudo ufw delete deny in on
    PHYSICAL_INTERFACE proto tcp from 198.51.100.10 to any port 443 comment 'manual automation investigation'`. Verify
    its absence with `sudo ufw status numbered`.

- **`block_scan_source_ip`**
  - **Inspect current state:** `ip -br link` and `sudo ufw status numbered`
  - **Apply manually:** `sudo ufw insert 1 deny in on PHYSICAL_INTERFACE from 198.51.100.10 to any comment
    'manual-scan-block'` on each verified physical ingress interface
  - **Restore or disable the manual effect:** Confirm the exact manual comment, interface, and source, then run `sudo
    ufw delete deny in on PHYSICAL_INTERFACE from 198.51.100.10 to any comment 'manual-scan-block'` on each interface.
    Verify absence with `sudo ufw status numbered`.

- **`usb_block_device`**
  - **Inspect current state:** `sudo usbguard list-devices` and `sudo usbguard list-rules`
  - **Apply manually:** `sudo usbguard block-device --permanent DEVICE_ID`
  - **Restore or disable the manual effect:** If the recorded prior policy was allow, use `sudo usbguard allow-device
    --permanent DEVICE_ID` only after confirming the same port, vendor/product, and hash.

- **`login_terminate_session`**
  - **Inspect current state:** `loginctl list-sessions` and `loginctl show-session SESSION_ID`
  - **Apply manually:** `sudo loginctl terminate-session -- SESSION_ID`
  - **Restore or disable the manual effect:** No inverse exists. Terminating a session cannot be undone by an enable
    command.

- **Manual systemd operation; no automation rule currently stops a unit**
  - **Inspect current state:** `systemctl is-active UNIT.service` and `systemctl status UNIT.service`
  - **Apply manually:** `sudo systemctl --system stop -- UNIT.service`
  - **Restore or disable the manual effect:** If the unit was active before the stop and no other owner has since
    changed it, use `sudo systemctl --system start -- UNIT.service`. Check Service Watchdog ownership first.

`ufw status numbered`, `usbguard list-devices`, `loginctl list-sessions`, and `systemctl status` are also the direct
verification commands after an apply or restore. A successful command exit alone does not establish the final state.
These raw commands bypass automation's authentication, incident record, expiry, and audit lifecycle. Once an effect is
owned by automation, use its recorded incident and rollback path so its state stays consistent with the host.

LiteNAS web sessions are separate from `loginctl` sessions. An administrator can inspect them through `GET
/api/auth/sessions` and revoke a selected session through `POST /api/auth/sessions/{session_id}/revoke`. Nginx marks
direct HTTPS requests and overwrites the peer and local-listener address headers. The gateway stores a browser client IP
only when that local listener belongs to a physical interface. Auth's service-scoped `auth.rpc.session.revoke-matched`
RPC revokes only sessions with both the Log Analysis account and that client IP; Tunnel and Tailscale sessions do not
supply this evidence. Sudo alerts include account and TTY, which are matched against exactly one Session Manager
inventory entry before termination.

The UFW helper limits an operator-requested or verified Web/API source block to 30 minutes on inbound TCP 443 on
verified physical interfaces. The horizontal-then-vertical scan rule uses a separate 30-minute deny for all inbound
traffic from the matched source on those interfaces. Both insert before existing allows and verify the numbered rule and
incident comment. Tunnel and Tailscale interfaces are outside these rules. Two distinct USB connection alerts for the
same port and vendor/product within ten minutes can permanently block one currently allowed USBGuard device; use
`security-automation-cli active` and `rollback INCIDENT_ID` for manual restoration. Loginctl termination and web-session
revocation are separate one-way rules. `integrity_service_disruption` has no host action and remains disabled because
its input cannot be correlated to a verified unit.
