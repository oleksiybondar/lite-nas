# lite-nas-file-writer

`lite-nas-file-writer` performs one upload-stage filesystem operation as its
effective non-root Unix user. File Service owns authentication, RBAC, path
policy, upload metadata, chunk sequencing, and retries. The writer uses only
filesystem paths and expected identities supplied as command parameters.
It has no daemon, HTTP listener, or NATS identity.

Run `lite-nas-file-writer --help` or invoke it without arguments for readable
usage. The installed executable is
`/usr/libexec/lite-nas/lite-nas-file-writer`; `/usr/bin/lite-nas-file-writer`
is a convenience symlink. File Service must invoke the fixed libexec path.

## CLI contract

Every `--target` is an absolute, clean, existing-parent path ending in
`.upload`. The final destination is that path without `.upload`. The writer
rejects symlinks in the parent path and stage entry.
It also rejects a stage with another hard link.

```text
lite-nas-file-writer create --target /home/alice/file.upload
lite-nas-file-writer write --target /home/alice/file.upload --expected-device 123 --expected-inode 456 --expected-uid 1000 --offset 0 --length 4096
lite-nas-file-writer stat --target /home/alice/file.upload --expected-device 123 --expected-inode 456 --expected-uid 1000
lite-nas-file-writer checksum --target /home/alice/file.upload --expected-device 123 --expected-inode 456 --expected-uid 1000
lite-nas-file-writer truncate --target /home/alice/file.upload --expected-device 123 --expected-inode 456 --expected-uid 1000 --offset 4096
lite-nas-file-writer publish --target /home/alice/file.upload --expected-device 123 --expected-inode 456 --expected-uid 1000 --destination-absent
lite-nas-file-writer publish --target /home/alice/file.upload --expected-device 123 --expected-inode 456 --expected-uid 1000 --destination-device 123 --destination-inode 789
lite-nas-file-writer cleanup --target /home/alice/file.upload --expected-device 123 --expected-inode 456 --expected-uid 1000
```

`create` exclusively creates the stage with mode `0600`, syncs the file and
directory, and returns its identity. All later operations require that
identity. `write` reads exactly `--length` raw bytes from stdin, up to
16 MiB. The caller closes stdin after the bytes. It truncates any uncommitted
tail at `--offset`; an incomplete or extra input restores the committed
prefix. `stat` reports size. `checksum` reports SHA-256. `truncate`
discards an uncommitted tail and syncs it. `cleanup` removes only the matching
stage. File Service must not infer authorization from any writer result.

`publish` checks an expected absent or regular destination and atomically
renames the stage in the same directory. `--destination-absent` uses
`RENAME_NOREPLACE`. Existing-file replacement checks its device and inode
immediately before rename. Linux does not offer an atomic compare-and-replace
rename, so a concurrent change in that brief interval can replace a different
inode. File Service must account for that risk when it authorizes publication.
The stage pathname is also rechecked before rename or cleanup, but another
process with write access to its directory can replace that entry immediately
after the check. Linux provides no descriptor-based rename or unlink for that
entry; File Service must control concurrent access to the upload directory.
The stage's UID, group, mode, and ACL become those of the published file.

Success prints one JSON object on stdout with `operation`, `device`, `inode`,
`uid`, `size`, `bytes_written`, and `next_offset`. `checksum` also returns
`sha256`. Failure prints `{"error":{"code":"operation_failed","message":"..."}}`
on stderr and exits nonzero. A failed publish attempts stage cleanup and
returns `publish_failed_cleaned`, `publish_failed_cleanup_failed`, or
`publish_failed_cleanup_skipped` with a `cleanup` field. A directory sync
failure after rename returns `publish_sync_failed` with
`cleanup:"not_possible"`; the destination may already have changed.

## Deployment and confinement

`scripts/build-file-writer-binary.sh` builds the executable.
`scripts/deploy-file-writer.sh` installs the root-owned binary, sudoers rule,
and AppArmor profile. The single LiteNAS package includes them as well.
The sudoers rule allows only the `lite-nas-file-service` account to invoke
the fixed writer path as a non-root user; the writer also rejects effective
UID 0.

The AppArmor profile grants write access under `/home` by default. File
Service deployment must provision each additional ZFS write root in
`/etc/apparmor.d/local/usr.libexec.lite-nas.lite-nas-file-writer` and reload
the profile with `apparmor_parser -r
/etc/apparmor.d/usr.libexec.lite-nas.lite-nas-file-writer` whenever roots
change. For a provisioned `/tank/uploads` root, the local include contains:

```text
/tank/uploads/ r,
/tank/uploads/** rw,
owner /tank/uploads/**.upload k,
```

The profile stays separate from File Service's profile. The writer
does not read or interpret File Service's write-root or protected-path policy;
File Service validates those rules before invoking it.

The File Service implementation and its user-identity/RBAC integration remain
in [its own task](../../services/file-service/task.md).
