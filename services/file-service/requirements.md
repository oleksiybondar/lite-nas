# File Service — Requirements

This document defines software requirements for the LiteNAS File Service.
Requirement IDs use the `file-service-svc/<TYPE>-<NNN>` format so tests and
implementation changes can trace back to their source.

## Overview

The File Service provides authenticated file browsing, upload, download,
directory creation, copy, move, and deletion, including resumable transfers.
It validates the caller's LiteNAS access token and delegates every
caller-specific filesystem permission decision to RBAC. It runs as a non-root
service and uses shared LiteNAS infrastructure and native Linux/Unix
facilities where appropriate.

## Functional Requirements

### file-service-svc/FR-001 Authenticate transfer requests

#### file-service-svc/FR-001 Description

The service MUST require a valid access token for every operation that creates,
queries, pauses, resumes, writes, finalizes, cancels, or downloads a transfer,
lists a directory, mutates a file or directory, or reads a job.

#### file-service-svc/FR-001 Input

- Access token presented using the documented service transport
- Transfer operation and request data

#### file-service-svc/FR-001 Output

- Authenticated principal context or an authentication failure

#### file-service-svc/FR-001 Acceptance Criteria

- Missing, malformed, invalid, expired, or rejected tokens are denied before
  protected file processing.
- Browser requests can authenticate with the existing `lite-nas-at` access
  cookie, configured `Secure`, `HttpOnly`, `SameSite=Lax`, and `Path=/`; token
  validation remains the service's responsibility.
- Nginx forwards the access cookie over the HTTPS `/f/` and `/d/` routes. The
  File Service does not issue a replacement browser session cookie.
- The authenticated username comes only from verified token claims.
- Token roles or scopes do not grant file access by themselves.
- The service follows the existing Auth/shared token-validation contracts and
  documents the behavior for token revocation/current status.

### file-service-svc/FR-002 Resolve the RBAC subject

#### file-service-svc/FR-002 Description

For an authenticated principal, the service MUST resolve the principal through
RBAC and use the returned UID for path capability checks.

#### file-service-svc/FR-002 Input

- Username from verified token claims
- RBAC subject-resolution request

#### file-service-svc/FR-002 Output

- RBAC subject UID or a failure outcome

#### file-service-svc/FR-002 Acceptance Criteria

- Subject resolution uses the existing `rbac.rpc.subject.roles.get` contract.
- A missing subject or failed/invalid RBAC response denies the operation.
- The service never accepts a client-supplied UID as the authorization subject.

### file-service-svc/FR-003 Authorize downloads through RBAC

#### file-service-svc/FR-003 Description

Before opening or reading a file, the service MUST ask RBAC whether the token
owner may read the actual resolved target path.

#### file-service-svc/FR-003 Input

- RBAC UID for the token owner
- Safely resolved file path

#### file-service-svc/FR-003 Output

- RBAC allow/deny decision

#### file-service-svc/FR-003 Acceptance Criteria

- The read check uses `rbac.rpc.path.read.check` and the actual target path.
- Before the RBAC check, non-administrators may read only within `/home`,
  configured storage write roots, and mounted ZFS filesystems outside system directories.
  Members of the `admin` or `sudo` group may request system paths, subject to
  the same RBAC read check.
- Denial, timeout, transport error, malformed response, or unresolved subject
  prevents file access.
- Every new, resumed, or continued download request is authenticated and
  authorized again.
- Client path traversal and encoding tricks cannot escape the configured
  storage boundary.
- The first version may use fixed `/bin/cat` through `sudo -u` as the verified
  non-root Unix user after RBAC allows the read. Its pathname replacement race
  is explicitly accepted for this version and tracked in
  [READ_RACE_REVIEW.md](READ_RACE_REVIEW.md). This exception does not relax
  token validation, RBAC checks, or storage-root checks.

### file-service-svc/FR-004 Authorize uploads through RBAC

#### file-service-svc/FR-004 Description

Before creating or replacing a destination, the service MUST ask RBAC whether
the token owner may mutate its directory entry. The user-scoped writer checks
actual write access to the staged file.

#### file-service-svc/FR-004 Input

- RBAC UID for the token owner
- Safely resolved intended target path

#### file-service-svc/FR-004 Output

- RBAC allow/deny decision

#### file-service-svc/FR-004 Acceptance Criteria

- Upload initialization, resumed/chunk writes, and finalization use
  `rbac.rpc.path.operation.check` with `create` for an absent destination or
  `replace` for an existing regular destination. Each request needs a fresh
  decision for the final intended target path. A changed target is rejected.
- The existing `rbac.rpc.path.write.check` applies only to writing the content
  of an existing regular file; it does not authorize upload publication.
- Denial, timeout, transport error, malformed response, or unresolved subject
  prevents file writes.
- Client paths and filenames are validated before affecting storage paths.
- Upload creation and finalization also reject targets in built-in or
  configured protected path trees. Check protection before accepting upload
  data and recheck immediately before publishing, so configuration or path
  changes cannot expose a protected target to overwrite.
- Mutation targets MUST be beneath `filesystem.write_roots` as well as outside
  `filesystem.protected_paths` and authorized by RBAC. The default write root is
  `/home`; ZFS mount paths are added explicitly by deployment/configuration.
  RBAC must still allow the token owner to write the selected target.
- An existing regular file may be replaced only after the complete new upload
  passes size/checksum validation and RBAC authorization. Publish by atomic
  same-filesystem replacement; until that succeeds, the previous file remains
  intact. If atomic replacement cannot be guaranteed, fail and preserve the
  existing file.
- Existing directories and symbolic links are not upload replacement targets;
  reject them without following or replacing them.
- The destination-adjacent `.upload` file is created as the authenticated Unix
  user by [lite-nas-file-writer](../../apps/file-writer/README.md) after
  RBAC approval. Both conventional and resumable upload bytes are written as
  that same user through the writer; the service account does not write or
  read the stage directly or require a host write group. A newly published
  file therefore has that user's UID without a
  post-upload ownership change. Its group, mode, and ACL come from the stage,
  including when an existing file is replaced. Do not copy the old file's
  permissions or ACL onto the replacement.

### file-service-svc/FR-005 Support pausing and resuming uploads

#### file-service-svc/FR-005 Description

The service MUST let a client pause an upload and continue it later without
resending bytes already accepted by the service.

#### file-service-svc/FR-005 Input

- Authenticated upload create, pause, progress, resume, content, finalize, and
  cancel operations as defined by the documented transfer protocol

#### file-service-svc/FR-005 Output

- Transfer identity and state sufficient to continue or finish the upload

#### file-service-svc/FR-005 Acceptance Criteria

- The protocol defines upload creation, explicit pause/resume, progress query,
  content submission, finalization, cancellation, and expiration.
- Two upload modes are provided: conventional `POST /f/upload` with
  `multipart/form-data` for browser/form clients, and a separate resumable app
  protocol under `/f/proprietary/upload`. The conventional mode does not promise
  resume support; its multipart request carries a `destination` field and one
  `file` part. Both modes replace an existing regular-file destination on
  successful completion, subject to RBAC and protected-path checks.
- Both upload modes stream request data to the destination-adjacent `.upload`
  file. The conventional multipart mode parses parts incrementally with bounded
  working memory. Neither mode spools a complete upload to `/tmp`, the host SD
  card, or tmpfs before writing it to the destination filesystem. The multipart
  `destination` part precedes the `file` part so path validation and RBAC
  authorization precede direct streaming. Conventional file content is limited
  to 10 GiB. The proprietary mode has no configured whole-file size limit;
  storage and 64-bit offsets impose practical bounds.
- The resumable service routes are:
  - `POST /proprietary/upload` to create an upload resource.
  - `GET` and `HEAD /proprietary/upload/{upload-id}` to query progress.
  - `PATCH /proprietary/upload/{upload-id}` to append the next chunk.
  - `PUT /proprietary/upload/{upload-id}/completion` to complete it.
  - `DELETE /proprietary/upload/{upload-id}` to cancel it.
  Nginx exposes these under `/f/` by preserving the suffix after `/f`.
- `POST /proprietary/upload` creates an upload resource using destination, original
  filename, total size, whole-file SHA-256, supported file metadata, an optional
  `chunk_size` in bytes from 1 MiB through 16 MiB, and a
  client `Idempotency-Key`. Success returns `201 Created` and `Location`.
  Repeating the same key and metadata returns the existing resource with
  `200 OK`; reusing that key with different metadata returns `409 Conflict`.
  Creation validates the destination, rejects protected paths, and obtains RBAC
  `create` or `replace` authorization before accepting the upload.
- Resumable data uses `PATCH /proprietary/upload/{upload-id}` with ordered chunks and the
  committed byte offset in `Upload-Offset`. When creation omits `chunk_size`,
  the service setting applies; it defaults to 1 MiB and is configurable from
  1 MiB through 16 MiB. Initialization records and returns the effective
  chunk size for that upload; resume uses the recorded size even after
  configuration changes. Every non-final chunk equals the recorded size; only
  the final chunk may be shorter. Success returns `204 No Content` and the
  updated `Upload-Offset` and `Upload-Next-Chunk-Index` headers. An identical
  replay of the last committed chunk is idempotent; stale offsets, conflicting
  replay, and out-of-order chunks return an error with the current offset/index.
- `GET /proprietary/upload/{upload-id}` returns progress JSON.
  `HEAD /proprietary/upload/{upload-id}`
  returns progress metadata in headers without a body. Pausing means the client
  stops sending; it resumes by querying GET or HEAD.
  `DELETE /proprietary/upload/{id}`
  cancels the upload. `PUT /proprietary/upload/{id}/completion` verifies declared
  size and SHA-256, rechecks protected-path policy and RBAC `create` or `replace` authorization,
  then atomically publishes the completed target. If a regular file already
  exists, publication atomically replaces it; any failure beforehand leaves the
  existing file intact. Repeating the request is idempotent.
- Authenticated `GET /proprietary/upload-window` reports available chunk-buffer bytes and
  maximum accepted chunk size. A valid PATCH that cannot reserve capacity
  returns `503 Service Unavailable`, `Retry-After`, and JSON with
  `error.code=upload_window_full`; it accepts no bytes and advances no offset.
  Clients wait and probe the window before retrying, and still handle capacity
  changing between the probe and PATCH.
- The service reserves buffer capacity after validating PATCH request headers
  and before reading the body. If capacity is unavailable, it returns the
  on-hold response without accepting or buffering chunk content.
- Resumable success responses are JSON and report `upload_id` and
  `next_chunk_index` where applicable. Errors return a non-success HTTP status
  and JSON with a stable `error.code`; chunk sequencing errors report the
  current expected index. The complete error-code and HTTP status mapping is
  documented before implementation, and errors do not leak inaccessible host
  paths.
- Concurrency is limited by chunk-buffer bytes rather than a fixed upload
  count. A 64 MiB size-limited tmpfs reserves 4 MiB for metadata working files
  and 60 MiB for incoming chunk buffers. Admission control MUST account for
  each request's maximum buffered bytes and never reserve more than 60 MiB for
  chunk bodies. At the configured 1 MiB default this permits up to 60
  simultaneous receiving requests in theory; at 16 MiB it permits up to 3,
  before implementation overhead and other limits. Paused uploads consume no
  chunk-buffer reservation. Tmpfs holds only in-flight data and bounded
  working metadata; accumulated file bytes are durable on the destination
  filesystem, so a 100 GiB upload does not require a 100 GiB temporary volume.
- The 4 MiB metadata reservation is a hard limit. New upload initialization is
  rejected when its bounded metadata would exceed the reservation; cancellation
  and expiration release the associated reservation.
- Both upload modes store accumulated bytes in the destination directory at
  `<destination>.upload`; for `/path/file.ext`, use `/path/file.ext.upload`.
  Reserve that path for one active upload to the target. Reject an unrelated
  existing file or symlink. Create the stage exclusively as the authenticated
  Unix user after an RBAC create decision, and write every chunk as that same
  user. A literal `touch` is insufficient because it can modify an existing
  path; use exclusive creation semantics. Check the stage's owner and identity
  before each resumed write and publication. If user-scoped creation fails,
  accept no upload bytes. Bind continuation to the recorded
  upload owner and target. Host filesystem clients may see the incomplete
  `.upload` file; it is not the completed destination.
- Keep durable resume metadata in a private, permission-restricted service
  area associated with the upload ID and canonical destination. Do not create
  a `.meta` sidecar beside the target. Durable metadata remains authoritative
  across process restart and reboot; the tmpfs metadata reservation is only
  for working copies. Each record stores the owner, canonical target, stage
  identity, declared size and checksum, effective chunk size, committed offset
  and index, idempotency key, and expiry. An incomplete upload expires 24 hours
  after its last committed activity. A separate worker scans at startup and
  every three hours using the shared timer infrastructure; it verifies stage
  identity before cleaning expired content and metadata. Any in-memory index
  is rebuilt from durable records after restart.
- Each accepted chunk is appended at the committed offset to the `.upload`
  file by the user-scoped operation; the implementation does not rewrite all
  earlier bytes or append
  partial bytes to the completed destination. Before acknowledging a
  chunk, the service makes its data durable and atomically updates the
  committed offset/chunk metadata. Recovery truncates uncommitted trailing
  bytes to that offset as the user; staged checksum verification also reads as
  the user. A partial write can be retried at the same index.
- A transfer failure before a chunk is committed requires retrying only that
  chunk. The server discards any partial tail after the last committed offset;
  it does not rewrite earlier accepted chunks.
- The service detects invalid offsets, conflicting content, unknown/expired
  transfers, and incomplete finalization with defined failure outcomes.
- Transfer continuation is bound to its authenticated owner and intended
  target; another principal cannot continue or finalize it.
- Interrupted or incomplete data is never exposed as a completed target.
- After full validation and fresh authorization, attempt the same-directory
  atomic rename of `.upload` to the target as the authenticated Unix user.
  Unix directory and sticky-bit permissions may still deny publication. An existing
  target remains intact until publication succeeds. On a failed rename,
  attempt stage cleanup and report failure. Destination-pool quota or space
  exhaustion fails the upload without publishing partial content; available
  capacity is not presumed to be reserved for the full declared size.

### file-service-svc/FR-006 Support pausing and resuming downloads

#### file-service-svc/FR-006 Description

The service MUST let clients pause a download and continue from the last
confirmed byte position using documented continuation semantics.

#### file-service-svc/FR-006 Input

- Authenticated file identity and requested continuation position
- Optional HTTP byte range and strong file-version validator

#### file-service-svc/FR-006 Output

- File length and the content/range actually served, or a defined failure

#### file-service-svc/FR-006 Acceptance Criteria

- The response gives clients sufficient range and length information to resume.
- `HEAD` and full `GET` responses expose the file length, `Accept-Ranges: bytes`,
  and a strong `ETag` identifying the exact file representation. A single-range
  `GET` returns `206 Partial Content`, `Content-Range`, the range length, and the
  same ETag.
- Authorized `HEAD`, full `GET`, and single-range `GET` responses use consistent
  `Content-Type`, `Content-Disposition`, and `X-Content-Type-Options: nosniff`
  headers for the opened file within each mount. `HEAD` has no body.
  `Content-Length` describes the full representation on `HEAD` and full `GET`,
  and the selected range on `206`.
- The same file is available for `GET` and `HEAD` at both
  `/f/{absolute-path...}` and `/d/{absolute-path...}`. Both mounts use the same
  token validation, RBAC read authorization, file identity, strong ETag, and
  range semantics; each request is independently authorized.
- The service accepts multiple independent concurrent `GET` requests for
  different byte ranges of the same file. A download manager can use these
  ordinary REST requests to fetch segments in parallel; no server-side download
  session or multipart response is required.
- Parallel segment requests carry `If-Match` with the strong ETag learned from
  `HEAD` or an initial `GET`. If the target changes, the service returns
  `412 Precondition Failed` instead of serving a segment from a different file
  version. The client restarts against the new ETag.
- The service supports up to 16 simultaneous range responses per authenticated
  principal and file by default. Configuration may set a lower or higher cap;
  global stream/resource limits still apply. Excess requests receive a
  documented retryable response; per-principal cap exhaustion returns `429`
  with `Retry-After`.
- A single request containing multiple ranges is not required; the parallel
  download client uses one byte range per `GET`.
- Invalid or unsatisfiable ranges receive a defined failure and never read
  outside the file.
- Pausing stops transfer work and releases request-scoped resources.
- Each resumed request revalidates the token and repeats the RBAC read check.
- Every parallel range request independently validates the token, resolves the
  subject, and obtains RBAC read authorization before opening/reading the file.
- Continuation cannot change file identity or bypass path validation.

### file-service-svc/FR-007 Preserve arbitrary file content

#### file-service-svc/FR-007 Description

The service MUST treat file data as opaque bytes and support transfers
independently of media format or filename extension.

#### file-service-svc/FR-007 Input

- Client file content and a valid authorized file identity

#### file-service-svc/FR-007 Output

- The original bytes accepted by upload and returned by download

#### file-service-svc/FR-007 Acceptance Criteria

- Content is not rejected or transformed solely because its format or
  extension is unfamiliar.
- Completed upload and subsequent download preserve bytes exactly.
- On `/f/`, reliably identified PDF, JPEG, PNG, GIF, and WebP content uses its
  matching media type and `Content-Disposition: inline`. The inline decision
  requires file-content identification; the extension or a client-supplied
  media type alone cannot make content inline.
- On `/f/`, archives and content outside that inline allowlist use
  `Content-Disposition: attachment` without redirecting to `/d/`. An archive
  uses its accurate media type when reliably identified; otherwise the
  response uses `application/octet-stream`. Unknown or potentially active
  content, including HTML and SVG, uses `application/octet-stream` and
  attachment.
- On `/d/`, every file uses `Content-Type: application/octet-stream` and
  `Content-Disposition: attachment`, including files that `/f/` could display
  inline. Both mounts return identical file bytes.
- `Content-Disposition` provides a sanitized basename as the suggested
  filename, with non-ASCII names encoded according to HTTP header rules. Path
  components, header-control characters, and client-provided metadata cannot
  affect response headers. Metadata is determined for the authorized, opened
  file and is not disclosed before authentication and RBAC authorization.
- Media type and disposition do not affect authorization. Browsers decide
  whether they can display an inline file.

### file-service-svc/FR-008 List only RBAC-accessible entries

#### file-service-svc/FR-008 Description

The service MUST provide an `ls` operation that filters directory results
through a storage visibility boundary and the authenticated token owner's
effective path access as decided by RBAC.

#### file-service-svc/FR-008 Input

- `GET /ls/{absolute-path...}` request with authenticated target directory
- Presence-based `recursive` and `files-only` query flags
- RBAC UID for the authenticated token owner

#### file-service-svc/FR-008 Output

- Authorized directory entries encoded as JSON `files` and `directories`
  arrays of absolute host paths, or an error JSON response

#### file-service-svc/FR-008 Acceptance Criteria

- The initial directory must pass RBAC read and execute/search checks, and all
  ancestor directories from host root `/` must pass
  execute/search checks.
- Non-administrators may list `/home`, configured storage roots, mounted ZFS trees
  outside system directories, and the ancestors needed to reach those trees.
  Root listings omit system entries such as `/etc`, `/var`, `/sys`, and `/proc`.
  Members of `admin` or `sudo` may list system paths when RBAC permits them.
- A file is included only when RBAC allows read access to it and its ancestor
  directories from `/` are searchable by the subject.
- A directory is included and traversed only when RBAC allows both read and
  execute/search on that directory and execute/search on its ancestors.
- Without `recursive`, the service returns eligible direct children only;
  including `recursive` descends only into eligible directories.
- Including `files-only` omits directory entries while including eligible
  regular files within the requested traversal scope.
- Symlink targets are resolved and checked as absolute host paths; symlink
  listing behavior is documented.
- Denied entries are omitted without exposing their names or metadata. An RBAC
  dependency/evaluation error fails the listing rather than returning an
  apparently complete partial result.
- Presence of `recursive` or `files-only` enables that flag; absence disables
  it. `/ls/abs/path?recursive&files-only` enables both on the public listing
  route, separately from the `/f/` file-operation mount.
- A successful response is JSON with `files` and `directories` arrays. Both
  contain absolute host paths; `files-only` returns an empty `directories`
  array.
- A failed request returns a non-success HTTP status and a JSON object with an
  `error` key containing a documented stable code. Error output does not reveal
  denied paths or entry names.

#### file-service-svc/FR-008 Notes

The current RBAC contracts expose separate path read and execute decisions.
Use execute access for directory search/traversal, read access for directory
contents and file content, and evaluate ancestor directory execute access as
well as the leaf path. The service process's own permissions are not evidence
of caller access.

### file-service-svc/FR-009 Delete authorized files and directories outside protected paths

#### file-service-svc/FR-009 Description

The service MUST allow an authenticated user to delete a file or directory
only when RBAC authorizes the token owner to delete it and the target is
outside all configured protected path trees.

#### file-service-svc/FR-009 Input

- `DELETE /f/{absolute-path...}` request, mapped by Nginx to the service path
- RBAC UID for the authenticated token owner
- `filesystem.protected_paths` configuration

#### file-service-svc/FR-009 Output

- `204 No Content` for a regular file, or `202 Accepted` with a job ID for a
  directory

#### file-service-svc/FR-009 Acceptance Criteria

- The public route `DELETE /f/{absolute-path...}` maps the path suffix to the
  absolute host path using the same safe path handling as other file routes.
- The operation requires a valid access token and a fresh
  `rbac.rpc.path.operation.check` decision with `delete` for the authenticated
  token owner's UID and resolved target. Upload/write permission does not
  authorize deletion.
- The configured protected path list is read from
  `filesystem.protected_paths`, contains canonical absolute paths, and
  lets operators add protected roots. It applies to upload creation, upload
  publication/replacement, and deletion. Operator configuration MUST NOT remove
  built-in protected OS/runtime roots. Installed defaults MUST include `/etc`
  and platform-critical paths such as `/usr`, `/bin`, `/sbin`, `/lib`, `/lib64`,
  `/boot`, `/proc`, `/sys`, `/dev`, `/run`, and `/var/lib` where present.
- `filesystem.write_roots` is a canonical absolute-path allowlist for all
  service mutations, including uploads, replacements, and deletes. Its default
  includes `/home`; ZFS mount paths must be explicitly configured. The service
  denies writes outside this list even when RBAC allows them.
- Protected-path matching uses path-component boundaries after safe
  normalization/resolution: a protected root and every descendant are denied,
  while a sibling with a similar string prefix (for example `/etc-backup`) is
  not treated as a child of `/etc`.
- Protected-path denial takes precedence over RBAC allow. Neither a successful
  RBAC decision nor service-process filesystem permissions can override the
  protected list.
- Directories are removed recursively by an in-memory job after each descendant
  receives policy and RBAC authorization. Symbolic links and special files are
  rejected and never followed.
- The implementation authorizes the target and required parent-directory
  operation according to the agreed RBAC contract, then removes the entry
  through fixed user-scoped commands.
- Missing paths, paths outside service policy, RBAC denial/errors, protected
  paths, and unsupported file types return documented stable JSON error codes
  with non-success HTTP statuses and do not disclose protected host details.
- Cookie-authenticated DELETE requests pass the service's CSRF checks.

#### file-service-svc/FR-009 Notes

The RBAC `delete` operation checks parent write/search and sticky-directory
rules for an existing file or directory. File Service enforces protected paths.

### file-service-svc/FR-010 Create directories and transfer trees

#### file-service-svc/FR-010 Description

The service MUST create authorized directories and recursively copy or move
authorized files and directory trees. Existing destination directories are
merged and existing regular files are replaced.

#### file-service-svc/FR-010 Input

- `POST /f/mkdir` with `{ "path": "/absolute/new-directory" }`
- `POST /f/copy` or `POST /f/move` with
  `{ "source": "/absolute/source", "destination": "/absolute/target" }`
- Authenticated subject resolved through Auth and RBAC

#### file-service-svc/FR-010 Output

- `201 Created` with the created path for `mkdir`
- `202 Accepted` with a job ID and initial progress for copy or move
- Stable non-success JSON error for invalid or denied requests

#### file-service-svc/FR-010 Acceptance Criteria

- `destination` names the complete target path, including the copied entry's
  filename or directory name. Source and destination paths are clean absolute
  host paths. The source MUST exist and MUST be a regular file or directory;
  symlink and special-file sources are rejected.
- `mkdir` MUST create exactly one absent directory. Copy and move MUST
  recurse through directories, merge existing destination directories, and
  replace existing regular files. File-versus-directory type conflicts and
  copying or moving a directory into itself MUST fail.
- Before mutation, the service MUST enforce write-root and protected-path
  policy, caller read permission on every copied source entry, the appropriate
  RBAC create or replace decision for destination entries, and delete
  permission on entries removed by Move. A policy or RBAC failure MUST NOT be
  treated as permission to continue.
- Copied file bytes MUST be written as the verified non-root Unix user through
  the staged writer, checked for byte integrity, and atomically published.
  Existing destination content MUST remain in place until publication.
- Move MUST copy the complete source tree before removing source entries from
  leaves to root. A partial failure MUST be reported by the job and MUST NOT
  claim an atomic whole-tree move.

### file-service-svc/FR-011 Report queued operation progress

#### file-service-svc/FR-011 Description

The service MUST execute copy, move, and recursive directory deletion as
server-side jobs that continue after the initiating browser request ends.

#### file-service-svc/FR-011 Input

- An accepted copy, move, or directory-delete operation
- `GET /f/jobs/{id}` by an authenticated caller

#### file-service-svc/FR-011 Output

- `202 Accepted` with `job_id`, `state`, `percent`, `completed_bytes`,
  `total_bytes`, `completed_items`, and `total_items`
- Current status from `GET /f/jobs/{id}`, including `error_code` on failure

#### file-service-svc/FR-011 Acceptance Criteria

- States MUST distinguish pending, running, done, and error. Percent MUST
  reflect server-side completed bytes and items and reach 100 on success.
  Large file copies MUST advance byte progress as chunks finish.
- The service MUST queue jobs in memory and process them sequentially with Go
  goroutines and input/output channels. An accepted job MUST continue after a
  browser reload or client disconnect while File Service remains running.
- Job lookup MUST require a fresh authenticated subject and disclose a job
  only to its submitting username and UID. A missing job ID MUST return a
  stable non-success response.
- In-memory jobs do not survive a File Service restart. Clients MUST be able
  to distinguish a missing job from a still-running job.

## Interface Requirements

### file-service-svc/IR-001 Define the transfer interface

#### file-service-svc/IR-001 Description

The service MUST expose a documented interface for upload/download workflows
and transfer continuation.

#### file-service-svc/IR-001 Input

- Authenticated client transfer requests

#### file-service-svc/IR-001 Output

- Transfer responses and file content appropriate to each documented
  operation

#### file-service-svc/IR-001 Acceptance Criteria

- The API contract defines transport, authentication transport, request and
  response fields, error mapping, pause/resume semantics, transfer expiry, and
  configured size/part limits before implementation.
- The contract supports upload create/chunk/pause/resume/status/finalize/cancel,
  download pause/resume/continuation and concurrent byte-range downloads, and
  `ls` with `recursive` and `files-only` query parameters without treating
  transfer or job IDs as credentials. It also defines `mkdir`, `copy`, `move`,
  file and directory deletion, and job progress lookup.
- Each download mount uses the same resource URL for `HEAD` metadata and `GET`
  requests. Parallel segments are separate single-range `GET`s; the API does
  not require a download-session endpoint or multipart-range response.
- Public `GET` and `HEAD /d/{absolute-path...}` request an attachment for the
  same file selected by `/f/{absolute-path...}`. Nginx forwards them to the
  File Service's explicit download handler. Other methods on `/d/` are denied.
- HTTP methods follow resource semantics: `POST /proprietary/upload` creates
  an upload,
  `GET` retrieves its state, `HEAD` retrieves resume metadata without a body,
  `PATCH` appends bytes at the committed offset, `PUT` creates the idempotent
  completion resource, and `DELETE` cancels an incomplete upload. The ordinary
  browser endpoint uses `POST` with `multipart/form-data`.
- `GET /f/proprietary/upload-window` is authenticated and reports current
  buffer availability.
  If a valid PATCH cannot reserve capacity, it returns `503` with `Retry-After`
  and a retryable JSON error, accepts no body bytes, and leaves the upload
  offset unchanged. The client probes the window and handles capacity races.
- Listing tests verify the successful JSON keys and absolute-path arrays,
  empty `directories` with `files-only`, and non-success status plus an `error`
  code for failures without leaking denied paths or entry names.
- Public `GET /f/docs` serves Swagger UI and `GET /f/openapi.json` serves the
  maintained OpenAPI document. The service routes are `/docs` and
  `/openapi.json`; Nginx strips `/f` when forwarding. `api_docs.enabled`
  controls both routes. Production configuration sets it to `false`; development
  and test configurations set it to `true`. Operators may override it. When
  false, both routes return `404`. The UI and schema endpoints use the same HTTPS origin and require a
  valid access token. The UI's API server URL and spec fetch path preserve the
  public `/f` prefix so interactive requests return through Nginx.
- Swagger UI assets are served locally by the service; runtime page rendering
  does not depend on a third-party CDN.
- The public `ls` route is `GET /ls/{absolute-path...}`; for example,
  `/ls/abs/path` lists host directory `/abs/path`. It is not prefixed by `/f/`.
- The public mutation routes are `POST /f/mkdir` with `{path}` and
  `POST /f/copy` or `POST /f/move` with `{source,destination}`. The latter
  return `202` and a job status body. `DELETE /f/{absolute-path...}` returns
  `204` for a regular file and `202` with a job body for a directory.
  `GET /f/jobs/{id}` returns the submitting owner's current job status.
- The public HTTP paths are mounted by Nginx at `/f/` and `/d/`; path forwarding
  and mapping each route suffix to the same absolute host path are documented.
  Nginx forwards `/f/docs`, `/f/openapi.json`, `/f/upload`,
  `/f/proprietary/upload...`, `/f/proprietary/upload-window`, `/f/mkdir`,
  `/f/copy`, `/f/move`, `/f/jobs/{id}`, and generic
  `DELETE /f/{absolute-path...}` requests to their matching service routes
  after stripping the `/f` prefix. Explicit API routes take precedence over the
  generic file-path handler.
- Nginx forwards `GET` and `HEAD /d/{absolute-path...}` to the explicit
  download handler. The `/d/` mount exposes no upload or delete route.
- Nginx uses `proxy_request_buffering off` for both upload routes. It streams
  PATCH bodies so the service can apply buffer admission before reading chunk
  content, and streams conventional multipart bodies without writing a
  complete request body to Nginx's client-body temporary directory. Proxy HTTP
  version and request-size settings support the documented maximum upload size.
- The path suffix after `/f` or `/d` maps to an absolute host path; for
  example, `/f/home/alice/photo.jpg` and `/d/home/alice/photo.jpg` both map
  to `/home/alice/photo.jpg`. RBAC evaluates the resolved host path; mapping
  alone does not grant access.
- For the listing route, the suffix after `/ls` identifies the absolute host
  directory to list; for example, `/ls/abs/path` lists `/abs/path`.
- HTTP DTO validation follows the repository's Huma requirements; non-HTTP
  message inputs use the repository's Go boundary-validation conventions.

### file-service-svc/IR-002 Use protected LiteNAS transports

#### file-service-svc/IR-002 Description

The service MUST use LiteNAS's established TLS, NATS messaging, and HTTP proxy
topology for internal and external communication.

#### file-service-svc/IR-002 Input

- Service transport configuration and provisioned service certificates

#### file-service-svc/IR-002 Output

- Authenticated, encrypted connections at the documented trust boundaries

#### file-service-svc/IR-002 Acceptance Criteria

- NATS uses the shared messaging configuration, dedicated client certificate
  identity, trusted CA, and private key.
- NATS authorization grants only required Auth and RBAC subjects.
- Internal service-to-service calls use NATS; HTTP is the client-facing
  transport and is not used as an internal service RPC path.
- The service HTTP(S) listener uses its own configured port and binds only to
  loopback; it is not directly reachable from non-local interfaces.
- Nginx exposes the service at `/f/` and `/d/` using the existing TLS and
  reverse-proxy configuration conventions.
- Browser access is through the HTTPS Nginx origin so Secure cookies are sent;
  the service's loopback listener is not a browser-facing endpoint.
- Required certificate paths and TLS configuration are validated at startup.

## Reliability Requirements

### file-service-svc/RR-001 Fail closed on dependency or evaluation failure

#### file-service-svc/RR-001 Description

The service MUST deny protected operations when authentication, RBAC, or safe
path resolution cannot produce a trustworthy result.

#### file-service-svc/RR-001 Acceptance Criteria

- No dependency failure is interpreted as allow.
- No file data is read or written after a failed authorization check.
- Interrupted transfer state is either valid and resumable or safely rejected
  and cleaned up.

### file-service-svc/RR-002 Handle transfer interruption safely

#### file-service-svc/RR-002 Description

The service MUST preserve filesystem consistency on client disconnect, process
shutdown, or dependency failure.

#### file-service-svc/RR-002 Acceptance Criteria

- Partial uploads are not marked complete.
- Client disconnects stop request-scoped reads/writes and release resources;
  already accepted server-side jobs continue under File Service ownership.
- Incomplete transfer metadata and temporary data follow the documented
  expiration and cleanup policy.
- Accepted in-memory copy, move, and directory-delete jobs continue after
  client disconnect. A service restart loses their status and does not claim
  they completed; copied files already published remain subject to normal
  filesystem state.

## Operational Requirements

### file-service-svc/OR-001 Run non-root with least-privilege file access

#### file-service-svc/OR-001 Description

The service MUST run as a dedicated non-root account with only the host file
access required to perform its service role.

#### file-service-svc/OR-001 Acceptance Criteria

- Runtime effective UID is nonzero; operation does not depend on root.
- Local deployment and Debian install/upgrade flows idempotently create
  `lite-nas-fs-rd`. File Service and Stream Service join that group. It grants
  file read and directory read plus search/traverse, with no write permission,
  within provisioned file-backed storage roots. It grants no access to excluded
  system-critical paths. Group membership alone is insufficient unless host
  group bits or ACLs actually provide read/search access, including on newly
  created private user files that the services must later download or stream.
- RBAC decides the requesting user's potential access before an operation.
  The restricted `lite-nas-file-writer` exclusively creates the stage, writes
  every conventional or resumable chunk, attempts the final rename, and handles
  user-owned cleanup; Unix permission checks can still reject these
  operations. The caller cannot select the Unix identity or bypass RBAC
  through service group membership.
- Read/search access is limited to supported file-backed filesystems; virtual
  kernel interfaces are not treated as ordinary file storage unless explicitly
  supported and configured.
- Standard OS/runtime paths, including `/etc`, `/usr`, `/bin`, `/sbin`, `/lib`,
  `/boot`, `/proc`, `/sys`, `/dev`, and `/run`, remain non-writable through
  service policy and host confinement. The HTTP listener MUST NOT receive
  broad `CAP_CHOWN` or run as root.
- A newly created `.upload` file is owned by the authenticated Unix user, so
  publication does not require `chown`. All writes also run as that user, so
  no service write ACL is needed. Group, mode, and ACL follow the stage on new
  files and replacements; the old destination's metadata is not preserved.
- Required storage and temporary upload permissions are documented and
  installed through LiteNAS deployment/package configuration.
- The File Service AppArmor profile is packaged as
  `configs/etc/apparmor.d/usr.libexec.lite-nas.file-service`, attaches to
  `/usr/libexec/lite-nas/file-service`, and is installed, validated, and loaded
  in enforce mode by local deployment and Debian install/upgrade flows through
  the repository's managed AppArmor profile loader.
  The profile limits access to the service's configured logs, certificates,
  tmpfs and durable staging, and host file paths needed for authorized
  operations. Fixed user-scoped remove and directory commands use a confined
  child profile; the writer has its own enforce-mode profile at
  `configs/etc/apparmor.d/usr.libexec.lite-nas.lite-nas-file-writer` and the
  transition through `sudo` must keep it confined. The service invokes only
  fixed commands as the verified non-root Unix user, without a shell; a
  restricted sudoers rule is not a substitute for path and policy checks.
- Deployment provides a 64 MiB bounded tmpfs available only to the service
  account, with 4 MiB reserved for metadata working files and 60 MiB for
  incoming chunk buffers. Use size-limited `tmpfs` rather than unbounded
  `ramfs`.
- Acknowledged chunks remain in the destination-adjacent `.upload` file and
  resume metadata remains in protected durable storage so an accepted chunk
  survives a process restart or host reboot. The tmpfs is only a bounded
  request-buffer area and is not the durable upload store.
- Weighted admission control rejects or waits on buffer requests that would
  exceed the 60 MiB chunk-buffer budget; tmpfs exhaustion is handled safely.
  The service never acknowledges a chunk that has not been persisted to
  the destination-adjacent `.upload` file and committed metadata.
- Host permissions do not replace the per-request RBAC allow decision.
- Staged file publication uses the writer's descriptor-relative checks.
  Fixed user-scoped read and directory commands still use pathnames; File
  Service applies policy and RBAC checks before invoking them. The accepted
  first-version `cat` pathname race is documented in
  [READ_RACE_REVIEW.md](READ_RACE_REVIEW.md).

### file-service-svc/OR-002 Integrate local build, test, deployment, and Debian

publication

#### file-service-svc/OR-002 Description

The service MUST be integrated into LiteNAS's local development and single
Debian-package workflows.

#### file-service-svc/OR-002 Acceptance Criteria

- Dedicated build, deploy, and test scripts follow existing repository
  conventions and are included in aggregate local workflows.
- Debian package build includes the service binary, config, systemd unit,
  AppArmor policy, and required runtime assets.
- Local deployment and Debian install/upgrade flows create the service account
  and read group and install the restricted user-scoped execution policy
  idempotently. Provision existing storage and newly created ZFS mounts with
  documented read/search group rules, while excluding system-critical paths.
- Package and install validation verify that the File Service AppArmor profile
  is installed and loaded in enforce mode.
- The package includes the root-owned `lite-nas-file-writer` executable,
  restricted sudoers configuration, and separate AppArmor policy. The writer
  is a short-lived application with no service/socket unit; the HTTP File
  Service remains non-root.
- Package analysis, contents validation, PR install validation, and supported
  architecture release workflows include the service.
- The service remains part of the existing single LiteNAS package.

## Security Requirements

### file-service-svc/SR-001 Protect credentials and transfer state

#### file-service-svc/SR-001 Description

The service MUST protect bearer tokens, private keys, transfer identifiers,
temporary content, and protected host path details.

#### file-service-svc/SR-001 Acceptance Criteria

- Tokens, private key material, and file contents are not written to logs.
- Transfer identifiers do not independently authorize access and are bound to
  owner and target metadata.
- Job identifiers do not independently authorize access; job lookup requires
  authentication and ownership checks.
- Cookie-authenticated upload and delete mutations enforce a documented CSRF
  defense at the service boundary; cookie attributes provide defense in depth.
- Temporary upload data and metadata are protected from unrelated users and
  cleaned up according to policy.
- Certificate permissions allow the non-root process to read only what it
  needs and do not expose CA private keys or unrelated service identities.

## Integration Requirements

### file-service-svc/IR-003 Share common service infrastructure

#### file-service-svc/IR-003 Description

Behavior that is identical across File Service and Stream Service MUST use
shared LiteNAS infrastructure.

#### file-service-svc/IR-003 Acceptance Criteria

- Common token verification, RBAC subject/path authorization, and safe
  file-access/path-resolution behavior reuse existing shared packages or the
  smallest new shared package required by both services.
- Upload-specific workflow remains owned by File Service.
- Shared behavior and contracts are explicit and are not duplicated across
  service modules.

### file-service-svc/IR-004 Prefer native Unix and mainstream Go facilities

#### file-service-svc/IR-004 Description

The service SHOULD use native Linux/Unix facilities and mainstream Go libraries
that keep implementation and operations simple.

#### file-service-svc/IR-004 Acceptance Criteria

- The Go standard library and native filesystem/process facilities are used
  where they safely meet the requirements.
- Additional dependencies are simple, mainstream, maintained, and minimal.
- Databases, external object stores, and specialized transfer frameworks are
  added only for documented requirements that existing platform facilities
  cannot safely meet.

### file-service-svc/IR-005 Rotate service logs

#### file-service-svc/IR-005 Description

The File Service log MUST be included in LiteNAS-managed log retention.

#### file-service-svc/IR-005 Acceptance Criteria

- The service writes its application log to
  `/var/log/lite-nas/file-service.log` using the shared logging configuration.
- The packaged LiteNAS logrotate drop-in rotates this log daily, retains seven
  rotations, compresses older rotations, and tolerates a missing or empty log.
- Rotation uses `copytruncate` while the shared logger keeps the file open, so
  rotation does not require restarting the service. If the logger later gains
  a reliable reopen mechanism, the rotation strategy may be changed with its
  behavior documented and covered by operational validation.
- Log rotation never changes log ownership or grants access beyond the service
  account and its authorized log readers.
- The logrotate configuration is included in the existing single Debian
  package and installed by the repository's config deployment flow.

## Testability Requirements

### file-service-svc/TR-001 Verify authorization and transfer behavior

#### file-service-svc/TR-001 Description

The service MUST be testable without production storage, live host accounts, or
production Auth/RBAC instances.

#### file-service-svc/TR-001 Acceptance Criteria

- Service-local tests cover missing/invalid tokens, RBAC allow/deny/errors,
  unsafe paths, transfer ownership, pause/resume, offsets, interruption,
  finalization, byte preservation, both upload modes, ordered chunk handling,
  idempotent chunk replay, resume after lost connection/response, checksum and
  size mismatch, weighted concurrency at 1 MiB and 16 MiB chunk settings,
  upload-window probing and on-hold responses, tmpfs exhaustion, and
  publish-only-after-finalize behavior. Download tests cover HEAD metadata,
  valid/invalid ranges, concurrent segments, matching and changed ETags, the
  concurrency cap, complete reassembly of one file version, and equal file
  bytes with the same ETag on `/f/` and `/d/`. Tests verify inline PDF/image
  headers on `/f/`, attachment headers for unsupported content on `/f/`,
  unconditional octet-stream attachments on `/d/`, and denial of mutation
  methods on `/d/`. Tests verify existing-file replacement for both upload
  modes, atomic visibility, preservation of old content after failed
  uploads/publish, and rejection of protected, directory, or symlink targets.
  Upload tests also verify user-owned stage creation, user-scoped chunk writes,
  user-scoped final rename, failed-rename cleanup, and that the final UID
  matches the RBAC subject UID. Group/mode/ACL follow the chosen creation and
  replacement policy. Tests cover delete RBAC allow/deny and
  dependency failures, protected-root and descendant matching, similar-prefix
  siblings, path races, missing targets, directory jobs and symlink rejection,
  stable error
  responses, and DELETE CSRF handling. Tests also cover enabled and disabled
  API docs routes, token enforcement, and `/f` prefix handling in Swagger UI.
- Operation coverage verifies `mkdir` path and RBAC validation; recursive
  copy and move of files and directories; existing-directory merge and
  existing-file replacement; type conflicts, symlinks, protected paths, and
  self-descendant destinations; byte preservation and staged publication;
  recursive deletion with `202` versus regular-file deletion with `204`;
  pending/running/done/error job transitions; byte and item progress; job
  ownership; and continued execution after client disconnect.
- Dependencies can be controlled or replaced in tests.

### file-service-svc/TR-002 Verify transport, certificates, and runtime access

#### file-service-svc/TR-002 Description

Transport security and non-root runtime access MUST have automated or repeatable
integration coverage.

#### file-service-svc/TR-002 Acceptance Criteria

- Configuration checks cover valid and missing/invalid TLS certificate paths.
- Host-access integration checks verify read-group file/directory read and
  search without write, no service-account write to host files,
  user-scoped creation/rename, service-account confinement, write-root and
  protected-path enforcement, system-path write denial, and final upload
  UID/group/mode/ACL.
- Protected-delete configuration validation rejects relative or malformed
  paths and verifies that installed defaults include the required OS/runtime
  protected roots.
- NATS integration verifies the service certificate identity, trusted CA, and
  restricted Auth/RBAC subject access; invalid credentials fail closed.
- The loopback-only listener, Nginx path mount, HTTP path forwarding, TLS
  boundary, and certificate bootstrap/rotation are checked.
- Browser integration coverage verifies that `lite-nas-at` is sent to `/f/`
  and `/d/` over HTTPS, is not sent over HTTP, and cannot be read by frontend
  JavaScript.
- Deployment checks verify the non-root account can access required files and
  certificates, but not unrelated certificate material.
- Host-access checks verify that directory creation, removal, and staged
  copies execute as the verified non-root Unix user under their AppArmor
  profiles, while the File Service account cannot directly mutate host files.
- Listing tests cover allowed/denied files and directories, ancestor search
  permissions, recursive traversal, `files-only`, symlinks, and RBAC failures.
- Cookie-authenticated upload and DELETE mutation tests cover accepted
  same-origin requests and rejected cross-origin requests according to the
  documented CSRF policy.
