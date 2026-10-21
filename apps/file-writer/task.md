# File Writer Application — Implementation Task

## Goal

Build a short-lived Go application that performs File Service upload-stage
operations with the verified Unix user's effective identity. File Service
authenticates each HTTP request, resolves its subject through RBAC, checks the
applicable path operation, and enforces configured write roots and protected
paths before invoking the writer. The writer performs the final Unix filesystem
operation under that user's UID and its own AppArmor confinement. It has no
HTTP listener, NATS identity, token parser, or independent RBAC decision.

This task implements the upload path in
[File Service requirements](../../services/file-service/requirements.md).
Use `file-writer/<TYPE>-<NNN>` for requirement-traceable checks. The repository
application is `apps/file-writer`; its installed binary is
`lite-nas-file-writer`.

## Invocation and responsibility

- File Service invokes the fixed, root-owned executable
  `/usr/libexec/lite-nas/lite-nas-file-writer` through noninteractive `sudo`
  with `-n`, `-u <verified-user>`, `--`, and the absolute executable path as
  separate arguments using Go `os/exec` argument
  vectors, never a shell or an `exec` shell builtin. The Unix account must be
  derived from successful Auth validation and matched to the RBAC subject UID;
  client input cannot select it. Reject a missing, renamed, or mismatched Unix
  identity before invoking the writer.
- Sudoers permits only the File Service account to run this exact executable
  as approved non-root user accounts. It does not permit a shell, generic
  `touch`/`tee`/`mv`/`rm`, arbitrary executables, or root as the target user.
  Validate the effective run-as restriction and argument handling on the
  installed sudo version; a sudoers command path alone does not validate file
  paths or prove RBAC authorization.
- The writer accepts only explicit upload-stage operations: exclusive create,
  bounded write at an expected committed offset, stat/identity check, staged
  checksum read, recovery truncate, same-directory publish/rename, and
  cancellation/failed-publication cleanup. It receives upload bytes on stdin
  and returns bounded, machine-readable results and a nonzero exit status on
  failure. Define and document the exact CLI arguments/result contract before
  implementing the File Service integration. Do not add general file deletion
  to this initial helper contract; File Service's delete operation needs its
  separate RBAC contract and execution decision.
- File Service streams the conventional multipart file part directly to the
  writer after authorization and supplies each resumable chunk through a
  bounded stream. It owns private durable resume metadata and records an
  acknowledged offset only after the writer reports a durable write. The
  writer does not keep a second full-file copy in `/tmp` or tmpfs.

## Filesystem behavior

- After RBAC authorizes creation, the writer creates
  `<destination>.upload` with exclusive, no-follow semantics as the user.
  This actual create is the Unix permission check. Do not run a separate
  `touch`: it changes timestamps on an existing file and cannot reserve the
  path or guarantee a later write. A conflicting file or symlink is an error.
- For every subsequent operation, verify the stage is the expected regular
  file, user-owned and still bound to the recorded upload. Do not follow
  symlinks or allow path substitution between validation and operation. Use
  race-resistant directory-relative operations, including for the destination
  and final rename. Reject paths outside File Service's configured write roots
  or inside protected paths; the helper must enforce an equivalent filesystem
  boundary because its arguments are supplied by another process. Read the
  boundary from root-owned configuration that the run-as user cannot change.
- Write only the expected bytes at the committed offset; handle short writes,
  cancellation and interrupted child processes. Make accepted data durable
  before reporting success. Support recovery truncation to File Service's
  committed offset and reading the staged bytes for checksum verification.
- After File Service reauthorizes publication and verifies size/checksum,
  attempt the same-directory atomic rename as the user. On failure, preserve
  the existing destination and attempt user-scoped stage cleanup. Report
  cleanup failure distinctly. The staged inode becomes the destination: its
  user UID, group, mode, and ACL become the completed file's metadata, including
  on overwrite. The old destination's metadata is not copied. No `chown`,
  `chgrp`, or service write group is needed.

## Confinement and deployment

- Add a dedicated profile at
  `configs/etc/apparmor.d/usr.libexec.lite-nas.lite-nas-file-writer`, attached
  to the fixed executable and loaded in enforce mode. Permit only required
  upload operations under provisioned write roots, including configured ZFS
  mounts; deny writes to protected/system paths and prevent execution of
  shells or arbitrary tools. Define how installation updates the profile when
  write roots change. Keep the File Service profile separate and verify the
  `sudo` to writer transition retains confinement. Do not use an unconfined
  fallback if the writer profile is missing.
- Install the root-owned binary, restricted sudoers drop-in, AppArmor profile,
  and configuration through existing local build/deploy flows and the single
  LiteNAS Debian package. Validate sudoers syntax and profile loading during
  deployment. The writer is invoked per operation and needs no daemon or
  systemd service.
- Cover valid and denied identities, a mismatched RBAC UID, conflicting or
  symlink stage, write offset and durability, changed path identity, failed
  rename and cleanup, cancellation, sudoers denial, AppArmor confinement, and
  write roots on `/home` and configured ZFS mounts. Verify the resulting file
  UID and the absence of service-account write access.

## Decisions before implementation

1. Specify the exact writer CLI/result contract and how File Service conveys
   a stage identity across separate writer processes without trusting a path
   or upload ID alone.
2. Specify how the writer receives the effective write-root/protected-path
   policy and how the AppArmor profile covers configured ZFS roots without
   expanding access to system paths.
3. Settle cleanup/retry responses and RBAC create/replace semantics before
   implementing final publication.
