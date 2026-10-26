# Moments Gallery — Design Reference and Remaining Work

## High-Performance Virtualized Mixed-Media Gallery

**Version:** 2.1 — Integrated with LiteNAS planning
**Status:** Fixture gallery implemented; `/ls/` gallery discovery, metadata integration, and scale validation remain
**Technology:** React, TypeScript, Material UI, Motion for React  
**Architecture:** Bounded media loading, virtualized views, original-media-first transport
**Primary constraint:** NAS-friendly performance with bounded media requests

---

This task describes the intended Moments gallery frontend and user experience.
Read it with the [LiteNAS Moments architecture](../docs/moments-architecture.md),
the [Moments app task](moments/task.md), and the
[Moments service task](../services/moments/task.md). The latter documents backend
ownership and decisions still required. TypeScript examples below are proposed
view-local models, not implemented HTTP DTOs or approved API routes.
The implemented transport contracts are in [File Service requirements](../services/file-service/requirements.md)
and the [Stream Service README](../services/stream-service/README.md).

The fixture-backed gallery demo is implemented. It reads authorized media
through File and Stream Service, including optional HEIC/HEIF previews and video
transcode fallback. Its item list comes only from bundled
`public/dev-fixtures/gallery.json`; there is no gallery `/ls/` request, Moments
media response adapter, album loading, or stable backend media ID yet. Shared
file/folder dialogs elsewhere in the app use `/ls/` for selection but do not
feed the gallery. The detailed sections below preserve original-media-first
design goals and candidate structures; they are not a checklist of missing
components or a statement of the current transport contract. The
[app README](moments/README.md) describes current behavior. Albums, persisted
layout settings, tags, and stored covers require the metadata API described in
the [Moments service task](../services/moments/task.md).

## 1. Project Overview

Implement a reusable, high-performance mixed-media gallery in an existing React application.

The gallery must support large collections of original media files stored on a resource-constrained NAS.

Supported gallery content includes:

- Photographs
- Videos
- Apple Live Photos
- Mixed image and video collections
- Portrait, landscape, and square media

Typical gallery characteristics:

- Approximately 1,000 media items.
- Average original image size around 10 MB.
- Approximately 10 GB of potential image data per gallery.
- Files stored on a ZFS pool attached to a Raspberry Pi.
- File access routed through Cloudflare, Nginx, and a file service.
- Storage also used for CIFS/SMB and NFS shares.

The gallery must not overwhelm the storage infrastructure or unnecessarily consume browser memory.

## 1.1 Infrastructure

Typical delivery path:

```text
React Browser Application
         |
         v
      Cloudflare
         |
         v
        Nginx
         |
         v
    File Service
         |
         v
   ZFS Storage Pool
    Raspberry Pi
```

The diagram shows the original-file route. The deployed Nginx mounts also
expose File Service `/ls/` for authorized listings, `/d/` for downloads, and
Stream Service `/s/` for original playback and format fallbacks. Moments
metadata APIs described below are planned; the current gallery uses a bundled
fixture.

**Planned Moments metadata backend (contract undecided):**

- Maintains gallery metadata.
- Returns gallery configuration.
- Provides album/source metadata; the browser uses File Service `/ls/` for
  album media. A collection-specific reference response needs its own contract.
- Must not imply that a gallery reference grants read access to its file.
- Does not process or transform media.

**File service:**

- Exposes existing original files.
- Returns authorized `/ls/` entries with `path`, `name`, `kind`, and nullable
  size, owner, group, permissions, and modification time. `recursive` and
  `files-only` are presence-based flags; the response is one JSON array with
  no pagination, media dimensions, EXIF tags, or Live Photo association.
- Reads files directly from the NAS.
- Does not produce derived media.
- Does not generate previews or thumbnails.
- Does not transcode videos.

The gallery remains original-first and does not require persistent derived
files. Stream Service can generate a HEIC/HEIF preview or transcode fallback
on request when the browser cannot display the original.

---

## 2. Fundamental Infrastructure Constraints

These are original-first design constraints. The current Stream Service
fallback routes are documented in the [app README](moments/README.md).

## 2.1 Original Files Only

The frontend requests original media first when the browser can display it.
Current gallery classification covers JPG/JPEG, PNG, WebP, GIF, AVIF,
HEIC/HEIF, MP4, MOV, WebM, M4V, AVI, and MKV. Classification does not promise
browser decoding: HEIC/HEIF images use Stream Service JPEG preview, and a
video can use the explicit transcode fallback after native playback fails.
Files outside the current gallery allowlist remain available through File
Service but are not gallery media items. Adding TIFF or another source format
requires a separate browser rendering decision.

There must be no assumption that an optimized browser-friendly variant exists.

## 2.2 One Primary URL Per Item

Every gallery item has one primary download URL.

Do not require:

- `thumbnailUrl`
- `previewUrl`
- `posterUrl`
- `optimizedUrl`
- `mobileUrl`
- Alternative resolution URLs

Use only established File and Stream Service media endpoints.

Apple Live Photos may require a logical association between an original still photograph and an original
motion file. Such an association must use existing file references or metadata relationships, not generated
preview files.

## 2.3 No Server-Side Media Processing

The baseline gallery must not require persistent derivatives or background
processing. Optional Stream Service previews and transcode fallback are
available for browser compatibility. The baseline does not require:

- Thumbnail generation.
- Image resizing.
- Preview generation.
- Image transcoding.
- Video transcoding.
- Video poster generation.
- Background media conversion.
- Sidecar metadata files.
- Preview caches stored alongside media.

The Raspberry Pi is not intended to operate as a media-processing server.

## 2.4 NAS Storage Protection

The ZFS pool is shared with other services, including SMB/CIFS and NFS.

The gallery must minimize unnecessary storage reads.

Requirements:

- Strictly bounded image download concurrency.
- No eager downloading of entire galleries.
- No mass loading of offscreen originals.
- No speculative loading of hundreds of images.
- No uncontrolled video probing.
- Avoid duplicate requests.
- Prioritize currently needed media.
- Avoid aggressive retry loops.
- Allow conservative concurrency settings.

**Protecting NAS responsiveness is a first-class performance objective.**

## 2.5 No Additional NAS Files

Do not create thousands of preview files, metadata sidecars, or derived assets on the NAS.

The database may retain gallery metadata, including image dimensions, but this must not require creating one
filesystem metadata file per media item.

---

## 3. Primary Objectives

## 3.1 Browser Performance

The implementation must:

- Support thousands of gallery records.
- Render only a bounded subset of wall media.
- Keep a bounded carousel neighborhood.
- Control active image downloads.
- Avoid unnecessary decoded-image retention.
- Avoid application-level Blob caching of all media.
- Prevent uncontrolled React rerenders.
- Avoid expensive full-gallery DOM operations.
- Maintain responsive scrolling.
- Preserve scroll position while switching views.

## 3.2 Network and Storage Performance

The implementation must:

- Use original media URLs directly.
- Take advantage of browser HTTP caching where available.
- Avoid downloading unused files.
- Deduplicate in-flight image requests.
- Provide configurable transport concurrency.
- Coordinate wall and carousel requests.
- Deprioritize inactive views.
- Minimize unnecessary requests reaching the ZFS pool.

## 3.3 User Experience

The implementation must provide:

- Progressive initial rendering.
- Responsive, intelligent gallery arrangements.
- Mosaic and justified layouts.
- Configurable virtualization.
- Loading skeletons.
- Optional bottom-up tile assembly animations.
- Smooth scrolling.
- A modal carousel.
- Image/video/Live Photo support.
- Video autoplay.
- Configurable slideshow auto-next.
- Graceful unsupported-format handling.
- Consistent MUI styling.

---

## 4. Technology and Implementation Standards

Use:

- React
- TypeScript
- Material UI
- Motion for React
- Native `HTMLImageElement`
- Native HTML5 video APIs
- ResizeObserver
- IntersectionObserver where appropriate
- Browser HTTP caching

Prefer the existing application's dependencies and conventions.

Do not introduce another state-management library unless justified.

## 4.1 MUI Requirement

Use MUI for nearly all presentation.

Preferred components:

- Box
- Stack
- Typography
- Skeleton
- Dialog
- Modal
- IconButton
- Tooltip
- CircularProgress
- Slider
- Switch
- ToggleButtonGroup

Native `<img>` and `<video>` elements are appropriate for media rendering.

Avoid raw HTML layout elements when MUI equivalents are available.

Use the existing theme, responsive breakpoints, typography, and palette.

---

## 5. High-Level Architecture

The implemented gallery composes three principal contexts:

1. `GalleryContext`
2. `WallContext`
3. `CarouselContext`

The existing `GalleryImageQueue`, `GalleryWindow`, `GalleryLayout`, and media
components cover request coordination, virtualization, layout, and playback.
The design examples below do not require splitting these into new layers.

## 5.1 Context Hierarchy

```tsx
<GalleryProvider>
  <WallProvider>
    <CarouselProvider>
      <GalleryWall />
      <GalleryViewer />
    </CarouselProvider>
  </WallProvider>
</GalleryProvider>
```

The carousel can be conditionally mounted.

The wall must remain mounted while the carousel is open.

The MUI modal may use a React portal without losing context access.

## 5.2 Separation of Responsibilities

**GalleryContext:** What items exist?

**WallContext:** Which items should appear on the wall, where, and when?

**CarouselContext:** Which media item is selected, which nearby items are needed, and how playback/navigation behaves?

**GalleryImageQueue:** How image requests are bounded and marked ready across
wall and viewer.

**GalleryWindow and GalleryLayout:** How visible tiles and geometry are
calculated.

**Presentation components:** How state is rendered.

These boundaries must remain explicit.

---

## 6. Data and Transport Contracts

There is no implemented Moments gallery DTO. The current fixture contains
`defaultLayout` and an ordered array of `/f/` file URLs. The browser derives
index-prefixed item IDs, filenames, and media kind from those URLs; those IDs
are valid only for that fixture presentation. The fixture also applies an
unambiguous same-folder HEIC/HEIF and MOV filename pairing for Live Photo
motion. None of these rules defines persistent backend identity.

## 6.1 File Service listing

`GET /ls/{absolute-host-path...}` returns only entries the authenticated user
may list. `?recursive&files-only` enables both flags by presence. Each entry
has an absolute host `path`, `name`, `kind` (`file` or `directory`), and nullable
`size`, `owner`, `group`, `permissions`, and `last_modified`. The response is a
single JSON array; it has no pagination or media-specific fields. Denied
entries are omitted and an authorization dependency failure fails the listing.
Validate the response before adapting paths to local gallery items. For the
agreed album flow, call `/ls/<album-directory>?recursive&files-only` once,
derive any nested grouping from each absolute `path`, and compose the full
ordered frontend list. A large recursive response is currently materialized in
full by File Service and held in browser memory; measure this separately from
the bounded image queue and virtualized wall.

## 6.2 Media access

The same absolute host file maps to `/f/<path>` for an inline original read,
`/d/<path>` for an attachment download, and `/s/<path>` for authorized original
playback. Stream Service also exposes `/s/preview/<path>` for an HEIC/HEIF JPEG
preview and `/s/transcode/<path>?start=<seconds>` for an explicit compatibility
stream. The current `StreamContext` derives these routes from validated `/f/`
URLs. Every request is separately authenticated and authorized; metadata or a
gallery membership reference is not a file access grant.

## 6.3 Planned Moments metadata

The Moments API must define album/collection identity, ownership, source
references, filters, ordering, and settings before the gallery consumes it.
For albums, the browser obtains the full file list directly from File Service
`/ls/`; Moments does not need to return media pages for that workflow. If dimensions or
associations become available from an agreed extraction contract, they can
improve layout and Live Photo handling. Until then, keep fallback geometry and
present unassociated files safely. Scrolling must not trigger repeated source
metadata extraction or require sidecar files.

An explicit on-demand Moments rescan will use a new user-authorized File
Service media-tag endpoint, preferably over NATS. Its internal Go job will
rebuild the selected album's or collection's actual media-type and extracted
tag associations and expose progress. The initiating user's access JWT stays
in Moments process memory; progress polling can replace it only with a valid
JWT for the same user. A normal `/ls/` refresh updates the
visible file list only; it does not implicitly start extraction. The gallery
must distinguish tags from the last successful scan from newly listed files
that have not been rescanned.

Moments visibility and sharing govern which gallery metadata appears in the
app; they do not change Unix ACLs or File Service authorization. A user may
see a shared or public gallery while `/ls/`, `/f/`, `/d/`, or `/s/` denies some
source files. The UI must present that state safely without treating a Moments
membership as a file grant.

The four implemented layout identifiers are `mosaic`, `justified`, `masonry`,
and `grid`. Layout selection is currently local page state. Persist only an
agreed gallery-specific preference using an agreed Moments contract; the shared
light/dark appearance preference already uses File Service `/f/profile/ui`.
Viewport dimensions and request concurrency remain runtime values.

---

## 7. GalleryContext

GalleryContext owns:

- Gallery identity.
- Complete ordered item list.
- Backend configuration.
- Shared image resource registry.
- Shared request coordinator.

It must not own:

- Wall scrolling.
- Wall window rotation.
- Carousel navigation.
- Slideshow timers.
- Wall geometry.
- Layout animations.

GalleryContext should expose stable metadata and service references.

Avoid rerendering the entire context consumer tree whenever one image finishes loading.

Use granular subscriptions, stable external stores, or equivalent techniques where appropriate.

---

## 8. Image Loading Design Reference

The interfaces and class names in this section predate the implemented
`GalleryImageQueue`. Keep them as performance behavior examples; do not add
parallel loader, registry, or coordinator classes solely to match these names.

Implement a reusable, headless image-loading service.

This component is independent of the wall and carousel.

## 8.1 Responsibilities

The loader must:

1. Accept image URLs.
2. Create native Image objects.
3. Queue requests.
4. Respect configured concurrency.
5. Track request states.
6. Track individual image readiness.
7. Track batch completion.
8. Notify consumers of success or failure.
9. Deduplicate equivalent in-flight requests.
10. Release loader-owned temporary bookkeeping when no longer needed.

The loader must not:

- Track scrolling.
- Manage carousel navigation.
- Calculate virtual windows.
- Calculate geometry.
- Decide when the next wall group is needed.
- Trigger slideshow navigation.
- Manage HTML5 video playback.
- Remove wall/carousel components.

## 8.2 Request State

```text
idle
  |
  v
queued
  |
  v
loading
  |
  +------> ready
  |
  +------> error
```

Optional cancellation/abandonment states may be introduced.

## 8.3 TypeScript Contract

```typescript
type ImageLoadStatus =
  | "idle"
  | "queued"
  | "loading"
  | "ready"
  | "error";

interface ImageRequest {
  id: string;
  url: string;
  priority?: number;
}

interface ImageLoadResult {
  id: string;
  status: "ready" | "error";
  image?: HTMLImageElement;
  error?: unknown;
}

interface ImageBatchResult {
  results: ImageLoadResult[];
  completed: number;
  failed: number;
  total: number;
}

interface ImageLazyLoader {
  load(
    requests: ImageRequest[]
  ): Promise<ImageBatchResult>;
}
```

This is an illustrative contract.

The implementation should support progressive individual notifications or subscriptions in addition to batch completion.

## 8.4 Concurrency vs Batch Size

These are different parameters.

Example:

```typescript
const loadingConfig = {
  maxConcurrent: 5,
  batchSize: 10,
};
```

The loader may receive 10 requests while allowing only 5 downloads simultaneously.

The remaining requests stay queued.

As downloads settle, available concurrency slots are reused within the pending group.

The group is complete when all entries have either succeeded or failed.

One failed image must not block the group indefinitely.

## 8.5 Batch Completion Ownership

The LazyLoader reports that the current group has settled.

It does not automatically decide which group is next.

That decision belongs to:

- WallContext for wall requests.
- CarouselContext for carousel requests.

## 8.6 Readiness

The loader tracks whether an image is ready for presentation.

Possible implementation:

```typescript
const image = new Image();

image.src = url;

await image.decode();
```

The implementation must handle browser decoding failures and loading errors appropriately.

`decode()` can be part of the readiness strategy but must not cause already usable images to become permanently
stuck due to recoverable decoding/API behavior.

Track readiness by stable resource identity and request generation.

## 8.7 Resource Release

Releasing loader-owned references does not mean discarding wall or carousel resources.

Separate the following:

- Loader request bookkeeping.
- Shared resource state.
- Active consumer references.
- React component lifetime.
- Browser HTTP cache.
- Browser decoded image cache.

A completed group may be released from the loader's bookkeeping while its images remain displayed.

The wall and carousel independently control their own resource rotation.

Do not retain completed Image objects indefinitely just because they were once loaded.

---

## 9. Shared Image Request Coordinator

Wall and carousel share one global image download budget.

Do not create independent unrestricted queues.

Example:

```typescript
interface CoordinatorConfig {
  maxConcurrent: number;
}
```

## 9.1 Priority Rules

When the wall is active:

1. Images required for the current viewport.
2. Images near the viewport.
3. Additional wall overscan resources.

When the carousel opens:

1. Currently selected carousel image.
2. Immediate carousel neighbors.
3. More distant carousel neighbors.
4. Wall requests deferred until the carousel closes.

WallContext and CarouselContext decide which resources need which priorities.

The transport coordinator merely enforces those supplied priorities.

## 9.2 Fast Navigation

If the user navigates rapidly:

- Reprioritize requests.
- Remove obsolete queued requests.
- Reuse in-flight requests where applicable.
- Prevent obsolete callbacks from changing the active slide.
- Avoid unnecessary NAS reads.

Do not assume an already-started image download can always be cancelled immediately.

---

## 10. Browser Cache and Network Policies

The frontend must rely primarily on standard browser HTTP caching.

Image resources loaded through `new Image()` can use the browser's caching mechanisms.

However, previously loading a URL does not guarantee it remains cached.

## 10.1 Recommended Server Headers

Configure caching at Nginx, Cloudflare, or the file service where appropriate.

Prefer:

- Stable original file URLs.
- Correct Content-Type.
- Suitable Cache-Control.
- ETag or Last-Modified where supported.
- Conditional requests when useful.

Use long-lived immutable caching only when the URL changes if file contents change.

## 10.2 Frontend Restrictions

Do not:

- Download all originals into an in-memory Map.
- Store all files as Blobs.
- Build a custom persistent cache without a demonstrated need.
- Assume browser cache retention is guaranteed.
- Re-download originals simply to create local thumbnails.

Use browser-managed resources wherever practical.

---

## 11. WallContext

WallContext controls:

- Virtualization.
- Scrolling.
- Window rotation.
- Overscan.
- Loading triggers.
- Requesting media images.
- Layout selection.
- Planner invocation.
- Scroll anchoring.
- Initial-loading animation state.
- Resource release when tiles leave the window.

It must not implement actual network transport.

## 11.1 Virtual Window

Initial suggested configuration:

```typescript
const wallConfig = {
  windowSize: 50,
  overscanBefore: 30,
  overscanAfter: 30,
};
```

The exact configuration must be adjustable.

The nominal window is 50 items, with up to 30 extra items on either side.

The active materialized count is therefore typically bounded near 110, subject to boundaries and
planner-specific grouping.

## 11.2 Initial Load

When the gallery is opened:

1. Load metadata.
2. Read layout configuration.
3. Determine available wall width.
4. Calculate initial geometry.
5. Materialize initial virtual items.
6. Display placeholders.
7. Request image groups using the shared loader.
8. Reveal loaded media progressively.

On a large desktop viewport, multiple loading groups may be required before the visible area is fully populated.

The wall should support this without eagerly loading hundreds of originals.

## 11.3 Incremental Loading

As scrolling approaches the window boundary:

- Determine upcoming items.
- Materialize the new segment.
- Request missing resources.
- Retire items outside the configured window.
- Maintain scroll position.

Use planned item geometry rather than assuming all tiles occupy equal heights.

## 11.4 Rotation

The wall must support bidirectional rotation.

Do not permanently append every item to the DOM.

After continued scrolling, old items must be unmounted.

The implementation must preserve stable scroll geometry using absolute placement, calculated offsets, spacers,
or a similarly correct approach.

## 11.5 Carousel Interaction

While the carousel is open:

- Wall remains mounted.
- Scroll position remains stable.
- Background scrolling is locked.
- New wall downloads are paused or deprioritized.
- Existing active resources are not automatically discarded.

When carousel closes:

- Wall becomes interactive.
- Loading resumes based on the current wall window.
- Scroll position is preserved.

---

## 12. Dedicated Wall Layout Planner

The wall uses an independent headless planner.

The planner supports:

1. Mosaic layout.
2. Justified layout.

Layout mode comes from gallery metadata stored in the database.

## 12.1 Planner Interface

```typescript
interface TileLayout {
  id: string;
  index: number;

  x: number;
  y: number;

  width: number;
  height: number;
}

interface LayoutConfig {
  targetSize: number;
  gap: number;
  allowCropping: boolean;
}

interface LayoutStrategy {
  calculate(
    items: readonly GalleryItem[],
    containerWidth: number,
    config: LayoutConfig
  ): TileLayout[];
}
```

The planner must be:

- Headless.
- Deterministic.
- Independent of React.
- Independent of image loading.
- Responsive.
- Unit-testable.

Geometry must be calculated from metadata and container dimensions.

Image readiness must not trigger unnecessary geometry recalculation.

---

## 13. Intelligent Mosaic Layout

Implement a responsive, intelligently packed quilted gallery.

This is not a fixed two-row layout.

Desired visual characteristics:

- Asymmetric composition.
- Portrait and landscape media interleaved.
- Tall portraits may span multiple rows.
- Wide landscapes may span multiple columns.
- Square media fills smaller spaces.
- Minimal unused gaps.
- Balanced tile sizes.
- Controlled cropping when enabled.
- Responsive base dimensions.

Potential tile spans:

```text
1 x 1
2 x 1
1 x 2
2 x 2
```

Additional spans may be used when appropriate.

## 13.1 Intelligent Planning

The planner should select placements based on:

- Source aspect ratio.
- Available row/column space.
- Cropping penalty.
- Space utilization.
- Balance across the layout.
- Stable ordering.

The goal is visually optimized packing without unpredictable rearrangement.

The planner may reorganize locally for better fit, but should avoid arbitrary global shuffling that changes on rerender.

## 13.2 Cropping

When cropping is enabled:

```css
object-fit: cover;
object-position: center;
```

Cropping should remain visually reasonable.

When cropping is disabled, prefer dimensions that preserve the media's natural ratio.

The original file must never be resized or transformed server-side to achieve a tile shape.

CSS scaling and cropping happen in the browser.

---

## 14. Justified Layout

Implement a justified photo gallery.

Behavior:

- Images retain natural aspect ratios where practical.
- Each completed row fills available width.
- Row height varies around a configured target.
- Row item count is calculated dynamically.
- Spacing remains consistent.
- The final incomplete row can remain partially filled.
- No fixed number of images per row.

The layout should work with portrait, landscape, and square media.

The same virtualization and loader must work in both layout modes.

Changing layout mode should not require downloading media again solely because the geometry changed.

---

## 15. Handling Missing Dimensions

Because the backend is thin, source dimensions may occasionally be unavailable.

The wall must handle this gracefully.

Recommended approach:

1. Use known dimensions when provided.
2. Otherwise assign a deterministic placeholder aspect ratio.
3. Load the original only when the item becomes eligible.
4. Read actual natural dimensions after decoding if available.
5. Update the necessary geometry conservatively.

Do not trigger a full-gallery rearrangement for a single dimension correction.

Preserve scroll anchoring during geometry changes.

Dimension discovery should not trigger background downloads for all 1,000 files.

---

## 16. Wall Loading UX and Tetris-Inspired Animation

Implement a visually engaging initial-loading experience.

The intended effect resembles Tetris assembly, except images appear to rise upward into their positions rather
than fall from above.

This effect is primarily intended for situations where the gallery does not yet have sufficient ready media to
populate the viewport and its initial buffer.

## 16.1 Bootstrap Mode

When the initial viewport is not sufficiently populated:

- Show skeleton placeholders.
- Download controlled image groups.
- Reveal images progressively.
- Animate new tiles upward into position.
- Use subtle fading.
- Optionally stagger newly completed groups.

Suggested animation:

```typescript
const bootstrapAnimation = {
  translateY: 35,
  durationMs: 300,
  staggerMs: 35,
};
```

Values must be tunable.

## 16.2 Steady-State Mode

When a sufficient buffer is available:

- Stop special assembly animation.
- Continue ordinary virtualization.
- Preserve stable positions.
- Reveal newly ready images without major reflow.
- Avoid repeatedly animating previously shown tiles.

## 16.3 Recovery Mode

When the user jumps into a region without available media:

- Display placeholders.
- Prioritize required resources.
- Optionally reactivate restrained assembly animations.
- Return to steady state when the buffer recovers.

## 16.4 Animation Rules

Use Motion for React.

Animations should primarily affect opacity and transform.

Avoid expensive whole-grid rearrangements on every image completion.

Respect `prefers-reduced-motion`.

Do not allow animation to degrade scrolling performance.

---

## 17. Mixed-Media Wall Representation

The wall contains images, videos, and Live Photos.

Presentation must work with original media constraints; HEIC/HEIF preview and
video transcode fallback are available through Stream Service when needed.

## 17.1 Images

For browser-supported image formats:

- Render using the original file URL.
- Respect calculated tile geometry.
- Show a skeleton while loading.
- Apply fit/crop rules.
- Show a fallback on error.

CSS downscaling does not reduce transferred file size.

Therefore, lazy loading and concurrency control are mandatory.

## 17.2 Videos

Video items must not automatically download full video data to populate wall tiles.

If no poster image exists:

- Show a styled placeholder.
- Show a video/play icon.
- Optionally show the filename or known duration.
- Preserve planned tile dimensions.

Do not create 50 active video elements to generate thumbnail frames.

Avoid uncontrolled metadata probing and video frame extraction.

## 17.3 Live Photos

Use the original still image when a still-image resource is explicitly available through the existing metadata model.

Show a Live Photo indicator.

Do not download the motion component simply because the tile enters the wall.

If a Live Photo association cannot be resolved, degrade gracefully to an individually supported image, video,
or file placeholder.

---

## 18. Original File Format Compatibility

The NAS may contain formats not universally supported by browsers.

The frontend must detect and handle unsupported rendering gracefully.

Typical considerations:

| Format | Expected support |
|---|---|
| JPEG | Broad |
| PNG | Broad |
| HEIC/HEIF | Browser-dependent |
| TIFF | Limited/browser-dependent |
| MP4 | Codec-dependent |
| MOV | Container and codec-dependent |
| AVI | Generally limited |

Do not assume a MIME extension guarantees that the browser can decode the file.

## 18.1 Unsupported Media UX

When media cannot be presented:

- Show a meaningful file icon.
- Display filename or format when available.
- Provide an option to open or download the original.
- Avoid infinite loading indicators.
- Avoid aggressive retries.

Do not automatically implement client-side transcoding of multi-gigabyte originals.

Do not require server transcoding as a fallback.

## 18.2 MIME and Browser APIs

Use file metadata and native browser capabilities as appropriate.

Browser support detection must not require eagerly downloading every original.

---

## 19. CarouselContext

CarouselContext owns:

- Current media index.
- Nearby virtual range.
- Navigation operations.
- Carousel loading requests.
- Playback lifecycle.
- Slideshow timers.
- Media error handling.
- Modal close behavior.

CarouselContext does not own the wall's scroll position or layout.

## 19.1 Initial Handoff

When opening the carousel:

1. Capture the selected wall item.
2. Determine nearby items.
3. Inspect shared resource availability.
4. Reuse loaded Image references when valid.
5. Reuse active in-flight requests.
6. Request missing nearby image resources.
7. Initialize independent navigation state.

No image binary payloads should be copied.

## 19.2 Carousel Window

Default:

```typescript
const carouselWindow = {
  before: 10,
  after: 10,
};
```

This creates a logical neighborhood of at most 21 items.

The number of actual concurrent downloads remains constrained by the shared coordinator.

The logical range must not imply immediate downloading of all 21 original files.

## 19.3 Navigation

Support:

- Previous.
- Next.
- Keyboard controls.
- Touch/swipe navigation where appropriate.
- Navigation controls.
- Close modal.

When a slide changes:

- Cancel stale slideshow timers.
- Pause previous video.
- Update current index.
- Reprioritize nearby resources.
- Ignore callbacks from obsolete slides.

---

## 20. Carousel Modal UX

Use MUI Dialog or Modal.

Requirements:

- Responsive viewport-filling presentation.
- Appropriate dark/media-viewing background.
- Visible close control.
- Previous/next controls.
- Loading feedback.
- Media centering.
- Aspect-ratio-aware rendering.
- Keyboard interaction.
- Background scroll lock.
- Focus management.

The carousel should preserve the original media's visible composition.

Unlike the wall, the carousel should generally prefer `object-fit: contain` to avoid unnecessary cropping.

The original file remains the source.

---

## 21. HTML5 Video Handling

HTML5 video is separate from ImageLazyLoader.

The loader does not count video playback readiness as ordinary image readiness.

## 21.1 Video Mounting

Only active or explicitly required video elements should be mounted.

Inactive neighboring videos must not aggressively download data.

Use conservative video preload behavior.

## 21.2 Video Autoplay

When selecting a video slide:

- Attempt playback if video autoplay is enabled.
- Display loading/buffering feedback.
- Handle successful playback.
- Handle autoplay permission rejection.
- Provide manual playback controls.

Playback cannot be assumed to start automatically with audio on every browser.

## 21.3 Manual Slide Navigation

When moving away from a video:

- Pause it immediately.
- Stop associated playback handling.
- Ignore stale ended/error events.
- Prevent continued audio.
- Start the next slide's lifecycle.

## 21.4 Video Errors

On unrecoverable loading or playback error:

- Show an appropriate fallback.
- Offer original-file access.
- If slideshow auto-next is enabled, advance to the next media item.

Autoplay restrictions are not equivalent to unrecoverable media errors.

---

## 22. Slideshow Autoplay

The carousel supports two independent capabilities:

1. Video autoplay.
2. Automatic slide advancement.

## 22.1 Configuration

```typescript
interface CarouselAutoplayConfig {
  videoAutoplay: boolean;
  slideshowEnabled: boolean;
  imageDelayMs: 1000 | 3000 | 5000;
}
```

The UI must allow selecting:

- 1 second.
- 3 seconds.
- 5 seconds.

Use MUI controls.

## 22.2 Image Slides

When slideshow auto-next is enabled:

1. Wait for the active image to become ready.
2. Start the configured delay.
3. Advance when the delay expires.
4. Reset timing for the next slide.

Do not consume the display delay while the original is still downloading.

If loading fails, use a defined fallback that avoids indefinite slideshow blockage.

## 22.3 Video Slides

When slideshow auto-next is enabled:

- Ignore the 1/3/5-second image delay.
- Start playback if permitted.
- Wait for `ended`.
- Advance when video playback completes.
- Advance on unrecoverable video error.

Do not advance while normally buffering.

## 22.4 Paused Slideshow

When auto-next is disabled:

- Images remain until manual navigation.
- Videos may still autoplay if independently enabled.
- Ending a video does not automatically navigate away.

## 22.5 End-of-Gallery

Default behavior:

- Stop at the final item.
- Do not wrap unless looping is explicitly configured.

---

## 23. Apple Live Photo Handling

Live Photos are compound media, commonly consisting of a still image plus a motion resource.

The gallery metadata model must support association between original files where the backend supplies that relationship.

Do not assume a single original file universally contains both elements in a web-playable form.

## 23.1 Wall

- Present the still image when available.
- Show a Live Photo indicator.
- Avoid loading the motion file during ordinary wall scrolling.

## 23.2 Carousel

- Show the still image as the primary frame.
- Allow motion playback when the related original motion file is accessible and browser-compatible.
- Use native HTML5 video when possible.
- Consider Apple's LivePhotosKit only when necessary and compatible with the application.
- Provide fallback behavior for unsupported motion formats.

## 23.3 Slideshow Policy

Recommended initial behavior:

- Treat inactive Live Photo motion as a still image.
- Use image delay for still presentation.
- Do not interrupt explicitly active motion playback.
- Avoid automatic looping unless configured.

The exact motion interaction may be refined during implementation.

---

## 24. UI Component Structure

The current app keeps gallery files directly under `apps/moments/src/`. Its
relevant composition is:

```text
Gallery.tsx
├── GalleryContext.tsx
├── WallContext.tsx
├── CarouselContext.tsx
├── GalleryWall.tsx
└── GalleryViewer.tsx
GalleryImageQueue.ts
GalleryWindow.ts
GalleryLayout.ts
gallery-model.ts
StreamContext.tsx
```

Add a File Service listing adapter and a Moments metadata adapter near the
existing API/model code when their contracts are ready. Keep the current
component tree unless a measured issue requires a structural change.

---

## 25. Critical State Ownership Rules

1. GalleryContext owns the adapted media list and shared image request queue.

2. WallContext owns scrolling, virtualization, loading triggers, and wall presentation state.

3. CarouselContext owns navigation, media selection, and playback timing.

4. GalleryImageQueue owns bounded image transport and request readiness.

5. GalleryWindow and GalleryLayout own the existing virtualization and pure geometry calculations.

6. Presentation components must not create independent uncontrolled loading queues.

7. Releasing queue bookkeeping must not invalidate resources still required by the wall or viewer.

8. No view is permitted to eagerly download every original media file. The
   agreed `/ls/` flow does receive a complete metadata/path list.

9. No requirement may silently assume a persistent thumbnail, poster, or
   transcoded file exists. Stream Service's on-demand fallback routes are explicit.

10. Original-file compatibility limitations must be handled explicitly.

---

## 26. Edge Cases

The implementation must handle:

- Empty galleries.
- Galleries with thousands of items.
- Missing dimensions.
- Extreme aspect ratios.
- Unsupported images.
- Unsupported videos.
- Missing original files.
- Expired browser sessions.
- Large downloads.
- Slow NAS response.
- Network interruption.
- Duplicate requests.
- Rapid scrolling.
- Rapid carousel navigation.
- Carousel closing during loading.
- Browser cache eviction.
- Browser autoplay rejection.
- Video decoding failure.
- Live Photo component incompatibility.
- Viewport resizing.
- Layout changes.
- Reduced-motion preferences.
- Component unmounting during active asynchronous work.

Failures must be contained to affected media items wherever possible.

---

## 27. Performance Validation

Test using:

- 50 items.
- 1,000 items.
- 5,000 simulated metadata records.
- Mixed media collections.
- High-latency connections.
- Limited-bandwidth connections.
- Mobile browsers.
- Large desktop viewports.
- Rapid scrolling.
- Repeated carousel opening.
- Long carousel sessions.

Measure:

- Maximum concurrent image requests.
- Total unnecessary network transfers.
- First useful viewport rendering time.
- Memory behavior.
- DOM node count.
- React rerender frequency.
- Layout calculation overhead.
- Scroll responsiveness.
- Modal open/close performance.
- Browser cache reuse.
- Video buffering behavior.

Where possible, profile the impact on NAS storage responsiveness while other file-sharing workloads are active.

Performance tuning must be based on measurements, not assumptions.

---

## 28. Accessibility

Provide:

- Keyboard-operable carousel.
- Escape-to-close behavior.
- Accessible media controls.
- Correct dialog semantics.
- Focus management.
- Focus restoration.
- Touch-friendly navigation.
- Reduced-motion support.
- Meaningful media labels.
- Graceful loading and error states.

Do not make animation necessary for understanding or operating the application.

---

## 29. Remaining Gallery Work

The fixture gallery already has bounded image loading, virtualized wall layouts,
a modal viewer, video playback, Live Photo motion, EXIF display, and slideshow
controls. Keep those as implemented behavior rather than new implementation
tasks. The browser-only thumbnail maker is documented separately in the
[app README](moments/README.md).

- Replace fixture collection records with validated, authorized album/media
  data once the Moments metadata contract is defined. Add a gallery adapter
  for direct File Service `/ls/<album-directory>?recursive&files-only`; this
  integration does not exist today. Settle media identity, ordering,
  large-directory costs,
  dimensions, and Live Photo association first.
- Persist gallery layout settings through an agreed service or profile contract.
- Integrate explicit album/collection full rescan progress, extracted media
  tags, filtering, uploads, and stored cover descriptors through
  the respective app and service tasks.
- Measure scrolling, media request volume, browser memory, and NAS response on
  realistic large galleries and constrained networks. Tune based on those
  results; the design's numeric examples are starting points.
- Complete focused accessibility and lifecycle validation for the implemented
  gallery, including keyboard navigation, focus restoration, reduced motion,
  playback cleanup, and error recovery.

## 30. Remaining Acceptance Criteria

- [ ] The gallery no longer relies on the bundled fixture for production
  content; full recursive File Service listings are validated and mapped to
  the local view model. Persisted layout is covered by an agreed setting.
- [ ] An explicit Moments rescan rebuilds one album's or collection's
  extracted media-type and tag metadata through a user-authorized File Service
  operation, and the UI reports progress, completion, and failure.
- [ ] Large-gallery and constrained-network measurements confirm bounded DOM,
  media requests, and application-owned memory without unacceptable NAS impact.
- [ ] Accessibility and lifecycle behavior has been checked across wall,
  viewer, video, and Live Photo interactions.

The backend remains responsible for metadata and authorization; File Service
owns original files and downloads, and Stream Service owns media delivery.
The gallery should request media only when the user can use it and keep the
resource budget bounded.
