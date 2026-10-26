# Moments App — Remaining Frontend Work

## Current state

The authenticated app shell, embedded and standalone mounts, shared profile and
appearance settings, fixture-backed gallery, and browser-only thumbnail maker
are implemented. The gallery includes virtualized mosaic, justified, masonry,
and grid layouts; bounded image requests; a modal viewer; video playback;
filename-based Live Photo pairing for the demo; downloads; EXIF details; and
slideshow controls. These are existing behavior to preserve, not missing work.
The demo loads `public/dev-fixtures/gallery.json` and converts its `/f/` URLs to
local presentation items. It does not consume a Moments media response or call
File Service `/ls/` for gallery contents. The header's shared file/folder
selection dialogs do use File Service, but currently only display the selected
path; they do not populate the gallery.
See the [app README](README.md),
[implemented requirements](../../requirements/moments-skeleton.md), and
[Moments architecture](../../docs/moments-architecture.md). The implemented
transport contracts are in [File Service requirements](../../services/file-service/requirements.md)
and the [Stream Service README](../../services/stream-service/README.md).

## Remaining work

1. **Agree on the metadata contract first.** Coordinate with
   [the service task](../../services/moments/task.md) to settle validated
   request/response schemas, stable media identity, authorization behavior,
   metadata ordering, errors, and File/Stream URL construction. Keep runtime
   Zod validation at API/provider boundaries. Do not derive production API DTOs
   from examples in the design documents.
2. **Connect actual content discovery.** For each album, call File Service
   `GET /ls/{album-directory}?recursive&files-only` directly from the browser,
   validate its full response, and compose the gallery's complete visible
   file list from the returned absolute paths. The listing has basic
   filesystem metadata but no gallery IDs, dimensions, EXIF tags, Live Photo
   links, or pagination. Define deterministic client ordering and report
   listing failures separately from an empty album. Treat omitted entries as
   outside the caller's File Service visibility, even when gallery metadata is
   visible. Preserve current wall
   and viewer behavior, including layout switching and bounded loading. Handle
   missing or inaccessible files and stale references without breaking the
   rest of the gallery. Keep the fixture available only for the agreed demo or
   development use. The full listing response and item list occupy memory;
   virtualized rendering and bounded media requests do not bound their size.
   Measure this cost on target hardware and document a practical limit or
   future File Service paging need if it becomes unacceptable.
3. **Add album and collection workflows.** Implement the agreed metadata
   operations: create, rename/update, remove, and curated collection reference
   editing after its path and permission contract is defined. Removing metadata or a
   collection reference must not delete source files. Add Moments visibility
   and sharing controls with the agreed public audience and member rights.
   Enforce owner/shared affordances in the UI while relying on the service
   for authorization.
4. **Connect refresh, upload, and rescans.** Upload original files
   through File Service's `POST /f/upload` destination-first multipart route
   or its existing resumable `/f/proprietary/upload` protocol. Define filename
   selection, overwrite/collision behavior, and progress using those actual
   contracts. After upload, repeat the album's direct `/ls/` request; do not
   treat that as an
   implicit metadata rescan. Add an explicit full-rescan action for albums and
   collections. Poll the Moments job progress with the existing authenticated
   request flow, showing discovery, extraction counts, completion, and failure.
   The authenticated progress request may supply a current access JWT to the
   running job only after Moments verifies the same initiating user ID. A
   different user cannot renew or take over the job, and an expired JWT
   cannot authorize further extraction.
5. **Add tags, filters, and covers.** Display and filter by the service's
   extracted media-type and actual media tags after a successful rescan.
   Distinguish stale or unavailable tag data from an empty tag set. Persist a
   cover only using the agreed versioned
   descriptor and stable media references or an explicitly supported
   app-owned image upload. The current thumbnail maker's exported recipe,
   local uploads, and data URLs remain browser-only unless a separate
   compatible contract is agreed.
6. **Persist gallery-specific preferences.** Use the existing shared profile
   settings contract for general appearance preferences. Persist wall layout
   or album-specific choices only after the service contract defines scope,
   precedence, and defaults. Map to the four existing wall layout identifiers;
   do not use proposed strategy names from the gallery reference.
7. **Validate the existing gallery under realistic use.** Measure wall
   scrolling, media requests, browser-owned memory, and NAS responsiveness on
   representative large collections and constrained networks. Check keyboard
   navigation, focus restoration, reduced motion, playback cleanup, and error
   recovery across wall, viewer, video, and Live Photo interactions.

## Boundaries and conflict resolution

The [gallery task](../gallery_taks.md) is a design/performance reference, not a
second implementation checklist. Its numeric limits and proposed structures
are starting points unless verified against current code and measurements.
The [backend design discussion](../../services/moments/design-discussion.md)
contains proposed routes, models, and policy options; use the service task and
the verified LiteNAS contracts as the integration authority.

- File Service owns source-file authorization, listing, upload, and download;
  Stream Service owns playback and media fallbacks. The browser-facing paths
  are `/ls/` for listings, `/f/` for original reads and upload, `/d/` for
  attachment downloads, and `/s/` for original streaming, `/s/preview/` for
  HEIC/HEIF JPEG previews, and `/s/transcode/` for compatibility playback.
  Each file or stream request still requires the caller's authorization.
  Moments must not send original media through its API.
- Moments visibility and sharing are virtual application permissions. Showing
  a gallery to another user does not change its files' Unix ACLs or File
  Service access. A visible gallery can contain media that the viewer cannot
  read; handle those media failures without changing File Service policy.
- The fixture's same-folder filename Live Photo pairing is demo behavior, not
  a durable identity contract. Use a backend association only after its source
  and semantics are defined.
- The demo's local layout state and thumbnail recipes are not persisted API
  formats. Keep current local behavior until compatible persistence is
  specified.
- Do not add profile/settings endpoints, remote thumbnail storage, or extra
  collection behavior solely because a
  source design document proposes them. Include them only when the LiteNAS
  contract and product workflow are agreed.

## Completion criteria

- The app uses validated Moments metadata for album/collection workflows and
  a validated direct File Service recursive listing for production album media
  instead of relying on the fixture.
- Uploads and media access continue through File and Stream Service, with
  refresh and errors presented clearly.
- Album and collection rescans have explicit progress and successful scans
  replace the displayed media-type and extracted-tag metadata for that gallery.
- Existing wall, viewer, thumbnail maker, authentication, and appearance
  behavior remain intact unless an agreed contract requires a focused change.
- Gallery request and memory bounds, accessibility, and lifecycle behavior
  have been checked against representative data and documented results.
