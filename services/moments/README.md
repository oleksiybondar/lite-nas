# Moments metadata service

Moments is the loopback Go API for gallery metadata. It reuses the existing
LiteNAS login, refresh, logout, and current-user routes. Original media and
filesystem permissions remain with File Service; Stream Service serves browser
playback and previews. The Moments database contains paths and derived metadata,
not original media bytes or user tokens.

## API and Swagger UI

On the standalone Moments host, open `/api/docs` for the bundled Swagger UI or
`/api/openapi.json` for Huma's generated contract. File Service provides
`/f/docs` and `/f/openapi.json` when `api_docs.enabled=true`; Stream Service
provides `/s/docs` and `/s/openapi.json`. All three UIs use local assets and a
same-origin schema. File Service docs require an authenticated user.

| Route | Purpose |
| --- | --- |
| `POST /api/auth/login?redirect_to=...` | Issue session cookies and echo a validated app-local path in `data.redirect_to`; unsafe paths resolve to `/`. The browser performs navigation. |
| `GET /api/config` | Return the authenticated user's configured `workingDir` used to start file and folder pickers. |
| `POST /api/galleries` | Create an `ALBUM` or `COLLECTION`. Albums require an authorized, existing `uploadDirectory`. |
| `GET /api/galleries` | List accessible albums and collections together, newest first. `page` defaults to 1; `size` defaults to 20 and is capped at 500. The response has `items` and `metadata` (`page`, `size`, `total_count`, `total_pages`). Repeated `ownerId`, `type`, `visibility`, `mediaType`, and `location` filters use OR within a field and AND across fields. Use `type=ALBUM` or `type=COLLECTION` for a narrower search. `q` searches name, description, user tags, and every published scan tag. |
| `GET/PATCH/DELETE /api/galleries/{galleryId}` | Read, owner-edit, or delete gallery metadata. Deletion never deletes original files. |
| `GET /api/collections/{galleryId}/references` | List explicit file and directory references. |
| `POST/DELETE /api/collections/{galleryId}/files` | Add or remove `{ "path": "/absolute/path" }`. |
| `POST/DELETE /api/collections/{galleryId}/directories` | Add or remove a directory reference with the same body shape. |
| `GET /api/galleries/{galleryId}/shares` | Owner-visible member IDs. |
| `POST /api/galleries/{galleryId}/shares` | Owner adds a member by `{ "username": "oleksiybondar" }`; Moments resolves the login through RBAC and stores its stable UID. |
| `PUT/DELETE /api/galleries/{galleryId}/shares/{userId}` | Owner-controlled read-only sharing. |
| `PUT/GET/DELETE /api/galleries/{galleryId}/thumbnail/image` | Store, serve, or remove an optional uploaded PNG/JPEG cover. PUT accepts `image/png`, `image/jpeg`, or `application/octet-stream`. |
| `POST/GET /api/galleries/{galleryId}/scans` | Start a full rescan or page the owner's job history. |
| `GET/DELETE /api/scans/{scanId}` | Same-user progress polling with token renewal, or cancellation. |
| `GET /api/galleries/{galleryId}/media-metadata` | Page the last successfully published scan records. |
| `GET /api/galleries/{galleryId}/tags` | Return both user tags and aggregated scan tags. Each tag has `type=USER` or `type=AUTO`, a `category`, `value`, and `count`. For user tags, `category=USER` and `count=1`; for auto tags, count is the number of media records carrying that tag. |
| `GET/POST /api/galleries/{galleryId}/user-tags` | List user tags or create one with `{ "value": "Family" }`. Reads follow gallery visibility; creation requires the owner. |
| `GET/PATCH/DELETE /api/galleries/{galleryId}/user-tags/{tagId}` | Read, rename with `{ "value": "Friends" }`, or remove one user tag. Reads follow gallery visibility; only the owner can mutate. Values are unique per gallery after case folding and Unicode normalization. |

Huma validates typed route and query inputs. Gallery IDs are UUIDs. Paths use
File Service's clean absolute Unix path namespace. At creation or reference
addition, Moments asks File Service `/f/stat/<path>` to resolve the canonical
path, check its kind, and apply the initiating user's Auth/RBAC read permission.
The canonical path is the scan identity; a move creates a new identity on the
next scan.
The browser continues to load an album's files directly from
`GET /ls/<uploadDirectory>?recursive&files-only`; Moments does not proxy the
full album file listing.

Owners may set `wallRenderStrategy` to `mosaic`, `justified`, `masonry`, or
`grid` on creation or PATCH. Moments stores the choice per album or collection,
returns it in gallery metadata, and defaults to `mosaic` when omitted.

Every gallery operation requires a valid LiteNAS access token. The owner may
mutate metadata. Explicit members have read-only metadata access. `PUBLIC`
means readable by **authenticated** LiteNAS users; it never grants Unix file
access. New albums and collections default to `PUBLIC` unless creation
specifies `PRIVATE`. Unknown and hidden private galleries both return 404. Only the owner
can see shared-member IDs and scan jobs. The service does not store a second
profile or appearance settings document; those remain with File Service.

Gallery create, detail, patch, and list responses include a `tags` array with
both tag types. User tags belong to the gallery itself and survive rescans and
source-directory changes. Scan tags are read-only and track the currently
published media; rescan publication replaces only those media associations.

`thumbnailPath` is an optional clean absolute image path on create or PATCH.
Moments checks its read permission and image class with File Service, then
stores the canonical path. Clients render it through File Service or Stream
Service. `thumbnailGeneration` is an optional JSON value on create or PATCH,
limited to 16 MiB and stored unchanged. It may contain frontend-generated data
URLs. An uploaded image is a separate optional field, exposed as
`thumbnailImageUrl`. Uploading a cover clears `thumbnailPath`; clients show a
selected path first, then an uploaded image, then the first accessible gallery
image. Uploaded images are limited to 16 MiB, PNG/JPEG, 8192 pixels on either
side, and 50 million pixels. An upload never reads an original file path
through Moments.

## File Service media operations

File Service adds `GET /f/stat/<absolute-host-path>` and
`GET /f/metadata/<absolute-host-path>` (internal paths `/stat/...` and
`/metadata/...`). Both authenticate every request and repeat the normal scope
and RBAC read checks. The metadata response contains a media class, a safe
content type, dimensions when the built-in image decoder supports the file,
and source-backed tags. The current extractor reads JPEG EXIF description,
camera make, camera model, and original capture timestamp. It returns an empty
tag array when these fields are absent. It does not infer named locations from
coordinates or filenames. Format classification is derived from the file
extension; callers must still handle malformed or unsupported source content.
The scan uses File Service's loopback HTTP API with the initiating user's
Bearer token because that existing adapter already enforces caller-specific
ACL checks; service identity alone never grants source access.

## Full rescan

A scan resolves album or collection sources through File Service using the
initiating user's access JWT. It deduplicates overlapping paths, extracts each
file through File Service, and builds media-type, extension-based `FILE_TYPE`,
and source-backed tags.
Same-directory, case-sensitive basename pairing between an image and MOV marks
the still as `LIVE_PHOTO` and records its motion path; it never rewrites either
original. Scans are on demand; uploading or refreshing `/ls/` does not start one.

At most two scans run per process, one per gallery. A scan accepts at most
10,000 files and has a 20-minute deadline. Progress is durable in SQLite. The
access JWT is kept only in process memory. A same-user progress request can
replace it after session validation. Expiry without a replacement fails the
scan; restart marks interrupted jobs failed. A successful scan replaces only
that gallery's published media and tags in one SQLite transaction. A partial
scan leaves previous published results active and reports `PARTIAL`. A cancelled
scan also leaves previous results active. No token is persisted or logged.
Current Auth behavior keeps an issued access JWT valid until its expiry after
logout; cancelling a running scan is available through the scan DELETE route.
Changing an album source directory or adding/removing a collection reference
clears that gallery's published scan results; the owner starts a new scan to
rebuild them. A scan whose source metadata changes before publication fails
without restoring stale results.

## Storage and operations

The installed configuration is `/etc/lite-nas/moments.conf`. The database
is `/var/lib/lite-nas/moments/moments.db`, and uploaded covers are under
`/var/lib/lite-nas/moments/thumbnails`. Deployment provisions these directories
for the nonroot `lite-nas-moments` account and updates the systemd and AppArmor
write scope. The `[storage] working_dir` setting is a clean absolute host path; it defaults to
`/home` and can be set to a media root such as `/NASPool/Photos`. It controls the
picker's initial directory and does not change File Service authorization.

SQLite foreign keys, WAL, and a busy timeout are enabled. The
schema is created idempotently on startup; there was no earlier Moments
metadata schema to migrate.

Back up the database using SQLite's online backup API or a consistent
`VACUUM INTO` snapshot; copying only the main file while WAL has pending writes
is incomplete. Back up the thumbnails directory with the database so stored
asset names remain valid. Restore both while the Moments service is stopped,
then keep ownership with `lite-nas-moments`. Original files require their own
File Service/storage backup policy.
