# Operational Safety Automation feature design

**Status:** The service, gateway operator routes, local build/deploy flow, linked CPU stage FSM, and CPU
frequency helper are implemented. No operational safety rules are packaged or enabled. The [production rule
guide and four-rule example](../docs/operational-safety-cpu-rules.md) use the packaged monitor event prefixes.
CPU policy changes have been tested against simulated sysfs policies, not yet against a deployed Raspberry Pi
or PC. The existing Security Automation CLI provides a direct, verified manual cap command after a restart; a
separate Operational Safety CLI is not implemented.

## Packaged configuration

The default `/etc/lite-nas/operational-safety-automation.conf` names
`/etc/lite-nas/operational-safety-automation/rules.json`. Its packaged `rules` array is empty. The backup spec
selects the config and entire rules directory. Config deployment preserves an installed rules directory across
updates. Startup accepts an empty rule set and leaves the host CPU policy unchanged.

## Purpose

Operational Safety Automation is a separate Go service that observes validated system events, evaluates
configured automation FSMs, applies a verified host effect, and publishes a system alert describing its own
decisions and effect state. Its emitted alerts are ordinary, normalizable system alerts so other configured
FSMs can correlate them. A timer worker polls the existing System Logging Manager `getAlert` RPC for tracked
event IDs; the FSM advances recovery only after it observes each required alert in the `normal` state. It
reuses the automation detector plugins in `shared/go/automationdetection` and the table driven FSM in
`shared/go/fsm`.

The service is one OS process with goroutine workers. NATS callbacks validate alert creations and feed bounded
channels; a timer worker polls tracked alert IDs for normalization and cooldown deadlines; a dispatcher owns
compiled FSM instances; an outbound worker handles effects and alert publication. Compile configured instances
at startup. Rules select reviewed effect profiles and cannot provide arbitrary executables or command
arguments. No example in this document is a shipped default rule.

## Example: CPU frequency protection

A configurable rule may combine power and CPU temperature observations for one machine. For the reported
Raspberry Pi 5, the example requests an 1800 MHz warning cap, a 1600 MHz error cap, and a 1500 MHz critical
cap. The most restrictive active request wins. A more severe observation applies its tighter limit promptly.
The example's normalized state uses the operator-selected 1900 MHz cap.

The linked detector is implemented in `shared/go/automationdetection/chained_stage.go`. It returns uncommitted
effect intents; the owning service must apply and verify each cap before storing the next FSM state. Recovery
is staged. A completed cooldown may relax only one level, and only after all triggers requiring the current
level have normalized. From a 1500 MHz critical cap, the next recovery cap is 1600 MHz, then 1800 MHz, then
1900 MHz. A transition from critical through error and warning to the 1900 MHz normalized cap therefore takes
three successful cooldowns. If a trigger returns or grows more severe during a cooldown, the FSM must account
for it before any relaxation. The interval is configured per rule. An effect level is confirmed only after the
host command and verification succeed; a failed command remains visible rather than being reported as a
successful transition. The older `staged.go` detector remains available for `cpu_frequency` rules.

The example describes supported configuration and behavior; it does not enroll a CPU rule. The implemented
helper writes and verifies the Linux `cpufreq/policy*/scaling_max_freq` attributes after validating each
policy's limits. Actual Raspberry Pi behavior and restoration remain unverified.

## Operator surface

Authenticated NATS RPCs and a companion CLI must let an operator list configured rules and active tasks,
temporarily enable or disable a rule, force a configured task or effect, and stop a task. The CLI help and
reference must describe command arguments, temporary versus persistent state, effect verification, and
recovery behavior. Mutating requests require the same authorization, audit, input validation, and per-instance
serialization patterns as Security Automation. Force set applies a rule-declared level to an existing task
identified by task ID. Force stop starts stepwise recovery of that task. It must reject a relaxation while a
power or temperature alert still requires the current cap. Both commands use the same cooldown and
verification path as automatic recovery. RPC inputs and responses must carry a validated task ID and report
blocked, pending, completed, or failed outcomes.

## Implemented contracts

### Linked-stage requirements

- **FR-CHAIN-001:** An error or critical stage waits for its configured parent's active published event.
- **FR-CHAIN-002:** A stage restores its parent cap only after its source alerts clear and its cooldown expires.
- **FR-CHAIN-003:** A new bad value cancels the current cooldown and requires a fresh full interval after clearing.
- **FR-CHAIN-004:** An active child stage prevents its parent from recovering.
- **FR-CHAIN-005:** A stage normalizes its owned event only after its recovery cap is verified.
- **FR-CHAIN-006:** A direct critical sensor alert applies warning, error, then critical caps through the event chain.
- **FR-CHAIN-007:** Recovery verifies critical, error, then warning parent caps in that order and records the normal state.
- **FR-CHAIN-008:** A returned bad value during cooldown leaves the owned event active until the renewed cooldown completes.
- **FR-CHAIN-009:** A failed cap restoration leaves the stage and its owned event active for retry.
- **FR-CHAIN-010:** A failed normal-record normalization remains queued for retry.
- **FR-CHAIN-011:** Task force-stop cannot bypass a source or child alert that still requires the stage cap.
- **FR-CHAIN-012:** After restart, an authorized operator can directly request a verified cap through the CLI
  while no in-memory task exists.
- **FR-CHAIN-013:** All active power and temperature sources must clear before the shared warning stage starts recovery.

Rules use a validated `steps` array. Each step selects an `id_prefix` and may combine `message.equals`,
`message.contains`, `message.regex`, numeric/text/boolean occurrence `value` conditions, and a lifecycle
`state`. A rule with `action.kind` set to `emit` uses the shared `sequence` or `count_within_window` FSM and
publishes a system alert. A rule with `action.kind` set to `cpu_frequency` uses the staged recovery FSM and a
configured cap for each of `normal`, `warning`, `error`, and `critical`. Its single step specifies the
requested protective `level`. Rule output config supplies `id_prefix`, `category`, `severity`, and `message`;
an emitted alert can be matched by another rule through the same system-alert subscription. Publication cycles
are rejected at startup.

The production chain uses `cpu_stage` for each of warning, error, and critical. Each stage has its own
`active_khz`, `normal_khz`, `cooldown`, and (except warning) `parent_rule`. Its `steps` are alternative source
alert selectors. The stage waits for its parent's published active event before applying its own cap, keeps
its event active while a child stage is active, and resets cooldown when a bad value returns. It normalizes
its owned event only after its recovery cap is verified. `emit_normal` records and normalizes the final
verified normal state. A direct critical sensor alert descends through all protective stages because warning
and error also select more severe sensor events. The documented four-rule example is tested through the
service effect worker with a virtual clock.

For example, an `emit` rule may match `{"id_prefix":"power","message":{"contains":"undervoltage"}}` and
publish an event with output prefix `opsafe`. A CPU rule can then match
`{"id_prefix":"opsafe","message":{"equals":"Power condition"},"level":"warning"}`. A selector that needs all
three input dimensions can use
`{"id_prefix":"thermal","message":{"contains":"CPU"},"value":{"num_gte":80},"state":"active"}`; it matches
once those fields have been observed for the same event ID, regardless of whether the occurrence or state
update arrives first. Output prefixes must be unique among all publishers using the System Logging Manager.

Output event IDs use the configured prefix and a persisted, bounded sequence in the System Logging Manager's
accepted `<prefix>_<sequence>` shape. The source field is `operational-safety-automation`. Rule state changes
and verified CPU transitions also publish system alerts. Alert lifecycle state is polled through
`system-logging-manager.getAlert`; normalization starts the configured cooldown.

The service provides security-role NATS RPCs for rules, tasks, temporary rule enablement, force set, and force
stop. The gateway exposes them at `/api/system/automations/rules`, `/api/system/automations/tasks`,
`/api/system/automations/rules/{rule_id}/enabled`, `/api/system/automations/tasks/{task_id}/force-set`, and
`/api/system/automations/tasks/{task_id}/force-stop`. Force set and stop target an existing in-memory task ID.
A relaxation is blocked while an active alert needs a stronger cap, and recovery proceeds one level per
verified cooldown.

CPU caps are expressed in kHz. The fixed privileged helper discovers every `cpufreq/policy*`, checks its
hardware and configured minimum limits, writes `scaling_max_freq`, reads it back, and attempts rollback if any
policy fails. A first application may only tighten an existing host limit; a later relaxation requires the
policy to equal the service's last verified cap. Rule caps must fit each deployed host, so the example's 1900
MHz normal cap is not a portable default.

Task and effect state stays in memory. Startup does not restore or reconcile tasks and does not change an
existing CPU cap. After a restart, operator task actions apply only to new tasks; the host cap may require
manual review and a direct operator cap request.
Manual host frequency changes can use the standard Linux `cpupower frequency-set --max FREQUENCY` command,
followed by `cpupower frequency-info` to inspect the policy. The CPU rule guide documents these commands and
the supported Pi 5 caps.
