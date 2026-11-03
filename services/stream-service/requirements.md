# Stream Service — Requirements

This document defines software requirements for the LiteNAS Stream Service.
Requirement IDs use the `stream-service-svc/<TYPE>-<NNN>` format so tests and
implementation changes can trace back to their source.

## Overview

The Stream Service delivers authorized media files in their original formats
and supports pause, resume, and seeking. Where codec processing is needed for
playback compatibility, it uses codecs installed and maintained by the host OS
where practical, rather than codec implementations bundled in Go. It validates
the caller's LiteNAS access token and delegates every file permission decision
to RBAC. It runs as a non-root service and uses shared LiteNAS infrastructure
and native Linux/Unix facilities where appropriate.

## Functional Requirements

### stream-service-svc/FR-001 Authenticate playback requests

#### stream-service-svc/FR-001 Description

The service MUST require a valid access token for every request that can return
media bytes or protected media metadata, including playback, seek, resume, and
range requests.

#### stream-service-svc/FR-001 Input

- Access token presented using the documented service transport
- Requested media identity and optional byte range

#### stream-service-svc/FR-001 Output

- Authenticated principal context or an authentication failure

#### stream-service-svc/FR-001 Acceptance Criteria

- Missing, malformed, invalid, expired, or rejected tokens are denied before
  media file access.
- Browser requests can authenticate with the existing `lite-nas-at` access
  cookie, configured `Secure`, `HttpOnly`, `SameSite=Lax`, and `Path=/`; token
  validation remains the service's responsibility.
- Nginx forwards the access cookie over the HTTPS `/s/` route. The Stream
  Service does not issue a replacement browser session cookie.
- Principal username comes from verified token claims, not client input.
- Token roles/scopes or an issued media URL do not independently grant access.
- The service follows the existing Auth/shared token-validation contracts and
  documents behavior for token revocation/current status.

### stream-service-svc/FR-002 Resolve the RBAC subject

#### stream-service-svc/FR-002 Description

For an authenticated principal, the service MUST resolve the principal through
RBAC and use the returned UID for file capability checks.

#### stream-service-svc/FR-002 Input

- Username from verified token claims
- RBAC subject-resolution request

#### stream-service-svc/FR-002 Output

- RBAC subject UID or a failure outcome

#### stream-service-svc/FR-002 Acceptance Criteria

- Subject resolution uses the existing `rbac.rpc.subject.roles.get` contract.
- Failed or invalid subject resolution denies the operation.
- The service never accepts a client-supplied UID as the authorization subject.

### stream-service-svc/FR-003 Authorize every media read through RBAC

#### stream-service-svc/FR-003 Description

Before opening or reading media content, the service MUST ask RBAC whether the
token owner may read the actual resolved path.

#### stream-service-svc/FR-003 Input

- RBAC UID for the token owner
- Safely resolved media path

#### stream-service-svc/FR-003 Output

- RBAC allow/deny decision

#### stream-service-svc/FR-003 Acceptance Criteria

- The read check uses `rbac.rpc.path.read.check` and the actual target path.
- Denial, timeout, transport error, malformed response, or unresolved subject
  prevents file access.
- Every playback, resume, seek, reconnect, and byte-range request repeats
  authentication, safe path resolution, and RBAC authorization.
- Path traversal and encoding tricks cannot escape the host root `/` or the
  service's filesystem mount namespace.

### stream-service-svc/FR-004 Support pause, resume, and seeking

#### stream-service-svc/FR-004 Description

The service MUST progressively deliver media over HTTP and support client pause,
resume, and seeking without requiring the entire file to be buffered first.

#### stream-service-svc/FR-004 Input

- Authorized media identity
- Optional standard HTTP byte-range request

#### stream-service-svc/FR-004 Output

- Progressive media response with correct range metadata or a defined failure

#### stream-service-svc/FR-004 Acceptance Criteria

- The public contract defines byte-range request and response behavior,
  including valid bounds and relevant content length/range metadata.
- Invalid or unsatisfiable ranges receive a defined response and never cause an
  out-of-bounds read.
- A client can pause by stopping a request and resume from the desired byte
  range; stopping the request releases request-scoped resources.
- Streaming uses bounded memory and stops unnecessary reads on disconnect.

### stream-service-svc/FR-005 Deliver original media formats and use system codecs

#### stream-service-svc/FR-005 Description

The service MUST deliver media without rejecting a file only because its
container, codec, image format, or extension is unfamiliar. Original-byte
delivery is the default. When a documented playback requirement calls for
codec processing or transcoding, the service MUST prefer compatible codecs
provided by the host OS over codec implementations bundled into the Go binary.

#### stream-service-svc/FR-005 Input

- Authorized media file of any stored format

#### stream-service-svc/FR-005 Output

- Original media bytes and safe playback metadata

#### stream-service-svc/FR-005 Acceptance Criteria

- The service preserves stored source files. It serves original bytes by
  default; any negotiated/transcoded playback representation is explicit and
  never overwrites or silently replaces the source.
- HEIC/HEIF images, HEVC/H.265 video, and QuickTime MOV video can be served in
  their original representation.
- If codec processing is needed to make a format playable, the implementation
  uses OS-maintained codec packages where practical and documents supported
  input/output formats.
- Codec libraries or tools are installed and updated through the supported OS
  package manager where practical; Go releases do not bundle codec updates.
  Application releases are required when codec integration or the playback
  contract changes incompatibly, not for compatible security/bug-fix updates
  to installed OS codec packages.
- Codec processing runs without root, is bounded for CPU, memory, and temporary
  storage, and is only started after authentication and RBAC authorization.
- Original-file byte ranges remain the baseline pause/resume and seek contract.
  Any transcoded playback mode defines its own time-based seek and continuation
  behavior instead of treating transformed bytes as source-file byte offsets.
- An unfamiliar media format is not rejected solely by extension or format.
- Delivery does not promise playback on a client that lacks the required image
  or video decoder; this limitation is documented.
- Content type follows a documented safe policy and is not used for
  authorization.

### stream-service-svc/FR-006 Preserve and deliver Apple Live Photos

#### stream-service-svc/FR-006 Description

When storage exposes the components and association, the service MUST deliver
an Apple Live Photo as its related still-image and motion-video pair.

#### stream-service-svc/FR-006 Input

- Authorized still-image and movie components
- Available Live Photo content identifier or equivalent source metadata

#### stream-service-svc/FR-006 Output

- Both original components and their association, or individual media
  responses when no trustworthy association is available

#### stream-service-svc/FR-006 Acceptance Criteria

- HEIC/HEIF or JPEG stills and MOV/QuickTime motion videos can be served in
  original form.
- The service preserves the still-to-video association from source metadata;
  it does not infer pairing from similar filenames alone.
- The interface exposes enough information for a compatible client to retrieve
  both components as one logical Live Photo.
- Missing or malformed association metadata does not cause unrelated files to
  be represented as a Live Photo; each component may still be served
  individually if authorized.

### stream-service-svc/FR-007 Return safe media metadata

#### stream-service-svc/FR-007 Description

The service MUST return only media metadata needed by clients and MUST not
disclose protected file information before authorization.

#### stream-service-svc/FR-007 Input

- Authorized media file metadata

#### stream-service-svc/FR-007 Output

- Safe response metadata such as media type and length

#### stream-service-svc/FR-007 Acceptance Criteria

- Metadata is derived from trusted file information or a documented policy, not
  blindly copied from request values.
- Error responses do not reveal sensitive host paths or protected-file
  existence beyond the documented response policy.
- Response headers cannot be injected through filenames or metadata.

## Interface Requirements

### stream-service-svc/IR-001 Define the playback interface

#### stream-service-svc/IR-001 Description

The service MUST expose a documented HTTP media interface for authorized
playback and range requests.

#### stream-service-svc/IR-001 Input

- Authenticated playback, range, seek, and media-component requests

#### stream-service-svc/IR-001 Output

- Progressive original-format media responses and defined errors

#### stream-service-svc/IR-001 Acceptance Criteria

- The contract defines media identity, request path/method, auth transport,
  byte-range behavior, content type, cache policy, Live Photo component
  association, and error responses before implementation.
- Huma DTO validation is used for HTTP inputs according to repository
  conventions.
- The service HTTP(S) listener uses its own configured port and binds only to
  loopback; it is not directly reachable from non-local interfaces.
- Nginx exposes the service at `/s/` using the existing TLS and
  reverse-proxy configuration conventions. The path suffix maps to the
  corresponding absolute host path; for example, `/s/media/movie.avi`
  maps to `/media/movie.avi`. RBAC checks the resolved host path.
- Browser access is through the HTTPS Nginx origin so Secure cookies are sent;
  the service's loopback listener is not a browser-facing endpoint.

### stream-service-svc/IR-002 Use protected LiteNAS transports

#### stream-service-svc/IR-002 Description

The service MUST use LiteNAS's established TLS, NATS messaging, and HTTP proxy
topology for internal and external communication.

#### stream-service-svc/IR-002 Input

- Service transport configuration and provisioned service certificates

#### stream-service-svc/IR-002 Output

- Authenticated, encrypted connections at the documented trust boundaries

#### stream-service-svc/IR-002 Acceptance Criteria

- NATS uses the shared messaging configuration, dedicated client certificate
  identity, trusted CA, and private key.
- NATS authorization grants only required Auth and RBAC subjects.
- Internal service-to-service calls use NATS; HTTP is the client-facing
  transport and is not used as an internal service RPC path.
- Required certificate paths and TLS configuration are validated at startup.

## Reliability Requirements

### stream-service-svc/RR-001 Fail closed on authorization and transport errors

#### stream-service-svc/RR-001 Description

The service MUST deny media access when token validation, RBAC, path resolution,
or required transport security cannot be established safely.

#### stream-service-svc/RR-001 Acceptance Criteria

- No token or RBAC failure is interpreted as allow.
- No media bytes are read or written before a trustworthy RBAC allow.
- A failed stream closes file and response resources without leaving persistent
  authorization state.

### stream-service-svc/RR-002 Bound stream resource use

#### stream-service-svc/RR-002 Description

The service MUST bound memory and release resources when streams complete,
fail, or are cancelled.

#### stream-service-svc/RR-002 Acceptance Criteria

- The service does not buffer the entire media file in memory.
- Client disconnects stop unnecessary file reads and close open file handles.
- Concurrent streams and read buffers follow documented operational limits.

## Operational Requirements

### stream-service-svc/OR-001 Run non-root with least-privilege media access

#### stream-service-svc/OR-001 Description

The service MUST run as a dedicated non-root account with only the host read
access required for its service role.

#### stream-service-svc/OR-001 Acceptance Criteria

- Runtime effective UID is nonzero; operation does not depend on root.
- Required media and certificate permissions are documented and installed
  through LiteNAS deployment/package configuration.
- Local deployment and Debian install/upgrade flows idempotently create
  `lite-nas-fs-rd`. Stream Service joins that group, which grants file read and
  directory read plus search/traverse in provisioned file-backed storage
  roots, with no write permission. It does not receive a host write grant.
  Define the root scope, system-critical exclusions, and existing/new ZFS
  mount ACL provisioning before implementation. RBAC remains the
  caller-specific decision maker; service group access does not authorize a
  caller.
- The Stream Service AppArmor profile is packaged as
  `configs/etc/apparmor.d/usr.libexec.lite-nas.stream-service`, attaches to
  `/usr/libexec/lite-nas/stream-service`, and is installed, validated, and
  loaded in enforce mode by local deployment and Debian install/upgrade flows
  through the repository's managed AppArmor profile loader. The profile limits
  access to configured logs, certificates, required codec tools/libraries, and
  host media paths needed for authorized reads.
- Host permissions do not replace per-request RBAC authorization.
- File access uses race-resistant descriptor-relative or equivalent path
  handling supported by the target Linux platform.

### stream-service-svc/OR-002 Integrate local build, test, deployment, and

Debian publication

#### stream-service-svc/OR-002 Description

The service MUST be integrated into LiteNAS's local development and single
Debian-package workflows.

#### stream-service-svc/OR-002 Acceptance Criteria

- Dedicated build, deploy, and test scripts follow existing repository
  conventions and are included in aggregate local workflows.
- Debian package build includes the service binary, config, systemd unit,
  AppArmor policy, required runtime assets, and declared OS codec runtime
  dependencies when codec processing is enabled.
- Local deployment and Debian install/upgrade flows create the service account
  and provision its read/search group idempotently, without granting host
  write access.
- Package and install validation verify that the Stream Service AppArmor
  profile is installed and loaded in enforce mode.
- Package analysis, contents validation, PR install validation, and supported
  architecture release workflows include the service.
- The service remains part of the existing single LiteNAS package.

## Security Requirements

### stream-service-svc/SR-001 Protect credentials and media metadata

#### stream-service-svc/SR-001 Description

The service MUST protect bearer tokens, private keys, media bytes, and sensitive
host path details.

#### stream-service-svc/SR-001 Acceptance Criteria

- Tokens, private keys, and protected media bytes are not written to logs.
- Stream URLs, identifiers, and continuation positions do not independently
  authorize a request.
- Certificate permissions allow the non-root process to read only what it
  needs and do not expose CA private keys or unrelated service identities.

## Integration Requirements

### stream-service-svc/IR-003 Share common service infrastructure

#### stream-service-svc/IR-003 Description

Behavior that is identical across File Service and Stream Service MUST use
shared LiteNAS infrastructure.

#### stream-service-svc/IR-003 Acceptance Criteria

- Common token verification, RBAC subject/path authorization, and safe
  file-access/path-resolution behavior reuse existing shared packages or the
  smallest new shared package required by both services.
- Media playback and range handling remain owned by Stream Service.
- Shared behavior and contracts are explicit and are not duplicated across
  service modules.

### stream-service-svc/IR-004 Prefer native Unix and OS-maintained codecs

#### stream-service-svc/IR-004 Description

The service SHOULD use native Linux/Unix facilities and mainstream Go libraries
for service logic. Codec processing SHOULD use compatible codecs installed and
updated as host OS packages where practical.

#### stream-service-svc/IR-004 Acceptance Criteria

- The Go standard library and native file/process facilities are used for
  service logic where they safely meet the requirements.
- Codec libraries/tools are not embedded in the Go binary when equivalent
  maintained system packages are suitable. Runtime OS package dependencies are
  declared and installed through LiteNAS's supported package workflow.
- A compatible OS codec package update does not by itself require a new LiteNAS
  application release. An application release is required for incompatible
  changes to the codec integration or service playback contract.
- Additional Go dependencies remain simple, mainstream, maintained, and
  minimal. A media server or transcoder is introduced only for a documented
  playback need, and its codec dependency follows the same OS-package policy.

### stream-service-svc/IR-005 Rotate service logs

#### stream-service-svc/IR-005 Description

The Stream Service log MUST be included in LiteNAS-managed log retention.

#### stream-service-svc/IR-005 Acceptance Criteria

- The service writes its application log to
  `/var/log/lite-nas/stream-service.log` using the shared logging
  configuration.
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

### stream-service-svc/TR-001 Verify authentication, RBAC, ranges, and formats

#### stream-service-svc/TR-001 Description

The service MUST be testable without production media files, live host accounts,
or production Auth/RBAC instances.

#### stream-service-svc/TR-001 Acceptance Criteria

- Service-local tests cover missing/invalid tokens, RBAC allow/deny/errors,
  path safety, valid/invalid ranges, pause/resume, seeks, disconnects,
  original-format delivery, system-codec integration/failure behavior where
  enabled, and Live Photo association.
- Token validation, RBAC, file access, and response behavior can be controlled
  or replaced in tests.

### stream-service-svc/TR-002 Verify transport, certificates, and runtime access

#### stream-service-svc/TR-002 Description

Transport security and non-root runtime access MUST have automated or repeatable
integration coverage.

#### stream-service-svc/TR-002 Acceptance Criteria

- Configuration checks cover valid and missing/invalid TLS certificate paths.
- NATS integration verifies the service certificate identity, trusted CA, and
  restricted Auth/RBAC subject access; invalid credentials fail closed.
- The loopback-only listener, Nginx path mount, HTTP path forwarding, TLS
  boundary, and certificate bootstrap/rotation are checked.
- Browser integration coverage verifies that `lite-nas-at` is sent to
  `/s/` over HTTPS, is not sent over HTTP, and cannot be read by frontend
  JavaScript.
- Deployment checks verify the non-root account can access required media and
  certificates, but not unrelated certificate material.
