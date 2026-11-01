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
logrotate files. AIDE has its own spec for its configuration and database.
Selections are individual files unless an application-owned asset directory is
included. The service validates packaged specs at startup.
Add one `<application>.conf` per additional host application:

```ini
[backup]
plugin=file
restart_units=example.service

[file.main]
path=/etc/example.conf

[directory.rules]
path=/etc/example.d
```

`[file.*]` selects one file or symlink. `[directory.*]` selects the entire
directory tree. Paths must be clean absolute paths. A symlink snapshot includes
its chain and final regular file, including targets outside `/etc`.

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
