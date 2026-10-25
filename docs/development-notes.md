# Development Notes

## Why Moments is in this slice

The security and platform skeleton establishes authentication, file access,
deployment, and browser hosting. The Moments gallery is outside that skeleton's
security scope. It is included as a near-term product priority because photo
browsing is the most useful LiteNAS application for the maintainer's family.

The fixture-backed gallery demo now provides bounded media loading, responsive
wall layouts, a viewer, video playback, and Live Photo motion through the
existing File and Stream Service routes. A browser-only thumbnail maker uses
the same fixture. The Moments shell supplies authentication and hosting.
Album, tag, pagination, upload, and persistence contracts still need Moments
service work. See the
[Moments architecture](moments-architecture.md).

## Why system-metrics came first

The `system-metrics` service is not the highest business-value capability in the
intended LiteNAS scope. By itself, it delivers only a small part of the overall
platform vision.

It was still chosen as the first implemented service on purpose.

`system-metrics` is simple enough to let the project establish the foundational
building blocks first:

- repository structure
- shared Go modules
- service runtime patterns
- configuration loading
- logging
- messaging integration
- test conventions
- build scripts
- deployment scripts
- packaging
- CI/CD workflow structure

This follows the intended LiteNAS development approach: build the platform from
small, understandable bricks first, and only then expand into higher-value
business logic and broader product capabilities.

The goal is to avoid jumping directly into complex domain behavior before the
base infrastructure is reproducible, testable, and maintainable.

In practice, `system-metrics` acts as an infrastructure-seeding service rather
than a statement of final product priority.

## Why the web skeleton is still infrastructure-first

The current web implementation adds:

- `web-gateway` as the browser-facing Go HTTP boundary
- `auth-service` as a dedicated PAM-backed auth authority
- `admin-panel` as a Vite + React browser app with alerts and telemetry pages
- packaging and assembly wiring so web assets are consumed by backend packaging
- CI/CD updates so JS/TS build artifacts are part of the reproducible build flow

This is still intentionally lower immediate business value than the later NAS
feature set, but it is no longer just a shell.

The purpose is not to ship a feature-rich user-facing product yet. The purpose
is to complete the next missing infrastructure boundary:

- browser-facing app wiring
- service-to-web integration points
- frontend build integration in CI/CD
- asset handoff into packaging
- reproducible installation and release flow beyond local scripts or manual tweaks

This stage should make the platform feel operationally complete enough that the
remaining effort is mostly product and domain implementation rather than
continued build and deployment bootstrapping.

## Platform-web-skeleton branch outcome

- The platform now has several services and apps that build and install through
  the same repository-owned flows.
- Internal service communication is wired over NATS rather than direct browser
  access to backend internals.
- The browser gateway serves packaged admin-panel assets and exposes documented
  browser-facing auth, alert, and telemetry surfaces.
- The frontend build output is produced under `.build/admin-panel`, normalized
  for gateway serving, and included in Debian package assembly.
- CI/CD scripts cover frontend dependencies, frontend build, JS/TS analysis,
  duplication checks, package analysis, and package install validation.
- The branch remains intentionally small in end-user behavior, but it completes
  enough of the skeleton that future branches can focus on actual NAS product
  development rather than platform bootstrapping.

## Why the alerting slice now has a browser surface

The alerting work now spans both CLI and browser access patterns. The browser
app can show and manage both system and security alerts, while the existing CLI
paths still matter for headless operation and administrative verification.

That branch added:

- `zfs-metrics` as a second metrics producer beside `system-metrics`
- `resources-monitor` as a rule-based alert source consuming system and ZFS
  metric events
- service-to-service authentication for internal callers
- `rbac-service` role and capability lookup used by auth and target services
- auth-token and role enforcement in logging-manager message handling

This gives LiteNAS an end-to-end alerting path that is useful in a home-lab
deployment and now has a browser-facing control surface:

- metric-producing services publish snapshots
- `resources-monitor` evaluates rules and emits alert lifecycle updates
- logging managers consume those updates and keep a user-friendlier state-based
  view of active alerts
- the admin panel exposes system and security alert dashboards for review and
  management
- raw stateless events still remain available in log files as a separate source
  of operational detail

For the current use case, the important milestone is that resource and
condition-change alerting now works through the real service boundaries that
both CLI and UI consumers can reuse.

## File Service and planned media service

File Service is implemented and locally deployed. Stream Service remains a
separate planned boundary:

- File Service owns authenticated uploads/downloads, explicit pause/resume, and
  transfer state.
- Stream Service owns progressive media delivery, HTTP range requests for
  pause/resume/seeking, and Live Photo component association.
- Both services validate the caller's token and use RBAC for path permission
  decisions. They share common token/RBAC and safe file-access infrastructure
  when the behavior is identical.
- Each service exposes an HTTP(S) listener on its own loopback-only port. Nginx
  mounts File Service at `/f/` for content/file operations and `/d/` for
  explicit downloads, and Stream Service at `/s/`; internal
  service-to-service calls continue to use NATS. The route suffix maps to the
  corresponding absolute host path: `/f/home/alice/photo.jpg` and
  `/d/home/alice/photo.jpg` map to `/home/alice/photo.jpg`, and
  `/s/media/movie.avi` maps to `/media/movie.avi`. RBAC checks the resolved
  host paths.
- The packaged `/etc/logrotate.d/lite-nas` policy rotates File Service and
  Stream Service logs daily, retaining seven compressed rotations.
- Local deployment and Debian installation create the shared `lite-nas-fs-rd`
  group idempotently. File Service uses it for file read and directory read
  plus search/traverse in provisioned file-backed storage roots, with no write
  access. Stream Service also joins this group. System-critical paths
  are excluded. Existing content and new mounts need read/search group/ACL
  provisioning. File Service has a dedicated enforce-mode AppArmor profile.
- Browser requests use the gateway's existing `lite-nas-at` access cookie over
  the same HTTPS origin. Its `Secure`, `HttpOnly`, `SameSite=Lax`, and `Path=/`
  attributes allow both path mounts while keeping the token unavailable to
  frontend JavaScript. Nginx forwards the cookie; each service validates it.
- Because upload requests use an automatically attached cookie, File Service
  enforces CSRF protection on state-changing operations.
- File Service has two upload modes: conventional multipart upload at
  `/f/upload` for browser/form clients, and a resumable app protocol under
  `/f/proprietary/upload` with initialization, ordered chunks, progress,
  resume, and finalization. Chunk size defaults to 1 MiB and is configurable
  from 1 MiB through 16 MiB; the effective size is saved per upload.
  Concurrency is limited by a 60 MiB chunk-buffer budget in a 64 MiB tmpfs,
  with 4 MiB reserved for metadata working files. A failed chunk retries from
  the last committed offset; if buffers are full the service returns a
  retryable on-hold response and the app probes
  `/f/proprietary/upload-window`. Both upload modes exclusively create and
  write a user-owned `<destination>.upload` file in the destination directory
  through the dedicated `lite-nas-file-writer` Go application, invoked as the
  verified Unix user through restricted noninteractive sudo. The writer's
  exclusive create is the Unix permission check; no preliminary `touch` or
  service write group is required. After size and SHA-256 verification and a
  fresh RBAC decision, the writer attempts the final rename as that user. Its
  separate AppArmor profile must remain enforced across sudo. A failed rename
  leaves the old destination intact and
  triggers staged-file cleanup. Successful uploads atomically replace existing
  regular files. Configured `filesystem.protected_paths` are
  excluded from uploads and deletes.
- Its authenticated Swagger UI is available at `/f/docs`, with the OpenAPI
  document at `/f/openapi.json`; development and test configurations enable
  them, while production disables them through `api_docs.enabled`.
- File Service adds public `GET /ls/{absolute-path...}` with presence-based
  `recursive` and `files-only` query flags, separately from its `/f/`
  file-operation mount. For example,
  `/ls/abs/path?recursive&files-only` lists authorized files below host
  `/abs/path`. Success JSON has `files` and `directories` arrays of absolute
  host paths; `files-only` leaves `directories` empty. Failures return a
  non-success status and JSON with an `error` code. Listing and traversal use
  RBAC read and execute/search decisions for the subject, including ancestor
  directories.
- File Service supports `DELETE /f/{absolute-path...}` for regular files when
  RBAC explicitly authorizes deletion. The packaged configuration protects
  OS/runtime paths including `/etc`; operators can edit
  `filesystem.protected_paths`. Directory and symlink deletion are not
  part of the initial scope.
- File downloads use ordinary HTTP byte ranges. The future file-explorer UI can
  fetch up to 16 ranges concurrently by default and reassemble them; each range
  request is authorized independently and carries a strong ETag condition so
  segments cannot silently mix file versions. No separate download-session
  protocol is needed. `/f/` can present recognized PDFs/images inline and
  returns attachment headers for other content; `/d/` always requests a
  download as `application/octet-stream`. Neither route redirects to the other.
- Both run as non-root services. RBAC authorizes the caller; the read group
  gives the processes storage read access, and the verified Unix user performs
  upload creation, writes, and final publication. Completed uploads retain that user's
  UID. The stage's group, mode, and ACL become the final file's metadata even
  on overwrite; the old file's permissions and ACL are not copied.
  The stage stays `0600` through rename; publication then applies `0664`
  for shared uploads or `0660` for private uploads. Directory
  creation defaults to `0775`, with `0770` for private requests and `0771`
  for configured known-path directories.
- The design favors native Unix facilities, the Go standard library, and small
  mainstream Go libraries. Media is delivered in original formats by default;
  where playback requires codec processing, the Stream Service prefers
  OS-maintained codec packages over codecs bundled in Go so compatible codec
  fixes can arrive through OS package updates. Client compatibility still
  depends on the playback client's codec support unless a documented conversion
  mode is provided.

File Service is wired into the local build and deployment flows and the
single Debian package build. Its first-version reader uses fixed `/bin/cat`
through `sudo -u` after RBAC allows the path. The non-root service unit grants
`CAP_SETUID` and `CAP_SETGID` only within its capability bounding set; the
AppArmor profile also permits those capabilities and sudo's PAM reads. The
pathname replacement race and required follow-up tests remain documented in
`services/file-service/READ_RACE_REVIEW.md`. The system test suite covers
gateway-facing authentication and read/list access cases, while broader
installed upload, package, and hardening verification remains open in
`services/file-service/task.md`. Stream Service requirements and plan remain
in `services/stream-service/`.

## Logging Manager Channel Split

The split between `system-logging-manager` and `security-logging-manager` is
intentional in the same way separate storage or service slices are intentional.

At this stage, the system side still has the richer set of monitoring producers
and alert sources. The security side now has browser and CLI management surfaces
for its alert state, even though its upstream monitoring producers remain more
limited than the system path.

That is acceptable for the current version line because the important platform
work is already in place:

- separate manager contracts and storage domains
- distinct CLI surfaces
- service authentication and role checks at manager boundaries
- package/runtime wiring for both domains

This keeps the architecture ready for future security-monitoring producers
without forcing fake divergence before the behavior exists.

## Why email notification still uses local Postfix

The current notification slice intentionally keeps delivery split across two
layers:

- `system-email-notifier` and `security-email-notifier` own alert consumption,
  template rendering, and local SMTP handoff
- local Postfix owns outbound mail transport

This is intentional.

The notifier services should stay small and transport-focused:

- subscribe to alert subjects
- render email from packaged templates
- deliver to `127.0.0.1:25`

They should not also own:

- third-party SMTP provider integration details
- SASL credential storage
- relay retry behavior
- machine-local mail transport policy

Keeping Postfix as the host mail boundary makes the deployment model more
practical:

- the same Go code works in local development and packaged installs
- provider credentials stay machine-local under `/etc/postfix/...`
- direct local rendering tests can switch to capture mode without changing the
  notifier code path
- production delivery can move to Mailgun, SES, or another authenticated relay
  without changing the notifier service contract

This is still infrastructure-first work in the same sense as earlier slices.
The immediate user-facing value is modest, but it completes an important
platform boundary: alert creation can now flow into real notification delivery
through the packaged runtime.

## Why simple system tests still carry more value now

The top-level system tests remain intentionally focused on simple, observable
verification points.

That simplicity no longer means the checks are shallow.

For example, a logging-manager CLI mutation test now exercises a longer
integration chain:

- CLI/runtime credential and local permission access
- service-to-service authentication
- RBAC-backed role retrieval in auth flows
- access-token validation in the logging manager
- manager-side role authorization before accepting a state change

So the tests still read as operationally small checks, but they now validate
more of the real system integration stack as a side effect. That is a useful
tradeoff for the current stage: one clear verification point can still prove a
meaningful amount of runtime wiring.

## Boundary Validation Direction

LiteNAS treats validation as a boundary concern. User-facing and
service-facing inputs should be validated before any meaningful processing
happens inside the application.

For HTTP and HTTPS boundaries, `web-gateway` should keep using Huma DTO
validation because those DTOs already describe browser-facing transport
contracts and OpenAPI metadata.

For non-HTTP Go boundaries, including CLI tools, NATS request/reply handlers,
NATS subscription handlers, and other message handlers, the default validation
library should be `go-playground/validator/v10`. These handlers should validate
incoming command or message structs before mapping them into service calls.

For frontend runtime validation, the admin panel should use Zod by default.
TypeScript DTOs document compile-time shapes, but Zod schemas should enforce
runtime input constraints before requests are sent. Form components may own
user-facing field state and messages, but providers and API actions should still
reuse the same schemas before accepting submitted payloads.

This split keeps HTTP/OpenAPI concerns in Huma, gives internal service messages
and CLI inputs a transport-neutral Go validation path, and gives browser code a
shared runtime validation layer.

## Packaging and Runtime Flow

For packaging/runtime orchestration policy and CI parity expectations, see
[Packaging Runtime Flow](./packaging-runtime-flow.md).

## Logging Managers Duplication (Current Intent)

`system-logging-manager` and `security-logging-manager` are intentionally kept
structurally identical in this stage, including their paired CLI apps.

Current expected differences are limited to:

- subscribed and requested NATS subjects/events
- NATS identity permissions
- service and CLI configuration files/paths

This is intentional for the current version line. Divergence is expected in
higher versions when security-specific behavior moves beyond contract-level
differences.

To keep duplication checks clean, shared runtime/CLI logic is centralized under
`shared/go/loggingmanagerservice` and `shared/go/loggingmanagercli`, while
system/security binaries remain thin adapters over different subjects and
configs.
