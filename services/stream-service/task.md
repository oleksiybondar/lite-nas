# Stream Service — Implementation Task

## Goal

Create a LiteNAS Stream Service that delivers media files in their original
formats by default and supports efficient seeking and continuation. Where
codec processing is needed for playback compatibility, prefer codecs supplied
and maintained by the host OS over codec implementations bundled with Go. The
service MUST validate the caller's access token and MUST rely on RBAC for file
read authorization. It MUST fail closed if the caller is unauthenticated,
lacks access, or authorization cannot be verified.

Implement the requirements in [requirements.md](requirements.md). Use its
source-qualified IDs (for example, `stream-service-svc/FR-001`) in
requirement-traceable tests. This task follows the classification in
`requirements/requirement-types.md` and the boundaries in
`requirements/rbac-service-boundaries.md`.

## Service Boundaries

- Auth Service owns access-token issuance and token lifecycle, including its
  revocation policy. The existing web-gateway authentication flow owns browser
  cookie issuance and clearing. Stream Service does not issue, refresh,
  revoke, or set browser tokens or cookies.
  For each protected HTTP request, Stream Service sends the presented access
  token to the existing `auth.rpc.token.validate` NATS contract and uses only
  the username returned when `valid` is true, `status` is authenticated, and
  `username` is nonempty. Token roles or scopes do
  not grant media access. The current Auth validation contract verifies token
  validity and expiry; it does not provide immediate access-token revocation
  checks.
- RBAC resolves the verified token's username through
  `rbac.rpc.subject.roles.get`, returning the UID for its path decision.
- RBAC decides file read access through `rbac.rpc.path.read.check`. RBAC owns
  host path permissions, groups, ACLs, and symlink evaluation.
- The Stream Service owns video lookup, HTTP media delivery, byte-range
  handling, response metadata, and stream lifecycle.
- The process MUST run as a dedicated non-root service account. It joins the
  provisioned `lite-nas-fs-rd` group for file read and directory read plus
  search/traverse within configured file-backed storage roots, without write
  permission. RBAC returns each caller's decision; group access never replaces
  that decision. Stream Service receives no host write grant. System-critical
  paths are excluded from group access.
- The Stream Service has a dedicated AppArmor profile at
  `configs/etc/apparmor.d/usr.libexec.lite-nas.stream-service`, attached to
  `/usr/libexec/lite-nas/stream-service` and loaded in enforce mode through the
  repository's managed AppArmor profile loader during local deployment and
  Debian installation.
- Common token verification, RBAC subject/path authorization, and safe
  file-access/path-resolution behavior MUST use shared repository
  infrastructure where those contracts and lifecycle are identical across
  File Service and Stream Service. Keep media streaming behavior local to this
  service.
- The client-facing interface uses HTTP on a service-owned loopback-only port.
  Nginx owns browser-facing HTTP/HTTPS routing and mounts this service at `/s/`.
  For example, `/s/media/movie.avi` maps to host path `/media/movie.avi` after
  removing the `/s` mount prefix. The existing `lite-nas-at` browser cookie is
  `Secure` and therefore accompanies browser requests over HTTPS, not HTTP.
  Stream Service communicates internally with Auth and RBAC only through their
  existing shared NATS RPC contracts; it introduces no new cross-service API.
- Prefer native Linux/Unix file and process facilities and the Go standard
  library for opening files, streaming bytes, permissions, cancellation, and
  resource cleanup. Preserve and serve source bytes by default. Where a
  documented playback requirement needs codec processing, prefer compatible
  OS-packaged codec libraries/tools over Go-bundled codec implementations.
  Declare these runtime package dependencies so compatible codec fixes and
  security updates can be delivered through OS package updates without a new
  LiteNAS application release. A breaking codec integration or playback
  contract change still requires an application release.
- Prefer simple, mainstream Go libraries for capabilities the standard library
  does not cover well. Keep the dependency set small, favor actively maintained
  libraries with broad Go ecosystem use, and document the specific need before
  adding a specialized or heavyweight dependency.
- No video bytes or metadata that reveals protected content may be returned
  unless authentication, safe path resolution, and an RBAC allow decision
  succeed.

## Moments integration

Moments uses Stream Service for authorized media delivery, including HTTP
range requests and Live Photo components. Keep album, tag, SQLite, and collage
metadata ownership in the Moments Go service; do not duplicate those features
here. Make the existing `/s/` media route reachable from both the admin-panel
origin and the configured standalone `moments.<host>` origin so playback works
in both embedding modes. Both routes must preserve the existing authorization,
cookie, and Nginx path-forwarding requirements below.

## Requirements

### IR-001 Configure HTTP, Nginx, NATS, and service certificates

The service MUST use LiteNAS's established transport-security and certificate
configuration for every internal or externally reachable connection.

Acceptance criteria:

- The HTTP listener uses a service-owned port and binds only to loopback; it
  is not directly reachable from non-local interfaces.
- Nginx exposes the service at `/s/` through the existing HTTP/HTTPS
  entrypoints and reverse-proxy conventions. Browser cookie authentication on
  this route works through HTTPS; an HTTP request without a valid token is
  denied. Path forwarding preserves the suffix used to derive the absolute
  host path and does not collide with gateway/API routes.
- Browser requests over HTTPS use the existing `lite-nas-at` access cookie
  managed by the web-gateway authentication flow. Nginx forwards that request
  cookie to the service. Stream Service validates
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

### FR-001 Authenticate playback and continuation requests

Every request that starts playback or seeks/continues a stream MUST require a
valid access token.

Acceptance criteria:

- Missing, malformed, invalid, expired, or otherwise rejected tokens are denied
  before file access. An Auth RPC timeout, error, or invalid response also
  denies the request.
- The authenticated username comes from the successful Auth validation
  response, not request parameters or unverified claims.
- The service resolves that username to a UID through the RBAC subject-role
  contract. Token roles and scopes do not replace the file read check.
- Validation uses `auth.rpc.token.validate` on every playback and continuation
  request. The current Auth handler verifies token validity and expiry; this
  RPC does not add immediate access-token revocation checks.

### FR-002 Authorize each stream request through RBAC

Before opening or reading the video, the service MUST request a read decision
from RBAC for the resolved file path and verified token owner's UID.

Acceptance criteria:

- A denied decision, timeout, RPC error, invalid response, or unresolved subject
  prevents opening/serving the video.
- Every new playback, seek, byte-range, or reconnect request is authenticated
  and authorized again.
- File identifiers and paths are safely resolved; traversal, encoding tricks,
  and symlink behavior cannot bypass the storage boundary or RBAC.
- Authorization checks do not treat a previously issued stream URL or range
  position as continuing authorization.

### FR-003 Stream video with pause, resume, and seeking

The service MUST deliver video progressively in a form supported by standard
HTTP media playback clients and support pausing, resuming, and seeking without
requiring the entire file to be downloaded first.

Acceptance criteria:

- The service defines and implements byte-range request and response behavior
  compatible with browser playback and seeking, including correct range
  boundaries and content-length metadata where applicable.
- Invalid or unsatisfiable ranges receive a defined response and never cause an
  out-of-bounds read.
- Repeated range requests for the same file independently repeat
  authentication, path validation, and RBAC read authorization.
- A client can pause playback by stopping its current request and resume later
  from the requested byte range; the service stops work and releases resources
  for the stopped request.
- The service streams from storage without buffering the full video in memory.
- Client disconnects stop unnecessary reads and release file/stream resources.

### FR-004 Deliver original media formats and use system codecs when needed

The service MUST deliver media files without rejecting them solely because
their container, codec, image format, or extension is unfamiliar. It serves
original bytes by default. If a documented playback requirement needs codec
processing or transcoding, it MUST prefer suitable codecs installed and
maintained by the host OS over codec implementations bundled in Go.

Acceptance criteria:

- The service can deliver arbitrary file bytes as an authorized file response;
  format support does not bypass authentication or RBAC checks.
- The service preserves stored source files. Original-byte delivery is the
  default; any transcoded representation is explicit and does not overwrite or
  silently replace the source.
- HEIC/HEIF images, HEVC/H.265 video, and QuickTime MOV video can be served in
  original form. Where conversion is needed for client playback, supported
  codecs come from declared OS package dependencies where practical.
- Compatible codec security and bug-fix updates are delivered through OS
  package updates without requiring an application release. An incompatible
  codec integration or playback-contract change requires an application
  release.
- Codec processing runs as the service's non-root account, starts only after
  authentication and RBAC authorization, and has bounded CPU, memory, and
  temporary-storage use.
- Byte ranges over original files remain the baseline pause/resume and seek
  behavior. Any transcoded mode defines time-based seeking and continuation
  rather than applying original-file byte offsets to transformed output.
- Successful original-byte delivery does not guarantee that every client can
  decode every format; client compatibility and any supported conversion modes
  are documented.
- Content type is determined by a defined safe policy and is not trusted as an
  authorization input.

### FR-005 Preserve and deliver Apple Live Photos

The service MUST support delivery of Apple Live Photos as a related still-image
and motion-video pair, when the source storage exposes both components and their
association.

Acceptance criteria:

- HEIC/HEIF and JPEG still-image components and MOV/QuickTime video components
  can be delivered in original form.
- The service preserves the still-to-video association using the available
  Live Photo content identifier or equivalent source metadata; it does not
  assume that an arbitrary still image and video with similar filenames are a
  Live Photo pair.
- The public contract exposes enough information for a compatible client to
  retrieve both components as one logical Live Photo asset.
- If pair metadata is missing or malformed, each component may still be served
  individually, but the service does not claim that the pair is a Live Photo.

### FR-006 Return safe media metadata

The service MUST return only the metadata required for media playback and MUST
not expose host filesystem paths or metadata for unauthorized files.

Acceptance criteria:

- Media type and file size are derived from trusted file metadata or a defined
  safe policy, not blindly copied from untrusted request values.
- Missing or unauthorized resources do not disclose whether a protected path
  exists beyond the service's documented response policy.
- Response headers and content disposition follow the documented playback
  contract and do not permit header injection.

### SR-001 Protect playback access

Authentication and RBAC failures MUST fail closed, and stream access MUST NOT
become an alternate route around File Service/RBAC policy.

Acceptance criteria:

- The Stream Service uses the same RBAC decision model as File Service and does
  not implement host permission checks itself.
- Stream identifiers, URLs, or continuation values are not authorization
  credentials.
- Errors and logs do not expose bearer tokens, protected video bytes, or
  sensitive host paths.
- The service runs with a nonzero effective UID and does not depend on root or
  unrestricted host filesystem access.
- Its host-level read access is documented and least-privilege; a file being
  readable by the service process never replaces the required per-request RBAC
  allow decision.
- File opening and path handling use safe descriptor-relative or equivalent
  race-resistant facilities supported by the target Linux platform.

### RR-001 Handle stream and dependency failures safely

The service MUST terminate or reject the stream when it cannot establish
authentication, authorization, or a valid file handle before content delivery.

Acceptance criteria:

- RBAC and token-validation dependency failures never cause content delivery.
- A client disconnect closes underlying file and response resources.
- Stream errors do not leave persistent authorization grants or unbounded
  resource usage.

### TR-001 Keep stream behavior testable

Token validation, RBAC, file access, and stream response behavior MUST be
replaceable or controllable in service-local tests without production files or
live host accounts.

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

## Implementation Actions

Complete these actions in order. Do not select a specialized streaming stack
unless a demonstrated requirement needs it; ordinary HTTP range delivery is
the baseline playback mechanism to evaluate first.

### Decisions required before playback implementation

1. Define the configured file-backed roots for `lite-nas-fs-rd`, the excluded
   system-critical paths, and how read/search ACLs are installed on existing
   content and newly created ZFS mounts. RBAC must evaluate ancestor traversal
   and the actual file before bytes are served; the service group is not a
   caller permission decision.
2. Define the original-media `GET`/`HEAD`, range, content-type,
   `Content-Disposition`, and cache behavior. A client-supplied extension or
   MIME type cannot authorize inline media. Keep original bytes as the
   initial delivery mode.
3. Specify how Live Photo association metadata is obtained and which public
   response lets a client discover both authorized components. Do not claim
   pairing from similar filenames or expose the partner before its own read
   authorization succeeds.
4. Decide whether any transcoding mode is part of the initial release. If it
   is, define its output formats, resource limits, and time-based seeking
   contract before implementation; original-byte ranges cannot be reused as
   transformed-media offsets.
5. Use the RBAC `rbac.rpc.path.read.check` contract, which checks every
   supplied and resolved ancestor directory as well as the existing file.
   Bind the allow to the opened inode with race-resistant file operations;
   a path-only decision does not prevent a later target change.

6. Inspect Auth token validation, service-to-service NATS access, RBAC contract
   registration, shared TLS/messaging configuration, certificate issuance and
   rotation scripts, NATS authorization, HTTP server conventions, configuration,
   logs, aggregate local build/deploy/test scripts, Debian packaging, package
   validation, and release workflows.
7. Identify and coordinate reusable infrastructure with File Service,
   including token validation, RBAC subject/path checks, safe path resolution,
   and file access. Reuse existing shared packages; if common behavior is not
   already shared, define and implement the smallest shared repository package
   needed by both services. Keep ownership and API explicit and keep media
   streaming behavior local.
8. Define the public HTTP playback contract and Nginx mount at `/s/`:
   service-owned loopback port, path forwarding behavior, file lookup identity,
   request method
   and path, accepted auth transport, byte-range behavior for pause/resume and
   seeking, original-format delivery, content-type and cache policy, and error
   responses. Document these before implementation.
9. Validate the presented access token on every protected HTTP request through
   the existing `auth.rpc.token.validate` NATS contract. Use its verified
   username for RBAC subject resolution and deny on any invalid, absent, or
   failed Auth response. Document that the current Auth handler verifies token
   validity and expiry but does not check an immediate access-token revocation
   list. RBAC does not validate JWTs.
10. Implement mapping of the path suffix after `/s/` to the corresponding
   absolute host path. Define safe normalization and symlink handling so RBAC
   checks the actual resolved target before it is opened. Define how stored
   Live Photo still and video components and their content-identifier
   association are discovered and
   represented to clients; do not infer pairing from filenames alone.
11. Define and implement the non-root runtime account and `lite-nas-fs-rd`
   read/search group provisioning in local deployment and Debian
   install/upgrade flows. Do not grant host write access. Write and enforce
   the Stream Service AppArmor profile, and verify access without running the
   service as root.
12. Implement source playback using native file handles, bounded reads, and
   standard HTTP range support. Avoid full-file buffering.
13. If a documented playback requirement needs codec processing, evaluate the
   maintained codec packages available on supported Linux distributions and
   use those through a narrow system integration. Do not bundle a Go codec
   implementation when a suitable OS-maintained codec is available. Declare
   runtime packages and ensure compatible package updates do not require an
   application rebuild; document any incompatible ABI or contract changes.
   For other capabilities not handled well by the standard library, select a
   simple, mainstream Go library, record why it is needed, and keep
   dependencies minimal.
14. Add only required service configuration, contracts, and runtime wiring,
   following existing service patterns. Validate request inputs at the
   transport boundary.
15. Implement subject resolution and RBAC read authorization for every playback
   or range request. Reject all failures closed.
16. Implement progressive video reads and standards-compatible range responses,
   with bounded memory, cancellation handling, and correct file cleanup.
17. Add focused service-local tests for missing/invalid tokens, RBAC allow/deny
   and errors, path safety, valid/invalid ranges, seeks, disconnects, metadata,
   original-format delivery, Live Photo pairing, and bounded streaming
   behavior, HTTP listener/proxy path handling, and TLS/certificate
   configuration and failure cases. Add system-level coverage for deployed
   certificate permissions and the Nginx mount where applicable. Keep top-level
   `tests/` for system-level HyperionTF tests only.
18. Add/update dedicated and aggregate local build/deploy/test scripts, service
    configuration, documentation, runtime integration, NATS authorization,
    certificate bootstrap/rotation, and the shared LiteNAS logrotate
    configuration for `/var/log/lite-nas/stream-service.log`. Include the
    AppArmor profile and account/group setup in both deployment and package
    installation flows; load the profile in enforce mode and validate the
    installed groups, memberships, permissions, and profile state.
19. Integrate the service binary and runtime artifacts into the single Debian
    package, including declared OS codec runtime dependencies if codec
    processing is enabled. Update package analysis/content validation, install
    validation, and architecture-specific PR/release workflows.
20. Add the service's loopback-only HTTP listener and Nginx reverse-proxy mount
   at `/s/`. Verify proxy path forwarding, TLS
   termination, and that direct non-loopback access to the service port is
    unavailable. Verify the existing secure `lite-nas-at` cookie reaches the
    service on the HTTPS mount and is not sent by browsers over HTTP.
21. Run relevant service checks and repository CI checks; report executed
    commands and any unavailable checks.

## Completion Criteria

- All requirements above have implementation and requirement-traceable tests.
- Every request that can produce video content passes token validation and a
  current RBAC read decision.
- Original media files can be delivered without format-based rejection, and
  Live Photo pairs preserve their association for compatible clients. Any
  documented conversion mode uses declared OS-maintained codecs where
  practical, runs non-root with bounded resources, and leaves source files
  unchanged.
- Playback supports pause/resume and seeking through documented range semantics
  without full-file buffering; client codec limitations and any transcoded
  playback behavior are documented.
- The service is integrated with the repository's normal build and deployment
  flows.
- The installed service runs non-root and can stream authorized files through
  its documented least-privilege filesystem access.
- Its transport configuration, NATS certificate identity, authorization rules,
  and certificate rotation/bootstrap are integrated with LiteNAS operations.
- Local build/deploy/test workflows and the single Debian build, validation,
  and publishing workflows include the service.
