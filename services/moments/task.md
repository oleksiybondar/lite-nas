# Moments Service — Remaining Metadata Backend Work

## Current state

The Go service currently provides the browser session API only: login, refresh,
logout, and authenticated session inspection through the existing LiteNAS Auth
Service integration. Same-origin checks, cookie behavior, loopback binding, and
embedded/standalone deployment are implemented. Album and gallery metadata
routes, persistence, and scans are not implemented. See the [architecture
overview](../../docs/moments-architecture.md) and
[implemented skeleton requirements](../../requirements/moments-skeleton.md).
Use [File Service requirements](../file-service/requirements.md) and the
[Stream Service README](../stream-service/README.md) for implemented media
transport behavior.

## Goal and ownership

Add the metadata capabilities the Moments app needs while preserving LiteNAS
service boundaries:

- Moments owns album/collection metadata, membership references, scan-derived
  tags, and a versioned cover descriptor.
- File Service remains authoritative for file access, ACL checks, directory
  listing, uploads, and basic filesystem metadata. Stream Service remains the
  media playback/preview path. Neither service currently exposes a JSON API
  for EXIF, named locations, dimensions, or durable Live Photo association.
  Moments must not read NAS paths directly or
  proxy/store original media.
- Derive the caller identity from the validated LiteNAS session. Enforce
  gallery authorization on every operation. Moments visibility and sharing
  are application-level metadata permissions; they do not change Unix ACLs
  or File Service access to a source file.
- Reuse shared profile and appearance settings where they already meet the
  app's needs. Do not add a second profile/settings store without a documented
  gap.

## Work to define and implement

1. **Repository and contract audit.** Inspect existing Moments, File Service,
   shared auth, configuration, persistence/migration, deployment, and frontend
   conventions. File Service already exposes authorized
   `GET /ls/{absolute-path...}` with presence-based `recursive` and
   `files-only` flags. Its response is one JSON array of absolute host paths
   and basic filesystem fields; it has no pagination. Define source-qualified
   requirements and a typed, validated metadata API contract (including
   errors, IDs, ordering, and gallery-list pagination)
   before implementing adapters. Do not copy proposed route names or DTOs from
   the design discussion without this verification.
2. **Gallery metadata.** Define and implement the smallest model that supports
   named albums backed by a source directory and curated collections backed by
   file/directory references. Support the app's required create, read, update,
   delete, listing, and filtering workflows. Deleting metadata or a reference
   must never delete an original file. Keep reference path interpretation
   consistent with File Service; do not invent canonicalization rules.
3. **Persistence and operations.** Persist metadata and scan results using the
   established LiteNAS storage/configuration and migration conventions.
   `/etc/lite-nas/moments.conf`, the `lite-nas-moments` nonroot systemd unit,
   and its AppArmor profile already exist. The current unit permits writes to
   the service log only; any database or owned cover asset requires deliberate
   provisioning and narrow unit/profile changes. The discussion's
   `/etc/lite-nas/moments/moments.conf` path conflicts with the installed file.
   Choose a database location consistent with LiteNAS conventions before
   adding it. Provide safe fresh-install and upgrade behavior, bounded
   transactions, and documented backup/restore implications.
4. **Authorized discovery and tags.** The browser will load each album by
   calling File Service `/ls/<album-directory>?recursive&files-only` with its
   shared cookie and compose the complete visible list. Moments does not need
   to proxy or duplicate that list for the agreed album flow. Resolve any
   backend collection references and scan inputs only through File
   Service operations that enforce the caller's ACL. The existing `/ls/` route
   returns no media tags or image dimensions, and recursive responses are not paged.
   Define stable media identity and a new File Service media-tag extraction
   contract explicitly. It must return actual extracted tags and media type
   for accessible files; do not infer named locations or Live Photo identity
   from paths when the extractor cannot supply them. Define refresh separately
   from metadata rescan.
5. **On-demand full rescans.** Start each album or collection rescan from an
   authenticated Moments request, then run bounded work in an internal Go
   goroutine. Resolve its current source files, request media tags through a
   newly defined File Service endpoint (NATS preferred, HTTP acceptable after
   contract review), and report durable progress for same-user polling.
   Rebuild the target gallery's media-type and extracted-tag associations as
   a complete replacement; do not append only new files or clear another
   gallery's tags. Stage results and atomically publish them on successful
   completion. Define partial failure, cancellation, concurrent-scan, and
   restart behavior. Keep file authorization tied to the initiating user for
   each extraction call. Hold the initiating user's current access JWT only
   in process memory. A progress request may replace it only after validating
   a fresh JWT with the same user ID. The worker cannot continue extraction
   with an expired JWT. Never use a service token as the user's identity or
   persist raw user bearer/refresh tokens.
6. **Covers and preferences.** Persist only an agreed versioned cover
   descriptor that refers to stable media identity or an explicitly
   app-owned image asset. The browser thumbnail recipe is a demo format and may
   contain uploaded data URLs; it is not a service DTO. Reuse existing shared
   profile/appearance settings for applicable preferences and add only
   gallery-specific values required by the app.
7. **Moments visibility and sharing.** Include owner, shared-member, and
   public visibility in the gallery metadata model. Enforce these only for
   Moments views and metadata operations; do not alter Unix permissions or
   File Service ACLs. Apply one policy to lists, aggregates, references,
   covers, and scan status so hidden galleries do not leak through counts or
   filters. Define the public audience and member mutation rights before
   implementing the exact API.

## Decisions required before implementation

The [design discussion](design-discussion.md) and
[gallery reference](../../apps/gallery_taks.md) are inputs, not authoritative
API specifications. Resolve these items against code and LiteNAS conventions:

- Exact routes, DTOs, response/error envelope, gallery-list pagination, stable media IDs,
  sort order, and File Service URL construction.
- Phase albums and collections against the agreed app workflow. Both are in
  the target model; collection reference mutation and path identity still
  need exact contracts.
- The browser will request one recursive `files-only` array per album. Confirm
  the practical response size and latency on target hardware. File Service
  currently materializes the entire array; any future paging extension needs
  a separate agreed File Service contract.
- Specify the new File Service media-tag operation, with NATS preferred.
  File Service `/ls/` supplies file kind, size, owner, group, permissions,
  and modification time, not actual extracted tags. Define exactly which
  media types and tag values its extractor can return.
- Define the exact JWT handoff and replacement DTO for NATS or HTTP scans.
  Existing Auth refresh RPC requires the browser's refresh token; Moments
  jobs do not receive it. Service tokens identify a service, not the
  initiating user. Progress checks from the same validated user may replace
  the job's in-memory access JWT. Existing access JWTs remain valid until
  expiry after logout. Define bounded expiry/pause behavior, partial failures,
  cancellation, concurrency, and restart handling without accepting a
  different user as job owner.
- Cover descriptor/version, source-move behavior, and whether explicit image
  upload/storage is needed. The gallery reference's "one primary URL" and
  original-media-first guidance does not define a cover API.
- Whether public means all authenticated LiteNAS users or anonymous visitors,
  and whether shared members can mutate metadata or start scans. Moments
  visibility is independent of File Service access in either case.
- Storage path, SQLite/migration mechanism, and any deployment changes, based
  on the installed `/etc/lite-nas/moments.conf`, nonroot systemd unit, AppArmor
  profile, and single-package workflow.

Do not implement proposed account profiles, independent settings APIs, public
anonymous gallery access, arbitrary wall-rendering JSON, or a new File Service
contract merely because they appear in the discussion. First establish a
concrete LiteNAS app requirement and a compatible existing contract.

## Acceptance

- The documented metadata API supports the agreed frontend workflows with
  request and response validation at the HTTP boundary.
- Metadata survives service restart and supported upgrades; authorization is
  derived from the validated session and checked on each operation.
- The browser can discover nested album files through File Service `/ls/`;
  backend album and collection rescans obtain real media tags through the new
  authorized File Service contract and report progress. A successful scan
  atomically replaces only that gallery's media-type and extracted-tag set.
  Failure behavior cannot silently erase previously valid metadata.
- Moments never reads, uploads, moves, deletes, or serves original media.
- Existing login/session behavior, service wiring, and admin-host integration
  continue to work.
- Focused unit/integration coverage and API/configuration documentation match
  implemented behavior. Do not claim gallery metadata is complete while the
  app still depends on the fixture.
