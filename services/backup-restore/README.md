# Backup and restore first slice

The root-run service owns selected host files and `/var/lib/lite-nas/backup`.
The CLI sends authenticated NATS requests; it does not read source files or
archives. File Service is not used.

## Install locally

Run `sudo ./scripts/deploy-backup-restore.sh` from the repository root. The
script builds both binaries, deploys configuration and the NATS ACL, provisions
the CLI identity, rotates missing certificates, loads AppArmor policy, and
starts the service. It has not been run on a live LiteNAS host in this change.

Packaged specs cover the LiteNAS service and CLI configuration files plus
repository-managed UFW, Postfix, rsyslog, systemd, AppArmor, and logrotate
files. They select files individually unless an application-owned asset
directory is included. The service validates 36 packaged specs at startup.
Add one `<application>.conf` per additional host application:

```ini
[backup]
plugin=file

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

Run the CLI through `sudo` so only root or an authorized host administrator can
read its identity credentials:

```text
sudo backup-restore-cli applications
sudo backup-restore-cli backed-up
sudo backup-restore-cli backup lite-nas
sudo backup-restore-cli backup all
sudo backup-restore-cli versions lite-nas
sudo backup-restore-cli diff lite-nas VERSION
sudo backup-restore-cli restore lite-nas VERSION replace
sudo backup-restore-cli restore lite-nas VERSION merge
```

`diff` reports changed paths without their contents. `replace` removes live
entries absent from selected directory snapshots; individual file selections
never remove neighboring files. `merge` preserves extra live entries while
restoring archived content at selected paths. Key-aware text merging and live
application reloads are outside this file-based slice.

Versions are ordinary `tar.gz` archives with host paths stored without the
leading slash, such as `etc/logrotate.d/lite-nas`. They can be inspected and
extracted manually on another host. Treat archives as sensitive: they may
contain secrets. Restores validate archive paths against the current installed
spec and reject an older version whose selected paths no longer match. The
service limits a regular file to 128 MiB and total uncompressed content to
1 GiB. Stored versions remain listable if a spec is removed, but restoring
them requires a compatible installed spec. It does not automatically restart
affected applications after restore.
