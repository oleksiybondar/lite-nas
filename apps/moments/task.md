# Moments App — Remaining Frontend Work

## Current state

The authenticated app shell, embedded and standalone mounts, shared profile and
appearance settings, and live gallery feed are implemented. Albums resolve
their media through File Service `/ls/`; collections use saved references.
The gallery includes virtualized mosaic, justified, masonry, and grid layouts;
bounded image requests; a modal viewer; video playback; filename-based Live
Photo pairing; downloads; EXIF details; and slideshow controls. The former
fixture gallery and standalone picker playbook pages have been removed.
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
   and viewer behavior, including the saved layout and bounded loading. Handle
   missing or inaccessible files and stale references without breaking the
   rest of the gallery. The full listing response and item list occupy memory;
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
6. **Persist gallery-specific preferences.** The owner now selects a wall
   layout for each album or collection through the existing Moments metadata
   contract. Use the shared profile settings contract for general appearance
   preferences. Define scope, precedence, and defaults before adding any other
   album-specific choices.
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
- Same-folder filename Live Photo pairing is a browser fallback, not
  a durable identity contract. Use a backend association only after its source
  and semantics are defined.
- The wall uses the saved `wallRenderStrategy` from gallery metadata. Layout
  changes belong to the owner create/edit workflow.
- Do not add profile/settings endpoints, remote thumbnail storage, or extra
  collection behavior solely because a
  source design document proposes them. Include them only when the LiteNAS
  contract and product workflow are agreed.

## Completion criteria

- The app uses validated Moments metadata for album/collection workflows and
  a validated direct File Service recursive listing for production album media
  from the configured gallery source.
- Uploads and media access continue through File and Stream Service, with
  refresh and errors presented clearly.
- Album and collection rescans have explicit progress and successful scans
  replace the displayed media-type and extracted-tag metadata for that gallery.
- Existing wall, viewer, authentication, and appearance
  behavior remain intact unless an agreed contract requires a focused change.
- Gallery request and memory bounds, accessibility, and lifecycle behavior
  have been checked against representative data and documented results.
