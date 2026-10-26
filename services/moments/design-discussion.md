<!-- markdownlint-disable MD041 MD013 MD029 -->
<!-- Historical contract notes retain long examples and section numbering. -->

TASK — Moments gallery metadata backend

Status: Historical design discussion; implemented routes and DTOs are documented in README.md and OpenAPI
Prepared: 2026-10-25
Scope: Moments backend, new user-authorized File Service media-tag operation (NATS preferred; HTTP acceptable after contract review), frontend integration, Linux service deployment, security, tests.
Authority: Product ideas and candidate policies. Existing LiteNAS code and the
Moments [service task](task.md), [app task](../../apps/moments/task.md), and
[gallery reference](../../apps/gallery_taks.md) describe the current integration
boundary. This document does not establish an implemented API contract.
The implemented transport behavior is documented in
[File Service requirements](../file-service/requirements.md) and the
[Stream Service README](../stream-service/README.md).

Implementation update (2026-10-26): The implemented Moments API is documented
in [README.md](README.md) and served at `/api/openapi.json`; the candidate
contracts below are retained as design history. The gallery list uses
`page`/`size` and returns `items` plus `metadata` with page, size, total count,
and total pages. Its default search includes ALBUM and COLLECTION, matches
name, description, user tags, and all scan-derived tag values; `type` narrows
the gallery kind. Gallery responses carry typed tag entries. User tags have
their own gallery-scoped CRUD routes and storage, so media rescans and source
changes cannot remove them. `AUTO` tags include source metadata, generated
`MEDIA_TYPE`, and extension-based `FILE_TYPE` values. Their association count
is the number of matching media; a gallery-level `USER` tag has count one.

Verified repository state (2026-10-25): Moments implements `/api/auth/login`,
`refresh`, `logout`, and `me` only. The demo gallery loads a bundled fixture,
does not consume Moments gallery responses, and does not call `/ls/` to populate
its media list. The header's shared file/folder dialogs use File Service
listing for selection only. File Service already implements authorized
`GET /ls/{absolute-path...}` with presence-based `recursive` and `files-only`
flags. It returns a single unpaged JSON array of basic filesystem entries,
not dimensions, EXIF/location tags, or Live Photo associations. File Service
also owns `/f/` original reads and uploads and `/d/` attachment downloads;
Stream Service owns `/s/` original playback, `/s/preview/` HEIC/HEIF JPEG
previews, and `/s/transcode/` fallback streams. All these requests recheck
user authorization. Actual installed configuration is
`/etc/lite-nas/moments.conf`, with an existing nonroot systemd unit and AppArmor
profile. Shared appearance/profile files already use File Service
`/f/profile/ui` and `/f/profile/avatar`.

Conflicts resolved by the repository audit and the agreed album-loading flow:

| Earlier draft assumption | LiteNAS-aligned requirement |
| --- | --- |
| The demo already adapts File Service listings or a Moments media DTO | The gallery currently adapts only a bundled fixture. Add direct `/ls/?recursive&files-only` loading for Albums. |
| Moments must paginate Album media | The browser composes one complete list from File Service's unpaged `/ls/` response. Measure listing memory and latency separately from virtualized rendering. |
| File Service already extracts media tags, dimensions, named locations, and Live Photo identity | Its current `/ls/` response provides basic filesystem metadata only. Add a separate authorized media-tag extraction operation for required full rescans. |
| Moments needs separate profile/settings routes | Shared appearance and avatar already use File Service profile routes; add only agreed gallery-specific settings. |
| Use `/etc/lite-nas/moments/moments.conf` and new service hardening assets | Keep installed `/etc/lite-nas/moments.conf`, systemd unit, and AppArmor profile; extend their narrow storage permissions if needed. |
| Original-only delivery forbids all server transforms | Stream Service already provides authorized on-demand HEIC/HEIF JPEG previews and video transcode fallback. Persistent derivative files remain outside the baseline. |

Moments visibility and sharing are application-level metadata permissions;
they do not change Unix ACLs or File Service authorization. Full on-demand
rescans are required. The public audience, shared-member mutation rights,
scan expiry/pause policy, exact extraction DTO, and persisted cover asset
contract remain decisions below.

0. Instructions to implementing agent / team

1. Inspect the existing Moments backend, File Service, authentication integration, liteNAS deployment conventions, and demo frontend before editing. Do not replace working login/logout, file transfers, or frontend layouts. Record relevant code locations and current contracts in the implementation PR.
2. Implement the functional behavior below; preserve established language, framework, API prefixes, error conventions, identity formats, .conf parser, and migrations tooling where viable. All routes below are proposed canonical names; map them consistently to existing routing conventions and document any differences.
3. Define a new File Service media-tag operation and its DTO before implementation; prefer NATS when its caller identity can be bound to the initiating user, otherwise use a reviewed HTTP contract. Do not invent auth-token refresh facilities or bypass File Service’s ACL.
4. Maintain an explicit decision log for design choices determined from repository evidence. Unresolved security policy decisions listed in §16 must be settled before exposing access publicly, using the conservative defaults stated there if product-owner input is unavailable.
5. Deliver executable migrations, running service, documented API contract, updated frontend integration where impacted, automated tests, deployment assets, and verification results—not merely model definitions.
6. Minimize breaking changes. Where existing clients use old contracts, add transition adapters or migrations and document them. No silent loss of existing data.

1. Product intent and ownership boundaries

Moments is a gallery metadata and organization service, not a general media file provider. File Service handles original media upload, access, and authorized directory traversal. Its current listing provides basic filesystem fields, not media tag extraction; add a separate authorized operation for that. The frontend uploads through File Service, then refreshes the agreed gallery discovery path. A full metadata rescan is an explicit Moments operation. Moments never uploads, moves, deletes, or serves original media.

Moments owns:

• Gallery entities (ALBUM and COLLECTION), association to referenced files/directories, metadata aggregation and search, ownership/privacy/sharing.
• On-demand full metadata rescans run in bounded internal Go goroutines, request extraction from File Service through a new user-authorized contract (NATS preferred), and expose progress. A successful scan replaces the selected Album's or Collection's complete media-type and extracted-tag associations.
• User-facing gallery configuration (thumbnail generation JSON, wall-rendering strategy) treated as opaque frontend data.
• Gallery-specific settings only where the shared File Service profile contract does not already cover the need.
• Only filesystem exception: explicitly uploaded, application-owned gallery thumbnail image assets under the configured Moments storage root (not original media).

Non-goals: rebuilding login/logout; managing original file bytes; generic file-browser functionality; server-side collage generation or image-layout rendering; custom file ACL engine in Moments; geocoding coordinates by Moments; introducing another database server; shell-based filesystem operations.

Architectural interaction

Browser / frontend
├─ existing auth/login/logout
├─ File Service: upload and original media access using authenticated client
└─ Moments REST (proposed): gallery CRUD/list/metadata, collection refs,
cover descriptor, optional thumbnail assets, and required scan jobs
│
├─ SQLite (Moments metadata and jobs)
├─ proposed Moments-owned thumbnail assets (path and need undecided)
└─ File Service: existing authorized `/ls/` listing plus a new user-authorized
media-tag extraction operation over NATS or HTTP (contract not yet defined)

2. Core behavior and terminology

• Gallery: one primary galleries entity with type = ALBUM | COLLECTION, name (required), description (optional), owner stable UID, PUBLIC | PRIVATE visibility, sharing associations, thumbnail data, and wall-rendering configuration.
• Album: one required upload_directory. The browser calls File Service `/ls/<upload_directory>?recursive&files-only` and composes the full visible file list. Uploads are performed through File Service, not Moments. Directory contents may change independently of Moments; reloading the album repeats the listing and does not launch metadata extraction.
• Collection: explicitly curated list of referenced files and directories from possibly disparate paths. Each item has stable association ID, reference_type = FILE | DIRECTORY, and path. Directory items can be recursively traversed when enumerating/scanning contents. No mandatory upload directory.
• Media type/classification: IMAGE, VIDEO, LIVE_PHOTO, etc., distinct from the ref kind FILE | DIRECTORY. No current File Service JSON extractor supplies these values. Define the source and classification rules before persisting them.
• Location tags: named entities (e.g., Copenhagen, Køge, Drammen, Kyiv), not latitude/longitude. Neither File Service `/ls/` nor Stream Service currently returns these names. This feature needs an agreed extraction source; never infer names from coordinates or filenames without a separate contract.
• Sharing: explicit users allowed to see a private gallery in Moments; owners always retain access. This is virtual application access only. File Service and Stream Service still check the user's source-file rights. Read-only for shared members is a proposed default pending policy confirmation.

3. Candidate SQLite relational design

Use SQLite with migrations, PRAGMA foreign_keys=ON on all connections, transactional writes, and WAL where supported. Schema below is a logical baseline, not a mandate for exact SQL types/column spelling; adapt to existing ORM and naming conventions while preserving semantics.

3.1 Tables

galleries

|Column                     |Meaning / constraints                                                                     |
|---------------------------|------------------------------------------------------------------------------------------|
|id                       |Stable PK (use existing identifier convention)                                            |
|type                     |Required ALBUM or COLLECTION                                                          |
|name                     |Required nonblank bounded string                                                          |
|description              |Nullable bounded string                                                                   |
|owner_id                 |Required immutable authenticated stable UID, indexed                                      |
|visibility               |PUBLIC or PRIVATE; new galleries default to PUBLIC when omitted                      |
|upload_directory         |Required nonblank path for ALBUM; null for COLLECTION, with CHECK constraint if compatible|
|wall_render_strategy_json|Nullable, opaque JSON (could be a simple string strategy identifier)                      |
|created_at, updated_at |UTC timestamps                                                                            |

gallery_references (explicit Collection membership; do not materialize an album’s entire directory as fixed membership unless existing architecture requires it)

|Column          |Meaning / constraints                                         |
|----------------|--------------------------------------------------------------|
|id            |PK                                                            |
|gallery_id    |FK to galleries, indexed                                    |
|reference_type|FILE / DIRECTORY                                          |
|path          |Canonical namespace-qualified file-service path representation|
|created_at    |UTC timestamp                                                 |

Unique (gallery_id, reference_type, path). Only Collections accept manually added/removed refs. The path canonicalization algorithm must match File Service semantics, avoiding unsafe equivalence assumptions (case, Unicode, trailing separators, namespaces). Deleting a ref must not delete source files.

media (optional global deduplicated file identity; conditional design)

|Column          |Meaning / constraints                                       |
|----------------|------------------------------------------------------------|
|id            |PK                                                          |
|file_namespace|File Service namespace/tenant/volume if such identity exists|
|path          |Canonical file path                                         |
|content_type  |Derived classification when scanned                         |
|updated_at    |UTC                                                         |

Only make media globally unique by (namespace,path) if File Service guarantees stable shared identity independent of user permissions. Otherwise use per-gallery resolved file references to prevent cross-user metadata disclosure/collisions. This is an explicit verification gate.

tags (unique metadata definitions)

|Column          |Meaning / constraints                                               |
|----------------|--------------------------------------------------------------------|
|id            |PK                                                                  |
|category      |e.g. LOCATION, MEDIA_TYPE, FEATURE                            |
|value         |Canonical human-readable value (Copenhagen, IMAGE, LIVE_PHOTO)|
|normalized_key|Deterministic uniqueness key within category                        |

Unique (category, normalized_key); never duplicate the same canonical tag entity. Preserve display spelling/diacritics. Implement case/Unicode normalization consistently. For MEDIA_TYPE, can instead use a typed column with indexed gallery aggregate; choose one path and keep filter/API behavior identical.

gallery_media_tags (recommended owner of materialized scan results)

|Column           |Meaning / constraints                                                 |
|-----------------|----------------------------------------------------------------------|
|gallery_id     |FK                                                                    |
|file_key       |Canonical file identity scoped appropriately (or FK to scoped media)|
|tag_id         |FK tags                                                             |
|scan_generation|Optional staging/version marker when using generation swap            |

Unique (gallery_id, file_key, tag_id) per active generation. Gallery-scoped associations are mandatory: scanning gallery A must not clear gallery B’s tags, even when referencing the same file. An optional global cache must never determine access to private content or bypass File Service authorization. Aggregated gallery tags should derive from/efficiently index these associations.

gallery_shares

gallery_id FK, user_id (stable UID), created_at, optionally created_by; unique (gallery_id,user_id); index user. Owner cannot meaningfully share with self. Owner controls membership; deleting a gallery cascades association cleanup.

gallery_thumbnails (one per gallery)

gallery_id PK/FK, mode (NONE, GENERATED_LAYOUT, GALLERY_MEDIA, UPLOADED_IMAGE), nullable source_media_path, nullable generation_json (opaque frontend JSON), nullable stored_asset_name (backend-generated safe filename), updated_at. External URL only if explicitly supported by frontend and security policy. Validate structural envelope; do not interpret generation instructions. Permit a stored rendered preview alongside generation JSON if needed. When a mode changes, clean obsolete owned assets safely after committed metadata changes.

user_profiles

user_id PK (external authenticated identity), optional display_name, avatar_url and timestamps. Profile editable fields must be checked against the current identity provider; do not allow arbitrary editing of authenticated UID, roles, or credentials.

user_settings

user_id PK; settings_json (bounded JSON), updated_at. Decide PUT replacement vs PATCH merge rules and document; never accept settings that modify authorization outside the identity system.

media_scan_jobs

id PK; gallery_id FK (or archival snapshot policy), requested_by UID; status (PENDING,DISCOVERING,RUNNING,COMPLETED,PARTIAL,FAILED,CANCELLED); total_items nullable during discovery; processed_items,successful_items,failed_items nonnegative; created_at,started_at,finished_at, sanitized error_summary; optional credential expiry deadline and run generation marker without storing bearer tokens in plaintext. Index gallery/status/created time. If per-file failures must be shown, add bounded media_scan_errors with sanitized, access-controlled messages.

3.2 Indexes and migrations

• Index gallery owner_id, visibility, type, creation time; shares user_id, gallery_id; references by gallery/type/path; tags category/key; tag associations gallery/tag and gallery/file; jobs by gallery/date/status.
• Prefer indexed EXISTS-based gallery metadata filtering over joins that duplicate results. Check SQLite query plans on representative data.
• Include upgrade and fresh-install migrations, foreign keys, uniqueness checks, stale association cleanup, and legacy-data migration if old tables exist.
• Never overwrite an existing database. Back up/migrate existing data according to liteNAS conventions.

4. Candidate Moments API contract — common behavior

Canonical draft prefix: /api (adapt to current service conventions). Use JSON camelCase unless existing API has an established format; generate an OpenAPI spec or equivalent machine-readable contract. Authenticate via established liteNAS mechanism. Use stable IDs as path params; do not accept arbitrary owner UID as authority.

• Pagination: limit with strict cap, offset or cursor as existing convention; deterministic stable sort; total when affordable. Document maximums and defaults.
• Filtering multiple values: repeated query params, e.g. ?ownerId=u1&ownerId=u2&mediaType=IMAGE&mediaType=VIDEO&location=Copenhagen&location=K%C3%B8ge. OR within a filter, AND across filters. Invalid values/duplicates/large arrays have bounded, documented behavior.
• Standard errors: consistent structured response, 400/422 malformed input, 401 unauthenticated, 403 forbidden, 404 unknown or hidden private resource (prefer non-enumerating behavior), 409 conflict, 413 oversized, 429 throttled, 5xx internal with correlation ID. Never echo credentials or sensitive paths in public error details.
• Ownership, visibility, sharing, and scan authorization must be enforced on every detail/list/mutation/media tag/thumbnail/job endpoint; counts and filters must not leak hidden gallery metadata.

4.1 Gallery CRUD and listing

|Method  |Suggested route                   |Behavior                                                                                          |
|--------|----------------------------------|--------------------------------------------------------------------------------------------------|
|POST  |/api/galleries                  |Create Album or Collection; owner derived from authenticated caller                               |
|GET   |/api/galleries                  |Default accessible union: own OR public OR explicitly shared; deduplicate; filter/search/paginate |
|GET   |/api/galleries/{galleryId}      |Full operational metadata for frontend                                                            |
|PATCH |/api/galleries/{galleryId}      |Owner-only editable fields; validated type-specific fields                                        |
|DELETE|/api/galleries/{galleryId}      |Owner-only delete metadata/references/tags/shares/owned thumbnail; **never delete original files**|
|GET   |/api/galleries/{galleryId}/media|Not part of the agreed album loading path; only reconsider for a separately specified collection need|

GET /api/galleries filters: ownerId repeated, mediaType repeated, location repeated, type repeated, visibility repeated, and single q for name/description. Apply caller-visibility predicate before aggregate filtering/pagination. Expose a default sorting rule; optionally allow a whitelisted sortable set.

Illustrative details DTO only; there is no current frontend Moments gallery DTO:

{
"id": "gallery-123",
"type": "ALBUM",
"name": "Summer 2026",
"description": "Family summer photos",
"ownerId": "user-123",
"visibility": "PRIVATE",
"uploadDirectory": "/photos/summer-2026",
"sharedMembers": [{ "userId": "user-456" }],
"mediaTags": {
"types": ["IMAGE", "VIDEO", "LIVE_PHOTO"],
"locations": ["Copenhagen", "Køge"]
},
"thumbnail": {
"mode": "GENERATED_LAYOUT",
"generation": { "preset": "stacked-cards", "items": [] },
"sourceMediaPath": null,
"assetUrl": null
},
"wallRenderStrategy": { "strategy": "EXISTING_FRONTEND_VALUE" },
"createdAt": "2026-10-25T10:00:00Z",
"updatedAt": "2026-10-25T10:00:00Z"
}

The sample preset and wall strategy are illustrations only, not real supported
enum values. The implemented wall layouts are `mosaic`, `justified`,
`masonry`, and `grid`; any persisted setting must map to these identifiers.
Return `uploadDirectory: null` for Collections if this candidate DTO is
adopted. Do not return thousands of file paths in a gallery detail response;
the browser obtains an Album's file list from File Service `/ls/`.

4.2 Collection references: explicit route type, shared implementation

|Method  |Route                              |Body               |
|--------|-----------------------------------|-------------------|
|POST  |/api/collections/{id}/files      |{ "path": "..." }|
|DELETE|/api/collections/{id}/files      |{ "path": "..." }|
|POST  |/api/collections/{id}/directories|{ "path": "..." }|
|DELETE|/api/collections/{id}/directories|{ "path": "..." }|

Implement all four through one internal service, e.g. modifyCollectionReference(galleryId, action, referenceType, path, caller). Type-check that target is a COLLECTION; owner-only mutation (unless more granular sharing is explicitly authorized). Idempotent add/remove; unique association. Removal never deletes files, no automatic full tag scan unless explicitly requested. Choose one consistent delete representation; if clients/proxies reject DELETE bodies, support a documented path/ID-based alternative while keeping explicit files/directories segments.

4.3 Sharing

Proposed:

• GET /api/galleries/{id}/shares
• PUT /api/galleries/{id}/shares/{userId} (idempotent add)
• DELETE /api/galleries/{id}/shares/{userId} (idempotent remove)

Owner-only membership mutations. Validate user IDs against existing identity convention; optionally verify user existence if identity API supports it. Detail response must include shared members for callers permitted to view that list; avoid exposing unnecessary identities to public/anonymous viewers.

4.4 User profile and settings

Existing profile and appearance settings are handled by the shared browser
profile components and File Service `/f/profile/ui` and `/f/profile/avatar`
contracts. Do not add the proposed `/api/profile` or `/api/settings` routes for
those existing needs. A future Moments-only setting needs a named schema,
scope, and precedence rule before it becomes an endpoint. Login/logout already
exist and must be reused.

4.5 Thumbnails and wall rendering

This section is a cover-storage proposal. Its image upload routes and opaque
generation JSON are not part of the current Moments API. Adopt only the
descriptor and storage operations required by the agreed frontend workflow.

• Owner updates gallery’s thumbnail descriptor and opaque wallRenderStrategy via gallery PATCH or dedicated metadata endpoint (choose existing style).
• PUT /api/galleries/{id}/thumbnail/image — authenticated owner uploads explicitly selected/custom/generated image using bounded multipart or binary body. Store in Moments-owned thumbnails directory with backend-issued safe name; update asset reference atomically.
• GET /api/galleries/{id}/thumbnail/image — stream only owned thumbnail asset with correct media type, caching headers, authorization matched to gallery visibility; never allow generic path-based retrieval.
• DELETE /api/galleries/{id}/thumbnail/image — remove explicitly stored asset, with clearly defined descriptor fallback/reset behavior.
• JSON generation configurations (collage layout, selected files, transforms, seed, source media) are generated and interpreted only by frontend; backend validates JSON syntax, size/depth, persists, and returns unchanged. A saved rendered preview is optional; no backend collage, resize, EXIF processing, or wall rendering.
• Candidate modes: NONE, GENERATED_LAYOUT, GALLERY_MEDIA, UPLOADED_IMAGE (proposed), with optional rendered asset. The thumbnail generation string mentioned in initial design may be plain string or structured JSON per current frontend; preserve compatibility.
• Ensure thumbnails do not become a bypass to serve unauthorized source images. Do not use a backend-provided image URL as an unrestricted filesystem read.

5. File Service integration and media scans

5.1 File Service listing and new extraction contract

File Service remains the authority for physical file ACL. It already exposes
`GET /ls/{absolute-path...}`: `?recursive&files-only` returns authorized regular
files below an album directory as one flat JSON array. Each entry contains
`path`, `name`, `kind`, and nullable `size`, `owner`, `group`, `permissions`,
and `last_modified`. The browser will call this route directly for album media
loading and compose the full list from the returned paths. The response is not
paged, and File Service currently materializes it in memory. Bounded original
image requests and DOM virtualization do not bound this listing response;
measure the resulting cost on target hardware and define a separate contract
if it proves too large. The existing `/ls/` contract does not return media
type, dimensions, EXIF/location tags, or Live Photo association. Define a new
File Service operation that extracts actual media type and media tags for
authorized files. Prefer NATS for the internal scan call if it can carry the
initiating user's JWT securely; HTTP is an acceptable fallback after contract
review. File Service must validate that JWT and apply its normal RBAC checks
for each extraction request. Do not guess the new request or response DTO.
If NATS is chosen, define the request and response in shared Go contracts,
validate every handler input with the repository's non-HTTP validator
convention, and grant only the required subjects to Moments and File Service.
NATS transport identity authenticates the calling service; the user JWT
remains the source of file authorization.

5.2 Initiate and monitor asynchronous scan

On-demand full rescans are required for Albums and Collections. Moments starts
bounded work in an internal Go goroutine and reports durable progress through
an authenticated job API. The route names below are proposals; their DTOs,
limits, and failure policy still need an agreed contract. Existing File
Service `/ls/` alone cannot produce media tags.

|Method  |Suggested route            |Behavior                                               |
|--------|---------------------------|-------------------------------------------------------|
|POST  |/api/galleries/{id}/scans|Owner initiates FULL scan; return 202 with job ID/URL|
|GET   |/api/galleries/{id}/scans|Authorized job history, paginated                      |
|GET   |/api/scans/{scanId}      |Authorized status/counters/error summary               |
|DELETE|/api/scans/{scanId}      |Authorized cancellation (if implemented)               |

GET /api/scans/{id} example:

{
"id": "scan-123",
"galleryId": "gallery-123",
"status": "RUNNING",
"totalItems": 500,
"processedItems": 340,
"successfulItems": 332,
"failedItems": 8,
"createdAt": "2026-10-25T10:00:00Z",
"startedAt": "2026-10-25T10:00:01Z",
"finishedAt": null
}

During directory enumeration, totalItems may be null; UI should indicate an indeterminate discovery phase. Counts invariant: processed = successful + failed for attempted items; never exceed known total. Frontend polls with bounded interval and stops on terminal status. Server enforces queue size/per-user and per-gallery concurrent scan limits, deadlines, retry/backoff, cancellation where offered, and rate limiting.

5.3 FULL replacement semantics (required)

Every requested scan fully rebuilds that gallery’s scan-derived metadata; no incremental/only-unscanned mode required.

1. Authorize caller’s gallery management action.
2. Snapshot/resolve Album recursive directory or Collection file/directory references through File Service with the initiating user's authority; deduplicate files in a job, even with nested/overlapping directories.
3. Extract actual media type and media tags through the new File Service NATS or HTTP operation, always enforcing the initiating user's file-level authorization. Run bounded work in a Moments Go goroutine and persist progress without excessive SQLite writes.
4. Normalize/deduplicate tag definitions and media-type values, building gallery-scoped file/tag associations as the complete replacement set.
5. Atomically replace only target gallery’s active scan-derived tag associations in a short SQLite transaction after extraction; retain old tags until successful handover. Delete orphaned unused tag definitions only if safe and efficient.
6. Mark completion and expose updated aggregates. No File Service calls while holding SQLite write locks.

Failure policy: A scan with inaccessible files/failed extraction must not silently wipe all previously good metadata. Proposed default: if any errors, mark PARTIAL and publish only if the resulting set meets a documented explicit full-replacement policy; safer baseline is retain previous active set on partial/failed scans and report errors. If the product owner expects successful items to replace old tags despite failures, document and test that different policy before deployment. If zero legitimately accessible files are discovered without errors, a successful scan may atomically clear old associations.

Concurrent/race policy: one active scan per gallery, and last committed generation cannot be overwritten by stale jobs. Handle gallery deletion and sharing changes while jobs run; abort or reauthorize prior to commit as needed. Never share scan results across galleries without respecting privacy and file-level ACL.

5.4 User token lifecycle (mandatory security gate)

Use the current user access JWT when the scan starts. Keep it only in the
Moments process memory. A progress request may replace that job's JWT only
after Moments validates the new JWT and confirms that its user ID matches the
initiating user. A different user may not renew or inspect the job. File
Service validates the presented JWT through Auth and applies RBAC to each
extraction call; the Moments service token or NATS connection identity alone
must never stand in for the user.

• Do not persist plaintext access or refresh tokens in SQLite, job JSON, logs, or error traces. The job never receives or rotates the browser refresh token.
• Do not continue extraction with an expired JWT. Define bounded pause/failure behavior when progress polling does not supply a valid replacement before expiry; fail interrupted jobs after restart because their in-memory user credential is gone.
• Revalidate Moments gallery authorization at initiation, renewal, and commit. Treat File Service authorization failures as safe file/job errors without credential leakage.
• Existing Auth `auth.rpc.token.validate` verifies the signed JWT and expiry. Logout revokes the refresh token but does not immediately revoke an issued access JWT; this scan flow inherits that limit until Auth's revocation model changes.

6. Gallery content refresh vs metadata rescan

The frontend uploads original files through File Service `POST /f/upload` or its
resumable upload protocol. For an Album it refreshes the displayed file list by
calling `/ls/<uploadDirectory>?recursive&files-only` directly and adapting the
returned absolute paths. Refresh does not automatically run a Moments metadata
scan. Explicit full scanning repopulates extracted media-type and tag
associations. For Collections, explicit reference add/remove affects membership
immediately; scan-derived tag aggregates update on next full scan unless an
agreed invalidation scheme removes stale tags from removed refs. Removal must
not leave removed-file tags indefinitely presented as current.

7. Access-control policy and security boundary

• Existing authentication is authoritative; stable user UID from validated token/session, not request-provided UID. Never implement duplicate passwords/login/logout.
• Owner may read/create/update/delete their own gallery, shares, refs, config, thumbnails, scans.
• Explicitly shared users have read-only gallery access by default. No edit/share/delete/scan initiation absent an explicit later product rule.
• Public discoverability: policy confirmation required—public to authenticated liteNAS users vs anonymous internet users. Secure implementation default pending decision: public readable by authenticated users only. Do not silently open unauthenticated routes.
• Gallery-level authorization and File Service file-level ACL are separate. Moments visibility and sharing change only app metadata access and never call File Service to change Unix permissions. A user who may see a shared or public gallery is not automatically authorized by File Service or Stream Service to read its originals; the frontend must tolerate denied media. File permissions may be managed separately by operators or other tools.
• Apply authorization across any implemented list, detail, reference, search/filter, aggregate tag, share, thumbnail, and scan endpoint. Existing File Service profile routes enforce their own authorization. Ensure an unauthorized caller learns neither private existence nor tag counts or member identities.
• Owner authorization must survive ownership checks for gallery mutations. Never allow arbitrary owner reassignment by PATCH without an explicit, separately designed transfer flow.

Input validation and injection resistance

Centralized typed validation for every query param, route/path param, request body, multipart part, and security-relevant header. Enforce sizes, enum membership, number ranges, permitted arrays, valid JSON syntax/depth, text length, and file format. Normalize reference paths according to File Service contract, rather than using unsafe ad hoc path joins. Do not blanket-ban shell punctuation in legitimate filenames; instead eliminate unsafe interpretation.

• SQL: use parameterized queries; whitelist SQL identifier/sort selection; enforce uniqueness/FKs.
• Shell/command: never pass request data to shell interpreters or string-concatenated commands; prefer native libraries, subprocess argv only if justified.
• Path traversal: backend-issued thumbnail names, fixed configured storage root, symlink-safe open/write operations, prevent traversal/TOCTOU and path substitution; reference paths are inert metadata unless passed safely to File Service.
• XSS/code: names/descriptions/tags and frontend layout JSON are untrusted; safe escaping/rendering on frontend; no JSON evaluation as code.
• Upload: enforce size/type/actual signature limits, bounded memory, safe overwrite and atomic rename, no arbitrary filenames or directories, appropriate private filesystem modes, ensure content not active HTML/SVG unless explicitly sanitized and supported.
• API hardening: request timeouts, scan worker concurrency/rate limits, payload caps, no credentials in logs, safe CORS/CSRF handling based on actual auth mechanism, no private data in public caches, stable error format with request ID, security logging without secrets.

8. Configuration, storage, and operations

The installed INI file is `/etc/lite-nas/moments.conf`. The current nonroot
`lite-nas-moments` systemd unit writes only its log, and the existing AppArmor
profile has no database or thumbnail write grant. The proposed nested config
path above is incompatible with the installed service. Other LiteNAS stateful
services use `/var/lib/lite-nas/<service>/`; select and configure a Moments
state path using the same deployment convention before schema work. Add only
needed storage paths to systemd `ReadWritePaths` and AppArmor. SQLite's DB,
journal/WAL/SHM, and optional thumbnail assets need provisioned ownership,
migrations, and backup behavior. Keep the existing single-package deployment
workflow and loopback listener.

Linux systemd unit

• The dedicated unprivileged `lite-nas-moments` user/group and hardened systemd unit already exist. Provision only new service-owned storage through the existing deployment workflow.
• Preserve existing `User`, `Group`, `NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp`, `PrivateDevices`, and capability restrictions; extend only justified writable paths.
• Config read-only, DB+thumbnail root writable, other application/system data inaccessible; existing auth socket/cert requirements must be explicitly allowed and reviewed.
• Restart/shutdown handling: stop accepting work, cancel/mark in-flight jobs correctly, release SQLite connections, and fail stale/in-memory-credential jobs after restart.
• Health/readiness endpoint if established deployment conventions require it.

AppArmor

Extend the existing dedicated Moments AppArmor profile with only the needed
SQLite DB/WAL/SHM and optional owned-thumbnail permissions. Test startup,
migrations, reads, scans, and restart with the profile enforced. Do not grant
direct NAS traversal to Moments or broad `/var/**` write access. The existing
service unit already limits egress to loopback addresses.

9. Frontend integration and layout inspection

The frontend implements four wall layouts named `mosaic`, `justified`,
`masonry`, and `grid`. Its demo gallery consumes only a bundled fixture; its
file/folder picker uses `/ls/` for selection but does not supply gallery items.
Add direct `/ls/<album directory>?recursive&files-only` loading, validate the
File Service response, and map absolute host paths into existing local view
items. Replace fixture metadata with agreed Moments gallery metadata where in
scope. Keep the existing layouts and bounded media request coordinator.

Frontend touchpoints for the adopted feature scope:

• Gallery list: accessible-union pagination, search/filter, owner/type/media types/location and visibility.
• Gallery detail: agreed metadata and album uploadDirectory; optional sharing, thumbnail, layout, and tags only after their contracts are settled.
• Album media: browser calls File Service `/ls/` with `recursive` and `files-only`; it composes the full visible list from the returned files. Neither Moments nor the current File Service supplies paged media responses for this flow.
• Collection editor: explicit /files and /directories add/remove calls; preserve references across navigation.
• Album upload: use File Service (existing), then refresh displayed gallery. Do not upload original media through Moments.
• Metadata scan: start scan, poll dedicated progress endpoint, show discovering/running/success/partial/failure and retry; do not block upload flow.
• Thumbnails: custom upload via Moments; selected gallery image or generation layout handled by frontend; no rendering logic in backend.
• Profile/appearance: continue using shared File Service `/f/profile/ui` and `/f/profile/avatar`; define only missing gallery-specific settings in Moments.
• Permission-aware UI: only display mutating controls for owners; still enforce server-side.

10. Testing strategy: existing Go/TypeScript tests and installed-system HyperionTF tests

Use Go package tests for Moments and File Service units, the existing
TypeScript frontend test conventions for browser adapters, and top-level
`tests/` HyperionTF tests only for an installed or running LiteNAS system.
Follow `tests/README.md` and its exact `infra`, `cli`, `api`, and `ui`
categories. Do not add service-local contract tests under top-level `tests/`.

Tests must be black-box at integration/system level (public REST or running service behavior), with reusable API clients, fixtures, predictable identities, isolated SQLite databases/storage roots, deterministic seeded file trees, clean teardown, explicit condition polling instead of arbitrary sleeps, test markers by level, and actionable diagnostic artifacts.

Required scenarios

1. Create ALBUM/COLLECTION, reject invalid type-specific fields, update, get complete details, delete only metadata (not originals).
2. Collection file and directory add/remove endpoints share implementation; duplicate and nonexistent refs, mixed case/path encoding, owner vs forbidden user.
3. Recursive Album discovery; Collection recursive directory expansion; overlapping paths deduplicate, nested dirs; File Service 401/403/not found/timeouts.
4. Gallery list default union of own/public/shared without duplicates; multi-value owner/location/media type/type filters with OR/AND; search, sorting, paging, privacy non-leakage.
5. Shared members CRUD if adopted; policy matrix for own/shared/other (and anonymous if explicitly enabled); existing File Service profile behavior remains in place.
6. Thumbnail upload, retrieval, replacement and deletion; JSON config round-trip; wall layout strategy persistence; invalid MIME/signature/size/path attempts.
7. Full media metadata scan: create gallery + seed files + scan + poll discovery/progress + verify tag normalization/dedup + rescan with changed metadata + verify atomic replacement; preserve other galleries’ associations.
8. Scan interruption, partial failure, token expiry, restart, duplicate start, cancellation/race, gallery deletion or permission revoked during scan; no stale generation overriding newer one.
9. Input fuzz/negative tests: SQL injection patterns, command injection strings, path traversal/symlink, invalid enum/ID/arrays/JSON, excessive payload/depth, HTML/XSS-like strings as inert values.
10. Installed systemd service: nonroot UID, permissions on config/storage/SQLite, functional auth and File Service calls, enforcement of AppArmor profile and recorded denied unauthorized reads/writes; service restart and migrations.

Mock File Service only for scoped service failure simulations; installed-system
coverage must exercise the real File Service and its authorization. Cover the
selected feature scope rather than adding tests for proposed endpoints that
are not adopted.

11. Implementation phases and dependency order

Phase 0 — Repository audit and decision gate
Document actual backend framework, auth/UID and token lifetime/delegation, File Service routes/ACL, frontend layout identifiers/DTOs, .conf conventions, service files/AppArmor patterns, installed migrations/tests. Resolve §16 decisions and concrete DTOs. Record evidence before irreversible schema changes.

Phase 1 — Persistence/domain model
SQLite schema/migrations; galleries, references, shares, tags/associations, thumbnail descriptor, user data, scan jobs; repositories, constraints, indices, validations and unit tests.

Phase 2 — Base REST interface
Gallery CRUD/list/details/filtering and agreed collection references, typed DTOs and OpenAPI, authz coverage, integration tests. Add shares only after policy is decided; reuse existing File Service profile routes.

Phase 3 — Thumbnails and configuration
Opaque JSON persistence, owner-controlled image upload/download, atomic safe filesystem storage, .conf runtime paths, frontend contract alignment and negative tests.

Phase 4 — Required extraction and scan orchestration
Use existing authenticated `/ls/` for discovery. Define the new File Service media-tag operation, preferably as a user-JWT-bound NATS RPC with shared contracts and narrow NATS permissions. Add bounded Moments Go jobs, same-user JWT replacement on progress requests, snapshot and full replacement, counters/progress, expiry/restart behavior, and system tests against actual File Service.

Phase 5 — Frontend refinements
Bind gallery metadata to Moments and album contents directly to File Service `/ls/` with both query flags. Adapt the full authorized listing into the local gallery model, refresh after File Service upload, and add explicit Album/Collection full-rescan progress plus agreed cover and gallery-setting flows. Reuse demo wall strategies; test success, errors, and restricted access.

Phase 6 — Deployment and hardening
Systemd unit, AppArmor enforce mode, permissions, production .conf defaults, migration/backups, health checks/logging, API and system security suite, documentation.

12. Candidate acceptance criteria for the adopted scope

The checklist below preserves the wider discussion. Full rescans of media type
and actual extracted tags are required; their File Service DTO and credential
handoff still need definition. Stored thumbnail assets, named location tags,
and anonymous visibility apply only after their contracts and policy are
agreed. Direct album `/ls/` loading and existing transport boundaries apply
to the core work.

☐ A single unified gallery model supports both ALBUM and COLLECTION, with correct required fields and no original-file operations in Moments.
☐ Gallery details supply frontend-operable metadata including type, owner, upload directory (album), aggregate tags, sharing as authorized, thumbnail and wall strategy.
☐ Default gallery list is union of owned/public/shared with no duplicates; backend name search and repeated multi-value ownership/media/location filters function and do not leak private galleries.
☐ Collection file/directory add/remove endpoints are distinct on URL and use one parameterized internal implementation.
☐ Ownership/visibility/sharing permission matrix is enforced on all endpoints, including job progress, tags and thumbnail reads.
☐ Shared profile/appearance behavior continues through File Service; any new gallery-specific setting has a documented scope and contract.
☐ Frontend-owned wall render and thumbnail-generation JSON survive API round trips unchanged; existing demo layouts remain functional.
☐ User-uploaded standalone/generated thumbnail image can be safely stored and served from configured Moments-owned directory, never arbitrary filesystem path.
☐ Browser album loading uses existing authorized recursive `/ls/` with `files-only`; any separate media extraction contract preserves File Service ACL checks, and Moments gains no filesystem bypass.
☐ On-demand scan jobs report progress and fully replace gallery-scoped tag associations atomically, with unique normalized metadata entities, and do not erase other galleries’ tags.
☐ Token expiry/restart/partial scan and concurrent runs are safe and tested; no plaintext user bearer tokens persisted or logged.
☐ SQLite migrations, FKs, indexes, WAL behavior, .conf configuration, and portable service storage work on fresh and upgrade deployments.
☐ Dedicated low-privilege systemd unit and enforcing, denial-tested AppArmor profile run operational workflows successfully.
☐ All HTTP inputs use typed, bounded validation; SQL/shell/path/XSS injection regression tests pass.
☐ Go/TypeScript unit and integration coverage and installed-system HyperionTF suites pass for adopted functionality, including real File Service authorization.
☐ API documentation, configuration sample, migration instructions, security policy decisions, deployment/uninstall instructions, and known limitations are committed.

13. Design choices and rationale

|Decision                                               |Rationale / status                                                             |
|-------------------------------------------------------|-------------------------------------------------------------------------------|
|Unified galleries entity with discriminator          |Discussed model; confirm with final API contract                               |
|Albums directory-backed, Collections reference-backed  |Discussed model; Album directory is required for agreed `/ls/` loading         |
|File Service handles original media and file ACL       |Confirmed; avoid duplicate access logic                                        |
|Browser `/ls/` uses shared cookie; backend scan uses user JWT|Agreed; same-user progress requests may replace the job's in-memory JWT|
|Full scan replaces media type and tags for target gallery|Required on success; atomic handover prevents partial loss                |
|Unique tags + gallery-scoped associations              |Discussed model; scoped associations prevent cross-gallery corruption          |
|Frontend owns rendering JSON, Moments stores opaquely  |Proposed; no persisted descriptor contract exists                              |
|Moments-owned thumbnail assets                         |Optional filesystem exception requiring agreed need and safe storage           |
|SQLite metadata proposal, installed `/etc/lite-nas/moments.conf`, storage path TBD|Config path verified; choose state path using LiteNAS convention before implementation|
|Least privilege + AppArmor + strict validation         |Confirmed                                                                      |
|Go/TypeScript tests and installed-system HyperionTF    |Matches repository testing split                                                |
|Repeated query parameters / OR-within AND-between      |Proposed agreed direction; verify old frontend compatibility                   |
|Read-only sharing and public requires auth             |**Conservative defaults pending owner policy confirmation**                    |

14. Operational and security peculiarities

• PUBLIC gallery metadata can become discoverable without implying a File Service grant on referenced files.
• Existing File Service recursive listing skips symlinks and returns one fully materialized array. Measure its size and latency for target albums; no direct OS traversal in Moments.
• Uploaded thumbnail storage is a distinct trust boundary from original media. A user-controlled display path cannot become a write target.
• SQLite is single-writer: use short transactions, bounded worker write frequency, busy timeout, and safe crash recovery.
• Shared-media identity is not necessarily global across users; scoping prevents unauthorized cross-user derived information.
• Metadata scans should not erase current valid tags at job start; only swap on a successful, explicitly defined completion policy.
• Progress counters are snapshots and may be unknown during discovery; polling should not imply synchronous processing.
• Restart recovery must be honest about jobs that cannot resume with expired/in-memory credentials.
• Do not log full Authorization headers, arbitrary EXIF/location data, or hidden private filenames at broad log levels.

15. Required deliverables

• Agreed Moments schema, migrations, REST DTOs, OpenAPI, and documentation.
• Browser `/ls/?recursive&files-only` adapter and File Service upload/refresh integration using existing routes.
• New user-authorized File Service media-tag operation (NATS preferred), with a shared contract, Auth/RBAC checks, and narrow NATS permissions or reviewed HTTP integration.
• Required Moments Go scan runner, progress storage, same-user in-memory JWT replacement, atomic full tag replacement, and documented expiry/restart/failure policy.
• Optional cover asset storage only after a descriptor and storage contract are agreed.
• Provisioned state paths and narrow updates to the existing .conf, systemd unit, and AppArmor profile where needed.
• Go/TypeScript tests and installed-system HyperionTF coverage for adopted functionality.
• Decision log for remaining product policies and implementation deviations.

16. Repository-verification gates / unresolved product policy

These are not excuses to stop implementation; inspect code and use documented secure defaults. Escalate only where a safe choice cannot be made without product policy or auth infrastructure changes.

1. Public gallery scope: authenticated users only (default) vs anonymous read. Anonymous must be explicitly approved; if enabled, ensure routes/data and thumbnail access are safe and privacy-preserving.
2. Shared user privileges: read-only (default) vs additional actions; scan initiation owner-only by default.
3. User JWT lifecycle: reuse the current JWT at initiation; on a progress request, replace it only after verifying the same initiating user ID. Define bounded expiry/pause behavior and fail interrupted jobs after restart. Never persist plaintext tokens or use service identity as the file-access subject.
4. Scan PARTIAL behavior: retain prior active metadata (default) or commit successfully extracted subset; choose deliberately and test. Successful rescans fully replace one gallery's media-type and extracted-tag associations.
5. File identity/path namespace: determine uniqueness and canonicalization from real File Service; do not expose global tag cache across tenants/users.
6. File Service traversal is implemented as unpaged `/ls/` with `recursive` and `files-only`; media extraction is absent. Define a new authorized extraction operation for actual media type and tags, preferably over NATS; named-location source and durable Live Photo association require explicit source support.
7. Frontend wall strategies are `mosaic`, `justified`, `masonry`, and `grid`; thumbnail maker JSON is browser-only. Define any persisted cover descriptor, source references, and missing-file policy separately.
8. Collection DELETE body compatibility: ensure infrastructure supports it; if not, use typed ref ID in URL or encoded path in query while preserving explicit file/directory routes and shared implementation.
9. Existing backend uses Go, Huma, `/api/auth/*`, and `/etc/lite-nas/moments.conf`; preserve those conventions. New route/schema names still need an agreed contract.
10. Tag freshness after membership changes: choose synchronous removal of now-unreachable tag associations or an explicit stale-until-rescan contract. Preferred: clear removed-membership-derived tags so filters aren’t wrong, without automatically full-rescanning.
11. Shared profile and appearance already use File Service `/f/profile/ui` and `/f/profile/avatar`; define only unmet gallery-specific setting keys and scope.
12. Uploaded thumbnail read visibility and cache policy: protect private assets and invalidate cached assets on visibility changes; never expose private file via static unguarded web-server directory.

────────
