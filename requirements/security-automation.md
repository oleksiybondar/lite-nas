# Security Automation Feature Design Specification

**Status:** Verified Web/API direct HTTPS UFW blocking, account-and-source Auth revocation, account-and-TTY loginctl
termination, and repeated USB connection blocking are implemented alongside operator-requested UFW mitigation.
Installed-system effect and rollback tests are still required.

## Purpose and scope

Security Automation correlates existing security and system alerts, decides when a configured threat sequence has
completed, applies an approved mitigation, and restores normal operation when the threat has passed. It records each
decision and effect so an operator can understand and reverse it. Detection and mitigation are separate phases: no host
or service change occurs between steps of an incomplete sequence.

The service is one unprivileged Go process with goroutine workers. It uses the shared table-driven FSM package for rule
state and `shared/go/messaging` for communication with other processes. It does not replace Log Analysis, Network
Security, either Logging Manager, Auth, or the email notifier. The intended release includes every rule discussed in
this specification. Some require upstream alert enrichment or new observations before they can operate correctly; that
work is part of the release scope. Do not claim those rules work with today's alert payloads.

## Terms

- **Alert**
  - **Meaning:** A security or system event received from the respective Logging Manager path.

- **Rule**
  - **Meaning:** Configured sequence or threshold, its correlation key, timing, mitigation, and restoration policy.

- **Rule instance**
  - **Meaning:** One rule tracking one entity, such as a source IP or account. It owns its own FSM state and deadlines.

- **Incident**
  - **Meaning:** A completed rule match and its alert lifecycle, including mitigation and restoration.

- **Mitigation**
  - **Meaning:** An approved action taken after a complete match.

- **Restore**
  - **Meaning:** The operation that reverses a mitigation owned by this service.

- **Cooldown**
  - **Meaning:** A rule-specific period controlling repeated matches and state reset. It is distinct from a block's
    lifetime.

## Required behavior

1. Subscribe to security and system alert creation events. Define explicitly whether occurrences and state updates are
   also inputs; their current payloads differ from create events. Validate each NATS message before matching it.
2. Select events by alert type (`security` or `system`), event ID prefix, and message `equals`, `contains`, or `regex`.
   Rules may also constrain category, severity, or source if supported by the final schema. Compile regular expressions
   and all configured plugin instances at startup.
3. Let each rule select a compiled FSM plugin and define as many sequence steps as that plugin supports. Support ordered
   steps and count-within-window rules. Bound a sequence by an explicit deadline; define how unrelated and out-of-order
   alerts affect it.
4. Keep separate bounded state for each rule and correlation key. Extract the key from validated alert fields or a named
   regex capture. Reject a match when a required key is absent or malformed; never merge unrelated IPs, accounts, or
   actors into one instance.
5. On a completed match, create one immutable mitigation intent. Execute it through an outbound worker after FSM
   evaluation. Do not execute effects from intermediate transitions. A repeat match during cooldown may create an
   occurrence, extend a deadline, or be suppressed according to the rule's explicit policy.
6. Give each rule an explicit cooldown and reset policy. Restoration may be triggered by a finite mitigation lifetime,
   an explicit recovery signal, an operator request, or a combination specified by the rule. A quiet period alone does
   not prove every threat is gone; document its meaning for each rule.
7. Record the incident alert ID and every applied effect. Publish a warning security alert when the service decides to
   block or unblock, occurrences for subsequent material updates, and a normal state update when the incident has been
   restored and closed. Report failed and skipped effects accurately. Allocate producer-owned readable IDs as required
   by `CONTRIBUTING.md` and reuse an ID across retries, occurrences, and normalization.
8. Send a separate recovery email through a defined NATS RPC after successful normalization. The security email notifier
   needs a recovery template and a request/reply contract. Alert creation continues to use its existing subscription and
   email policy. Retry or deduplicate delivery so one normalization does not cause duplicate recovery mail.
9. Expose authenticated NATS RPCs to list configured rules and active automations, request rollback, and reset a rule
   instance's cooldown. Only `security-officer` and `admin` may perform the mutating operations. Record the actor and
   serialize operator requests with automatic transitions for the same instance.
10. Provide an operator UI for active automations and those controls when the UI is reachable. The UI must show rule,
    target, mitigation, start and expiry times, current state, and last outcome. Use the existing gateway/auth pattern;
    do not expose NATS directly to the browser. Follow Admin Panel language and validation conventions.
11. Give every rule a stable machine-readable `id`, human-readable `name`, and concise `description`. Return all three
    in rule-list and active-automation responses so the UI can explain the rule without parsing its JSON.
12. Let an authorized operator enable or disable a rule through the UI and an authenticated NATS RPC. This override
    lives only in service memory: startup restores each rule's configured enabled state. Show the effective state
    clearly and label the control as temporary. The service must record who toggled it and why, if a reason is supplied.

The packaged startup defaults enable sixteen detectors. Web/API direct HTTPS, horizontal-then-vertical scanning,
repeated sudo denial, and repeated USB connection rules have targeted automatic actions; two source-correlation
sequences publish security alerts without host effects. SSH and other generic network alerts do not prove direct HTTPS
ingress. An authorized operator may request a scoped direct HTTPS UFW action. `integrity_service_disruption` is
disabled. The UI and CLI may temporarily suspend or resume a startup-enabled rule in memory; they never write
`rules.json`. An administrator changes persistent enablement manually and restarts the service, since AIDE monitors that
file.

### Temporary rule disablement

Disabling a rule stops new observations from advancing its FSM and prevents new mitigation intents. It must not forget
an already applied effect. The service requests restoration of active UFW blocks owned by that rule and keeps failed
restores visible for retry or manual rollback. USB blocks remain until an operator explicitly rolls back their recorded
incident. One-way session actions have no inverse. Re-enabling a rule through the UI starts new detection state; it does
not recreate a prior incident. Restart removes the temporary override and reloads the configured enabled state. A
separate audit alert for each enable/disable request remains future work.

Enable, disable, rollback, cooldown reset, and automatic mitigation/restore must be serialized per rule instance. The
exact RPC request and response DTOs, including how a partial disable is reported, remain to be defined. A disabled rule
must remain visible in the UI until it is re-enabled or the service restarts.

### Operator UI

Keep the page simple and consistent with the existing Admin Panel `Loginctl`, Auth Sessions, and Services views: a
compact rule list with a colored rule name, short description, enabled/disabled state, active mitigation count, and last
outcome. Use icon buttons with visible tooltips and accessible labels for enable/disable, rollback, and cooldown reset.
Color must reinforce state rather than be the sole state indicator. Show a confirmation where an action changes an
active mitigation. Reuse existing page, table/list, button, and feedback patterns instead of introducing a new UI
framework. Localize UI chrome, tooltips, confirmations, errors, and accessibility text in English, Ukrainian, and
Danish; display operator-configured rule names and descriptions as authored unless the final rule schema defines
translated variants.

## Configuration modes

The service INI configuration must require `mode=strict` or `mode=permissive`; reject any other value. The mode is a
service-wide upper bound on actions, not a rule matcher. Every alert and decision remains observable in either mode.

- **`strict`**
  - **Intended use:** Production security response
  - **Action policy:** May run approved targeted mitigations and explicitly configured broad lockdowns after their
    guards pass.

- **`permissive`**
  - **Intended use:** Test machines where complete loss of access is unacceptable
  - **Action policy:** May run the approved targeted UFW, USB, loginctl, and Auth session actions after their target
    checks. Full Auth lockdown and guarded SSH lockdown remain unavailable.

The exact initial set of effects permitted in permissive mode must be listed in the service config or documented policy
before implementation. Changing mode must not erase active mitigations or pending restores. On startup and mode change,
reconcile recorded effects with actual state before resuming decisions.

## Rule and action model

The proposed command profiles, named arguments, and FSM plugin names are listed in [Security Automation action and FSM
reference](../docs/security-automation-reference.md).

The packaged configuration currently defines seventeen rules, sixteen enabled at startup. Detector enablement and
host-effect enrollment are separate: SSH and generic network rules have no automatic HTTPS UFW effect because their
alerts do not prove direct HTTPS ingress. The horizontal-then-vertical scan rule uses a distinct all-inbound source
block on physical interfaces for 30 minutes. The web, scan, sudo, and USB effects require the verified targets described
in the [action reference](../docs/security-automation-reference.md). Two sequence rules publish correlated security
alerts: SSH probing followed by a root login from the same IP, and repeated sudo denials followed by privileged group
membership for the same account. The latter does not identify who changed the membership. The integrity-to-service rule
remains disabled and has no host action because its input cannot identify a service unit. Installed-system testing must
verify each enrolled effect and restore path against real host events.

An action selects exactly one kind. `alert` publishes a correlated security alert through the existing Logging Manager
contract with configured severity and email policy, and no host effect or rollback. `alert_only` records a service-log
outcome. A `builtin` is a named operation implemented by the service, with typed inputs and a defined verification and
restore operation. Initial builtins are `block_source_ip` and `restore_source_ip` through UFW. A `command_profile` is a
configuration-owned, allowlisted executable and fixed argument layout for one of `/usr/bin/usbguard`,
`/usr/bin/loginctl`, `/usr/bin/systemctl`, or `/usr/sbin/ufw`; it is never a shell command. Profiles are compiled and
validated at startup. Rules name a profile and supply values through named placeholders such as `:username`,
`:requester_ip_address`, `:session_id`, `:device_id`, and `:unit_name`. Unknown placeholders, executable paths, verbs,
or argument positions fail startup validation. Substitution produces an argument vector passed directly to the
executable, never to a shell.

Each placeholder has a provenance and operand validator. `:requester_ip_address` requires a canonical, unicast IPv4 or
IPv6 address from a validated alert field or named capture; `:username` requires a verified account identity;
`:session_id` requires a session attributed to that account and incident; `:device_id` requires a USBguard device
identity verified against its current inventory; `:unit_name` requires a configured allowlisted service unit. Values
beginning with `-`, containing whitespace or control characters, or failing the invoked program's operand rules are
rejected. A display name or regex capture alone does not authorize a destructive operand. Where the program supports
`--`, place it before variable operands. The service must persist the expanded validated operands and exact resulting
external identity with the incident. A profile must define verification and a safe inverse; one-way actions such as
`loginctl terminate-session` require an explicit non-restorable policy and must not be presented as reversible
mitigation.

Builtins and command profiles use the same incident, mode, audit, retry, and restoration lifecycle. A command profile
cannot bypass the strict/permissive policy or gain sudo permission merely by appearing in JSON. The installed sudoers
file must grant only the reviewed profiles; if sudoers cannot constrain a dynamic operand safely, use a root-owned
helper with a narrow validated request contract. No generic `sudo <configured executable> ...` permission is allowed.

## Rule JSON

The executable rule schema is documented in [Security Automation action and FSM
reference](../docs/security-automation-reference.md) and packaged under
`configs/etc/lite-nas/security-automation/rules.json`. The following earlier sketch remains a readability example only;
use the packaged schema for actual configuration. `source_ip` is a named capture from alert text; a future structured
alert field would be preferable.

```json
{
  "fsm_plugin": "sequence",
  "rules": [
    {
      "id": "port_scan_then_ssh_failures",
      "enabled": false,
      "name": "Port scan followed by SSH probing",
      "description": "Detect a network scan followed by SSH failures from the same source IP.",
      "correlate_by": "source_ip",
      "sequence": [
        {
          "event": {
            "type": "security",
            "id_prefix": "secnet",
            "message": { "regex": "Blocked TCP .* from (?P<source_ip>[0-9a-fA-F:.]+)" }
          }
        },
        {
          "event": {
            "type": "security",
            "id_prefix": "secssh",
            "message": { "regex": "from (?P<source_ip>[0-9a-fA-F:.]+)" }
          },
          "within": "10m"
        }
      ],
      "cooldown": "30m",
      "mitigate": {
        "block_cmd": "block_source_ip",
        "restore_cmd": "restore_source_ip",
        "lifetime": "30m"
      }
    }
  ]
}
```

In this readability sketch, `block_cmd` and `restore_cmd` should be replaced by an action selector and its typed
operands under the final schema. They are not shell command strings. A `strict`-only effect such as guarded SSH lockdown
requires an explicit policy marker and guard result.

## Rules in release scope

All rules discussed below are intended to ship. Each must have verified input data, correlation, mitigation,
restoration, and alert lifecycle before it is enabled. Where current data is insufficient, implement the required
upstream observation or alert enrichment as part of this feature. The current Log Analysis create message is a summary
such as `Security rule triggered: <rule ID> for account <account> from <address>`; it does not contain the original log
line. Occurrence payloads do not carry that summary.

- **SSH brute force**
  - **Correlation and timing:** Ten or more SSH failures from one IP in five minutes
  - **Required data work:** Existing `secssh` Log Analysis plugin already performs this aggregation. An automation rule
    can react to its emitted alert; counting ten `secssh` alerts would mean something different.
  - **Intended mitigation:** Detect and report; the scoped UFW action does not block SSH.

- **SSH probing followed by root login**
  - **Correlation and timing:** Aggregated `ssh_preauth_probing` alert followed by `ssh_root_login` from the same IP
    within ten minutes
  - **Required data work:** The Log Analysis summaries carry the source IP but do not establish a common person behind a
    shared address.
  - **Intended mitigation:** Publish a critical correlated security alert with the source alert IDs; no host effect.

- **Account enumeration or password spraying**
  - **Correlation and timing:** Five or more distinct usernames from one IP
  - **Required data work:** Current `secssh` alert summarizes a source-based threshold, not distinct usernames. Needs a
    richer observation or upstream aggregate.
  - **Intended mitigation:** Detect and report until an SSH-specific action is defined.

- **Account under attack**
  - **Correlation and timing:** `secauth` authentication failure followed by `secauth` account-lockout event for the
    same account
  - **Required data work:** Match Log Analysis rule IDs in alert messages; confirm account and source address
    attribution. Do not apply global lockdown for this sequence.
  - **Intended mitigation:** Account-specific restriction after an owner contract is defined.

- **Reconnaissance then SSH attack**
  - **Correlation and timing:** `secnet` scan followed by `secssh` alert from the same IP
  - **Required data work:** Both create messages appear to expose source address; verify parsing and timing.
  - **Intended mitigation:** Detect and report; a direct HTTPS block requires separate verified ingress provenance.

- **Network reconnaissance**
  - **Correlation and timing:** `secnet` horizontal scan plus repeated blocked attempts from the same IP
  - **Required data work:** Distinct network categories exist; verify common source IP.
  - **Intended mitigation:** Direct HTTPS block only with verified physical ingress provenance.

- **Horizontal then vertical scan**
  - **Correlation and timing:** `secnet` horizontal-scan alert followed by a vertical port-scan alert from the same
    source IP within ten minutes
  - **Required data work:** Match the Network Security categories and canonical source address in both alert summaries.
  - **Intended mitigation:** Deny all inbound traffic from the source on verified physical interfaces for 30 minutes,
    then restore only the owned rules.

- **Repeated unauthorized sudo**
  - **Correlation and timing:** Three or more `secpriv` sudo denials for one user in five minutes
  - **Required data work:** Confirm each denial yields a usable create event and account extraction. Session termination
    requires separate, confident session attribution.
  - **Intended mitigation:** Account restriction; terminate only an implicated session.

- **Sudo denials followed by privileged group membership**
  - **Correlation and timing:** `repeated_sudo_policy_violation` followed by `privileged_group_membership` for the same
    account within ten minutes
  - **Required data work:** The summaries identify the affected account but not the actor who changed its groups.
  - **Intended mitigation:** Publish a critical correlated security alert with the source alert IDs; no host effect.

- **SSH attack and privilege probing**
  - **Correlation and timing:** `secssh` failures plus `secpriv` sudo denial attributed to the same account
  - **Required data work:** Account correlation may be possible; sudo denial may not include a source IP.
  - **Intended mitigation:** Detect and report until an account restriction owner is defined.

- **Privilege escalation or persistence**
  - **Correlation and timing:** `secpriv` privileged-group addition plus `secacct` new user by the same actor within ten
    minutes
  - **Required data work:** Existing summaries may identify affected accounts but not the actor. Requires verified actor
    data.
  - **Intended mitigation:** Urgent alert; validated account/group reversal.

- **Repeated file access denial**
  - **Correlation and timing:** `secfile` alert for one account
  - **Required data work:** Existing upstream plugin already aggregates 51 denials in five minutes.
  - **Intended mitigation:** Account-specific restriction after attribution.

- **New USB device plus suspicious activity**
  - **Correlation and timing:** `secusb` alert followed by activity tied to the same device
  - **Required data work:** Shared device identity across sources is unverified.
  - **Intended mitigation:** Restrict that device after correlation exists.

- **Integrity change plus service disruption**
  - **Correlation and timing:** `integrity_` alert followed by `syssvc` unavailability
  - **Required data work:** Current AIDE alerts do not identify a service unit, so this candidate sequence remains
    disabled.
  - **Intended mitigation:** None; no immediate host change follows an integrity failure.

- **Repeated service recovery failure**
  - **Correlation and timing:** Watchdog incident plus `syssvc` unavailability for the same unit
  - **Required data work:** Avoid competing with Watchdog's existing recovery loop.
  - **Intended mitigation:** Escalate and stop repeated attempts under a defined owner policy.

### Initial action assignment

- **SSH brute force**
  - **Initial action class:** Candidate builtin `block_source_ip`
  - **Enrollment condition:** The current SSH alert does not establish direct HTTPS ingress, so skip the
    physical-interface TCP 443 block. SSH protection requires a separate verified action.

- **SSH probing followed by root login**
  - **Initial action class:** `alert`
  - **Enrollment condition:** Publish a correlated security alert with email; do not claim both events have the same
    actor.

- **Account enumeration or password spraying**
  - **Initial action class:** Candidate builtin `block_source_ip`
  - **Enrollment condition:** The current SSH alert does not establish direct HTTPS ingress, so skip the
    physical-interface TCP 443 block.

- **Account under attack**
  - **Initial action class:** Alert only until a verified source and action owner exist
  - **Enrollment condition:** Account-only restriction needs a separate owner contract and remains unavailable until
    defined.

- **Reconnaissance then SSH attack**
  - **Initial action class:** Candidate builtin `block_source_ip`
  - **Enrollment condition:** Both steps must resolve to the same verified IP and direct HTTPS ingress; otherwise skip
    the block.

- **Network reconnaissance**
  - **Initial action class:** Candidate builtin `block_source_ip`
  - **Enrollment condition:** Both network categories must resolve to the same verified IP and direct HTTPS ingress;
    otherwise skip the block.

- **Horizontal then vertical scan**
  - **Initial action class:** Builtin `block_scan_source_ip`
  - **Enrollment condition:** Two distinct Network Security alerts must identify the same canonical source IP in order;
    block all inbound traffic on verified physical interfaces for 30 minutes and restore by recorded identity.

- **Repeated unauthorized sudo**
  - **Initial action class:** Command profile for an attributed `loginctl` session, if explicitly configured
  - **Enrollment condition:** Session termination is one-way and requires an explicit non-restorable policy. Web-session
    revocation is a separate rule using a scoped Auth Service RPC; both require verified identity and source
    attribution.

- **Sudo denials followed by privileged group membership**
  - **Initial action class:** `alert`
  - **Enrollment condition:** Publish a correlated security alert with email; do not infer the group-change actor.

- **SSH attack and privilege probing**
  - **Initial action class:** Alert only
  - **Enrollment condition:** Account-only restriction needs its own contract; current alerts do not establish direct
    HTTPS ingress.

- **Privilege escalation or persistence**
  - **Initial action class:** Alert only until a narrow reversal profile is defined
  - **Enrollment condition:** Actor, changed account/group, and exact inverse must be verified. Do not infer an inverse
    `systemctl` or `loginctl` operation.

- **Repeated file access denial**
  - **Initial action class:** Alert only until an account restriction owner is defined
  - **Enrollment condition:** Current toolset has no account restriction operation.

- **New USB device plus suspicious activity**
  - **Initial action class:** Command profile for `usbguard block-device`, with a corresponding allow/restore profile
  - **Enrollment condition:** Shared device identity and pre-incident USBguard policy must be verified; restore must
    recover the prior policy, not assume allow.

- **Integrity change plus service disruption**
  - **Initial action class:** Alert only, currently disabled
  - **Enrollment condition:** Do not stop a service in response to the current AIDE and service-unavailable alert pair.

- **Repeated service recovery failure**
  - **Initial action class:** Alert only until Watchdog defines a suppression/ownership contract
  - **Enrollment condition:** A generic `systemctl stop` would compete with Watchdog and does not stop its retry loop.

## Mitigation capabilities and privilege boundary

Run the process as a dedicated non-root service user. Package an explicit sudoers policy for only the host operations
actually implemented. Validate a configured action against an application and operation allowlist, then validate every
variable operand against that program's rules. Launch executable and arguments separately without a shell. Do not accept
arbitrary command lines, shell syntax, option-shaped targets, or unbounded unit names. Review whether sudoers can
constrain each operation sufficiently; a narrow root-owned helper is an option if it cannot.

- **Built-in source IP block and restore via UFW**
  - **Intended use and guard:** Validate an operator-supplied canonical IPv4/IPv6 address; require direct HTTPS ingress
    provenance for any future automatic block. Insert an identifiable service-owned deny for inbound TCP 443 on verified
    physical interfaces ahead of any allow; persist each exact rule identity and a 30-minute expiry; remove only those
    rules. Cloudflare Tunnel and Tailscale traffic remain outside this scope. UFW does not own the expiry timer.

- **Guarded SSH lockdown via UFW**
  - **Intended use and guard:** In strict mode only, block SSH on physical interfaces while preserving an approved
    active VPN path. Apply only after positive interface classification and VPN reachability checks. Record each created
    rule and its restore operation. Skip and report on uncertainty.

- **`systemctl` via sudo**
  - **Intended use and guard:** No Security Automation rule currently uses this capability. Any future use requires a
    named operation, configured units, and a defined relationship with Service Watchdog.

- **`loginctl` via sudo**
  - **Intended use and guard:** Terminate only a session or user that is confidently attributed to the incident.

- **Auth actions over NATS**
  - **Intended use and guard:** Reuse existing lockdown and session RPCs where their authorization and request contracts
    fit; any broader revocation operation needs an explicit contract.

- **Account restriction**
  - **Intended use and guard:** Candidate capability; define owner, exact operation, validation, and restoration before
    enabling.

Operator reversal is a first-class path. Persist the rule ID, target, applied command profile, exact validated operands,
resulting external identity, expiry, restore intent, and last outcome. Local logs must include a usable rollback command
or equivalent precise steps. Security alerts and emails contain an action summary and outcome, **never rollback command
text**. Never claim a block, restore, or normalization succeeded solely because it was requested.

## Incident and restoration lifecycle

1. Validate an alert and dispatch it to matching rules. A rule's FSM evaluates it within the `(rule ID, correlation
   key)` state pool.
2. On a complete match, allocate an incident ID and record a pending mitigation intent before applying any external
   effect. Repeated delivery of the same alert or intent must not duplicate a block.
3. Execute the approved effect and verify its result. Publish a warning security alert describing the decision and
   result. Preserve the incident ID for later occurrences and state changes.
4. While active, add occurrences for material repeat observations and effect updates according to the rule's cooldown
   policy. Keep the effect deadline separate from the detection cooldown.
5. At expiry or verified recovery, run the owned restore operation. A failed restore keeps the incident open, raises an
   observable failure, and remains available for operator rollback.
6. After verified restoration, publish a normal state update and request the separate recovery email. Record email
   delivery outcome without reopening a restored mitigation.
7. On restart, load pending effects, reconcile with external state, and resume deadlines. An in-memory FSM store alone
   cannot guarantee restoration across restarts.

Manual rollback and cooldown reset must be serialized with automatic work for the same instance. Define whether a manual
rollback also suppresses immediate reapplication; the UI must display that policy. No channel may be closed while an
active NATS callback can still send to it.

An operator's force-normalize request first asks the service to roll back its owned changes and verify restoration. It
keeps the incident alert and all prior occurrences in history, adds a normal state update only after verification, and
creates a separate security audit alert for the operator request and outcome. A failed or unverified restore leaves the
incident open. The CLI must call the same authenticated automation RPCs as the gateway rather than altering Logging
Manager records directly.

## Worker and integration design

Follow the repository's one-process worker pattern. Construct validated configuration, compiled plugins, rule instances,
and messaging clients at startup. NATS handlers validate and hand off typed observations through bounded channels. A
dispatcher feeds the FSM; an outbound effect worker owns privileged execution and alert publication. A timer worker
handles sequence deadlines, cooldowns, and restores. A separate operator RPC path serializes commands with the same
incident owner. For every channel, implementation documentation must name owner, producer, consumer, capacity,
backpressure, cancellation, and close policy. Route errors to logs and alert/incident state; never silently drop
security observations or effects. Stop producers, account for pending effects, and drain NATS where possible on
shutdown.

Reuse `shared/go/fsm` transition tables and keyed store semantics, `shared/go/workers` lifecycle and polling primitives,
`shared/go/messaging`, the existing logging manager create/occurrence/state contracts, service token authentication, and
Log Analysis startup compilation pattern. Add shared adapters only when more than one service demonstrates the same
need. Persist mitigation records outside the volatile FSM store. The storage mechanism and recovery contract are an open
design choice.

The security email notifier currently subscribes to alert creations and renders one `alert.html` template. Define a
recovery notification RPC, its validated payload, authorization, response and retry semantics, then add a dedicated
recovery template. Update NATS permissions accordingly. Automation operator subjects and DTOs are defined in
`shared/go/contracts/securityautomation`; only the scoped operator-requested UFW host mutation and restoration path are
implemented so far.

## Documentation and delivery work

- Add a traceable requirements document before implementation, with functional, integration, security, and operational
  IDs. This specification can be split or converted into that format once contracts are settled.
- Document the final JSON schema, every FSM plugin type, correlation-key extraction, window and cooldown semantics, mode
  restrictions, permitted actions, and restore behavior. Supply valid and invalid examples.
- Document installed service config, packaged rules, dedicated account, systemd unit, sudoers entries, NATS account
  permissions, and the recovery email template. Preserve operator-edited rules/config on deployment using existing
  repository conventions.
- Document command operand validation, UFW rule ownership and precedence, VPN guard checks, restart reconciliation, and
  manual recovery in the operator runbook.
- Extend security monitoring, email notifier, gateway, Admin Panel, and security tooling docs where those integrations
  change. Add release notes in the repository's release-qualified format when preparing a release.
- Implement requirement-traceable service tests and focused system tests for every shipped rule and the installed
  workflows: matching, per-key isolation, cooldown/reset, strict/permissive behavior, command validation, failure and
  retry, restart restoration, RBAC, UI, and email. Use the repository's existing test placement and fixture conventions.

## Open decisions before implementation

1. Enrollment-tested predicates and correlation captures for each rule in the finalized JSON schema.
2. Which alert subjects and lifecycle events are ingested, and how duplicate delivery and event ordering are handled.
3. Structured correlation fields versus regex extraction from summary messages; provenance needed for actor and session
   attribution.
4. Exact cooldown, threat-gone, mitigation expiry, and manual reset semantics for each shipped rule.
5. Durable incident store and reconciliation policy, including what happens if external state changes outside this
   service.
6. Specific permitted effects in permissive mode and whether mode changes can cancel active strict-mode lockdowns.
7. Recovery email RPC DTO, authorization, and gateway behavior; automation operator RPC subjects and routes are already
   defined.
8. UFW rule ordering and safe removal, VPN health criteria, protected interface inventory, and precise sudoers
   constraints.
9. Verified alert examples and upstream data changes for every rule in release scope. A rule must not be enabled until
   its required correlation data and mitigation are implemented and verified; the release is incomplete until all
   discussed rules meet that condition.
10. Whether a disable request may complete while one or more restores are pending, and how its audit alert reports that
    partial outcome. The requirement to keep affected incidents open remains fixed.
