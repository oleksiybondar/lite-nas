# RBAC Service — File Access Decisions Task

## Goal

Define and implement RBAC decisions that File Service and Stream Service can
use for writing existing files, creating or replacing files, deleting files,
and traversing directories. Keep RBAC as the decision service: it evaluates
the authenticated subject's host permissions but does not open, write, or
delete files. Update [RBAC requirements](../../requirements/rbac-service.md)
and [service boundaries](../../requirements/rbac-service-boundaries.md) when
the contracts are settled.

## Current State

- Shared NATS contracts provide `rbac.rpc.path.read.check`,
  `rbac.rpc.path.write.check`, and `rbac.rpc.path.exec.check`, each taking a UID
  and path and returning a boolean decision. There is no delete RPC contract;
  the RBAC requirements mention `modify` without defining a matching RPC.
- Path checks resolve the requested path with `filepath.EvalSymlinks`, read
  its ACL with `getfacl`, and evaluate that path's permissions. A new file path
  cannot be resolved this way, so the current write check cannot authorize
  creation at a nonexistent target.
- The current path checks do not evaluate the subject's execute/search
  permission on every ancestor directory. A leaf read or write allow therefore
  does not by itself establish that the subject could reach the leaf.
- File Service needs decisions for upload creation, replacement, listing, and
  regular-file deletion. File Service and Stream Service both need decisions
  that account for directory traversal before serving an existing file.

## Service Boundaries

- Auth verifies the access token. RBAC receives the resolved subject UID
  through its existing internal NATS contracts; it does not authenticate the
  caller or validate JWTs.
- RBAC owns host identity, mode-bit, group, ACL, symlink, and directory-search
  permission evaluation for the subject. It returns a decision and has no
  filesystem side effect on behalf of File Service or Stream Service.
- File Service owns upload, replacement, deletion, protected-path and
  write-root policy, and race-resistant filesystem operations. After an RBAC
  allow, File Service invokes its dedicated
  [user-scoped writer](../../apps/file-writer/task.md) to create
  `.upload`, write staged bytes, and attempt final rename as the verified Unix
  user. Unix permissions can still reject creation or publication. Deletion
  has a separate RBAC and execution decision. Stream Service owns media
  delivery. Both services enforce
  RBAC's decisions and their own workflow rules; process-level access is
  never caller authorization.
- Use the existing shared NATS contract package for any approved contract
  additions. Do not assume that write permission on a file grants permission
  to create or remove a directory entry.

## Required Contract Decisions

Record the following decisions in the RBAC requirements and shared contract
definitions before changing RPC behavior. If existing contracts do not express
an operation safely, define the smallest additional request and response
contract needed by the consuming services. Do not silently change the meaning
of an existing RPC subject.

1. **Write existing content:** Define what `rbac.rpc.path.write.check` means
   for an existing regular file, including required ancestor search rights,
   symlink resolution, and behavior for unsupported file types. Keep the
   decision distinct from creating, replacing, or deleting a directory entry.
2. **Create and replace:** Define the inputs and permission checks for a new
   filename under an existing parent and for atomic replacement of an existing
   regular file. Cover a missing target, parent write/search permission,
   relevant target permissions, symlinks, and a target that changes between
   checks. Distinguish permission to write existing file content from
   permission to replace its directory entry. Include creation and later
   cleanup of the user-owned destination-adjacent `.upload` file. Define the
   parent write/search and sticky-directory rules required when the verified
   user attempts the final rename. State which decision File Service requests
   at upload initialization, each resumed write, and publication; RBAC does
   not infer the user's right to publish from an earlier write decision.
3. **Delete:** Define a separate delete decision for an existing regular file.
   Specify the parent-directory and target checks, including directory
   write/search permissions, sticky-directory behavior, symlinks, and the
   meaning of missing or unsupported targets. Do not infer delete permission
   from `rbac.rpc.path.write.check`. RBAC returns only a decision; File Service
   enforces its protected-path policy and performs the unlink.
4. **Traverse:** Define how RBAC establishes execute/search permission on every
   directory from `/` to the requested path, including resolved symlink
   targets. Decide whether the existing execute RPC can express the complete
   traversal decision or whether a distinct contract is required. Document
   how read, listing, write, and delete decisions incorporate traversal so a
   caller cannot gain access solely because a service process can reach a path.
5. **Path identity:** Define the path form accepted by these decisions and how
   callers relate an RBAC decision to the object or parent they operate on.
   RBAC must not claim that a path-only decision prevents a
   later filesystem race; File Service and Stream Service must use their own
   race-resistant file operations and reject a changed target when the
   authorization no longer applies.
6. **Failure behavior:** Define denial for invalid UID/path inputs, unresolved
   identities, inaccessible or missing required ancestors, ACL evaluation
   errors, unsafe symlink resolution, and other cases where RBAC cannot make a
   trustworthy decision. Keep transport failures distinct from an allow
   decision; callers fail closed on either a deny or an RPC failure.

## Implementation Actions

1. Inspect the current RBAC checks, shared DTOs and subjects, NATS permissions,
   and the File Service and Stream Service task requirements. Write down the
   current behavior and each contract gap above.
2. Settle the operation semantics and update the RBAC requirements and shared
   contract definitions. Update the File Service and Stream Service tasks where
   they name an RPC whose approved semantics differ from their assumptions.
3. Implement the approved decision checks and NATS handlers. Validate each
   incoming request at the RPC boundary with the existing validator pattern,
   and keep permission evaluation in RBAC.
4. Add focused, requirement-traceable RBAC tests using controlled identities,
   directory trees, and ACL fixtures. Cover existing-file writes, creation,
   replacement, deletion, traversal through each ancestor, symlinks, sticky
   directories, malformed requests, and failures. Include cases where a leaf
   permits access but an ancestor denies search, and where a parent permits
   creation but the target does not yet exist.
5. Verify the shared contract and NATS authorization changes needed for File
   Service and Stream Service callers. Run the relevant RBAC checks and report
   the commands and results.

## Completion Criteria

- The documented contracts tell File Service which decision to request for
  writing existing content, creating, replacing, and deleting a regular file.
- The documented contracts tell both File Service and Stream Service how to
  establish caller traversal rights before accessing a file or directory.
- Missing required parents, missing targets for operations that require an
  existing file, denied ancestors, unsafe paths, and evaluation failures
  cannot produce an allow decision. A missing target is allowed only through
  an explicitly defined creation decision.
- RBAC remains a decision service and performs no caller file mutation.
