# File Service — Contract and Remaining Work

## Goal

Create a LiteNAS File Service for authorized file upload and download. The
service MUST validate the caller's access token and MUST use the RBAC Service
for every file-system access decision. It MUST fail closed when authentication,
RBAC, or safe path resolution fails.

Implement the requirements in [requirements.md](requirements.md). Use its
source-qualified IDs (for example, `file-service-svc/FR-001`) in
requirement-traceable tests. This task follows the classification in
`requirements/requirement-types.md` and the boundaries in
`requirements/rbac-service-boundaries.md`.

## First-version status

File Service, its user-scoped writer, the `/f/`, `/d/`, and `/ls/` Nginx mounts,
Auth/RBAC NATS calls, certificate identity, non-root service unit, restricted
sudoers rules, AppArmor profiles, local build/deploy scripts, and single-package
Debian wiring are implemented. A browser download of an authorized user file
through `/f/` has been confirmed on a local deployment. This confirmation does
not close the broader acceptance and security verification below.

## Service Boundaries

- Auth Service owns access-token issuance and token lifecycle, including its
  revocation policy. The existing web-gateway authentication flow owns browser
  cookie issuance and clearing. File Service does not issue, refresh, revoke,
  or set browser tokens or cookies. For each
  protected HTTP request, File Service sends the presented access token to the
  existing `auth.rpc.token.validate` NATS contract and uses only the username
  returned when `valid` is true, `status` is authenticated, and `username` is
  nonempty. Token roles or scopes do not grant file
  access. The current Auth validation contract verifies token validity and
  expiry; it does not provide immediate access-token revocation checks.
- RBAC resolves the verified token's username through
  `rbac.rpc.subject.roles.get` and supplies the UID used for file checks.
- RBAC decides path access through `rbac.rpc.path.read.check` for downloads and
  directory listings. Upload creation and atomic replacement use
  `rbac.rpc.path.operation.check` with `create` or `replace`, and deletion uses
  that subject with `delete`. The existing `rbac.rpc.path.write.check` covers
  writing an existing file's content, not upload publication. RBAC owns host
  mode-bit, group, ACL, and symlink permission evaluation.
- The File Service owns upload/download workflow, transfer state, storage
  operations, HTTP response behavior, and enforcing each returned decision.
- The HTTP process MUST run as a dedicated non-root service account. A
  provisioned `lite-nas-fs-rd` group grants read on files and read plus
  search/traverse on directories in configured file-backed storage roots, with
  no write grant. File Service and Stream Service join this group; neither
  receives a general host write group. System-critical paths are excluded.
  RBAC still decides each caller's potential access; service group membership
  never authorizes a caller.
- For uploads, the dedicated Go application
  [lite-nas-file-writer](../../apps/file-writer/README.md) exclusively
  creates `<destination>.upload`, writes conventional and resumable bytes,
  verifies/recovers the stage, and attempts the same-directory final rename as
  the verified Unix user after the relevant RBAC checks. File Service streams
  bounded request bytes to the writer and owns private resume metadata; it
  does not open the stage for writing. Unix permissions and ACLs make the final
  filesystem decision. Invoke the fixed executable through restricted
  noninteractive `sudo -u` without a shell; a general-purpose sudo grant is
  forbidden.
- The File Service has a dedicated AppArmor profile at
  `configs/etc/apparmor.d/usr.libexec.lite-nas.file-service`, attached to
  `/usr/libexec/lite-nas/file-service` and loaded in enforce mode through the
  repository's managed AppArmor profile loader during local deployment and
  Debian installation.
- Common token verification, RBAC subject/path authorization, and safe
  file-access/path-resolution behavior MUST use shared repository
  infrastructure where those contracts and lifecycle are identical across
  File Service and Stream Service. Keep upload workflow specific to this
  service; do not force transfer-specific behavior into shared infrastructure.
- The client-facing interface uses HTTP on a service-owned loopback-only port.
  Nginx owns browser-facing HTTP/HTTPS routing and mounts file operations at
  `/f/` and explicit downloads at `/d/`. For example, both
  `/f/home/alice/photo.jpg` and `/d/home/alice/photo.jpg` identify host path
  `/home/alice/photo.jpg` after removing their mount prefix. The existing
  `lite-nas-at` browser cookie is `Secure` and therefore accompanies browser
  requests over HTTPS, not HTTP.
  File Service communicates internally with Auth and RBAC only through their
  existing shared NATS RPC contracts; it introduces no new cross-service API.
- Prefer native Linux/Unix filesystem and process facilities and the Go standard
  library for file operations, temporary storage, permissions, and transfer
  handling. Add a database, external storage system, or specialized transfer
  framework only when a documented requirement cannot be met safely with the
  existing platform facilities.
- Prefer simple, mainstream Go libraries for capabilities the standard library
  does not cover well. Keep the dependency set small, favor actively maintained
  libraries with broad Go ecosystem use, and document the specific need before
  adding a specialized or heavyweight dependency.
- A missing token, invalid/expired token, failed subject resolution, denied
  RBAC decision, RBAC timeout/error, or unsafe path MUST never result in file
  bytes being read or written.

## Requirements

### IR-001 Configure HTTP, Nginx, NATS, and service certificates

The service MUST use LiteNAS's established transport-security and certificate
configuration for every internal or externally reachable connection.

Acceptance criteria:

- The HTTP listener uses a service-owned port and binds only to loopback; it
  is not directly reachable from non-local interfaces.
- Nginx exposes the service at `/f/` and `/d/` through the existing HTTP/HTTPS
  entrypoints and reverse-proxy conventions. Browser cookie authentication on
  both mounts works through HTTPS; an HTTP request without a valid token is
  denied. Path forwarding preserves the suffix used to derive the absolute
  host path and does not collide with gateway/API routes.
- Keep File Service proxy locations in a dedicated Nginx snippet included by
  both LiteNAS server blocks; the web-gateway site contains only the include.
  Local deployment and the Debian package install the snippet before Nginx
  configuration validation or reload.
- Nginx forwards `GET` and `HEAD /d/{absolute-path...}` to the File Service's
  explicit download handler. The `/d/` suffix identifies the same host file as
  the corresponding `/f/` suffix; `/d/` does not accept upload, delete, or
  other mutation methods.
- Nginx forwards `/f/docs`, `/f/openapi.json`, `/f/upload`,
  `/f/proprietary/upload...`, `/f/proprietary/upload-window`, and generic
  `DELETE /f/{absolute-path...}` requests to the corresponding File Service
  routes after removing the `/f` prefix. These explicit API routes take
  precedence over the generic file-path handler.
  The paths `/docs`, `/openapi.json`, `/upload`, and `/proprietary/upload...`
  are reserved API routes, not generic host-path aliases.
- Nginx uses `proxy_request_buffering off` for both upload modes so it does not
  write a complete request body to its client-body temporary directory.
  Configure the upstream HTTP version and upload body limits so large
  multipart requests and chunk PATCH requests reach File Service
  incrementally.
- Swagger UI and the maintained OpenAPI document are available at `/docs` and
  `/openapi.json`, publicly mounted as `/f/docs` and `/f/openapi.json`.
  Configuration exposes `api_docs.enabled`; when false, both documentation
  routes return `404`; it defaults to `true`. The documentation routes require
  a valid access token and use `/f`-prefixed public API URLs. Serve Swagger UI
  assets locally instead of relying on a runtime CDN.
- Browser requests over HTTPS use the existing `lite-nas-at` access cookie
  managed by the web-gateway authentication flow. Nginx forwards that request
  cookie to the service. File Service validates
  the token through `auth.rpc.token.validate` and does not issue, refresh,
  revoke, set, or weaken browser session cookies. Browsers do not send the
  `Secure` cookie over HTTP.
- NATS connections use the shared messaging configuration and the service's
  dedicated client certificate identity, trusted CA, and private key. The
  service can request only the NATS subjects required by its documented Auth
  and RBAC contracts.
- Configuration provides certificate paths using existing LiteNAS conventions
  and validates required files and TLS setup at startup. Private keys are not
  logged or copied into application config values unnecessarily.
- Certificate ownership and permissions allow the non-root service account to
  read only the certificates it needs, without granting access to CA private
  keys or unrelated service identities.
- Bootstrap, certificate issuance/rotation, NATS authorization, deployment
  configuration, and package contents are updated for the new service identity
  where required.

### FR-001 Authenticate every protected transfer request

The service MUST require a valid access token for upload and download
operations, including upload creation, chunk transfer, resume/status, finalize,
and download continuation requests.

Acceptance criteria:

- Missing, malformed, invalid, expired, or otherwise rejected tokens are denied
  before file processing. An Auth RPC timeout, error, or invalid response also
  denies the request.
- The username comes from the successful Auth validation response, never from
  request parameters or unverified claims. Token-provided roles or scopes
  alone never grant file access.
- The service obtains the RBAC subject UID for the authenticated username using
  the existing subject-resolution contract.
- Token verification and RBAC failures are handled as deny/error outcomes, not
  as implicit allow.

### FR-003 Authorize downloads through RBAC

Before opening or reading a requested file, the service MUST ask RBAC whether
the verified token owner can read the resolved path.

Acceptance criteria:

- The read decision uses the token owner's UID and the actual resolved target
  path, not an untrusted client-supplied identity.
- For non-administrators, reads are limited to `/home`, configured storage roots, and
  mounted ZFS trees outside system directories. Members of `admin` or `sudo`
  may request system paths, subject to RBAC.
- A deny or inability to get a trustworthy allow prevents opening/serving file
  contents.
- Each new request to resume or continue a download is authenticated and
  authorized again.
- Path traversal and alternate path encodings cannot bypass the configured
  storage boundary or the RBAC decision.
- The first version may serve private files by invoking fixed `/bin/cat` with
  `sudo -u` as the verified Unix user after the RBAC read allow. This has the
  authorized-path-to-open race documented in
  [READ_RACE_REVIEW.md](READ_RACE_REVIEW.md); tests and a reader replacement
  are required before claiming race-resistant download access.

### FR-004 Authorize uploads through RBAC

Before creating or replacing a destination, the service MUST request the
corresponding RBAC `create` or `replace` decision for the verified token owner.
The writer checks actual staged-file write access as that user.

Acceptance criteria:

- Upload authorization uses the token owner's UID and the final intended target
  path.
- Upload initialization, each resumed/chunk write, and finalization request a
  fresh `create` or `replace` decision for the destination and deny on a
  denied or unavailable result. A changed destination kind or identity is
  rejected; earlier stage writes do not grant publication.
- Client-provided filenames and paths are validated and safely resolved before
  they can affect storage locations.
- Uploads targeting a configured protected path are rejected
  before accepting content and checked again before final publication. The
  shared protected-path policy is `filesystem.protected_paths` and applies to
  new-file creation, replacement, and deletion.
- When the destination is an existing regular file, successful upload replaces
  it atomically after full validation. The old file remains intact until the
  replacement is ready; a failed upload or publish leaves it untouched.
  Existing directories and symbolic links are rejected as upload targets and
  are never followed or replaced.
- A failed or interrupted upload cannot expose a partially written file as a
  completed file.
- The writer creates `.upload` exclusively and writes it as the authenticated
  Unix user for both upload modes. That create is the Unix permission check;
  a separate `touch` is insufficient and may change an existing file. The
  service never writes directly into the completed
  destination. A newly published file is owned by that user without a
  post-upload `chown`. Its group, mode, and ACL follow the staged file, including
  on overwrite. The previous file's permissions and ACL are not preserved or
  copied onto the replacement.
- All mutations must be within `filesystem.write_roots`, outside
  `filesystem.protected_paths`, and allowed by RBAC. The default write root is
  `/home`; explicitly configured ZFS mount paths may be added. Standard
  OS/runtime paths remain non-writable.

### FR-005 Support pausing and resuming uploads

The service MUST let a client pause an upload and continue it later without
requiring the client to resend already accepted bytes. A client disconnect or
process interruption MUST leave the upload resumable when its state is valid.

Acceptance criteria:

- The upload protocol defines how a client starts an upload, identifies it on
  continuation, explicitly pauses it, reports/query progress, submits remaining
  content, and marks completion or cancellation.
- The service detects invalid offsets, conflicting chunks, incomplete uploads,
  and attempts to continue an unknown or expired upload, and returns a defined
  failure outcome.
- Resumed operations are bound to the authenticated owner and intended target;
  another principal cannot continue or finalize an upload they do not own.
- The service documents and enforces how long incomplete upload state is kept
  and how it is cleaned up.
- An upload becomes visible as a completed target only after successful
  finalization.

#### Upload modes

The File Service MUST expose two upload modes with separate HTTP routes:

- A conventional browser upload accepts `POST /upload` with
  `multipart/form-data`, a `destination` field, and one `file` part. Nginx
  exposes it publicly as `POST /f/upload`. It supports ordinary browser/form
  clients and reports success or failure for that request. It replaces an
  existing regular-file destination only after the full upload is received,
  authorized, and validated. It uses the same destination-adjacent `.upload`
  staging and final publication flow as resumable uploads. Parse the multipart
  body incrementally and write its file part to `.upload` with bounded working
  memory; do not use a parser that spools the complete large file to `/tmp` or
  another service-local disk location. The `destination` part MUST precede the
  `file` part so authorization finishes before file bytes are accepted. Reject
  the opposite order. File content is limited to 10 GiB. This mode does not
  promise resumability.
- A LiteNAS app protocol provides resumable initialization, chunk upload, and
  finalization operations under `/proprietary/upload`; Nginx exposes these as
  `/f/proprietary/upload`. The later frontend app will implement this protocol;
  its implementation is outside this service task.
- The File Service exposes its OpenAPI document at `/openapi.json`
  and Swagger UI at `/docs`, publicly mounted as `/f/openapi.json` and
  `/f/docs`, respectively. A boolean `api_docs.enabled` setting controls both
  routes; production configuration sets it to `false`, while development and
  test configurations set it to `true`. Operators may override it. When
  disabled both routes return `404`. Both routes require a valid access token.

The resumable protocol MUST follow this lifecycle:

1. `POST /proprietary/upload` creates an upload resource using JSON metadata that
   includes the destination, original filename, total size, whole-file SHA-256,
   supported file metadata, and a client `Idempotency-Key`. The service
   validates the target and obtains
   RBAC `create` or `replace` authorization before accepting the upload. It does not reserve a
   chunk buffer; buffer capacity is reserved only while a chunk request is
   processed. Success returns `201 Created` and a `Location` for the resource.
   Retrying the same key and same metadata returns the existing upload resource;
   reusing that key with different metadata returns `409 Conflict`.
2. Creation returns an opaque upload ID, the next chunk index expected,
   effective chunk size, maximum chunk size, and upload state. The client may
   request a chunk size in the inclusive 1–16 MiB range when creating the
   upload. Clients that omit it use the service's configured size, which has
   a 1 MiB default. Initialization records the effective size for that upload, so
   resume keeps using it even if service configuration later changes. Only the
   final chunk may be shorter. The proprietary protocol has no configured
   whole-file size limit; the destination filesystem and 64-bit offsets are
   its practical bounds. The client discovers existing progress through
   the upload resource rather than repeating creation with the same metadata.
3. `PATCH /proprietary/upload/{upload-id}` sends one raw next-chunk body with the committed
   `Upload-Offset` header. Chunks are accepted in order, beginning at
   index zero. A committed chunk returns `204 No Content` with the updated
   `Upload-Offset` and `Upload-Next-Chunk-Index` response headers. A request
   whose offset is not the current committed offset returns a conflict with
   the current offset and next index. An identical replay of the last committed
   chunk is acknowledged idempotently; a conflicting replay is rejected.
4. `GET /proprietary/upload/{upload-id}` returns the authenticated owner's upload state
   and progress as JSON. `HEAD /proprietary/upload/{upload-id}` returns upload length,
   committed offset, next chunk index, and effective chunk size in headers with
   no response body. Pausing is client-controlled: stop sending chunks, then
   use GET or HEAD later to continue. `DELETE /proprietary/upload/{upload-id}` cancels and
   removes the incomplete upload.
5. `PUT /proprietary/upload/{upload-id}/completion` marks the upload complete. It succeeds
   only after the declared size and whole-file SHA-256 match the staged bytes;
   it rechecks RBAC `create` or `replace` permission and protected-path policy, then attempts
   the same-directory atomic rename as the verified Unix user. If a regular
   file already exists, successful publication replaces it atomically. A
   failed user-scoped rename leaves the previous file intact, attempts removal
   of `.upload`, and returns failure without claiming completion. Define
   cleanup-failure and completion-retry responses before implementation.
   Repeating a successful completion returns the same completed result.

Authenticated `GET /proprietary/upload-window` returns JSON describing available
chunk-buffer bytes and the maximum accepted chunk size. If a valid `PATCH`
cannot reserve enough global buffer capacity, return `503 Service Unavailable`,
`Retry-After`, and JSON with `error.code` `upload_window_full`. Accept no bytes
and do not advance the committed offset. The client waits for the indicated interval and probes
`GET /proprietary/upload-window` before retrying. Capacity may change between probe and
PATCH, so the client must handle another 503. A chunk larger than its upload's
configured chunk size is a permanent validation error, not an on-hold response.
The service reserves each request's maximum chunk-buffer capacity after
validating PATCH headers and before reading its body; Nginx streams PATCH bodies
without fully buffering them ahead of the service's admission decision.

Creation, progress, completion, and error responses use JSON. Creation reports
`upload_id`, `next_chunk_index`, and effective chunk size; GET progress reports
total size, acknowledged bytes, and next chunk index; HEAD carries equivalent
progress in headers and has no response body. Completion reports the verified
checksum and completed resource. Errors use a non-success HTTP status and JSON
with a stable `error.code`. Chunk sequencing errors include the current
committed offset and next chunk index. Retryable failures identify whether and
when the client may retry. Errors must not expose host paths the caller cannot
access. Define the complete error-code catalog and HTTP status mapping before
implementation.

Concurrency is limited by consumed chunk-buffer capacity rather than by a
fixed number of uploads. The service MUST serialize writes per upload so
concurrent requests cannot corrupt chunk order or metadata. Accepted chunks
are appended sequentially to the destination-adjacent `.upload` file; the
service MUST NOT rewrite the whole file for each chunk or append incomplete
data to the completed destination.
For each chunk, the service writes at the committed offset, makes the data
durable, then atomically records the new committed byte offset/chunk index in
metadata before acknowledging success. Metadata updates use a temporary file
and atomic replacement in the private metadata directory. On recovery, bytes
beyond the last committed offset are truncated before accepting more data. A
failed partial write is truncated to that offset and the same chunk remains
expected.
An interrupted connection leaves only fully acknowledged chunks committed;
reconnecting clients query or re-initialize to obtain the next expected index.
The service reauthenticates and rechecks RBAC for every initialization, chunk,
status, and finalize request.

The service uses a dedicated 64 MiB size-limited tmpfs mount only for bounded
in-flight upload buffers and metadata working copies, not accumulated file
content. No upload mode requires free space for the whole file in `/tmp`, on
the service host's SD card, or in tmpfs.
Reserve 4 MiB for metadata files/working state and 60 MiB for concurrent
incoming chunk buffers. The metadata reservation is a hard limit; reject new
upload initialization when its bounded metadata would exceed 4 MiB, and release
that reservation when the upload expires or is cancelled. Admit chunk requests
using a weighted capacity limit based on the bytes each request may buffer;
active buffered chunk requests MUST never reserve more than 60 MiB. The
configured chunk size defaults to 1 MiB and may
be set up to 16 MiB, so the theoretical concurrent receiving-request capacity
is 60 requests at 1 MiB or 3 requests at 16 MiB, before implementation overhead
and other limits. Paused uploads do not consume chunk-buffer capacity.

For both upload modes, accumulated bytes are stored at the destination path
plus `.upload`: for `/path/to/original/file.ext`, the incomplete file is
`/path/to/original/file.ext.upload`. This keeps publication on the same
filesystem and uses the destination pool's capacity for the upload. The
`.upload` path is reserved for one active upload to that target.
Create it exclusively as the verified Unix user after RBAC approval and write
all conventional or resumable upload bytes as that same user. Reject an
unrelated existing file or symlink rather than replacing it, and bind resumed
requests to the recorded upload owner and target. Recheck its owner, identity,
and completed content before publication; do not trust a path that another
process could have replaced. The `.upload` file may be visible to other host
filesystem clients before completion; it is never presented as the completed
destination.

Keep durable upload metadata in a private, permission-restricted service area,
associated with the upload ID and canonical target. Do not create a
`destination.ext.meta` sidecar beside the target. Durable metadata remains
authoritative across process restart and reboot; the tmpfs metadata working
copy is bounded by its 4 MiB reservation. Each durable record includes the
authenticated owner, canonical target, stage identity, declared size and
checksum, effective chunk size, committed offset and chunk index, idempotency
key, and expiry. Incomplete uploads expire 24 hours after the last committed
activity. A separate cleanup worker uses the existing shared timer to scan at
startup and every three hours, verifies stage identity, and removes expired
stages and records. An in-memory index may speed lookups but is rebuilt from
durable records after restart. After full
validation and a fresh authorization decision, attempt to atomically rename
the `.upload` file to the destination as the verified Unix user. Unix
directory and sticky-bit permissions may deny publication. On a failed rename,
attempt to remove the stage and report failure; define cleanup recovery if
removal also fails. Never remove or truncate an existing destination before its complete
replacement is ready. Handle destination-pool quota or space exhaustion
without publishing a partial file or claiming that the full declared size was
reserved in advance.

### FR-006 Support pausing and resuming downloads

The service MUST let a client pause a download and continue it later from the
last confirmed byte position or equivalent standard continuation mechanism.
Pausing MUST stop further transfer work and release request-scoped resources.

Acceptance criteria:

- The transfer protocol unambiguously reports file length and the served byte
  range/continuation position so the client can pause and resume reliably.
- Implement ordinary HTTP `HEAD` and `GET` range behavior: advertise
  `Accept-Ranges: bytes`, return a strong `ETag`, and answer one-range requests
  with `206 Partial Content`, correct `Content-Range`, and range length.
- After authentication and RBAC authorization, return the same file
  `Content-Type`, `Content-Disposition`, and `X-Content-Type-Options: nosniff`
  on `HEAD`, full `GET`, and partial `GET` responses within each mount. `HEAD`
  has no response body; `Content-Length` describes the full representation on
  `HEAD` and full `GET`, and the selected range on `206`.
- Both `/f/{absolute-path...}` and `/d/{absolute-path...}` use the same
  authentication, RBAC read check, file identity, strong ETag, and range
  semantics. Each request is independently authorized, including requests
  following a link from one mount to the other.
- Accept multiple simultaneous independent range `GET`s for one file so a
  download manager can fetch separate segments over the same REST file
  resource. Do not add a download-session protocol or require a
  `multipart/byteranges` response.
- Require the client's strong `If-Match` validator on parallel segment requests.
  Return `412 Precondition Failed` if the file was replaced or otherwise
  changed, preventing a client from combining bytes from different versions.
- Support a configurable per-principal/per-file cap of 16 simultaneous range
  responses by default, subject to global resource limits. Return a documented
  retryable response (`429 Too Many Requests` with `Retry-After`) when the
  per-principal cap is reached.
- Each request, including every parallel segment, independently validates the
  token and repeats the RBAC read authorization before reading file data.
- Invalid, unsatisfiable, or out-of-file continuation positions receive a
  defined failure response.
- The service authenticates and performs an RBAC read check before serving each
  new, paused-and-resumed, or otherwise continued request.
- Download continuation cannot change the requested file identity or bypass
  path validation.

### FR-007 Preserve arbitrary file formats

The service MUST treat file content as opaque bytes so upload and download work
for supported storage objects regardless of media format or filename extension.

Acceptance criteria:

- Upload and download do not transcode, rewrite, or reject content based only on
  its media format or extension.
- The service preserves the original bytes exactly across a completed upload
  and subsequent download.
- On `/f/`, PDF, JPEG, PNG, GIF, and WebP files use the matching media type and
  `Content-Disposition: inline` so browsers can offer their built-in viewers.
  Supported passive audio and video formats (`mp4`, `m4v`, `mov`, `avi`,
  `mkv`, `webm`, `ogv`, `mp3`, `wav`, `m4a`, `aac`, `flac`, and `ogg`) also
  use their matching media type
  and inline disposition so browser media controls can request byte ranges.
  Active document formats such as HTML and SVG remain attachments.
- On `/f/`, archives and content outside that inline allowlist use
  `Content-Disposition: attachment` directly, without redirecting to `/d/`.
  Use an accurate archive media type when reliably identified; otherwise use
  `application/octet-stream`. Unknown or potentially active content, including
  HTML and SVG, uses `application/octet-stream` and attachment.
- On `/d/`, every file uses `Content-Type: application/octet-stream` and
  `Content-Disposition: attachment`, including files that `/f/` could display
  inline. The payload bytes are identical on both mounts.
- Include a safe display filename in `Content-Disposition`, derived from the
  authorized file's basename. Encode non-ASCII names according to HTTP header
  rules and remove path components and header-control characters. Do not use
  client-provided names or media types as trusted response metadata.
- These headers describe the authorized, opened file and are not exposed on
  denied requests. Media type and filename never grant access. `inline` and
  `attachment` communicate the intended browser behavior; the browser makes
  the final display decision.

### FR-008 List authorized files and directories

The service MUST provide public `GET /ls/{absolute-path...}` that returns only
entries the authenticated token owner can access according to RBAC. For
example, `/ls/abs/path` lists host directory `/abs/path`. The listing route is
separate from the `/f/` file-operation mount.

Acceptance criteria:

- `recursive` and `files-only` are presence-based query flags: a present flag
  enables the behavior and an absent flag disables it. For example,
  `/ls/abs/path?recursive&files-only` enables both.
- A directory is listable only when RBAC allows both read and execute/search
  access to that directory for the token owner's UID, with execute/search
  access on every ancestor from host root `/`.
- Non-administrators see `/home`, configured storage roots, and mounted ZFS trees, with only the
  ancestor directories needed to reach them. System directories such as
  `/etc`, `/var`, `/sys`, and `/proc` are hidden from their root listing and
  denied as direct listing targets. Admin and sudo groups retain RBAC-filtered
  system listings.
- Recursive traversal descends only into directories for which RBAC allows
  read and execute/search access; traversal never passes through an
  inaccessible directory.
- A regular file is included only when RBAC allows read access to that file and
  execute/search access to every ancestor. When `files-only` is present,
  directory entries are omitted while authorized regular files in the selected
  traversal scope remain eligible.
- Symlink behavior is defined and RBAC checks the resolved absolute host path.
- Denied entries and subtrees are omitted without exposing their names or
  metadata. RBAC dependency/evaluation errors fail the listing closed rather
  than returning a partial result as complete.
- Every listing request validates the token and resolves the token owner's UID.
- A successful response is a JSON array of objects with absolute `path`, child
  `name`, `kind` (`file` or `directory`), and nullable `size` (bytes), `owner`,
  `group`, `permissions` (nine symbolic mode characters), and `last_modified`
  (UTC RFC 3339). `files-only` omits directory objects.
- A listing failure returns a non-success HTTP status and JSON with a top-level
  `error` key containing a documented stable error code. Error responses do not
  disclose inaccessible paths or entry names.

#### Notes

The current RBAC RPC contracts provide separate read and execute path checks.
Use execute permission as directory search/traversal permission and read
permission for directory contents and file content. Do not treat the File
Service process's own access as caller authorization.

### FR-009 Delete authorized files and directories outside protected paths

The service MUST expose `DELETE /f/{absolute-path...}` for deleting an
authorized regular file or directory tree. Nginx strips the `/f` prefix before
proxying, as for other file operations.

Acceptance criteria:

- Every delete request validates the token, resolves the token owner's UID,
  checks `filesystem.write_roots` and `filesystem.protected_paths`, and obtains
  a fresh RBAC delete allow for the resolved target before unlinking it.
- If RBAC does not already define delete authorization, define and review that
  contract with the RBAC service before implementing deletion. Do not infer
  delete permission from upload/write permission.
- Configuration key `filesystem.protected_paths` is a list of
  canonical absolute paths for protected roots. The packaged defaults include
  `/etc` and platform-critical paths such as `/usr`, `/bin`, `/sbin`, `/lib`,
  `/lib64`, `/boot`, `/proc`, `/sys`, `/dev`, `/run`, and `/var/lib` where
  present.
- A configured path and all descendants are protected using
  component-boundary matching after safe normalization/resolution. Similar
  sibling prefixes such as `/etc-backup` are not descendants of `/etc`.
- Protected-path denial takes precedence over RBAC allow. Directory deletion
  checks each descendant and removes children before parents. Symbolic links
  and special files are rejected and never followed.
- Authorize the target and required parent-directory operation according to
  the agreed RBAC contract, then delete through fixed user-scoped commands.
- Regular-file success returns `204 No Content`. Directory requests return
  `202 Accepted` with an in-memory job ID; poll `GET /f/jobs/{id}` for progress.
  Missing, unauthorized, protected, invalid,
  and unsupported targets return documented non-success JSON errors with stable
  codes and without sensitive host-path disclosure.
- Cookie-authenticated DELETE requests pass the service's CSRF checks.

### FR-010 Create directories and queue recursive transfers

`POST /f/mkdir` accepts `{ "path": "/absolute/new-directory" }` and returns
`201 Created`. `POST /f/copy` and `POST /f/move` accept complete paths in
`{ "source": "/absolute/source", "destination": "/absolute/target" }`.
Copy and move return `202 Accepted` with `job_id`, `state`, `percent`,
`completed_bytes`, `total_bytes`, `completed_items`, and `total_items`.
`GET /f/jobs/{id}` returns the current state to the authenticated owner.
The service queues operations sequentially in memory using a Go worker and
progress channel. Jobs survive browser reload while the service stays running;
they do not survive a service restart. Existing destination directories are
merged and existing regular files are replaced. A directory cannot be copied
or moved into itself. All source tree entries and destination paths are checked
against RBAC and storage policy. Copy uses the user-scoped staged writer and
reports progress after each chunk. Move copies and then removes the source tree.

### SR-001 Fail closed and protect file data

The service MUST deny access on authentication/RBAC/path-resolution failures
and MUST avoid disclosing protected file content or sensitive host paths in
errors and logs.

Acceptance criteria:

- RBAC timeout, transport error, malformed response, or unavailable service
  never permits an operation.
- Transfer identifiers cannot be used as bearer credentials and are bound to
  authenticated owner and target metadata.
- Cookie-authenticated upload mutations enforce a documented CSRF defense at
  the service boundary; `Secure`, `HttpOnly`, and `SameSite` cookie attributes
  are defense in depth, not the entire CSRF policy.
- Temporary upload data and transfer state are protected from unauthorized
  reads and writes and are cleaned up according to the defined lifecycle.
- The HTTP service runs with a nonzero effective UID and is not granted root,
  broad `CAP_CHOWN`, or unrestricted host-file access. User-scoped file
  operations require a narrowly scoped execution policy and repeat token and
  RBAC checks for every protected request. The target Unix account must match
  the verified username and resolved RBAC UID; a client cannot select it.
- Keep virtual kernel interfaces outside normal file operations unless
  explicitly supported. The service's private upload state remains accessible
  only to the service, while staged content lives in the target directory.
- Filesystem operations use safe descriptor-relative or equivalent
  race-resistant path handling supported by the target Linux platform, except
  for the accepted first-version raw `cat` read race documented in
  [READ_RACE_REVIEW.md](READ_RACE_REVIEW.md). Completed uploads are published
  atomically where the filesystem permits.

### RR-001 Handle interrupted transfers and dependency failures

The service MUST leave filesystem state consistent when a client disconnects,
the process stops, or a required dependency fails during a transfer.

Acceptance criteria:

- Interrupted uploads remain resumable only when their recorded state is
  complete and valid; otherwise they are rejected or safely removed.
- Client disconnects stop unnecessary I/O and do not mark a partial transfer as
  complete.
- Authentication/RBAC outages result in an explicit failure and do not
  continue protected I/O.

### TR-001 Keep transfer and authorization behavior testable

Authentication, RBAC, storage, and transfer-state dependencies MUST be
replaceable in tests so allow, deny, error, and interruption behavior can be
verified without a live host account database or production storage.

### TR-002 Verify transport and certificate configuration

Transport security, service certificate configuration, and certificate access
MUST be covered by automated or repeatable integration checks.

Acceptance criteria:

- Configuration tests cover valid TLS/NATS settings, missing or invalid
  certificate paths, and startup failure when required transport security
  cannot be configured.
- An integration check verifies that the service connects to NATS using its
  dedicated client identity and that the configured CA is trusted; missing,
  untrusted, or mismatched credentials fail closed.
- Tests or operational checks verify that NATS authorization grants only the
  service's required Auth/RBAC subjects.
- The selected HTTP exposure/TLS boundary is exercised by an integration or
  system check, including a negative case for an unprotected network path when
  the service is network-reachable directly.
- Certificate bootstrap/rotation checks verify the service identity is
  provisioned and that the non-root service can read only its required
  certificate material.

### OR-001 Integrate local build, deployment, and test workflows

The service MUST be usable through LiteNAS's local build/deployment scripts and
local test workflows.

Acceptance criteria:

- The service has dedicated build, deploy, and test scripts following existing
  repository conventions.
- Aggregate build and deployment scripts include the service and its required
  configuration, certificate identity, runtime account, and dependencies.
- The service test script runs its focused unit/integration tests and is wired
  into the repository's aggregate local test and CI test workflows.
- The service can be built, deployed, exercised, and removed/updated using the
  established local workflows without manually assembling package contents.

### OR-002 Include the service in Debian build and publishing

The service MUST ship as part of LiteNAS's existing single Debian package and
release workflow.

Acceptance criteria:

- Package build scripts include the service binary, configuration, systemd
  unit, AppArmor policy, logrotate drop-in, and other required runtime files.
- Package content analysis and Debian content validation assert that required
  service artifacts are present and have expected permissions.
- Pull-request package build/install validation covers the service being
  installed and activated successfully.
- Release workflows build the service for each supported architecture, include
  it in the Debian package, and publish the resulting package through the
  existing release process.
- This work does not create a separate per-service Debian package.

## Remaining work

The first-version implementation actions are complete and have been removed
from this task list. Keep the detailed contract and acceptance criteria above
as the source of truth while closing these verification items:

1. Run the full installed-system API table in `tests/api/test_file_service.py`
   through Nginx. Confirm anonymous denial, operator access to a private owned
   file and directory, denial of critical paths, and the configured admin
   account's intended access. Resolve any mismatch between admin expectations
   and the current Unix-UID-based RBAC/read model before calling these cases
   complete; missing admin credentials must not silently pass the tests.
2. Add requirement-traceable service and installed-system tests for both upload
   modes, overwrite and failure cleanup, 24-hour expiry and three-hour cleanup,
   concurrent destination protection, delete, listing flags, byte ranges and
   ETags, documentation routes, and dependency-failure responses. Exercise
   writer confinement and ownership with a deployed non-root service.
3. Verify a freshly built Debian package on both supported architectures:
   content validation, package installation, service activation, Nginx routing,
   NATS certificate issuance/rotation, and AppArmor confinement. CI now builds
   File Service and the file writer as separate artifacts and passes both to
   the package job; the installed-system test gate still needs a full green run.
4. Complete the reader race and private-directory review in
   [READ_RACE_REVIEW.md](READ_RACE_REVIEW.md). Until then, do not describe the
   raw `sudo -u /bin/cat` path as race resistant or claim support for a private
   `0700` parent directory whose ACL prevents File Service path resolution.

## Completion Criteria

- All requirements above have implementation and requirement-traceable tests.
- No file read/write path bypasses authentication or RBAC.
- Uploads and downloads can be explicitly paused and resumed under documented
  continuation contracts.
- The service is integrated with the repository's normal build and deployment
  flows.
- The installed service runs non-root and can complete authorized file
  operations using its documented least-privilege filesystem access.
- Its transport configuration, NATS certificate identity, authorization rules,
  and certificate rotation/bootstrap are integrated with LiteNAS operations.
- Local build/deploy/test workflows and the single Debian build, validation,
  and publishing workflows include the service.

## Additional work — per-user LiteNAS profile files

The Admin Panel needs a user-owned `.lite-nas` directory in the user's home
namespace for an avatar and preferred UI settings. Implement this through File
Service so clients do not access host paths directly.

1. Define and document the stable per-user profile directory and file naming
   conventions with Admin Panel before implementation. The authenticated
   identity, not a client-supplied username or arbitrary absolute path, selects
   the user's profile location.
2. Inspect existing listing, file read, conventional upload, user-scoped
   writer, and RBAC contracts. Reuse existing operations when they support the
   profile workflow safely.
3. If the directory needs to be created on demand, define the smallest
   validated File Service operation and corresponding user-scoped writer/RBAC
   behavior that can create `.lite-nas` as the authenticated Unix user. Do not
   create it as the File Service account or grant a general-purpose mkdir
   capability. Update requirements and shared contracts before implementation;
   do not guess a public DTO or route.
4. Ensure profile file reads, avatar uploads/replacements, and settings writes
   remain confined to the authenticated user's `.lite-nas` directory, use the
   existing authentication and RBAC checks, and fail closed on policy or
   filesystem errors.
5. Document how the Admin Panel obtains and updates the avatar and preferred
   UI settings through File Service. Keep the settings document schema owned by
   the frontend contract and validate it before writes.
6. Add focused tests for cross-user denial, path traversal, directory
   provisioning, upload/replacement ownership, read authorization, and
   dependency failures. Keep service-local tests in the File Service module.
