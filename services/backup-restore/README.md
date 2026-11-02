# Backup and restore first slice

The root-run service owns selected host files and `/var/lib/lite-nas/backup`.
The CLI sends authenticated NATS requests; it does not read source files or
archives. File Service is not used.

## Install locally

Run `sudo ./scripts/deploy-backup-restore.sh` from the repository root. The
script builds both binaries, deploys configuration and the NATS ACL, rotates
missing transport certificates, loads AppArmor policy, and
starts the service. It has not been run on a live LiteNAS host in this change.

Packaged specs cover the LiteNAS service and CLI configuration files plus
repository-managed NATS configuration and server certificates, LiteNAS transport,
auth and identity certificates, UFW, Postfix, rsyslog, systemd, AppArmor, and
logrotate files. The separate `logrotate` target captures `/etc/logrotate.conf`
and the complete `/etc/logrotate.d` directory; the existing `lite-nas` target
still captures its own logrotate rule for a focused restore. AIDE has its own
spec for its configuration and database.
Netplan has a separate spec for `/etc/netplan`; `backup all` skips it when that
directory is absent. Restoring its files does not apply network changes to the
running host.
USBGuard has a separate spec for its default daemon configuration and rule
file. Both files must exist for a snapshot. Restoring them does not restart
USBGuard or re-evaluate connected devices.
The `pam` target captures `/etc/pam.d` and `/etc/security`, including any
password-history data such as `opasswd` present there. The `auditd` target
captures `/etc/audit` and is skipped by `backup all` when auditd is absent.
The `ssh` target captures `sshd_config` and optional server drop-ins; it does
not capture host keys or client configuration. The `journald` target captures
the main config and optional `/etc/systemd/journald.conf.d` drop-ins. An absent
optional path is omitted from the snapshot and preserved if it appears later
before restore. Restoring these targets writes files but does not reload or
restart PAM consumers, auditd, SSH, or journald.
The `grub` target captures `/etc/default/grub`, optional
`/etc/default/grub.d` drop-ins, and `/etc/grub.d` scripts. It excludes the
generated `/boot/grub/grub.cfg`; after restoring source configuration, an
administrator must validate it and run `update-grub` before it affects boot.
The `nginx` target captures `/etc/nginx` and LiteNAS nginx TLS assets under
`/etc/lite-nas/certificates/nginx`. The `privilege-policy` target captures
sudoers and polkit rules; `kernel-hardening` captures sysctl, modprobe, and
module-load configuration. The `storage-recovery` target captures `/etc/fstab`
and optional `/etc/crypttab`; ZFS is excluded. The `time-sync` target captures
timesyncd or chrony configuration when present. Missing optional paths are
skipped, and a target with no present selections is skipped by `backup all`.
Restoring these files does not validate or activate the affected services;
review the restored policy before applying it on the host.
The `package-trust` target captures local APT sources, preferences, keyrings,
and update policy under `/etc/apt`. It excludes dedicated `auth.conf` files
and package-managed keyrings under `/usr/share/keyrings`. APT policy files may
themselves contain credentials. Restoring files does
not refresh package indexes or run updates.
The `ufw-rules` target exports enabled state, logging, default policies, and
UFW-managed rules as `ufw.json` in the archive. Its JSON can be merged through
the existing three-pane UI editor. Applying a resolved document validates and
replays it through UFW commands, replacing the CLI-managed rules and defaults.
The CLI's normalized `show added` report is parsed strictly; ambiguous quoting
or unsupported output fails the backup. Direct `merge` restore keeps live-only
CLI rules after the archived rules, while applying UI-resolved JSON replaces
the CLI-managed rule set.
The existing `ufw` file target still backs up `/etc/ufw/ufw.conf` separately.
Custom framework rules and app-profile files are outside `ufw-rules`.
The `cloudflared` plugin snapshots the paths declared in its spec (packaged
defaults: `/etc/cloudflared` and its `config.yml`) for a locally managed
tunnel. The config must name a tunnel and a credential file also selected by
the spec. The plugin calls
`cloudflared tunnel --config ... ingress validate`
before backup and before restore, then the restore requests a
`cloudflared.service` restart. Installations using the default
`/root/.cloudflared` location needs an adjusted spec and corresponding
systemd/AppArmor access; remotely managed tunnels are outside this spec.
The `tailscale` plugin calls `tailscale get --set-flags` and stores the
validated preferences as `tailscale.json`; restore calls `tailscale set` with
those flags as separate arguments. At restore, the plugin checks each saved
flag against the names emitted by the installed CLI, so newly added
preferences do not require a LiteNAS flag list update. It does not copy identity state, log in,
or recreate Serve/Funnel configuration. If the CLI reports a preference the
plugin cannot safely represent, the backup fails for that target. `backup all`
skips either CLI target when its executable is missing; it also skips
`cloudflared` when `/etc/cloudflared/config.yml` is missing. Neither CLI
target supports the text merge editor.
Selections are individual files unless an application-owned asset directory is
included. The service validates packaged specs at startup.
The `journal-manager` file target selects `/etc/lite-nas/journal-manager.conf`
and requests a Journal Manager restart after restore. The shared `systemd` and
`apparmor` targets include its unit and profile. Journal records and the
package-managed sudoers rule are outside the Journal Manager file target.
The `before-startup` file target selects the root-owned
`/etc/lite-nas/before-startup.sh` host preparation script. Restoring it does
not rerun the one-shot unit; the script takes effect on the next boot or when
an administrator explicitly restarts `lite-nas-before-startup.service`.
Add one `<application>.conf` per additional host application:

```ini
[backup]
plugin=file
restart_units=example.service

[file.main]
path=/etc/example.conf

[directory.rules]
path=/etc/example.d
optional=true
```

`[file.*]` selects one file or symlink. `[directory.*]` selects the entire
directory tree. Paths must be clean absolute paths. A symlink snapshot includes
its chain and final regular file, including targets outside `/etc`.
`optional=true` permits a missing selected path; at least one selected path
must exist to create a version. An optional path absent from a snapshot is
left alone during restore, including replace mode.

Custom paths also need a systemd drop-in extending `ReadWritePaths` for restore
destinations and matching rules in the local AppArmor include
`/etc/apparmor.d/local/usr.libexec.lite-nas.backup-restore`. Permit temporary
`.restore-*` siblings where file replacement occurs. Restart the service after
changing specs or confinement rules; specs load only at startup. Keep the
backup store out of selected directories. Paths under `/home` also require a
deliberate `ProtectHome` override in the systemd drop-in.

## CLI

Run the CLI as your own user. It prompts for a username and password, then
keeps only the short-lived access token under `/run/user/<uid>/lite-nas` with
owner-only permissions. When that token is no longer valid, it prompts again.
The service authorizes the authenticated user's role for every operation:

```text
backup-restore-cli applications
backup-restore-cli backed-up
backup-restore-cli backup lite-nas
backup-restore-cli backup all
backup-restore-cli versions lite-nas
backup-restore-cli diff lite-nas VERSION
backup-restore-cli restore lite-nas VERSION replace
backup-restore-cli restore lite-nas VERSION merge
backup-restore-cli merge lite-nas VERSION
backup-restore-cli merge resume SESSION_ID
backup-restore-cli merge discard SESSION_ID
```

`diff` reports changed paths without their contents. `replace` removes live
entries absent from selected directory snapshots; individual file selections
never remove neighboring files. `merge` preserves extra live entries while
restoring archived content at selected paths. After a successful restore, the
service requests a restart for each `restart_units` entry through Services
Manager. Restart failures are returned as warnings because the files have
already been restored. An empty `restart_units` value explicitly selects no
restart for targets without a defined service action. Automatic key-aware INI
merging remains future work.
`merge APP VERSION` starts one service-owned text merge session. The CLI opens
private snapshot, result, and current copies in `$EDITOR` (`vi`, `vim`, `nvim`,
or `nano`); edit the result copy. After upload, choose `apply`, `save`, or
`discard`. `apply` writes resolved files live and requests configured service
restarts; `save` creates a new snapshot without changing live files. A session
remains available until one of those actions succeeds or the service restarts.
Use `merge resume SESSION_ID` after an interrupted editor, or `merge discard
SESSION_ID` to release it. Only one session can exist at a time. The service
limits text files to 1 MiB each, a session to 512 files and 32 MiB of inputs, and each NATS
transfer chunk to 64 KiB.
The CLI forwards the user's administrator access token to the backup service;
the service uses the same token for Services Manager restart requests.

Versions are ordinary `tar.gz` archives with host paths stored without the
leading slash, such as `etc/logrotate.d/lite-nas`. They can be inspected and
extracted manually on another host. Treat archives as sensitive: they may
contain secrets. Restores validate archive paths against the current installed
spec and reject an older version whose selected paths no longer match. The
service limits a regular file to 128 MiB and total uncompressed content to
1 GiB. Stored versions remain listable if a spec is removed, but restoring
them requires a compatible installed spec. The current installed spec controls
which services are restarted; archive metadata records the list for inspection.
