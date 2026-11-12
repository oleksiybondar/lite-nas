# File Writer Application Requirements

The writer performs one requested upload-stage filesystem operation as its
effective non-root Unix user. File Service owns authentication, RBAC decisions,
write-root and protected-path policy, upload state, chunk indexing, and retries.
Requirement IDs use `file-writer/<TYPE>-<NNN>`.

## Functional Requirements

### file-writer/FR-001 Create an upload stage

The writer MUST exclusively create a regular, user-owned
`<destination>.upload` file with no-follow semantics. An existing file or
symlink MUST cause failure. Successful creation MUST sync the file and parent
directory and return its device, inode, owner UID, and size as JSON.

### file-writer/FR-002 Write a bounded chunk

The writer MUST accept an expected committed offset and a declared stdin byte
count no greater than 16 MiB. It MUST reject an offset beyond the current
stage size, truncate any uncommitted tail at that offset, write exactly the
declared bytes, and sync before reporting `bytes_written` and `next_offset`.
An incomplete or oversized input MUST restore and sync the committed prefix.

### file-writer/FR-003 Verify stage identity

Every operation after creation MUST require the stage device, inode, and owner
UID recorded by File Service. The writer MUST compare those parameters with
the opened regular file, reject symlinks, and reject a changed path entry.
Stages with another hard link MUST also be rejected.
The owner UID MUST equal the process's effective UID. The writer MUST run as
a non-root user. Path-entry rechecks cannot prevent a concurrent process with
directory write access from substituting the entry immediately before a
path-based rename or unlink; File Service must control that concurrency.

### file-writer/FR-004 Inspect and recover a stage

The writer MUST report the stage size, compute SHA-256 from its opened
descriptor, and truncate an uncommitted tail to a supplied offset. It MUST
serialize modifications and sync a successful recovery truncation.

### file-writer/FR-005 Explain command usage

The writer MUST print human-readable usage and exit successfully when invoked
without arguments or with `--help` or `-h`.

### file-writer/FR-006 Publish a stage

The writer MUST attempt a same-directory atomic rename after File Service
supplies the expected destination state. An absent destination uses
`RENAME_NOREPLACE`; an existing destination must be regular and match its
expected device and inode immediately before rename. A successful rename
MUST sync the parent directory. The published file keeps the stage's Unix
owner, group, mode, and ACL. Linux cannot atomically compare the existing
destination inode while replacing it, so a concurrent directory-entry change
after the check remains possible; File Service must treat this as a concurrent
mutation risk.
The stage MUST remain `0600` through rename so another group member cannot
alter its bytes after checksum verification. The writer MUST accept only
`0664`, `0660`, or `0644` as the final mode. File Service MUST request `0644`
only for the authenticated user's fixed `.lite-nas/avatar.image` target. The
writer MUST apply the selected mode to the opened published inode and sync it
before reporting success. A mode or sync failure after rename MUST report
that cleanup is not possible.

### file-writer/FR-007 Clean up a stage

The writer MUST remove only the identified stage. On a failed publication it
MUST attempt cleanup and return a distinct machine-readable cleanup outcome.
The upload is failed regardless of whether cleanup succeeds; File Service
requires a new upload rather than retrying publication.

## Integration Requirements

### file-writer/IR-001 Run with the caller's Unix identity

The writer MUST use its effective Unix identity for filesystem access. File
Service invokes the fixed root-owned executable through noninteractive sudo
as a verified non-root account. The writer has no independent authentication,
RBAC decision, or upload metadata store.

### file-writer/IR-002 Confine and deploy the writer

The binary MUST run under a dedicated enforcing AppArmor profile. Deployment
MUST install the root-owned binary, a restricted sudoers rule, and the profile
through local and single-package flows, validating sudoers syntax and profile
loading. The default AppArmor write root is `/home`; File Service
deployment must add explicitly configured ZFS roots to the profile's local
include and reload it when those roots change.
