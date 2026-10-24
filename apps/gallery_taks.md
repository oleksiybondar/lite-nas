# Technical Task Definition

## High-Performance Virtualized Mixed-Media Gallery

**Version:** 2.0 — Consolidated specification  
**Status:** Ready for implementation  
**Technology:** React, TypeScript, Material UI, Motion for React  
**Architecture:** Context-driven, headless services, virtualized views, original-media-only transport  
**Primary constraint:** NAS-friendly performance with no server-side media processing

---

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

The backend and file service are separate responsibilities.

**Application backend:**

- Maintains gallery metadata.
- Returns gallery configuration.
- Returns the ordered list of media items.
- Provides file references/download URLs.
- Does not process or transform media.

**File service:**

- Exposes existing original files.
- Reads files directly from the NAS.
- Does not produce derived media.
- Does not generate previews or thumbnails.
- Does not transcode videos.

The gallery frontend must be designed around these limitations.

---

## 2. Fundamental Infrastructure Constraints

These constraints are mandatory.

## 2.1 Original Files Only

The frontend receives original media files.

Examples include:

- JPG/JPEG
- PNG
- TIFF
- HEIC
- HEIF
- MP4
- MOV
- AVI
- Other formats supported by the underlying storage and file service

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

Do not invent additional media endpoints.

Apple Live Photos may require a logical association between an original still photograph and an original
motion file. Such an association must use existing file references or metadata relationships, not generated
preview files.

## 2.3 No Server-Side Media Processing

The solution must not require:

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

Implement three principal contexts:

1. `GalleryContext`
2. `WallContext`
3. `CarouselContext`

Use independent headless services for:

- Image transport.
- Global request coordination.
- Resource tracking.
- Layout planning.
- Layout algorithms.
- Carousel playback behavior.

## 5.1 Context Hierarchy

```tsx
<GalleryProvider>
  <WallProvider>
    <GalleryWall />

    <CarouselProvider>
      <CarouselModal />
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

**ImageLazyLoader:** How requested images are acquired and marked ready.

**ImageRequestCoordinator:** How global network concurrency and priorities are enforced.

**ImageResourceRegistry:** How shared image request states and references are tracked.

**WallLayoutPlanner:** How wall geometry is calculated.

**Presentation components:** How state is rendered.

These boundaries must remain explicit.

---

## 6. Backend Metadata Contract

The application backend returns gallery metadata as JSON.

It is a thin metadata service, not a media-processing backend.

The JSON contains:

- Gallery identity.
- Gallery configuration.
- Ordered media items.
- Original file download URLs.
- Available descriptive and geometric metadata.

## 6.1 Gallery Item Model

```typescript
type GalleryMediaType =
  | "image"
  | "video"
  | "live-photo";

interface GalleryItem {
  id: string;

  type: GalleryMediaType;

  // Single primary original-file URL
  url: string;

  // Source dimensions, when known
  width?: number;
  height?: number;

  // Optional original file metadata
  filename?: string;
  mimeType?: string;
  fileSize?: number;
  durationMs?: number;

  // Optional relationship to another
  // existing original file
  associatedMediaId?: string;
}
```

`associatedMediaId` can be used for compound media such as Live Photos if the backend already exposes an
appropriate association.

This does not create or require a second URL on the same item.

If a Live Photo relationship is not available, the frontend must not guess unrelated files.

The backend's current DTO may be mapped into this internal representation without changing the backend API
unnecessarily.

## 6.2 Gallery Metadata

```typescript
interface GalleryMetadata {
  id: string;
  title: string;

  settings: GallerySettings;

  items: GalleryItem[];
}
```

The full media list can remain in GalleryContext.

It contains metadata, not decoded media payloads.

## 6.3 Source Dimensions

Width and height should be included when already available in the gallery database.

They are useful for calculating layout geometry before media downloads.

However:

- Do not generate preview images to obtain dimensions.
- Do not repeatedly inspect original files on every gallery request.
- Do not create filesystem sidecars.
- Do not force expensive metadata rescanning during wall scrolling.

If dimensions are unavailable, use a fallback geometry strategy.

The frontend must function without requiring backend-generated previews or additional files.

## 6.4 Backend Layout Configuration

The layout mode is stored in the database as part of gallery settings.

```typescript
type GalleryLayoutMode =
  | "mosaic"
  | "justified";

interface GallerySettings {
  wall: {
    layout: {
      mode: GalleryLayoutMode;
      targetSize: number;
      gap: number;
      allowCropping: boolean;
    };
  };

  carousel: {
    autoplay: {
      videoAutoplay: boolean;
      slideshowEnabled: boolean;
      imageDelayMs: 1000 | 3000 | 5000;
    };
  };
}
```

Only meaningful gallery-level choices need to be persisted.

Operational tuning parameters may also have application-level defaults.

Viewport width is a runtime input and must not be stored in gallery metadata.

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

## 8. ImageLazyLoader

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

Since no generated previews exist, presentation must work directly with original media constraints.

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

Suggested structure:

```text
gallery/
├── types/
│   ├── gallery.types.ts
│   ├── media.types.ts
│   ├── layout.types.ts
│   └── loader.types.ts
│
├── context/
│   ├── GalleryContext.tsx
│   ├── WallContext.tsx
│   └── CarouselContext.tsx
│
├── services/
│   ├── ImageLazyLoader.ts
│   ├── ImageRequestCoordinator.ts
│   └── ImageResourceRegistry.ts
│
├── layout/
│   ├── WallLayoutPlanner.ts
│   ├── MosaicLayoutStrategy.ts
│   └── JustifiedLayoutStrategy.ts
│
├── hooks/
│   ├── useGallery.ts
│   ├── useWall.ts
│   ├── useCarousel.ts
│   ├── useImageResource.ts
│   ├── useWallVirtualization.ts
│   └── useCarouselPlayback.ts
│
├── components/
│   ├── Gallery.tsx
│   ├── GalleryWall.tsx
│   ├── GalleryMediaTile.tsx
│   ├── GallerySkeleton.tsx
│   │
│   ├── carousel/
│   │   ├── CarouselModal.tsx
│   │   ├── CarouselContent.tsx
│   │   ├── CarouselControls.tsx
│   │   ├── CarouselSlide.tsx
│   │   ├── CarouselAutoplayControls.tsx
│   │   └── CarouselMediaRenderer.tsx
│   │
│   └── media/
│       ├── ImageRenderer.tsx
│       ├── VideoRenderer.tsx
│       ├── LivePhotoRenderer.tsx
│       └── UnsupportedMediaRenderer.tsx
│
└── tests/
    ├── loader/
    ├── layout/
    ├── virtualization/
    ├── media/
    └── carousel/
```

Adapt to existing repository conventions.

---

## 25. Critical State Ownership Rules

1. GalleryContext owns gallery metadata and shared infrastructure references.

2. WallContext owns scrolling, virtualization, loading triggers, and wall presentation state.

3. CarouselContext owns navigation, media selection, and playback timing.

4. ImageLazyLoader owns image transport, readiness tracking, and group completion.

5. ImageRequestCoordinator enforces the common network budget.

6. ImageResourceRegistry manages shared resource identity and consumer references.

7. WallLayoutPlanner owns pure geometry calculations.

8. Presentation components must not create independent uncontrolled loading queues.

9. A loader bookkeeping release must not invalidate resources still required by the wall or carousel.

10. No view is permitted to eagerly load the full gallery.

11. No requirement may silently assume a thumbnail, poster, or transcoded file exists.

12. Original-file compatibility limitations must be handled explicitly.

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
- Expired URLs.
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

## 29. Implementation Phases

## Phase 1 — Gallery Domain and Contexts

Implement:

- Gallery metadata types.
- Backend DTO adaptation.
- GalleryContext.
- WallContext.
- CarouselContext.
- Typed hooks.
- Configuration access.

## Phase 2 — Image Transport

Implement:

- ImageLazyLoader.
- Request coordinator.
- Resource registry.
- Concurrency.
- Batch completion.
- Readiness tracking.
- Deduplication.
- Cleanup.
- Unit tests.

## Phase 3 — Layout Planning

Implement:

- Planner interface.
- Mosaic strategy.
- Justified strategy.
- Responsive geometry.
- Missing-dimension fallbacks.
- Unit tests.

## Phase 4 — Virtualized Wall

Implement:

- Bounded virtual window.
- Overscan.
- Rotation.
- Scroll anchoring.
- MUI tiles.
- Skeletons.
- Controlled image acquisition.

## Phase 5 — Loading UX

Implement:

- Bootstrap animation.
- Bottom-up tile assembly.
- Steady-state behavior.
- Recovery loading state.
- Reduced-motion handling.

## Phase 6 — Carousel

Implement:

- Modal.
- Initial resource handoff.
- Independent virtual window.
- Navigation.
- Shared priorities.
- Wall preservation.

## Phase 7 — Mixed Media

Implement:

- Original-image rendering.
- Video placeholders.
- HTML5 video renderer.
- Live Photo relationships.
- Format fallbacks.

## Phase 8 — Playback and Slideshow

Implement:

- Video autoplay.
- Pause on navigation.
- Image auto-next delay.
- Video completion advance.
- Error handling.
- Timer cleanup.

## Phase 9 — Performance and Quality

Implement:

- Automated lifecycle tests.
- Large gallery testing.
- Network throttling tests.
- Browser profiling.
- NAS-load observations.
- Accessibility validation.
- Final tuning.

---

## 30. Definition of Done

The implementation is complete when:

- [ ] Gallery metadata is loaded independently of media payloads.
- [ ] The frontend uses one original-file URL per item.
- [ ] No generated thumbnail or preview endpoints are required.
- [ ] No media transformation is required from the backend.
- [ ] No filesystem sidecars are introduced.
- [ ] The image loader supports configurable batching and concurrency.
- [ ] All image requests follow the shared network budget.
- [ ] Individual image readiness is tracked.
- [ ] Loading groups settle despite individual failures.
- [ ] Image resource sharing works between wall and carousel.
- [ ] Loader cleanup does not destroy view-owned resources.
- [ ] Wall virtualization is bounded and bidirectional.
- [ ] Scroll position remains stable during rotation.
- [ ] Mosaic layout works with mixed aspect ratios.
- [ ] Justified layout fills rows dynamically.
- [ ] Layout selection is read from database metadata.
- [ ] Layout planning is independent of loading.
- [ ] Missing dimensions are handled.
- [ ] Initial loading uses the intended bottom-up assembly effect.
- [ ] Animation is restrained during normal scrolling.
- [ ] UI consistently uses MUI.
- [ ] Carousel opens over the still-mounted wall.
- [ ] Closing the carousel preserves the wall position.
- [ ] Carousel reuses available image resources where possible.
- [ ] Carousel navigation remains independent from the wall.
- [ ] Videos do not trigger mass wall-video downloads.
- [ ] Unsupported media displays a useful fallback.
- [ ] Video playback pauses on slide changes.
- [ ] Video autoplay restrictions are handled.
- [ ] Slideshow supports 1/3/5-second image delays.
- [ ] Videos advance on completion when slideshow mode is enabled.
- [ ] Live Photo still/motion relationships are handled where supported.
- [ ] No unbounded DOM or application-owned media-resource accumulation occurs.
- [ ] Performance has been evaluated on realistic constrained-network scenarios.
- [ ] Tests cover major layout, transport, navigation, and lifecycle behavior.

---

## 31. Explicit Non-Goals

Do not implement as part of this task:

- Media uploads.
- Media editing.
- Server-side image processing.
- Thumbnail generation.
- Preview generation.
- Video transcoding.
- Image transcoding.
- Automatic browser-compatible format conversion.
- Server-side poster extraction.
- Sidecar metadata generation.
- Persistent custom offline gallery storage.
- Full-gallery prefetching.
- AI-based photo analysis.
- Automated focal-point recognition.

Future support for optional optimized derivatives may be considered separately, but must not affect the
original-only requirements of this implementation.

---

## 32. Instructions for an AI Implementation Agent

When implementing this specification:

1. Inspect the repository before writing code.

2. Identify the existing React, TypeScript, MUI, routing, backend API, and dependency conventions.

3. Preserve existing application architecture where reasonable.

4. Implement the specified responsibilities without combining unrelated layers.

5. Do not introduce thumbnail or preview endpoints.

6. Do not introduce media conversion or transcoding services.

7. Do not add persistent NAS files.

8. Use only original file download URLs supplied by the backend.

9. Keep the loader reusable and unaware of wall/carousel navigation.

10. Keep layout planning independent of React rendering and image readiness.

11. Make network concurrency configurable and enforce it globally.

12. Ensure view-specific loading triggers remain inside their respective contexts.

13. Avoid retaining all downloaded media in application memory.

14. Use stable resource identities and deduplicate requests.

15. Treat resource release and view lifecycle as separate concerns.

16. Use MUI consistently for UI.

17. Prefer native media APIs for browser-supported formats.

18. Provide appropriate fallbacks for unsupported originals.

19. Implement video playback cleanup and guard against stale callbacks.

20. Maintain wall state when the carousel opens.

21. Keep layout algorithms deterministic and testable.

22. Do not create unnecessary React context updates.

23. Preserve scrolling behavior over decorative animations.

24. Respect reduced-motion settings.

25. Add tests for concurrency, failures, geometry, and lifecycle transitions.

26. Avoid premature abstractions that have no clear responsibility.

27. Document any deviation from this specification.

28. Do not silently substitute conventional media-gallery assumptions for the NAS-specific constraints.

29. Prefer conservative resource usage when performance choices are ambiguous.

30. Deliver a fully integrated and tested implementation rather than isolated demonstration components.

---

## 33. Final Architectural Principles

The design must maintain the following relationships:

**Gallery metadata describes the media collection.**

**The backend remains thin and does not process original files.**

**The NAS remains the authoritative source of original media.**

**WallContext determines what belongs in the virtualized wall and when loading should be requested.**

**CarouselContext determines what belongs in the carousel and how navigation/playback behaves.**

**The layout planner determines the visual arrangement of wall tiles.**

**The LazyLoader performs image acquisition and readiness tracking.**

**The shared request coordinator protects network and storage resources.**

**The resource registry supports reuse without duplicating large payloads.**

**MUI components present the resulting state.**

**Animations improve perceived loading quality without introducing excessive rendering or storage pressure.**

The finished system must provide a modern, responsive, visually attractive gallery experience while respecting
the operational constraints of original media stored on a Raspberry Pi/ZFS NAS.

**The primary design philosophy is controlled demand, bounded resources, clear ownership, and predictable
behavior—not eager media processing or aggressive prefetching.**
