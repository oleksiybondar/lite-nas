# Security Monitoring Skeleton — Implementation Task Specification

**Version:** 0.4 — repository-aligned implementation plan
**Status:** TASK-000 complete; security services and FSM work pending
**Scope:** Five new single-process/multi-worker Go services; mandatory channel-based worker/NATS architecture and
AGENTS.md rule; explicit instantiated-plugin injection architecture for Log Analysis; reusable table-driven FSM engine;
mandatory migration of applicable existing monitoring logic; and all previously agreed integrations, security,
packaging, CI/CD, tests, demos, and documentation
**Repository inspection:** TASK-000 worker and messaging walkthrough completed. The verified baseline is below;
remaining service-specific contracts and runtime choices still require inspection before implementation.

## 0. Mandatory first microtask — TASK-000: Worker pipelines, NATS channel adapters, and AGENTS.md

**Completed foundation:** Root `AGENTS.md` now defines the worker, channel,
NATS, and compiled-plugin construction rules. `docs/development-notes.md`
maps the concrete APIs and gives a two-process pipeline example. No shared
adapter was added: the existing email notifier handler demonstrates a channel
bridge, and no reusable gap has yet been demonstrated.

| Concern | Verified repository implementation |
| --- | --- |
| Worker lifecycle | `shared/go/workers`: `Worker.Start(context.Context)`, `BaseWorker`, `TimerWorker`, `NewPollingTimerWorker`, `PollingWorker[T]` |
| Outbound NATS | `shared/go/messaging.Client`: `Publish`, `Request`, `Drain`, `Close` |
| Inbound NATS | `shared/go/messaging.Server`: `Subscribe`, `RegisterRPC`, middleware, `Drain`, `Close`; callbacks receive `Envelope` |
| Channel bridge | `shared/go/emailnotifier/subscription.go` validates and forwards an alert into a bounded channel; no generic shared adapter exists |
| Channel pipeline | `services/system-metrics/modules` wires bounded polling and processing channels; its runtime publishes output through `messaging.Client` |
| Existing exception | `services/resources-monitor/runtime.go` subscribes a processor callback directly, so the planned worker pipeline is not an already universal pattern |

For new services, the runtime owns channel creation, producer/consumer direction,
capacity and closing. Bounded sends must define cancellation and backpressure.
Multiple receivers divide messages; explicitly fan out when each plugin needs
one. Security observations and effect failures need an observable, handled path.
Stop producers and callbacks before closing channels, account for pending work,
then drain messaging connections. `PollingWorker[T].EmitError` has a
nonblocking send and can discard errors, so security pipelines must not rely on
it as their sole failure path. Current messaging uses core NATS, with no durable
acknowledgment or replay guarantee. Any required delivery or reconciliation
strategy remains a per-service design task.

**TASK-000 acceptance: complete for architecture and documentation.** Future
implementations must follow these rules and validate their own topology and
shutdown behavior with tests.

## 1. Objective

Implement a lean foundational security monitoring layer using existing project infrastructure. Deliver five narrowly
scoped services:

| ID | Service | Primary responsibility |
| --- | --- | --- |
| SVC-001 | Log Analysis | Analyze selected rsyslog/audit logs with configurable patterns and thresholds; report security events; request auth-service lockdown where configured, after the auth prerequisite is implemented. |
| SVC-002 | Mount Monitor | Detect mount/unmount activity, report informational operational events, and publish mount lifecycle notifications via NATS. |
| SVC-003 | Antivirus | Consume mount events, select eligible mounts, scan with existing/appropriate ClamAV integration, report detections and notify by email. |
| SVC-004 | System Service Monitor | Periodically inspect configured systemd units, distinguish relevant states, report operational failures, and publish service-state notifications via NATS. |
| SVC-005 | Service Watchdog | Consume service-state notifications, protect configured critical units, recover persistent failures, check expected enablement, and report security incidents. |

The intent is a **security skeleton**, not a SIEM, general-purpose scripting/rules DSL, or automated remediation
platform. A small reusable FSM engine is expressly in scope.

### 1.1 Out of scope unless existing contracts demand otherwise

- Reimplementing the Security Logging Manager, System Logging Manager, NATS,
  Authentication Service lockdown inside SVC-001, or mail transport.
- A generic syslog analytics platform, large rules library, or complex detection DSL.
- Dynamically loaded executable Go plugins, user-defined code in configuration, and implementation of future
  force-logout/password-expiry/account-disable actions.
- New quarantine, deletion, or automatic unmount features in Antivirus.
- Arbitrary systemd service management, unconstrained root privileges, or new administrative API surfaces.

## 2. Governing implementation approach

1. **Inspect the repository first.** Locate existing service templates, project groups/users, NATS contracts and ACLs,
   logging managers and event lifecycle APIs, shared Go workers/configuration/CLI utilities, email implementation,
   Debian scripts, deployment scripts, AppArmor/systemd units, GitHub Actions workflows, and documentation
   conventions.
2. **Reuse established contracts.** Each new service adapts to existing integration patterns; do not change shared
   infrastructure merely for convenience.
3. **Minimal scope, production-quality implementation.** Keep features small and configuration readable, while ensuring
   correctness, security and operational resilience.
4. **Explicitly separate discovery from implementation.** Record gaps, decisions, and any justified new shared module
   APIs. Never guess message subjects or existing API schemas.
5. **Demonstrate incrementally.** In particular, refine log-rule configuration and authentication detection with sample
   data and team demos.

### 2.1 Existing architecture constraints

- **NATS** is the internal service messaging backbone. Use existing contracts, subjects, access-control and subscription
  conventions; update subscriptions/publication permissions for new services.
- **Security Logging Manager** owns security event registration, updates, and lifecycle.
- **System Logging Manager** is the existing operational alert manager; confirm
  its alert contract and severity mapping for mount and service events.
- **Authentication Service lockdown is not implemented yet.**
  `shared/go/contracts/auth` declares `SetLockdownRPCSubject`,
  `SetLockdownRequest`, and `LockdownChangedEvent`, but
  `services/auth/runtime.go` does not register the lockdown RPC or enforce
  lockdown. SVC-001 must not publish a request and claim enforcement until this
  prerequisite and its authorization/ACL path are implemented and tested.
- **Email capability** is implemented by `system-email-notifier` and
  `security-email-notifier` through `shared/go/emailnotifier`; they consume
  logging-manager alert events and deliver through local SMTP/Postfix.
- **Shared Go modules** contain polling/worker utilities and baseline configuration. Extend shared worker support only
  if genuinely needed and in existing project style.
- UNIX commands follow existing CLI execution conventions and secure process execution rules in §11.

## 3. Cross-service event and lifecycle rules

- Follow the exact existing manager contracts for event create/update/close and severity/type values.
- A single internal detection may require multiple external lifecycle operations. Each owning service **MUST** hold
  sufficient correlation/state to progress the manager lifecycle correctly. SVC-001 **MUST** use the shared
  table-driven FSM engine defined in §5; existing monitoring services **MUST** be assessed and suitable FSM-like
  implementations migrated under TASK-FSM-002. Other new services SHOULD reuse the shared engine only where it fits
  their workflows.
- Distinguish detection/observation state from manager event lifecycle state. Do not blindly create duplicate events for
  polling repeats.
- Decide and document whether intermediate state requires persistence; a volatile cache is **not** assumed sufficient
  for every pending event or side effect.
- For NATS-triggered actions, address duplicate delivery, reconnect/restart, out-of-order notifications, missed
  messages, and the need for state reconciliation where relevant.
- Record and correlate failures of outgoing operations. Logging or notification failures must not be silently treated as
  successful delivery.
- Avoid blocking primary event processing on email delivery; use the existing retry/deduplication conventions if
  present.
- Retain stable rule/service/mount/correlation identifiers and timestamps in events when supported by existing schemas.
- Preserve source attribution and relevant failure information without exposing passwords, tokens, or other secrets.

### 3.1 TASK-EVENT-LIFECYCLE-001 — Per-source event correlation and resolution

Refine alert lifecycle behavior for mount, antivirus, service-monitor,
watchdog, and Log Analysis events. Use the existing `resources-monitor`
lifecycle as a reference for stable source/rule correlation and normalization;
inspect its current contracts and behavior before reusing the pattern. Events
from independent mounts, scans, services, or protection conditions MUST
remain independently identifiable and independently resolvable. Do not emit
occurrence records for ordinary one-shot observations or repeated state
publishes. Emit occurrences only when an FSM intentionally aggregates multiple
observations into one incident.

Required refinements:

- **Mount Monitor:** correlate a mount alert with the matching mount identity
  and resolve that alert when that same mount is observed unmounted. Keep mount
  and unmount lifecycle notifications available to subscribers. Handle mount
  ID reuse, identity changes, monitor restart/reconciliation, duplicates, and
  out-of-order notifications according to the existing mount contract.
- **Antivirus:** report one malware event per scan with a unique event ID and
  the antivirus service as source; do not reuse an event ID for later scans or
  emit occurrences for multiple detections from one scan. Resolve a one-shot
  malware event immediately after successful registration. Keep clean,
  incomplete, and cancelled scan outcomes distinct.
- **System Service Monitor:** maintain an independent alert lifecycle per
  configured systemd unit. A unit becoming unavailable creates its incident;
  repeated polls do not create occurrences or replace/merge another unit's
  incident; restoration normalizes only
  that unit's incident. Preserve the relationship across service state changes,
  duplicate observations, failed probes, and monitor restart/reconciliation.
- **Service Watchdog:** keep each protected unit and protection-condition
  incident independent. Repeated recovery attempts append occurrences only
  when the watchdog FSM tracks those attempts; verified recovery normalizes
  only the corresponding incident. Distinguish availability recovery from enablement/protection
  restoration, and do not close one incident merely because a different
  condition or unit recovered.
- **Log Analysis:** resolve one-shot detections immediately after registration.
  Multi-observation FSM detections remain active until their configured quiet
  period expires. Each matching observation during that active period adds an
  occurrence and restarts the quiet-period deadline; cooldown expiry normalizes
  the event and clears its incident identity so a later burst creates a new
  event.
- **ZFS/network monitoring alignment:** inspect current source/rule-keyed alert
  and occurrence behavior in `resources-monitor` for ZFS and network metric
  rules. Document which stable source identity is used and ensure the lifecycle
  semantics above align where the same alert manager contract applies.
- Define stable correlation keys, event IDs, occurrence meaning, normalization
  conditions, persistence/restart reconciliation, and duplicate/out-of-order
  handling against the existing logging-manager contracts before changing
  implementation. Do not claim durable or exactly-once lifecycle behavior
  without corresponding storage or messaging guarantees.

Acceptance criteria:

- Two mounts, two systemd units, and two watchdog-protected units can each
  maintain separate incidents; transitions for one do not mutate another.
- Occurrences are emitted only by FSM paths that aggregate multiple
  observations; repeated polls and one scan's multiple findings do not create
  occurrence rows.
- Unmount or verified recovery resolves only the correlated incident; repeated
  polling, mount changes, and unsuccessful probes do not prematurely normalize
  an event.
- One-shot security detections are immediately normal after successful
  registration. Password-guessing incidents normalize only after the configured
  quiet period, which resets whenever another failure arrives.
- Restart/reconciliation, duplicate delivery, and out-of-order transition
  behavior are documented and covered at the service level; externally visible
  lifecycle behavior is verified through Python HyperionTF system tests and
  the relevant CLI.

## 4. SVC-001 — Log Analysis Service

### 4.1 Function

Receive or read **selected logs forwarded via rsyslog** using a repository-compatible input mechanism; apply a small set
of configuration-driven patterns and conditions; raise/update Security Logging Manager events. For configured critical
detections, use the existing security email notifier path. Lockdown-triggering
rules require the auth-service lockdown prerequisite described in §2.1.

### 4.2 Rules and configuration

A readable rule entity **MUST** support, as needed by initial demonstrations:

- Stable rule ID and enable/disable setting.
- Log source mapping (specific forwarded file/input and recognized format).
- Match criteria (patterns plus required/exclusion conditions), including safe validation of configured patterns.
- Trigger threshold: qualifying detections at least **X** times, with a configurable time window and/or idle counter
  reset after **Y**. The exact counting semantics **MUST** be documented and tested. Detection window,
  alert-suppression cooldown, and idle-state expiration are separate concepts and MUST NOT be conflated.
- Correlation dimensions relevant to the data, e.g. account, originating address, process/session, and audit-event
  identifier.
- Security event severity/type, email-on-critical configuration, and explicit `lockdown` action setting.
- Clear validation failures for invalid rule configuration; no silent fallback to unsafe behavior.

Use the existing shared Go configuration loader and **inspect prevailing project file formats** (JSON/YAML/etc.) before
choosing a human-friendly rules format. The exact schema is **intentionally open** and must be iterated against sample
logs in team demos. No sample key names in this document are normative.

### 4.3 Plugin-based ingestion, instance injection, routing, and FSM use

**Normative architecture: configured instances are constructed at startup and injected into the listener, NOT
instantiated or dynamically imported by the listener.**

- Organize built-in detection types under a service-local `plugins/` Go
  package/folder. Each is a statically compiled Go implementation of a common
  plugin API with the project worker lifecycle conventions.
- Config contains an explicit enabled plugin **instance list**; each entry selects an approved compiled plugin type,
  assigns a stable instance ID, and provides validated instance-specific settings. Instantiate exactly the configured
  instances using the startup composition root and a static type-to-constructor registry. Multiple configured
  instances of the same plugin type MUST be supported.
- Application initialization passes the **already-initialized plugin instance map/collection** to the
  listener/dispatcher at construction. The listener does not create plugin instances, dynamically import code, resolve
  constructors, or inspect Go files at runtime. The plugin registry is an initialization mechanism, not a runtime
  discovery mechanism inside listeners.
- Listener(s) implement existing worker lifecycle APIs. Plugin workers likewise conform to the same lifecycle contract,
  but their data input/output types may differ; avoid forcing sources and processors into an awkward identical data
  interface.
- Read a physical configured log source **once** where feasible. Build a configuration-derived route index mapping `log
  source -> category -> one or more rules -> initialized plugin instance reference`. Categories can share a log file
  while using distinct rules, parameters, methods/guards, and detection FSM plugins.
- On new entries, the listener emits a typed record to an internal channel; dispatcher worker(s) perform cheap, safe
  initial rule/pattern selection and fan out each qualifying normalized observation to **every applicable** plugin
  instance via that instance's channel. A single shared work-queue channel must not accidentally distribute messages
  when independent rule consumers must each inspect them.
- Plugin workers receive observations through typed, bounded input channels, invoke their shared FSM engine **within the
  worker**, and either continue waiting/accumulating state or emit a typed detection/effect intent on an output
  channel. Reporting and NATS publisher workers consume typed effect intents;
  the existing security email notifier consumes alert events separately.
- Plugin configurations contain rule/instance references and parameters, not dynamic code/import paths or arbitrary
  executable functions. Connection topology, subscriptions, and backpressure must be validated at startup and
  demonstrated with at least two categories sharing one source.
- The same channel-connected worker pattern MUST apply to Mount Monitor, Antivirus, System Service Monitor, and
  Watchdog, including their NATS sender/receiver integration. Existing NATS and shared worker packages remain
  authoritative.
- Initial plugin behaviors: **immediate detection** (one qualifying observation can alert), **threshold detection** (at
  least X observations in a configured window, followed by cooldown behavior), and **keyed threshold detection**
  (independent state/counters/deadlines for each validated correlation key such as UID/GID).
- A single reusable threshold implementation MAY support both global and keyed behavior, but all three configuration
  behaviors MUST be demonstrable.
- Rule-specific parsing and matching produces normalized observations; selected plugins translate observations into
  state transitions and typed effect intents. Rule configuration MUST NOT contain executable code or arbitrary
  function names that can invoke unregistered functions.
- The FSM engine is shared Go infrastructure (§5); it must not embed audit-log parsing, NATS subjects, mail transport,
  or hardcoded security reactions.

### 4.4 Initial authentication ruleset

- Monitor failed UNIX authentication/login attempts from actual project audit-log sources (the exact audit/rsyslog
  records must be inspected).
- Count **authentication attempts**, not every line from a multi-record audit event; correlate audit records where
  possible.
- Minimize false positives from ordinary mistyped passwords/logins, while surfacing more suspicious patterns such as
  repeated attempts, invalid usernames, cross-account probing, and repeated bursts when metadata permits.
- Treat source address, account, successful logins, service type, and rate as context, **not proof of attacker intent**.
  Never promise perfect human-versus-probing classification.
- Choose initial thresholds through examples, negative cases, realistic fixtures, and team feedback, rather than assume
  arbitrary production numbers.
- Do not enable disruptive lockdown by default. A threshold alone should not automatically imply permission to lock down
  the system.

### 4.5 Lockdown flow

1. Detection satisfies a rule with an explicit lockdown action.
2. SVC-001 validates the rule/action policy and deduplicates repeat requests for the same incident.
3. First implement and verify auth-service handling of
   `authcontract.SetLockdownRPCSubject` and `SetLockdownRequest` with its
   enforcement and authorization policy. The declared subject is
   `auth.rpc.lockdown.set`; declaration alone is not a working RPC.
4. SVC-001 sends the validated request via `messaging.Client.Request`, then
   handles the response or failure. Align NATS ACLs with the actual subject:
   `configs/etc/nats-server/authorization.conf` currently lists older
   `auth.cmd.lockdown.set` and `auth.events.lockdown` names.
5. Authentication Service owns enforcement and the state-change event.
6. Report and correlate RPC responses and failures as far as the implemented
   contract supports. A successful transport request alone does not prove
   enforcement.

**Acceptance examples:** isolated mistype; mistype then success; repeated failed attempts; many accounts from one
source; duplicate audit records; counter expiration; absent metadata; malformed log line; email threshold; lockdown
enabled/disabled; denied or unavailable NATS publication.

## 5. Shared Go table-driven FSM engine (TASK-FSM-001 — mandatory)

### 5.1 Purpose and ownership

Implement a reusable, generic shared Go package providing **typed, declarative, table-driven finite-state machines**.
SVC-001 is the required first consumer. The API SHOULD allow other event-monitoring and future security automation
components to reuse state-transition mechanics without embedding their domain-specific commands, managers, or messaging
into the engine.

### 5.2 Declarative FSM definition

- Each plugin defines immutable **named states**, **input/event identifiers**, an explicit **transition table**, and a
  registry of typed **guard** and **action/evaluation** functions identified by stable names.
- Table rows SHOULD express `current state`, `input`, optional `guard`, `next state`, and optional `action`/effect. The
  table is defined in Go code as typed constants/data and is easy to traverse, print, inspect, and unit-test.
- A rule's configuration selects an approved plugin and supplies validated parameters. Configuration MUST NOT provide
  arbitrary executable functions, process commands, or dynamic Go `.so` plugins.
- Distinguish **state identifiers** from **transition outcomes**: matched transition, stay/no change, not
  applicable/sequence break, reset, and evaluation error. A normal sequence break MUST NOT be recorded as an engine
  failure.
- Guards select whether a transition applies; actions compute next state data and typed effect intents. A complex
  condition MAY be evaluated by a registered Go function while transition destinations remain visible in the table.
- Resolve multiple eligible rows deterministically using a documented priority or reject ambiguous tables during
  initialization. Detect missing targets/functions, unreachable states when practicable, and unbounded automatic
  transitions.

**Illustrative, nonbinding table:**

| State | Input | Guard | Next state | Effect |
| --- | --- | --- | --- | --- |
| idle | qualifying_observation | always | counting | initialize counter/deadline |
| counting | qualifying_observation | below_threshold | counting | increment counter |
| counting | qualifying_observation | threshold_reached | cooldown | emit detection intent; set cooldown deadline |
| counting | window_expired | always | idle | reset window counter |
| cooldown | qualifying_observation | always | cooldown | correlate/update incident according to policy; suppress repeat notifications |
| cooldown | cooldown_expired | always | idle | reset/rearm; expire correlation when eligible |

This is an example only. The implementer MUST ensure the threshold-crossing observation is counted exactly once,
deadlines are consistent, and guard evaluation order is unambiguous.

### 5.3 Engine loop, instance data, and workers

- The generic engine is a **library called within plugin worker execution**, not a central manager that creates
  listeners or wires the channel graph. It accepts a validated table, function registry, a state-store abstraction,
  and injected time/deadline facilities; per-rule/per-key instances store state, bounded metadata/counters,
  deadline(s), event references, and effect delivery status as needed.
- Model timeout expiry as an **ordinary FSM input** traversing the **same transition engine**, not ad hoc reset logic.
  Integrate deadline polling with the **existing shared Go worker module**; do not create a goroutine or timer for
  every correlation key.
- Separate **detection window**, **post-trigger cooldown/notification suppression**, and **inactive instance expiry**.
  Define clock semantics, restart behavior, timeout processing, and treatment of observations arriving near time
  boundaries.
- Serialize state updates for the same instance/correlation key; allow safe concurrency between unrelated keys. Bound
  key cardinality, cache size, payload size, and expiry to resist memory exhaustion from adversarial inputs.
- Support keyed instances with explicit correlation scoping (e.g. rule, host/source, UID/GID and account namespace where
  applicable); missing/invalid keys follow a documented safe rule, not silent collisions.
- Track detection/alert state separately from the external Security Logging Manager lifecycle and from outgoing
  NATS/email side effects.

### 5.4 State persistence, effects, and failure handling

- Inspect existing state-storage conventions and choose/document whether counters/cooldowns are volatile or durable;
  **do not pretend in-memory state survives restart**. For high-impact lockdown or recovery effects, protect against
  replay/duplicate commands after restart using available incident identifiers, deduplication, and existing
  infrastructure guarantees.
- FSM transitions SHOULD produce **typed effect intents** (create/update event, email, auth lockdown request) that
  dedicated adapters deliver. Do not run CLI commands or invoke privileged operations from FSM guard functions.
- Ensure transition state and side-effect dispatch have defined failure/retry behavior. If atomic commit plus effect
  delivery is unavailable, document and test idempotency/reconciliation; never claim exactly-once delivery without
  supporting guarantees.
- Reject invalid definitions at initialization and surface transition/guard/action errors with rule ID, plugin ID,
  instance key (appropriately redacted), previous state, input type, and destination where safe.
- Include human-readable table dumps or diagnostics to support tracing, debugging, and team demos without exposing
  sensitive data.

### 5.5 Required engine/plugin tests and demonstrations

- Immediate, global-threshold, and keyed-threshold plugins behave independently and deterministically.
- Guards, deliberate nonmatch/sequence break, reset, error, ambiguous table, invalid function/state reference, timeout,
  cooldown, and event update semantics are covered.
- Test at threshold boundaries, duplicate observations, out-of-order input where relevant, restart/reset behavior,
  concurrent same-key requests, high-cardinality keys, and failed effect delivery.
- Demo the declarative transition table, named guard/action dispatch, shared worker deadline expiration, and
  logging/event lifecycle adaptation.

## 6. FSM adoption in existing monitoring services (TASK-FSM-002 — mandatory)

After TASK-FSM-001 is implemented and validated, the assignee **MUST**:

1. Inspect existing monitoring services for FSM-like state handling, repeated-state polling, counters, time windows,
   cooldowns, timeouts, and event lifecycle/correlation logic.
2. Produce a candidate inventory and assess applicability using complexity, behavioral equivalence, dependency
   constraints, and benefit.
3. **Refactor suitable existing implementations** to reuse the shared FSM package; treat this as part of normal delivery
   scope.
4. Preserve existing externally observable behavior, event/logging contracts, configuration compatibility, and
   error/retry behavior unless a deliberate, reviewed change is required.
5. Add transition-table, regression, integration, and documentation coverage for migrations.
6. Document exclusions where adoption would be unsuitable or introduce unnecessary complexity, with a concrete reason
   for each.

**Acceptance:** All suitable reviewed instances are migrated, unchanged external behavior is demonstrated by regression
tests, and any exclusions are documented. The assessment alone is insufficient if suitable candidates exist.

## 7. SVC-002 — Mount Monitor

- Detect relevant new mounts, unmounts, and changes to observed mount identity/state using an unprivileged mechanism
  compatible with existing OS/project conventions.
- Emit **informational** operational events for mounts and unmounts using the System Logging Manager.
- Broadcast mount lifecycle notifications over NATS for independent consumers, including SVC-003.
- Include enough validated metadata for consumers to identify the mount and determine applicability, subject to the
  project's existing event contract (e.g. location, source, filesystem type, stable or best-available identity,
  timestamp).
- Avoid duplicate events when repeated observations show no state change.
- Determine startup baseline and restart/reconciliation behavior; distinguish existing mounts from newly observed ones
  when possible.
- **Do not** perform antivirus scanning or acquire mount/unmount privileges.

## 8. SVC-003 — Antivirus

- Subscribe to SVC-002 mount lifecycle notifications over NATS; **do not** implement a second general-purpose mount
  monitor.
- Select mounts eligible for scanning by configuration and validated mount properties. Account for already-mounted
  filesystems and missed notifications through an agreed initial reconciliation policy.
- Inspect existing ClamAV deployment; choose and document `clamscan` / daemon-compatible integration according to
  project conventions. Do not assume both are available.
- Execute a controlled, bounded scan; track queued/running/completed/failed/cancelled states and react safely to
  unmounts or changed mount identity.
- Report malware findings as security events through the existing lifecycle. Report scan failures/incomplete scans
  separately from clean scans. Operational scan progress and result handling must follow existing logging conventions.
- Send malware-detection email using the existing security alert notifier path; deduplicate notifications per incident as
  appropriate.
- Prevent duplicate simultaneous scanning of the same eligible mount. Bound concurrency and resource use.
- Do not automatically remove/quarantine files or unmount filesystems without a separately approved requirement.
- Check read access, symlink traversal, mount replacement races, filesystem crossing, scanner flags, path-as-option
  injection, and limitations of host read permissions. Keep scanner privilege minimal.

## 9. SVC-004 — System Service Monitor

- Periodically inspect **configured systemd services** using the existing shared Go polling worker; polling interval is
  configurable.
- Use ordinary unprivileged systemd inspection (`systemctl` or established project integration). **Never sudo for
  polling**.
- Differentiate as accurately as systemd permits: active, inactive/stopped, failed, auto-restart in progress, persistent
  failure/start-rate-limit condition, and restored service.
- Report service-down/persistent-error conditions to the **System Logging Manager at error level** and publish
  relevant state changes over NATS.
- Include the observed condition, unit name, relevant systemd state/result fields, timestamps, and available restart
  counts. Avoid asserting an intentional user stop or root cause unless evidence actually supports it.
- Respect systemd's configured restart handling; detect when its attempts have failed or rate-limits prevent recovery
  instead of immediately competing with automatic restart.
- Correlate repeated polling observations into state transitions; avoid error/event floods; emit recovery signals.
- Do not start/stop services or require privileged service-management access.

## 10. SVC-005 — Service Watchdog

### 10.1 Configuration and recovery

- Maintain its **own configuration** identifying protected/critical systemd units, permitted recovery actions, polling
  interval, notification criticality, and retry/backoff limits.
- Listen to SVC-004 service-down/persistent-failure notifications on NATS and independently check a protected unit's
  current state before taking action.
- Periodically reconcile the configured protected set, independently of NATS notifications, so missed messages do not
  permanently suppress protection.
- Defer to systemd's normal `Restart=` policy while it is operating. When recovery is exhausted or the service remains
  unavailable, attempt **controlled, narrowly authorized** recovery, then verify outcome.
- Track **systemd restart attempts** separately from **Watchdog attempts**. Correlate failures and update the
  appropriate incident/event with failed-attempt counts and eventual recovery, rather than create one new alert per
  attempt.
- Limit retry loops; use backoff. Investigate systemd start-limit behavior and do not blindly repeat `reset-failed` or
  use it to bypass limits without an explicitly approved policy.

### 10.2 Critical-service enablement protection

- Independently check whether protected services maintain their **expected enablement state**. Understand systemd
  `enabled`, `disabled`, `static`, `indirect`, and `masked` distinctions and each unit's intended policy.
- Report unexpected disabling/masking or comparable interference to the **Security Logging Manager** with meaningful
  factual evidence; notify by existing email mechanism for services configured as critically protected.
- Identify observed change without automatically attributing malicious intent or assuming who changed it.
- Separate policy decisions for a manually stopped service, a failed service, and an unexpectedly disabled service.
  **Automatic re-enablement** is not authorized by this draft; make it an explicit future decision, mindful of planned
  maintenance.
- Watchdog must not trust a NATS payload as authorization to operate on a unit; cross-check its own protected-unit
  allowlist.

### 10.3 Privilege boundary

- Use unprivileged commands for status and enablement checks.
- If starting a protected unit genuinely requires escalation, use the project's existing secure mechanism or a
  **dedicated passwordless `/etc/sudoers.d/` rule restricted to exact permitted commands and specific unit
  arguments**. No shell access, blanket `systemctl`, wildcard bypasses, or broad service-management group access.
- Restrict execution paths, argument vectors and target allowlists; verify that the systemd/AppArmor policy permits only
  the intended action.
- Do not grant broad root/service-management access just to simplify Watchdog implementation.

## 11. Mandatory security baseline (all services)

### 11.1 Dedicated identities and confinement

- One **dedicated non-root system user and primary group per service**. Inspect existing project-specific and
  security-related groups and reuse established conventions when appropriate. Add only essential supplementary groups.
- Provision least-privilege ownership, mode bits, read/write access, log and runtime directories, NATS credentials,
  scanner resources, and sockets.
- Supply and enforce **service-specific AppArmor profiles** plus **systemd sandboxing**. Restrict capabilities,
  namespace/network access, writable paths, device access, privilege escalation, and access to protected host
  resources as appropriate for each unit.
- Test actual production identities **under both protections**; document unavoidable exceptions and the reason for each
  permission.
- Watchdog and Antivirus privilege edges require separate threat review; avoid running either as root.

### 11.2 UNIX commands, sudo and audit noise

- Reuse the project's CLI helper/conventions; execute ordinary UNIX commands as the service account.
- Use Go's argument-based process execution (`exec.CommandContext` or shared equivalent), explicit trusted executable
  path, execution deadline, controlled environment, and validated arguments.
- Do **not** interpolate log/NATS/config/mount data into `sh -c`, `bash -c`, or constructed shell command strings.
- A separated argument vector is **not sufficient** if user-controlled data may become an option or unauthorized target.
  Validate option syntax and use safe option termination where supported.
- Prefer narrowly scoped UNIX permissions / dedicated group membership before considering sudo. **Sudo MUST NOT be used
  for polling**; it creates avoidable privileged execution and audit noise.
- If sudo is unavoidable, create an explicitly reviewed passwordless sudoers drop-in specifying exact executables,
  operations, and arguments; validate with `visudo`. No unrestricted shells, wildcard command patterns, or unbounded
  privileged operations.
- ClamAV file access may need special consideration. Evaluate dedicated scanner account/group and existing
  daemon/service arrangements **before** contemplating a privileged scan process.

### 11.3 Input validation and trust boundaries

Validate before sensitive use, even with no user-facing API:

- Config files: schema, types, required fields, durations/limits, paths, rule pattern safety, allowed actions.
- NATS payloads: message version/schema, source/correlation identifiers, size and value limits; **authorization
  independently of message contents**.
- rsyslog/audit records: parsing boundaries, malformed/truncated records, message sizes, correlation uniqueness,
  expensive expressions.
- Mount/device paths: source trust, path validation, identity/TOCTOU risks, symlinks, option injection, filesystem
  boundaries.
- systemd unit names: valid form plus explicit service allowlist before any state-changing command.
- ClamAV responses: defensively parse exit statuses/output; never execute returned strings.

Fail safely: invalid input must not trigger privileged operations or lockdown. Do not log secrets or unlimited untrusted
payloads.

## 12. Configuration requirements

- Follow the project's existing **shared Go baseline configuration** and repository file-format convention; do not
  introduce a competing configuration subsystem.
- Keep separate, clearly documented configuration for each service; Watchdog owns its protected list independently from
  System Service Monitor.
- Validate configuration before starting unsafe actions. Favor predictable, explicit defaults and safe behavior when
  optional integrations are absent.
- Keep initial Log Analysis rules readable and small; exact JSON/YAML choice and rule structure finalized after
  repository inspection and team demos.
- Explicitly document all defaults, time units, polling intervals, counters, severity mapping, scanner
  inclusion/exclusion, action flags and service allowlists.
- Handle secret NATS/email credentials via existing protected mechanisms; never embed them in committed sample configs.

## 13. NATS messaging and authorization

- Discover/reuse existing subject naming, schema/versioning, subscriber conventions, reconnect behavior and NATS
  configuration/ACL structure.
- SVC-002 publishes mount lifecycle; SVC-003 subscribes.
- SVC-004 publishes relevant service state changes; SVC-005 subscribes.
- Once auth-service implements and registers its declared lockdown RPC, SVC-001
  requests lockdown using that RPC when authorized by configuration. Align its
  subject and ACL first; do not treat the DTO declaration as enforcement.
- Update publisher/subscriber permissions, subscriptions, deployment NATS config, and tests accordingly. Give each
  service only required subjects and rights.
- Ensure consumers handle duplicate messages and re-observe authoritative host state when necessary. Determine whether
  durable delivery, replay, or periodic reconciliation is used by existing infrastructure; record any trade-off.
- Treat data in messages as untrusted and never elevate its authority merely because it came through NATS.

## 14. Email notification integration

- Reuse `system-email-notifier` / `security-email-notifier` and
  `shared/go/emailnotifier` through the existing alert-event path. Check the
  actual notifier retry and deduplication behavior before promising delivery.
- SVC-001: configured critical rules. SVC-003: virus/malware detected. SVC-005: critical protected-service
  disablement/interference and any additional critical incidents explicitly configured.
- Deduplicate repeated alerts and avoid email floods. Notification failures must not suppress core event reporting.
- Do not create an independent SMTP stack without documented approval of a genuine gap.

## 15. Build, Debian packaging, deployment, and GitHub Actions

The feature is incomplete until integrated into the **existing** delivery workflows.

- Update Go build/test scripts for all five services using project conventions.
- Extend the existing main `lite-nas` `.deb` build scripts and package metadata
  for these Go services; update binary installation, service units, configuration
  defaults/examples, AppArmor profiles, directories, group/user provisioning, and package lifecycle hooks.
- Update deployment/install/upgrade scripts to provision users/groups and permissions **idempotently**; install valid
  security profiles and systemd units; reload policy; enable/start services according to existing deployment policy;
  verify running state.
- Update GitHub Actions workflows for compile, automated tests, lint/static checks if present, packaging, and suitable
  install/upgrade/security-validation jobs.
- Preserve administrator-managed configuration and existing persistent event/runtime state across upgrades. Removal
  should not unexpectedly delete shared logs, privileged groups, or operator-managed data.
- Ensure identity/profile/directories are established **before** service startup. Keep secret distribution within
  existing deployment facilities.
- Document install, upgrade, rollback/uninstall, service ownership/permissions, configuration updates, and
  troubleshooting.
- Test real package installation and upgrades (where CI supports them), not only local Go builds.

## 16. Testing and acceptance criteria

### 16.1 Cross-service acceptance

- [x] TASK-000 completed before the remaining implementation: mandatory `AGENTS.md` instruction, shared worker/NATS
  reuse, typed bounded Go channels, fan-out, and clearly documented backpressure/shutdown.
- [ ] Each service executes as one process with multiple internal goroutine workers where applicable, and inter-service
  communication uses the existing NATS package.
- [ ] SVC-001 constructs every configured plugin instance once at startup and injects its map/collection into listeners;
  one log file feeding two or more independent categories/rules/plugins demonstrably reaches all intended plugin
  instances.

- [ ] All five services compile, run, stop gracefully, and use the shared worker/config/CLI conventions where
  applicable.
- [ ] Existing manager and NATS contracts are used; events pass correct lifecycle/create/update/correlation behavior.
- [ ] **Event lifecycle (TASK-EVENT-LIFECYCLE-001):** mount, antivirus, system-service-monitor, and watchdog incidents
  use stable per-source correlation and source-specific normalization; occurrence rows are reserved for FSM paths that
  intentionally aggregate multiple observations. ZFS/network `resources-monitor` lifecycle behavior is inspected and
  documented as the comparison baseline.
- [ ] All run under separate non-root identities; only necessary group access is granted.
- [ ] AppArmor is enforcing and systemd sandboxing active; both normal operations and expected permission denials are
  demonstrated.
- [ ] No `sudo` in polling paths; any sudoers drop-in is exact, minimal, validated, and tested against forbidden
  commands.
- [ ] NATS publishing/subscribing permissions updated and verified with real service credentials.
- [ ] Inputs, including config and NATS payloads, are validated; malformed/hostile inputs cannot result in shell
  injection, unwanted scans, unauthorized unit starts, or lockdown.
- [ ] Temporary infrastructure failures, restarts, duplicate messages, and email failures do not create uncontrolled
  loops or alert floods.
- [ ] Debian build/install/upgrade and existing deployment/GitHub Actions pipelines cover every new service.
- [ ] User/operator/developer documentation updated; permissions and exceptions justified.

- [ ] **Shared FSM (TASK-FSM-001):** typed transition tables and registries, validations, effects, deadline worker,
  concurrency, bounded state, and transition diagnostics tested.
- [ ] **Migration (TASK-FSM-002):** existing monitoring services inventoried; all suitable FSM-like implementations
  migrated, with regression tests and justified exclusions.

### 16.2 Feature acceptance

- [ ] Pipeline tests verify channel backpressure, worker cancellation, fan-out semantics, NATS adapter behavior, and no
  silent security-event drops.
- [ ] Plugin tests verify duplicate plugin types instantiated with distinct settings, unknown/missing instance
  references fail at startup, and listener never constructs plugin types.

- [ ] **Log Analysis:** all three plugin behaviors, named table-driven transitions, timeout/reset and keyed correlation
  demonstrated; selected rsyslog/audit inputs mapped to rules; threshold/counter expiration/correlation tested;
  ordinary login mistakes create fewer false positives than repeated probing; event updates, email, and auth lockdown
  NATS request demonstrated.
- [ ] **Mount Monitor:** mount/unmount changes produce informative operational records and NATS events, with stable
  deduplication and startup handling.
- [ ] **Antivirus:** eligible mount leads to bounded ClamAV scan; clean, infected, error and disappearing-mount cases
  distinguished; malware event/email integration demonstrated.
- [ ] **System Service Monitor:** running, user-stopped/unknown-stop, failed, restarting, restart-exhausted and
  recovered states appropriately classified and reported, with no sudo polling.
- [ ] **Watchdog:** only allowlisted protected services recovered; normal systemd autorecovery not interfered with;
  retry limits/backoff and correlated failed-attempt updates demonstrated; unexpected disablement produces security
  event and critical email.

### 16.3 Suggested demo sequence

1. Repository reconnaissance and contract/identity/packaging walkthrough.
2. Existing monitoring FSM candidate inventory and shared-engine integration plan.
3. Shared FSM table/guard/action walkthrough: immediate, threshold, keyed threshold, cooldown timeout, and event update
   behavior.
4. Mount Monitor → NATS → Antivirus: scan, clean, malware and failed scan cases.
5. Service Monitor → NATS → Watchdog: systemd recovery and exhausted retries.
6. Disabled critical service → security incident/email; forbidden non-allowlisted recovery.
7. Audit-log fixtures → rule schema and thresholds; auth lockdown integration in a safe test environment.
8. Existing monitoring FSM migration and regression demonstration.
9. Debian install/upgrade, GitHub Actions build/test, and systemd/AppArmor confinement checks.

## 17. Implementation order (TASK-000 is a hard prerequisite; subsequent work can be staged)

0. **Completed — TASK-000:** worker/NATS walkthrough, root `AGENTS.md` rules,
   developer diagram, and channel-lifecycle guidance. No common adapter gap was
   demonstrated.
1. Finish service-specific inventory and map remaining contracts. Implement
   auth-service lockdown and align its NATS ACL before enabling SVC-001 lockdown.
2. Wire common config/worker/NATS/logging/email integrations and security-account/systemd/AppArmor conventions.
3. Implement SVC-002, then SVC-003 integration.
4. Implement SVC-004, then SVC-005 integration.
5. Implement shared FSM engine and its three detection behaviors (TASK-FSM-001); use it in SVC-001 with sample audit
   records, rule-format demos, and the verified auth lockdown integration.
6. Refine event correlation and lifecycle per source for SVC-002–SVC-005 (TASK-EVENT-LIFECYCLE-001), using the
   verified `resources-monitor` lifecycle and ZFS/network source behavior as comparison points.
7. Assess and migrate applicable FSM-like logic in existing monitoring services (TASK-FSM-002); verify behavioral
   compatibility.
8. Extend full packaging, pipelines, deployment, security and integration tests alongside each service; finish with
   end-to-end demos and documentation review.

Do not defer all deployment/security work until the end: each service should be validated in the expected confinement as
soon as it becomes runnable.

## 18. Required repository investigation checklist

The assignee **MUST** document the findings before implementation decisions are finalized:

- [x] Existing root/nested `AGENTS.md` and worker pipeline rules; update with mandatory single-process/multi-worker
  channels and plugin instance injection (TASK-000).
- [x] Existing shared worker/polling APIs, channel ownership/fan-out/backpressure, NATS channel bridges and reusable
  extension points.
- [ ] Existing state stores, FSM/correlation patterns, and monitoring-service migration candidates; document suitable
  and unsuitable consumers.
- [ ] Existing shared config baseline and dominant file formats.
- [ ] Existing CLI execution helpers and safe command conventions.
- [ ] Logging managers' concrete APIs, event types, lifecycle and update semantics.
- [ ] NATS subjects, payload schemas, client/permissions configuration, delivery semantics, and subscriptions.
- [ ] Authentication Service lockdown: declared subject and DTOs identified;
  implement and verify the missing RPC handler, enforcement and ACL alignment.
- [x] Existing email module/service integration (`shared/go/emailnotifier`,
  `system-email-notifier`, `security-email-notifier`).
- [ ] rsyslog routing, audit-log format and representative authentication records.
- [ ] ClamAV deployment and permissions/scanner execution model.
- [ ] Existing service users, primary/supplementary groups, group naming and provisioning approach.
- [ ] systemd unit examples and sandboxing requirements; operational-service list and restart policies.
- [ ] AppArmor conventions, deployment/activation and testing method.
- [ ] Restricted sudoers patterns, if any, and safest existing unit-recovery mechanism.
- [ ] Build, `.deb`, deploy/install/upgrade scripts, and GitHub Actions workflows.
- [ ] Existing docs and standard test environments.

## 19. Decisions, proposals and open choices

### 19.1 Confirmed architectural choices

- **AD-001:** New services adapt to established security/operational logging and event-lifecycle contracts; state/FSM
  only as needed.
- **AD-002:** Mount Monitor and Antivirus are separate NATS-connected services.
- **AD-003:** System Service Monitor and Watchdog are separate observation and recovery/protection services.
- **AD-004:** Five narrowly scoped services constitute the initial security skeleton.
- **AD-005/006:** Reuse/extend shared Go polling workers and existing configuration infrastructure.
- **AD-007/008/009:** Distinguish observable systemd statuses, let systemd autorecovery lead, use controlled Watchdog
  recovery and correlated attempt tracking.
- **AD-010:** Dedicated non-root users/groups and both systemd sandboxing and AppArmor confinement.
- **AD-011:** Integrate with existing builds, Debian packages, deployment scripts and GitHub Actions.
- **AD-012:** Unprivileged UNIX/systemd inspection; sudo only for exact, exceptional authorized operations, never
  polling.
- **SEC-003:** Validate all trust-boundary inputs and defend against shell/argument injection.
- **AD-013:** Lightweight configurable auth-failure rules; authorized lockdown
  requests use the declared auth-service RPC after its implementation and ACL
  alignment; rule schema evolves through demos.
- **AD-014:** Shared typed table-driven FSM library used *inside* statically compiled plugin workers; configured plugin
  **instances** constructed at startup, injected into listeners as map/collection, log-source/category/rule routing
  and channel-based detection outputs; named guards/actions, keyed state, deadline inputs.
- **AD-015:** Mandatory assessment and migration of suitable existing monitoring logic to the new FSM engine; only
  justified unsuitable candidates may be excluded.
- **AD-016 / TASK-000 (complete):** Each new service is one OS process with
  workers connected via typed/bounded channels where independently scheduled
  work requires them. Reuse existing worker and NATS communication packages;
  validated handlers and outbound workers bridge NATS. The mandatory rules are
  in `AGENTS.md`.

### 19.2 Implementation choices intentionally unresolved

1. Exact rsyslog transport/file selection, audit parsing/correlation, and detection thresholds. The shared worker and
   messaging APIs are mapped in TASK-000; a generic shared NATS channel adapter does not exist.
2. JSON/YAML/other rules file and full schema (follow existing project + team demos).
3. Whether event state is in-memory or must be persisted to ensure correct replay/update behavior, especially across
   cooldown, restart, and high-impact action effects.
4. Mount observation method, identity keys, baseline handling, and whether mount events require durable NATS delivery.
5. Antivirus scan engine mode, scope, resource bounds, and access to files not readable by the service user.
6. Exact service stop/restart exhaustion classification supported by the deployed systemd version and available
   properties.
7. Watchdog behavior on manually stopped services, authorized maintenance, auto re-enable policy, and restricted
   recovery command mechanism.
8. Delivery guarantees of the existing email notifier path, especially retry
   and deduplication for new security events.
9. Remaining service-specific NATS subjects/ACLs and manager lifecycle details.
   Auth lockdown constants exist, but the RPC handler and matching ACL do not.

**Rule:** Open choices must be resolved from code and evidence, with documented rationale and team review where
meaningful. They must not be silently invented or treated as permission to weaken security.

## 20. Definition of done

Work is complete only when **TASK-000** and its `AGENTS.md` instructions are implemented, worker/NATS channel topology
and instantiated-plugin injection are demonstrated, the shared table-driven FSM engine and three detection behaviors are
implemented, suitable existing monitoring services are migrated with regression coverage or exclusions justified, and
all five services are implemented, tested and deployed through existing project workflows; their NATS/logging/mail/auth
integrations operate end-to-end; rule demonstrations and operational recovery cases pass; each process runs as its
dedicated non-root user with active systemd/AppArmor restrictions; privileged paths are justified and narrowly scoped;
installation and upgrade work safely; GitHub Actions includes relevant build/test/package checks; and operator/developer
documentation describes configuration, permissions, recovery behavior, diagnostics, limitations, and any intentionally
deferred choices.

---

**Review note:** This document intentionally does not fabricate repository API names or prescribe production thresholds
without real log samples. During review, decisions marked open may be converted into firm requirements.
