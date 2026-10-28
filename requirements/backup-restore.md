# Backup and Restore — Feature Definition (Draft)

This draft records the agreed feature direction and repository constraints. It is
not a complete implementation contract yet: the decisions in **Open design
questions** must be resolved before the affected service, NATS, CLI, or HTTP
interfaces are implemented.
Requirement IDs use `backup-restore/<TYPE>-<NNN>` for traceability.

## First slice: confirmed direction

- Start with a compiled generic file plugin driven by application-specific
  backup specification drop-ins. A spec maps one application to its related
  files and directories at their normal absolute paths. A spec may mix
  individual files and directories from `/etc`, `/var`, or another location;
  the exact path allowlist and secret-file policy remain open.
- Each application target, including one that needs an application CLI, has a
  drop-in spec. The presence of a valid spec enables that target. The spec
  names a compiled plugin type; startup maps that name to a constructor and
  creates the configured instance before the listener starts. Unknown plugin
  names, duplicate application IDs, and invalid specs fail startup rather than
  silently disabling a target. Plugin-specific fields remain to be defined
  per type.
- Treat one backup version as a snapshot of all selected files for one
  application. Preserve each file's original path reference within that
  snapshot, so listing, diff, and restore can identify the same file. Include
  the application's drop-in spec in the snapshot so a restore can roll back
  the spec and the files it describes together.
- Define a `lite-nas` application for global package configuration. Assign
  files owned by a particular service or CLI to that application's spec rather
  than including every `/etc/lite-nas` file in the global snapshot.
- A source location under `/var` does not by itself make a live database or
  mutable application state safe to copy. Such targets need an
  application-specific export/restore plugin when plain file copying cannot
  produce a consistent snapshot.
- Use NATS through `shared/go/messaging` between the CLI and the backup and
  restore service. The CLI does not read or write privileged host files.
- Allow backup and restore operations only to administrators and security
  officers. In the current shared role model, this means `admin`, `sudo`, and
  `lite-nas-security`; `lite-nas-operator` alone is insufficient. Apply the same
  rule to inventory, version, and diff reads because those operations can
  disclose configuration and backup history.
- Expose an inventory of configured applications supported for backup, an
  inventory of applications with stored backups, and versions for a selected
  application. Keep supported targets distinct from targets with history;
  removing a drop-in does not erase stored versions.
- Support comparison of a selected stored version against current settings.
  The exact diff representation and handling of non-text or sensitive content
  are still open.
- A normal restore makes a selected directory match its stored snapshot:
  live entries within that selected directory which are absent from the
  snapshot are removed. A merge restore keeps those extra live entries.
- For a file present in both the snapshot and the live application, an
  automatic content merge should preserve current-only keys when a supported
  key-aware format permits that safely. A key present on both sides with
  different values is a conflict and must be surfaced for user resolution
  rather than silently overwritten. The CLI should let an authorized user
  resolve conflicts with a merge tool and submit the result through the
  service. INI is the first format selected for key-aware merging. Key names
  are matched case-sensitively within their sections. The merge retains live
  comments and order, appends snapshot-only keys, and sends duplicate keys or
  malformed INI to manual resolution. The exact CLI tool workflow remains
  to be defined.

The first slice excludes browser and admin-panel integration until its service
contract is settled. The agreed inventory and diff operations should remain
available to a later Web Gateway adapter without making the CLI own backup
logic.

## First slice: proposed service and CLI skeleton

The service startup loads and validates its application spec drop-ins, resolves
each plugin name against a fixed compiled registry, constructs one configured
instance per application, and then starts a coordinating listener. The listener
owns NATS request validation and authorization, resolves the requested
application to its initialized plugin, and calls the plugin operation. A file
plugin reads only spec-selected source paths and keeps application snapshots
as versioned, compressed data under the service-owned backup root. A CLI-backed
plugin implements its own validated export and restore behavior for its
application; a spec must not supply arbitrary shell commands. The user-facing
CLI validates command arguments and sends requests through the shared messaging
client.

Drop-ins live at `/etc/lite-nas/backup.d/<application>.conf`, using INI syntax.
The filename stem is the application ID. `[backup] plugin=<compiled-type>`
selects the compiled plugin constructor. For the file plugin, named
`[file.<name>]` and `[directory.<name>]` sections each contain one
`path=/absolute/path` entry and may coexist in one spec. A CLI-backed plugin's
spec contains only settings defined and validated by that plugin. Specs are
loaded and checked at startup, not interpreted from user-supplied paths during
a NATS request. The exact validation and traversal rules remain open.

The first packaged `lite-nas` spec is
`configs/etc/lite-nas/backup.d/lite-nas.conf`. It selects only
`/etc/logrotate.d/lite-nas` through the file plugin. Service and CLI configs,
NATS configuration, certificates, and mutable state are outside this first
global application snapshot and can be assigned to separate specs later.
The existing config deployment copies the packaged spec to `/etc` and
preserves an installed version of the same drop-in on later deployments.
Installed drop-ins are restricted to the LiteNAS group by the normal
configuration-permission pass.
Each stored version includes the exact drop-in used to create it. Restoring an
older version can therefore restore the older spec as well as its selected
files. The service must validate that stored spec against the compiled plugin
registry and current restore path policy before using it; an archived spec is
not an authorization bypass.

Proposed command families are `applications` (supported and backed-up
inventory), `versions <application>`, `backup <application>`,
`diff <application> <version>`, and `restore <application> <version>`.
These are review candidates, not established command names, NATS subjects, or
DTO contracts. Concrete interfaces depend on the remaining first-slice
decisions below.

The proposed restore request has a replace or merge mode. Merge mode preserves
live directory entries absent from the snapshot. Where a file exists on both
sides, the service can perform automatic merging only for formats with an
explicitly defined parser and conflict policy. For unresolved files, the CLI
can invoke a local merge tool with the material the service makes available,
then send the resolved content back through the service boundary. The service
still validates authorization, selected version, target path, and the
submitted content before applying it. The merge tool is not a privileged
restore executor.

Confirmed first INI merge rules: identify a key by its
section and key names with case-sensitive matching; retain the live file's
comments and order; add keys or sections that exist only in the snapshot; keep
keys or sections that exist only in the live file; keep an identical value
without conflict; and report different values for the same key as a conflict.
Duplicate keys, malformed INI, or constructs whose meaning cannot be preserved
without rewriting the file require manual resolution. The resolved
content becomes a new stored version as required by FR-006.

Existing CLIs authenticate with a service identity whose token carries roles.
For this CLI, deployment must also restrict access to its identity certificate
and private key to authorized host groups; NATS role checks alone would
authorize the CLI identity rather than identify the human who launched it.

## Goal and scope

Provide versioned, per-application backups and controlled restoration of LiteNAS
application configuration and state. A user can also merge a backed-up text file
with its current version, apply the merged result, and keep that result as a new
backup version. The initial target inventory below is a starting scope; exact
files, commands, and restore effects remain to be specified per target.

The feature has a Go backup and restore service, a CLI, browser-facing Web
Gateway integration, and a graphical workflow in the admin panel. The service
communicates with other services over NATS. Neither the CLI nor the gateway
should perform privileged restore work themselves.

## Confirmed functional requirements

- **backup-restore/FR-001 — Per-application history:** The service MUST keep
  separately identifiable, versioned backup copies for each configured
  application. One version is an application snapshot containing its selected
  files with their original path references. A user MUST be able to identify
  a version before restoring or merging it.
- **backup-restore/FR-002 — Generic file plugin:** A compiled, configured
  plugin MUST support backing up and restoring ordinary files and directories
  selected by an application-specific backup spec. A spec MUST be able to mix
  individual files and directories from `/etc`, `/var`, or another allowed
  location. Stored backup content MUST be compressed.
- **backup-restore/FR-003 — Application-specific plugins:** The design MUST
  permit compiled application-specific plugins where file copying cannot
  correctly export and restore application state. Such a plugin may use the
  application's CLI when its exact behavior and privilege rules are specified.
  UFW is now an initial target, but its plugin type and operations are not yet
  committed.
- **backup-restore/FR-004 — Direct restore:** A user MUST be able to select a
  stored version and request its restoration to the applicable application.
  The stored application spec MUST be restorable with the version. In normal
  restore mode, selected directories MUST be reconciled to the snapshot,
  removing live entries that are absent from that snapshot. Deletion MUST be
  limited to the selected directory roots.
- **backup-restore/FR-005 — Text merge:** A user MUST be able to compare and
  merge a backed-up text file with the current version through a graphical
  workflow, then submit the merged text for application through the backup and
  restore service.
- **backup-restore/FR-006 — Version merged content:** A completed merge MUST
  create a new backup version from the merged content held by the service,
  rather than treating the live `/etc` file as the source of that version.
- **backup-restore/FR-007 — Initial target inventory:** The design MUST account
  for LiteNAS configuration, firewall, USBGuard, AIDE, AppArmor, systemd,
  Postfix mail configuration, logrotate, rsyslog, IP routes, host network
  settings (including Netplan when present), and auditd. Inclusion does not
  imply that all targets use the generic file plugin or that their exact data
  boundaries have already been chosen.
- **backup-restore/FR-008 — All-application backup:** The coordinating listener
  MUST support an operation that visits configured application targets one by
  one. An application that is absent or uninstalled MUST be skipped so the
  operation continues to the next target. Skipping MUST NOT create that
  application's configuration or count as a backup failure.
- **backup-restore/FR-009 — Application inventories:** A caller MUST be able
  to list configured applications supported for backup and separately list
  applications for which at least one stored backup version exists.
- **backup-restore/FR-010 — Version comparison:** A caller MUST be able to
  request the differences between a selected stored version and the current
  settings for that application. The output format and eligible file types
  must be fixed before this operation is implemented.
- **backup-restore/FR-011 — Merge restore:** A user MUST be able to request a
  merge restore that preserves live entries absent from selected snapshot
  directories. INI MUST be the first supported key-aware text format; its
  automatic merge MUST retain keys found only in the current file. A key
  present in both files with different values MUST be reported as a conflict
  for user resolution; the CLI MUST support resolving conflicting files with
  a merge tool and submitting the resolved result to the service.

## Initial target inventory

| Target | Repository evidence | Scope still to define |
| --- | --- | --- |
| LiteNAS configuration | `configs/etc/lite-nas/` contains service configs, rules, templates, and certificate paths. | Which live files are included; how private keys and other credentials are handled. |
| Firewall | Deployment manages `/etc/default/ufw`, `/etc/ufw/ufw.conf`, and UFW commands that change rules. | Both file and live-rule state; safe export and restore sequence. |
| USBGuard | Installed dependency and service-watchdog protection are documented. | Policy files, generated state, and any required CLI operations. |
| AIDE | `litenas-integrity` invokes AIDE and maintains its baseline under `/var/lib/aide`. | Configuration versus baseline database, and whether baseline restore is appropriate. |
| AppArmor | Deployment manages profiles under `/etc/apparmor.d` and reloads them. | Profile selection, validation, and reload behavior. |
| systemd | LiteNAS unit files are shipped under `configs/etc/systemd/system/`. | Which unit files, drop-ins, enablement state, and reload/restart effects are included. |
| Postfix | Deployment manages `/etc/postfix` and may create restricted relay credential files. | Config and credential scope, generated maps, validation, and reload behavior. |
| logrotate | Repository ships `/etc/logrotate.d/lite-nas`. | LiteNAS-only or broader host policy. |
| rsyslog | Repository uses rsyslog output for log analysis. | Which host rules and state belong to the backup. |
| IP routes | No dedicated backup contract found. | Persistent configuration versus current kernel route table. |
| Host networking / Netplan | No Netplan configuration is shipped in this repository. | Detect the host's network manager and include Netplan only where applicable. |
| auditd | No auditd configuration is shipped in this repository. | Installed-host presence and configuration scope. |

## Confirmed interface and architecture requirements

- **backup-restore/IR-001 — Static plugin selection:** Plugin implementations
  MUST be compiled into the service. Startup configuration MUST select which
  instances to construct. A valid application drop-in enables that target and
  names its compiled plugin type. Plugin construction MUST finish before the
  coordinating listener starts. The main service MUST call that listener through the same
  operation contract exposed by plugins. The listener MUST delegate each
  operation to the applicable initialized plugin or coordinate multiple
  initialized plugins, according to the operation's defined scope. An operation
  naming one application or stored version routes to its owning plugin; an
  all-application backup visits configured plugins one by one. The listener
  MUST NOT load code or construct plugin instances.
- **backup-restore/IR-002 — NATS service boundary:** The backup and restore
  service MUST use the repository's shared messaging layer for communication
  with other services. NATS subjects, request and response DTOs, and ownership
  remain to be designed from actual consumers and existing contracts.
- **backup-restore/IR-003 — CLI:** A CLI application MUST expose the supported
  backup, version selection, restore, and merge-related workflows through the
  service boundary. Interactive merge editing in the CLI has not been agreed.
- **backup-restore/IR-004 — Browser integration:** Web Gateway MUST adapt
  authenticated browser requests to the service's NATS interface. The admin
  panel MUST provide the graphical version and text-merge workflow. Existing
  gateway layering is `routes -> controllers -> services`; the browser app is
  `apps/admin-panel`, whose assets Web Gateway serves.

## Confirmed security and operations requirements

- **backup-restore/SR-001 — Restricted stored content:** Backups MUST be
  accessible only to the service account and any explicitly authorized
  administrative access. Encryption at rest is not a requirement. Archive and
  directory permissions, ownership, and AppArmor access MUST enforce the
  selected storage policy.
- **backup-restore/SR-002 — Scoped privileged commands:** Any plugin requiring
  elevated CLI access MUST use a dedicated, narrowly scoped passwordless
  sudoers rule for the service account. Commands and variable arguments MUST
  be validated against each program's operand and option rules and invoked
  without a shell. The specific permissions depend on the eventual plugin
  inventory.
- **backup-restore/SR-003 — Validate and authorize requests:** CLI, NATS,
  HTTP, and frontend inputs MUST be validated at their adapter boundaries.
  Backup reads, version reads, merge submissions, and restores MUST be
  authorized before exposing data or changing host state. Inventory, version,
  and diff reads, backup creation, and direct restore require the administrator
  (`admin` or `sudo`) or security officer (`lite-nas-security`) role. Service
  identities and any future merge-specific policy remain open.
- **backup-restore/OR-001 — Service storage:** Use
  `/var/lib/lite-nas/backup` as the intended service-owned data root, alongside
  existing roots such as `/var/lib/lite-nas/moments`. The exact internal layout
  is undecided.
- **backup-restore/OR-002 — Deployment:** Deploy as one non-root Go service
  process with in-process workers, a dedicated service identity, systemd unit,
  AppArmor policy, configuration, NATS permissions, and local build/deploy
  integration. Follow the repository's single-package Debian workflow.
- **backup-restore/RR-001 — Optional target absence:** Before invoking a
  target's backup operation, the service MUST determine whether that target is
  available. An absent or uninstalled target MUST not stop an all-application
  backup. A target that is present but fails during backup MUST remain an
  observable failure rather than being silently treated as absent.

## Repository evidence and design boundaries

- Log Analysis uses a fixed constructor registry, constructs configured plugin
  instances at startup, and routes normalized records to those instances.
  Its log listeners are separate source readers; the current detector plugins
  are instances of one concrete Go type, not implementations of a common Go
  interface. The backup and restore coordinating listener described here is a
  distinct service design, while startup selection follows the Log Analysis
  pattern.
- `shared/go/workers` supplies lifecycle and polling primitives.
  `shared/go/messaging` supplies inter-process messaging. Worker-owned bounded
  channels need named producers, consumers, backpressure, cancellation, and
  shutdown ownership when the implementation design defines them.
- Moments provisions `/var/lib/lite-nas/moments` for its dedicated account and
  restricts writes with systemd and AppArmor. The backup service can follow
  this service-owned storage pattern.
- Web Gateway owns the browser HTTP boundary; its services layer performs
  backend NATS calls. The frontend lives in `apps/admin-panel`. Existing CLI
  applications call NATS through shared infrastructure.

## Open design questions

1. What operations and result shapes belong in the shared listener/plugin
   contract? How should the all-application result present completed, skipped,
   and failed targets?
2. Which operations initiate backups: explicit user requests, schedules,
   application events, or some combination? What is in scope for the first
   implementation?
3. How are application snapshots identified, retained, and deleted? If a
   drop-in has been removed, may a stored version use its included spec to
   restore the application, provided the current plugin and path policy accept
   that spec?
4. For generic file backups, what are the exact drop-in location and format,
   path allowlist, compression format, metadata requirements, ownership and
   mode restoration rules, and behavior for symlinks, missing files, changed
   file types, recursive directory contents, and overlapping file/directory
   entries?
   What fields identify a plugin type and configure each plugin instance?
5. Are sensitive values in configuration files shown to all authorized
   administrators and security officers? Which service identities may make
   requests, and what policy applies to future merge submissions?
6. For text merge, what is the comparison base, how are concurrent live-file
   changes detected, and does the new version become durable before or after
   successful rollout? What happens if rollout fails?
7. How will merged content and archive data cross the browser, gateway, CLI,
   and NATS boundaries? Existing NATS and HTTP contracts should be inspected
   before choosing DTOs, size limits, streaming, or temporary storage.
8. What restore guarantees are required: preview/dry run, confirmation,
   atomic replacement, service reload or restart, rollback, and partial-failure
   reporting? Which formats support key-aware merging, how are duplicate keys
   within one file handled, what base (if any) is used, and which local CLI
   merge tool workflow resolves files with conflicts?
9. For each target in the initial inventory, which exact files and live state
   are backed up, and what is the restore method? Which targets are required in
   the first release versus added later? Exact sudoers rules and live-state
   verification depend on those choices.
10. What are the limits for storage use, archive size, compression work,
    concurrency, retention, and recovery after a service crash?
11. Does “all LiteNAS configs” include transport and signing private keys,
    Postfix relay credentials, and generated credential maps? Those files have
    stricter access than ordinary configuration and affect whether a non-root
    service can read or expose their backup contents.

No API subjects, DTO fields, archive format, directory hierarchy, or privilege
allowlist is specified here until the relevant contracts and decisions are
confirmed.
