# Backup and Restore — Handoff and Feature Definition

This document records the implemented first slice and the remaining feature
direction. The service, NATS, and CLI contracts for first-slice backup, diff,
and restore are implemented. Browser integration, archive transfer, and
application-specific runtime operations remain future work. Open questions
below apply to those later slices unless stated otherwise.
Requirement IDs use `backup-restore/<TYPE>-<NNN>` for traceability.

## First slice: confirmed direction

- Start with a compiled generic file plugin driven by application-specific
  backup specification drop-ins. A spec maps one application to its related
  files and directories at their normal absolute paths. A spec may mix
  individual files and directories from `/etc`, `/var`, or another location.
  The file engine uses the absolute paths declared in the installed spec,
  regardless of whether LiteNAS installed those files or whether the host uses
  a standard location. Callers select a configured application; they cannot
  supply arbitrary paths in backup or restore requests. The secret-file policy
  remains open.
- The first implementation slice covers file-based settings for every
  applicable target in the initial inventory through file-plugin specs,
  including LiteNAS application configuration. Specs may explicitly select
  host settings files at any absolute path when the application uses them.
  Backups are invoked on demand from the CLI. Application CLI exports and
  restores, and live runtime state are later slices. A target whose complete state requires
  an application CLI can still have its settings files backed up in this
  slice; that snapshot alone does not claim to restore its live state.
- A file entry covers only its named path. A directory entry covers its whole
  tree. Shared configuration directories such as `/etc/lite-nas` should use
  individual file entries unless the application owns the whole directory.
  Restore mode is selected separately: replace removes live entries absent
  from a selected directory snapshot, while merge preserves them.
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
- Include a root-level `metadata.json` manifest in each archive. The manifest
  is JSON, matching Go's standard library support, and contains `application`,
  `version`, optional user-provided `name` and `description`, service-generated
  UTC `created_at` timestamp, and `created_by` from the verified caller's
  authenticated login claim. Do not accept `created_by` from the request.
  Reserve `metadata.json` as an archive member so a selected host path cannot
  collide with it. The CLI may omit name and description. For `backup all`,
  each application's archive receives the same supplied name and description,
  but its own application ID, version, and creation timestamp.
  The CLI now uses a user access token, so its snapshots identify the
  authenticated user.
- Store each version as a `tar.gz` archive with entries reflecting the host
  path tree: `/etc/path/to/file.conf` appears as `etc/path/to/file.conf` inside
  the archive. This permits manual inspection and extraction on another host.
  Preserve file metadata and symlink entries; when a selected path is a
  symlink, also capture and restore the linked file under its own path, even
  when it is outside `/etc`. The service must validate the resolved target
  path and treat it as an explicit restore destination. The exact metadata set,
  handling of symlink chains, and path policy still need definition. The service
  reads archive entries for comparison and merge without requiring the CLI to
  unpack them or access privileged source paths. Gzip alone compresses one
  stream; the tar layer supplies the file tree, entry types, and metadata.
  Stream source files into the archive and publish the completed archive from
  a temporary file under the service-owned backup root. A RAM filesystem is
  not required for staging and does not grant access to protected source or
  restore paths.
- Define a `lite-nas` application for global package configuration. Assign
  files owned by a particular service or CLI to that application's spec rather
  than including every `/etc/lite-nas` file in the global snapshot.
- A source location under `/var` does not by itself make a live database or
  mutable application state safe to copy. Such targets need an
  application-specific export/restore plugin when plain file copying cannot
  produce a consistent snapshot.
- Use NATS through `shared/go/messaging` between the CLI and the backup and
  restore service. The CLI does not read or write privileged host files.
- Require an administrator-equivalent NATS role for every operation. Apply the
  same rule to inventory, version, and diff reads
  because those operations can disclose configuration and backup history.
  Security-officer access can be reconsidered when browser integration is
  designed; `lite-nas-operator` alone is insufficient.
- Expose an inventory of configured applications supported for backup, an
  inventory of applications with stored backups, and versions for a selected
  application. Keep supported targets distinct from targets with history;
  removing a drop-in does not erase stored versions.
- The implemented diff compares a selected stored version with current settings
  and returns changed paths and statuses without file contents. A later diff
  operation should also compare two stored versions using the same path/status
  result shape.
- A replace restore makes a whole-directory target match its stored snapshot:
  live entries beneath that target which are absent from the snapshot are
  removed. For selected-file targets, replace changes only the listed paths;
  unlisted neighbors remain. A merge restore keeps extra live entries in both
  cases.
- Text or key-aware content merging is not included in the first slice. The
  implemented `merge` mode preserves additional live directory entries while
  restoring archived paths. INI content merging and conflict resolution can be
  specified as a later feature.

## Implemented shared merge-session contract

Text merge now uses a service contract intended to serve
both the CLI and the Web Gateway/admin panel. The service remains the sole
owner of privileged reads and writes. The CLI and browser are clients of the
same merge session operations; the Web Gateway adapts authenticated HTTP
requests to NATS and does not read host files or perform the merge itself.

- The service allows one active merge session globally. A second start request
  is rejected while a session is active. The session is held in service memory
  and contains the selected merge inputs and staged merged file contents.
- Starting a session selects an application and backup version. The
  service reads archived and current text files, returns a session ID
  and per-file paths, sizes, and comparison status, and makes the input contents
  available to the client for a local merge tool or browser merge editor.
- The client may submit resolved file contents to the session. The service
  validates each path against the selected application spec and stages those
  bytes in memory. The client cannot supply arbitrary destination paths.
- A session has one explicit terminal action: `apply` writes staged content to
  the live selected paths; `save` creates a new application snapshot from the
  selected snapshot with staged content substituted at resolved paths, without
  applying it live; `discard` drops staged content without applying or saving.
  An explicit discard is not required on shutdown.
- Applying checks that the live files have not changed since session start.
  Changed inputs invalidate the session for apply and require a fresh session.
- Session contents are lost on process restart. No timer expires a session.
  The service limits each file to 1 MiB, a session to 512 files and 32 MiB of inputs, and
  each NATS transfer chunk to 64 KiB. A partial apply error keeps the session
  for `save` or `discard` and prevents another apply.
- File contents are transferred as bounded chunks, not as a base64 copy of the
  whole archive. NATS is the service boundary for both clients; the CLI consumes
  the transfer directly, while the Web Gateway exposes an authenticated HTTP
  transfer to the frontend and bridges it to NATS. The archive itself remains
  in the service-owned store. This merge transfer is distinct from archive
  import/export, which is not part of the contract.
- The saved snapshot includes the merged application files and the applicable
  spec. It receives normal snapshot metadata: a new service-generated version
  and timestamp, user-provided name/description if supplied, and creator from
  the verified authenticated login claim. Creator identity must reflect the
  human caller if per-user attribution is required.

NATS subjects and DTOs are in `shared/go/contracts/backuprestore`. Chunks carry
base64 bytes with a full-content SHA-256 checksum; the service ties a session
to the verified token subject. The merge has two inputs: selected snapshot and
current file, with the result initialized from current content. HTTP shapes
remain for the later Web Gateway adapter.

The first slice excludes browser and admin-panel integration. Its implemented
NATS contract is available for a later Web Gateway adapter without making the
CLI own backup logic. Archive bytes stay in the service-owned store; there is
currently no API to import or export an archive as a byte stream or base64
payload. A future UI transfer contract must define transport, size limits,
authorization, and temporary storage.

The backup and restore service is the single owner of source-file reads,
metadata collection, stored-version reads and writes, comparisons with live
files, and restore writes. On-demand CLI requests go through NATS to that
service. The CLI presents results and collects user choices; it does not read
privileged source files, compare them independently, or write restored files.
Web Gateway is a later adapter to the same service operations and does not
duplicate file or backup logic.

## Implemented first-slice service and CLI

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
`path=/absolute/path` entry and may coexist in one spec. File-plugin
selections may also set `optional=true` for host-dependent configuration;
an absent optional selection is omitted from capture and left untouched by
restore, and a snapshot requires at least one present selection. A CLI-backed plugin's
spec contains only settings defined and validated by that plugin. Specs are
loaded and checked at startup, not interpreted from user-supplied paths during
a NATS request. A directory section selects its entire tree; a file section
selects only its named path. The exact validation and traversal rules remain
open.

The existing packaged `lite-nas` spec is
`configs/etc/lite-nas/backup.d/lite-nas.conf`. It selects only
`/etc/logrotate.d/lite-nas` through the file plugin. Additional first-slice
specs must assign selected settings files to their owning
applications. Mutable state is outside the first slice. Certificate and
credential inclusion needs an explicit policy before those files are selected.
The first implementation packages file specs for repository-managed LiteNAS
configuration, UFW, Postfix, rsyslog, systemd, and AppArmor paths. Additional
host-specific paths and targets remain administrator-configured.
The existing config deployment copies the packaged spec to `/etc` and
preserves an installed version of the same drop-in on later deployments.
Installed drop-ins are restricted to the LiteNAS group by the normal
configuration-permission pass.
Each stored version includes the exact drop-in used to create it. Restoring an
older version can therefore restore the older spec as well as its selected
files. The service must validate that stored spec against the compiled plugin
registry and current restore path policy before using it; an archived spec is
not an authorization bypass.

The CLI provides `applications`, `backed-up`, `versions <application>`,
`backup <application>`, `backup all`, `diff <application> <version>`, and
`restore <application> <version> <replace|merge>`. NATS subjects and DTOs are
defined in `shared/go/contracts/backuprestore`. Requests select configured
applications and versions; they do not accept arbitrary host paths.

The restore request has a replace or merge mode. In replace mode,
deletion is confined to directory trees explicitly marked as fully covered in
the stored spec; for selected-file coverage, unrelated files remain in place.
Merge mode preserves live entries absent from the snapshot while restoring
archived content at selected paths. Text content merging and submitting
resolved content are not implemented in this slice.

Each installed application spec declares an explicit `restart_units` list.
New archive metadata records that list for inspection. After restoring files,
the service requests each listed service restart through the Services Manager
NATS RPC using the caller's token. The installed spec, rather
than archive content, authorizes the restart targets. A failed restart is
reported as a warning because the file restore has already completed. Targets
without a defined restart action declare an empty list.
The CLI authenticates a user through Auth Service, caches only the access token
in that user's `/run/user/<uid>` directory, and does not refresh it. The backup
service and Services Manager both require an administrator role for restore and
restart. A failed restart remains a warning in the service response and causes
the CLI to exit with failure after stating that files were restored.

Future INI merge rules, if that feature is selected: identify a key by its
section and key names with case-sensitive matching; retain the live file's
comments and order; add keys or sections that exist only in the snapshot; keep
keys or sections that exist only in the live file; keep an identical value
without conflict; and report different values for the same key as a conflict.
Duplicate keys, malformed INI, or constructs whose meaning cannot be preserved
without rewriting the file require manual resolution. Resolved content should
become a new stored version.

This CLI authenticates the user and caches only a short-lived access token in
the user's runtime directory. The service stays running and handles requests only
when they arrive; it does not schedule backups in the first slice.

## Later feature direction

Extend versioned, per-application backups and controlled restoration of LiteNAS
configuration and state. Future work may add application CLI plugins, live
state handling, text merge, archive transfer, and browser integration. The
initial target inventory below guides packaged file specs; it does not mean all
runtime state is handled by this slice.

The implemented feature has a Go backup and restore service and a CLI. A future
Web Gateway adapter and admin-panel workflow should call the same service over
NATS; neither adapter should perform privileged file operations itself.

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
  location. A directory selection MUST cover its whole tree, while a file
  selection MUST cover only its named path. Each stored version MUST be a `tar.gz` archive
  whose entries use host paths without their leading slash, preserve selected
  symlinks, and include their linked files under their resolved host paths.
- **backup-restore/FR-003 — Application-specific plugins:** The design MUST
  permit compiled application-specific plugins where file copying cannot
  correctly export and restore application state. Such a plugin may use the
  application's CLI when its exact behavior and privilege rules are specified.
  The UFW rule plugin exports a versioned JSON document containing enabled
  state, logging level, incoming/outgoing/routed defaults, and ordered,
  structured CLI-managed rules. It MUST fail backup when a rule cannot be
  represented, validate edited JSON before any firewall mutation, and replay
  a resolved document through UFW commands. Applying a merged document
  replaces UFW-managed rules and defaults. Custom `before.rules` and
  `after.rules`, the full `/etc/default/ufw` file, and app-profile files are
  outside this JSON plugin.
  The Cloudflare plugin MUST archive a locally managed tunnel configuration
  and its credential file from a spec-declared directory, and validate ingress with
  the installed `cloudflared` CLI before capture and restore. The Tailscale
  plugin MUST archive local preferences exported by `tailscale get --set-flags`
  and restore them through `tailscale set` with individually validated operands.
  It MUST NOT copy device identity state. Both CLI targets MUST be skipped by
  all-application backup when their executable is absent.
- **backup-restore/FR-004 — Direct restore:** A user MUST be able to select a
  stored version and request its restoration to the applicable application.
  The stored application spec MUST be restorable with the version. In replace
  mode, selected directories MUST be reconciled to the snapshot by removing
  live entries absent from it. Individual file selections MUST leave unlisted
  neighboring files intact. Deletion MUST remain within explicitly selected
  directory roots.
- **backup-restore/FR-005 — Text merge:** A user MUST be able
  to compare and merge backed-up text with current files through a service-owned
  merge session shared by the CLI and Web Gateway/admin panel. The service
  stages resolved contents and owns apply, save-as-new-snapshot, and discard
  actions. The UFW JSON document MUST be available in this UI merge flow;
  applying its resolved content uses the UFW rule plugin's validation and
  replay path rather than writing the JSON as a live configuration file.
- **backup-restore/FR-006 — Merged-content versioning:** The merge
  session's `save` action MUST create a new backup version using the staged
  merged content, rather than reading the live file as the source for those
  merged paths. `apply` and `save` are separate terminal actions.
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
  request differences between a selected stored version and current settings;
  the first slice returns changed paths and statuses without file contents.
  A later extension SHOULD compare two stored versions with the same result
  shape.
- **backup-restore/FR-011 — Directory merge restore:** A user MUST be able to
  request a merge restore that preserves live entries absent from selected
  snapshot paths, including wholly covered directories. In the first slice,
  this means preserving extra filesystem entries while restoring archived
  entries; it does not merge file contents. Key-aware INI merge and conflict
  resolution are future work and need a separate requirement/API contract.
- **backup-restore/FR-012 — Restore by name:** A caller MUST be able to restore
  the newest stored version whose archive metadata name exactly matches a
  supplied name for one configured application. A separate all-application
  operation MUST visit configured applications in stable order, restore each
  newest matching version, skip applications without a matching name, and
  report completed, skipped, and failed outcomes with selected versions and
  restart warnings. Both operations MUST use the same replace or merge restore
  behavior and administrator authorization as direct version restore. The CLI
  and admin panel MUST expose both operations.

## Initial target inventory

| Target | Repository evidence | Scope still to define |
| --- | --- | --- |
| LiteNAS configuration | `configs/etc/lite-nas/` contains service configs, rules, templates, and certificate paths. | Which live files are included; how private keys and other credentials are handled. |
| Firewall | The existing `ufw` file spec selects `/etc/ufw/ufw.conf`; the `ufw-rules` plugin exports and replays CLI-managed state in JSON. | Custom framework rules and app-profile files remain outside the JSON target. |
| Cloudflare Tunnel | The `cloudflared` plugin archives a spec-declared local directory and validates ingress through the CLI. The packaged spec uses `/etc/cloudflared`. | Remote tunnel settings are excluded. Other local paths require matching confinement rules. |
| Tailscale | The `tailscale` plugin exports and replays local preferences through the CLI. | Identity state and Serve/Funnel configuration are excluded. |
| USBGuard | A file spec selects `/etc/usbguard/usbguard-daemon.conf` and `/etc/usbguard/rules.conf`. Both files must exist. Runtime device authorization state is excluded. | Safe policy reload behavior and optional `rules.d` directories remain separate work. |
| Journal Manager | A file spec selects `/etc/lite-nas/journal-manager.conf` and requests a Journal Manager restart after restore. Shared systemd and AppArmor targets include its unit and profile. | Journal records and the package-managed sudoers rule are outside this configuration target. |
| AIDE | `litenas-integrity` invokes AIDE and maintains its baseline under `/var/lib/aide`. | Configuration versus baseline database, and whether baseline restore is appropriate. |
| AppArmor | Deployment manages profiles under `/etc/apparmor.d` and reloads them. | Profile selection, validation, and reload behavior. |
| systemd | LiteNAS unit files are shipped under `configs/etc/systemd/system/`. | Which unit files, drop-ins, enablement state, and reload/restart effects are included. |
| Postfix | Deployment manages `/etc/postfix` and may create restricted relay credential files. | Config and credential scope, generated maps, validation, and reload behavior. |
| logrotate | The `logrotate` file spec selects `/etc/logrotate.conf` and the complete `/etc/logrotate.d` tree. The existing `lite-nas` target also selects its own rule. | Restore uses the normal file plugin's replace or merge mode. |
| rsyslog | Repository uses rsyslog output for log analysis. | Which host rules and state belong to the backup. |
| IP routes | No dedicated backup contract found. | Persistent configuration versus current kernel route table. |
| Host networking / Netplan | A file spec selects `/etc/netplan` when present; `backup all` skips an absent directory. Restore writes the files but does not apply them to the running network. | A validated, connectivity-safe apply operation is separate work. |
| PAM | The `pam` file spec selects `/etc/pam.d` and `/etc/security`. | Password-history data inside `/etc/security` is included; restore does not reload consumers. |
| auditd | The `auditd` file spec selects `/etc/audit` when installed. | Restore writes configuration and rules without activating them. |
| SSH server | The `ssh` file spec selects `sshd_config` and optional server drop-ins. | Host keys and client settings are excluded; restore does not restart sshd. |
| journald | The `journald` file spec selects the optional main config and local drop-ins. | Journal data is excluded; restore does not restart journald. |
| GRUB | The `grub` file spec selects `/etc/default/grub`, optional drop-ins, and `/etc/grub.d`. | Generated `/boot/grub/grub.cfg` is excluded; restore does not run `update-grub`. |
| nginx | The `nginx` file spec selects `/etc/nginx` and optional LiteNAS TLS assets under `/etc/lite-nas/certificates/nginx`. | Restore writes files without validating or reloading nginx. |
| Privilege policy | The `privilege-policy` file spec selects sudoers, sudoers drop-ins, and polkit rules when present. | Restore writes policy without validating it; administrative access can be affected when the policy is applied. |
| Kernel hardening | The `kernel-hardening` file spec selects sysctl configuration, modprobe drop-ins, and module-load settings. | Restore does not apply settings to the running kernel. |
| Storage recovery | The `storage-recovery` file spec selects `/etc/fstab` and optional `/etc/crypttab`. | ZFS export and restore are excluded by request; restore does not mount filesystems or unlock volumes. |
| Time synchronization | The `time-sync` file spec selects timesyncd and chrony configuration when present. | Restore does not restart either daemon. |
| Package trust | The `package-trust` file spec selects APT sources, preferences, local keyrings, and `apt.conf` policy files under `/etc/apt`. | Dedicated `auth.conf` files and package-managed `/usr/share/keyrings` are excluded; policy files may still contain credentials. Restore does not run `apt update`. |

Cloudflare Tunnel can have a local configuration file for a locally managed
tunnel, but the tunnel credential file is a separate secret and remotely
managed tunnel settings live in Cloudflare. A configuration-only snapshot would
therefore be useful for review but insufficient to restore every tunnel setup.
See [Cloudflare's local configuration](https://developers.cloudflare.com/tunnel/features/locally-managed-tunnels/configuration-file/)
and [remote tunnel tokens](https://developers.cloudflare.com/tunnel/reference/tunnel-tokens/).

Tailscale's Linux daemon has local settings such as `/etc/default/tailscaled`,
but its node identity and keys live in daemon state, which may be encrypted or
stored outside a local file. A settings-only snapshot would not restore a
node's authenticated identity. The state location and restore policy need
separate design before adding a Tailscale target. See the
[tailscaled reference](https://tailscale.com/docs/reference/tailscaled).

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
  with other services. First-slice subjects and DTOs are defined in
  `shared/go/contracts/backuprestore`, including merge-session subjects and
  chunk DTOs.
- **backup-restore/IR-003 — CLI:** The CLI MUST expose backup, version
  selection, restore, and merge-session operations through the service
  boundary. Merge inputs and resolved contents travel through the service; the
  CLI does not read privileged files or apply restored content itself.
- **backup-restore/IR-004 — Browser integration:** Web Gateway MUST adapt
  authenticated browser requests to the service's NATS interface. The admin
  panel MUST use the same merge-session lifecycle as the CLI. Gateway layering
  is `routes -> controllers -> services`; the browser app is `apps/admin-panel`.

## Confirmed security and operations requirements

- **backup-restore/SR-001 — Restricted stored content:** Backups MUST be
  accessible only to the service account and any explicitly authorized
  administrative access. Encryption at rest is not a requirement. Archive and
  directory permissions, ownership, and AppArmor access MUST enforce the
  selected storage policy.
- **backup-restore/SR-002 — Scoped privileged commands:** Any plugin requiring
  an application CLI MUST validate commands and variable arguments against
  each program's operand and option rules and invoke them without a shell.
  The root-run service does not use sudo for file access in the first slice.
  The Cloudflare local-tunnel plugin validates its selected files with the
  `cloudflared` ingress CLI before capture and restore. The Tailscale plugin
  exports local preferences with `tailscale get --set-flags` and replays
  validated flags with `tailscale set`; device identity remains outside its
  snapshot. Missing CLI installations are skipped by all-application backup.
- **backup-restore/SR-003 — Validate and authorize requests:** CLI, NATS,
  HTTP, and frontend inputs MUST be validated at their adapter boundaries.
  Backup reads, version reads, merge submissions, and restores MUST be
  authorized before exposing data or changing host state. Inventory, version,
  and diff reads, backup creation, and direct restore require the administrator
  (`admin` or `sudo`) role. The CLI caches only a user access token in an
  owner-only runtime file and does not refresh it. Merge RPCs also require
  the administrator role and the same verified subject that started the session.
- **backup-restore/OR-001 — Service storage:** Use
  `/var/lib/lite-nas/backup` as the intended service-owned data root, alongside
  existing roots such as `/var/lib/lite-nas/moments`. The exact internal layout
  is undecided.
- **backup-restore/OR-002 — Deployment:** As a first-slice exception, deploy
  one root-run Go service process with in-process workers, a dedicated NATS
  service identity, systemd unit, AppArmor policy, configuration, NATS
  permissions, and local build/deploy integration. Keep the process idle except
  while handling an authorized on-demand request. Constrain writable paths in
  systemd and AppArmor to the backup store and configured restore targets;
  drop capabilities not needed for file backup and restore and retain the usual
  systemd process hardening where compatible. Host-specific specs require
  matching policy updates. Reassess root execution
  after the first slice. Follow the repository's single-package Debian workflow.
- **backup-restore/RR-001 — Optional target absence:** Before invoking a
  target's backup operation, the service MUST determine whether that target is
  available. An absent or uninstalled target MUST not stop an all-application
  backup. A target that is present but fails during backup MUST remain an
  observable failure rather than being silently treated as absent.

## First-slice root-service risk assessment

Root execution is an explicit exception to the normal non-root service model.
It avoids changing ownership or ACLs on host configuration files and permits
restoring their ownership and modes. Idle operation limits how often the
service handles requests, but it does not reduce the privileges of a process
that is running and listening on NATS. The residual impact of a service
compromise remains high; this exception must be revisited after the file-based
slice is operational.

| Risk | First-slice control | Residual risk |
| --- | --- | --- |
| An unauthorized caller uses root file operations through NATS. | Authenticate requests, require the `admin` or `sudo` role on every subject, limit NATS permissions to intended identities, and keep user access tokens owner-only. | A compromised authorized identity can still request privileged operations. |
| A malicious or mistaken spec or archive writes outside the intended application. | Load specs at startup; accept application IDs and versions rather than caller-supplied paths; validate absolute paths, archive members, symlink targets, and restored spec before use. Confine write paths with systemd and AppArmor. | Root and policy mistakes can still affect host files; arbitrary custom paths require deliberate policy updates. |
| Replace restore removes unrelated configuration. | File selections affect only named paths. Directory selections cover the whole tree; require an explicit replace request and reconcile only the selected directory root. | An overly broad directory spec can still delete unrelated files; preview remains an open design question. |
| Backups disclose credentials or signing keys. | Keep archive storage root-owned and private, authorize metadata and diff reads, and avoid sending archive bytes or secret file contents over NATS in the first slice. | Root and authorized administrators can read stored secrets; encryption at rest is not planned. |
| A large or malformed archive exhausts resources or escapes extraction boundaries. | Bound archive sizes and processing time, reject unsupported entry types and unsafe paths, and validate a complete snapshot before restore. Stream backups to temporary files and publish only complete archives. | Limits and crash recovery still need exact values and implementation. |

Deployment must review the effective systemd and AppArmor rules against every
configured path; a denied path must produce an observable operation failure.
Security claims in this table are requirements for the implementation, not
evidence that the current code already enforces them.

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
   application events, or some combination beyond the on-demand first slice?
3. How are application snapshots identified, retained, and deleted? If a
   drop-in has been removed, may a stored version use its included spec to
   restore the application, provided the current plugin and path policy accept
   that spec?
4. For generic file backups, what path validation and special-filesystem
   exclusions apply to spec-selected absolute paths? Which metadata, ownership,
   and modes are restored? What happens with missing files, changed file types,
   recursive directory contents, symlink chains, and overlapping file and
   directory entries? What fields configure each plugin instance?
5. Are sensitive values in configuration files shown to all authorized
   administrators? Should security officers gain access in a later slice?
   Which service identities may make requests, and what policy applies to
   future merge submissions?
6. For future automatic key-aware merge, what conflict and parser rules should
   be added beyond the current two-input manual text workflow?
7. If browser archive import/export is added, how will archive bytes cross the
   browser, gateway, CLI, and NATS boundaries? Inspect existing NATS and HTTP
   contracts before choosing DTOs, size limits, streaming, or temporary storage.
8. What restore guarantees are required: preview/dry run, confirmation,
   atomic replacement, service reload or restart, rollback, and partial-failure
   reporting? Which formats support key-aware merging, how are duplicate keys
   within one file handled, what base (if any) is used, and which local CLI
   merge tool workflow resolves files with conflicts?
9. Which settings files should the repository ship in each initial spec, and
   which additional host paths remain administrator-configured? Which paths
   contain sensitive credentials? Which later plugin restores live state?
   Exact command permissions and live-state
   verification depend on those choices.
10. What are the limits for storage use, archive size, compression work,
    concurrency, retention, and recovery after a service crash?
11. Does “all LiteNAS configs” include transport and signing private keys,
    Postfix relay credentials, and generated credential maps? Those files have
    stricter access than ordinary configuration, and their backup contents
    require an explicit exposure policy even though the first-slice service
    runs as root.

The first-slice API subjects, DTOs, tar.gz archive format, per-application
directory hierarchy, and privilege checks are implemented and described above.
The metadata manifest and backup name/description inputs are a proposed
extension; they have not yet been added to the service or CLI contracts.
