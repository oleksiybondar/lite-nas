# LiteNAS Security Alerts — Findings and Fixes for Development

Environment: LiteNAS 0.16.0~beta.127.4cb3ea7
Repository: oleksiybondar/lite-nas
Branch: security-monitor-skeleton
Production: Raspberry Pi 5, Ubuntu 26.04.1 ARM64
Investigation date: November 3, 2026

1. Confirmed bug: Service Watchdog generates invalid event IDs
Confirmed source defect
File: services/service-watchdog/runtime.go

Current code:
digest := sha256.Sum256([]byte(effect.kind + ":" + effect.unit.Name))
id := "W" + hex.EncodeToString(digest[:8])

This produces IDs such as W4a6f3d1e8c2b0975.

However, Security Logging Manager validates event IDs against:
^(?:[A-Za-z0-9_]{1,10}_[0-9]{1,8}|t[0-9]{1,19})$

Therefore, watchdog alerts are rejected.

Fix already applied to the production source checkout, but not compiled or deployed:
digest := sha256.Sum256([]byte(effect.kind + ":" + effect.unit.Name))
id := fmt.Sprintf(
    "t%019d",
    binary.BigEndian.Uint64(digest[:8])%10_000_000_000_000_000_000,
)

Replace encoding/hex with encoding/binary.

Development action: Apply and test this change in the development repository. Add regression tests covering watchdog
alert creation, occurrences, and state normalization. All three operations must use the same valid deterministic event
ID.

## 2. Confirmed production failure: Security alerts rejected by validation

Confirmed runtime failure
Production logs show:
2026-11-03T21:22:06.111Z
security-automation:
AlertPayload.EventID failed loggingmanager_event_id validation

2026-11-03T21:22:06.113Z
security-logging-manager:
CreateEventInput.EventID failed loggingmanager_event_id validation

Both services rejected an alert at almost exactly the same time.

This occurred immediately after restarting the security producers.

Likely cause: Service Watchdog generated an invalid event ID.

Not yet proven: The receiving services do not identify the original sender or invalid ID in their logs.

Development actions:

- Add safe structured diagnostics for rejected security alerts, including the event ID and claimed source.
- Add integration tests for actual publisher payloads, not only internally constructed DTOs.
- Ensure all security-alert producers use one shared, validated event-ID generator or compatible deterministic
  generation logic.

1. Confirmed historical failure: Service JWTs lacked sufficient roles
Previously observed
Security Logging Manager repeatedly reported:
messaging: handler failed: insufficient role

Affected subjects included:
security-alert
security-alert-occurrence
security-logging-manager.updateAlertState

System Logging Manager also reported the same failure for:
system-logging-manager.updateAlertState

Relevant source files:

- services/auth/runtime.go
- services/rbac/runtime.go
- shared/go/loggingmanager/nats/auth_middleware.go
- shared/go/loggingmanagerauth/client.go
The role middleware checks roles contained in service JWTs.

Production Unix accounts already have the required groups:

- Resource monitor: lite-nas-operator
- Security producers: lite-nas-security
Therefore, missing Unix group membership is not the demonstrated cause.

A possible contributing issue is in services/auth/runtime.go: service-role resolution uses an RBAC request with a 300
ms timeout. If resolution fails, a token may be issued without roles.

Another relevant behavior is in shared/go/loggingmanagerauth/client.go: CurrentToken() returns an existing token
without checking whether its expiry is approaching.

Development actions:

1. Verify whether RBAC timeout or lookup errors can result in a successfully issued service JWT with missing roles.
1. Consider failing closed on role-resolution errors instead of silently issuing an underprivileged token.
1. Ensure token refresh handles role changes and impending expiry.
1. Test token renewal and authorization for both operator and security service accounts.
1. Improve error logging without exposing JWT contents.
Important: The precise reason for the historical insufficient role errors remains unproven.
1. Confirmed historical failure: NATS reported no responders
Previously observed
The log-analysis service detected a real event:
rule=usb_device_connected
event_id=secusb_93716312
error=messaging: publish failed:
nats: no responders available for request

That ID is valid.

Other services, including antivirus and watchdog, reported similar NATS availability errors during the earlier
interruption.

Development actions:

- Test startup and restart ordering between NATS, Auth, RBAC, Security Logging Manager, and producers.
- Test behavior when a producer publishes while the logging manager is unavailable.
- Verify retry behavior and whether queued alerts survive a producer shutdown.
- Ensure alerts are not silently marked delivered merely because a transport publish returned success.
Important: These errors were historical. There is no evidence they are still occurring under the latest production
processes.

1. Confirmed: Security Logging Manager works when given a valid alert
Verified working
A controlled CLI test successfully published:
event_id=secdiag_12345678
category=security.diagnostic
severity=info

The alert appeared in SQLite:
Total security events: 1
Diagnostic event stored: YES

Before the test, the database contained zero events. After subsequent producer restarts, it still contained one event.

This proves that the following path can work:
CLI

NATS

Security Manager

SQLite

Development implication: Do not redesign database storage or TLS based on this investigation. Focus on producer
contracts, authorization resilience, and delivery reliability.

## 6. Producer-specific findings

Producer Finding Development priority
Service Watchdog Invalid event-ID format High — fix
Log Analysis Valid IDs, previously failed to publish during NATS outage High — test delivery/retries
Network Security Valid shared ID generator, no observed publication errors Medium — integration test
Antivirus Valid secav_######## format, historical auth connectivity failures Medium — integration test
Security Automation Rejected invalid incoming event; its own generated ID format appears compatible Medium — test
incoming validation and audit publishing

The production log-analysis configuration also has nine enabled detection instances. It is not generally disabled.
Recommended development work order

P0 — Fix watchdog event IDs

Apply the already identified patch and add tests for create, occurrence, and normalization operations.

P1 — Make service-role resolution reliable

Investigate the 300 ms RBAC lookup timeout and ensure token issuance does not silently lose required service roles.

P1 — Verify NATS delivery resilience

Cover logging-manager outages, producer reconnects, retry queues, and loss during shutdown.

P1 — Add cross-service integration tests

Generate a valid event from every producer and verify that Security Logging Manager persists it with the expected
state.

Release acceptance criteria
A development build should not be deployed until:

- All security producer-generated event IDs pass the shared validator.
- Security alerts from each producer are stored successfully.
- Service JWTs contain the expected roles, including after renewal.
- NATS interruptions do not silently lose pending alerts under the tested recovery conditions.
- Security Automation accepts valid producer alerts.
- Both system and security alert lifecycle RPCs pass authorization tests.
- Tests run on the target ARM64 architecture or an equivalent validated environment.
Bottom line: We found one confirmed source defect (watchdog event IDs), two historical reliability/authorization
failures requiring investigation, and one working baseline (CLI → Security Logging Manager → SQLite). The other
producers are not proven defective solely because they have not emitted fresh alerts.

do we have deb construction installation isseu identified?

Not conclusively. From our production investigation, we have identified application-level defects, but we have not
demonstrated a specific .deb construction or installation defect.

What we know is:

- LiteNAS installed and its main services start successfully.
- The security logging manager accepts and stores a valid diagnostic alert.
- The installed watchdog has an event-ID incompatibility also present in the source branch we inspected.
- Service authorization and NATS connectivity failures occurred, but we have not established whether packaging or
  installation caused them.
- The attempted .deb build on the Pi did not produce a package because the development toolchain was absent. That is
  not evidence of a defective Debian build.
One packaging-related concern remains worth checking: whether the Debian package creates the correct service groups,
assigns service accounts to those groups, and starts services only after their dependencies are ready. The production
group memberships we checked were correct, so group creation is not a confirmed installation failure.
What to verify in the Debian build and installer
Area Status Development check
.deb compilation Not verified Confirm CI builds all services successfully for ARM64
Packaged binary versions Not verified Confirm the .deb contains binaries built from the intended commit
Systemd unit installation Mostly working Check installed units against packaged units
Service accounts and groups Verified on production Add automated post-install permission tests
NATS permissions Previously matched package configuration Add publish/subscribe integration tests
Service startup dependencies Suspected area Test fresh installation and restart ordering
JWT role initialization Suspected area Verify RBAC is ready before service tokens are issued
Configuration deployment Not verified Confirm upgrades preserve existing settings

The highest-value packaging test is a clean ARM64 virtual-machine installation of the exact .deb, followed by an
automated test that publishes alerts from each real service identity.

That would reveal whether installation ordering, RBAC initialization, or package contents contribute to the failures.

Conclusion: We have a confirmed application bug and possible installation/startup issues, but no confirmed .deb
construction or installation bug yet. I would classify packaging as requiring verification, not identified as broken.

hmm ih i can have a dev machine... can we fix issues there

Yes. And that is exactly how I recommend fixing LiteNAS. We can do the development, rebuilding, packaging, and testing
on a separate machine while keeping your production Raspberry Pi untouched.

We already have your GitHub repository, branch, and concrete findings, so we don't need to repeat the production
investigation.
Proposed development workflow

1. Fix the application code

Start with the confirmed watchdog event-ID defect. Then investigate service JWT role resolution, token refresh, and
alert-delivery reliability.

1. Add automated tests

Test all security alert producers against the logging manager's validator, including authorization and retry behavior.

1. Build a new ARM64 .deb

Use the repository's existing Debian packaging scripts and verify that the package contains all the corrected binaries
and configuration.

1. Test installation on staging

Install or upgrade LiteNAS on an isolated development environment. Verify service startup, group membership, JWT
roles, alert creation, and SQLite persistence.

1. Prepare a production release

Produce a tested Debian package with a clear upgrade procedure and rollback plan. Deploy to production only after
staging tests pass.

What kind of development machine?
An Ubuntu ARM64 machine is ideal for reproducing the Pi's deployment environment, but an Ubuntu x86-64 development
machine can also work with ARM64 cross-compilation and a separate ARM64 test environment.

One question: Will your development machine be Ubuntu ARM64 or x86-64?

That determines how we set up the compiler, build the .deb, and run the integration tests without touching production.

arm vm

Excellent. An Ubuntu ARM64 VM is a good choice for LiteNAS development because it matches your Raspberry Pi's CPU
architecture.

We can use it to fix the application, build ARM64 Debian packages, and test upgrades without installing development
tools on production.
Recommended development environment
Component Recommendation
OS Ubuntu 26.04.1 LTS ARM64
CPU 4–8 virtual cores
RAM 8 GB minimum, 16 GB preferred
Disk 60–100 GB
Go / Node.js Install versions required by the repository
Git Clone oleksiybondar/lite-nas
Branch security-monitor-skeleton
Production Pi Leave untouched

How we'll use the VM

1. Fix the confirmed watchdog bug — invalid security-alert event IDs.
1. Investigate JWT/RBAC problems — service tokens sometimes lacked the required roles.
1. Improve NATS recovery — investigate alert delivery failures during restarts.
1. Test every security-alert producer — verify valid IDs, authorization, and SQLite persistence.
1. Audit Debian packaging — check binary contents, systemd units, permissions, and installation scripts.
1. Build and test the ARM64 .deb — including installation and upgrade scenarios.
First step on the VM
Once Ubuntu ARM64 is installed, run:
sudo apt update
sudo apt install -y git ca-certificates curl build-essential

Then clone the development branch:
mkdir -p ~/src
cd ~/src

git clone \
  --branch security-monitor-skeleton \
  <https://github.com/oleksiybondar/lite-nas.git>

cd lite-nas

git status
git rev-parse HEAD

This does not modify your production machine.

Important: We should inspect the repository's toolchain requirements before installing Go and Node.js, rather than
selecting arbitrary versions.
VM testing limitation
An ARM64 VM can validate package construction, service authorization, NATS messaging, database storage, and most
security-monitoring workflows.

Raspberry Pi-specific hardware features, such as GPIO, physical USB hotplug, and certain device monitoring functions,
will still require separate hardware testing.

Next: Set up the VM and share the output of git status and git rev-parse HEAD. From there, we can work through the
fixes systematically and prepare a tested .deb for production.
