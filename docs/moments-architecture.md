# Moments Architecture

Moments is LiteNAS's personal photo and media application. It is a deliberate
family-use product slice alongside the platform's security and administration
work; the gallery is not a security-skeleton feature. The first useful target
is a gallery demo over existing, authorized files. Album management, tags,
uploads, and persistent gallery settings can follow through the separate
Moments service work.

## Current platform boundary

- The deployed Moments app and loopback Go service currently provide browser
  authentication, an empty home screen, profile settings, and embedded and
  standalone hosting. They do not yet provide album or gallery APIs.
- File Service already supplies authenticated `/ls/`, `/f/`, and `/d/` routes
  for listing, reading, downloading, and uploading files. RBAC and File Service
  remain responsible for path authorization; the browser does not scan NAS
  paths directly.
- Stream Service and its `/s/` playback route are planned for range delivery
  and Live Photo handling. Gallery design must not treat them as available in
  the initial demo.
- The Moments Go service is the planned owner of album identities, tags,
  pagination, SQLite metadata, layout preferences, and cover descriptors. Its
  [service task](../services/moments/task.md) requires explicit API decisions
  before those contracts are implemented.

## First gallery demo

Use the existing Moments login and the established File Service routes to
display a user-authorized directory as a responsive media wall and modal
viewer. The browser may adapt validated File Service listing records into a
view-local gallery model. Keep that adapter specific to the documented File
Service contract; do not introduce a speculative Moments API, DTO, or route.
The demo may use original file URLs for browser-supported still images. Video
playback and Live Photo motion require checking the existing file delivery
behavior or waiting for the Stream Service contract. Unsupported originals
need a useful file fallback.

The demo does not imply album CRUD, tags, recursive discovery, pagination,
uploads, SQLite settings, or collage covers are implemented. Those remain in
the [Moments app task](../apps/moments/task.md) and service task. Before
implementing an integration whose contract is not documented in existing code,
settle that contract explicitly.

## Gallery design direction

The [gallery task](../apps/gallery_taks.md) is the detailed source for the
frontend experience. Its principles apply to the full Moments gallery and
guide a smaller demo where relevant:

- Keep metadata separate from original media bytes. Only visible items and a
  small nearby range should create media requests or DOM nodes.
- Share one conservative image-request budget across wall and modal viewing.
  Prioritize the current viewport or selected item, reuse in-flight requests,
  and drop queued work that is no longer useful. This protects the Raspberry
  Pi and shared ZFS pool from gallery traffic competing with SMB/NFS users.
- Keep wall scroll position and layout stable while the modal opens. Plan wall
  geometry independently of image readiness, using known dimensions when
  available and a stable fallback when they are not.
- Use original media through existing authorized routes. The gallery must not
  require generated thumbnails, posters, transcoded files, or per-file
  sidecars. Browser cache behavior can help, but decoded media must not be
  retained without a bound.
- Provide a responsive mosaic or justified wall, loading placeholders, a
  keyboard-accessible modal, reduced-motion behavior, and explicit unsupported
  format and loading-error states. Video tiles should not probe many videos.
- Use the established React, TypeScript, MUI, and Zod conventions. Validate
  File Service responses before mapping them into gallery state and validate
  submitted settings when those operations exist.

The detailed task's contexts, loader, registry, and layout strategies describe
responsibilities, not a requirement to create every proposed abstraction for
the demo. Add each piece only when the implementation needs that behavior.
Fixed numeric examples in that task are tuning starting points, not backend
contracts or guaranteed performance targets; measure with realistic galleries
and constrained network conditions.

## Later service integration

When the Moments metadata API is defined, replace the directory adapter with
validated album/media data from the Moments service. Keep File Service as the
authorized file boundary and use Stream Service for its documented playback
contract when available. Resolve ordering, pagination, stable media identity,
URL construction, dimensions, Live Photo association, and persisted layout
settings in those service contracts before wiring the browser to them. The
gallery's display model may map from those contracts, but it does not define
their HTTP shapes by itself.

See [Moments skeleton requirements](../requirements/moments-skeleton.md) for
implemented scope and [development notes](development-notes.md) for priority
context.
