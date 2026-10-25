# Moments Architecture

Moments is LiteNAS's personal photo and media application. Its current browser
experience includes a fixture-backed gallery and a browser-only thumbnail maker.
Album, tag, upload, and persistent gallery metadata remain future service work.

## Current boundaries

- The React app and loopback Go service provide authentication, shared profile
  and appearance settings, and embedded and standalone hosting. The Go service
  does not yet expose album or gallery metadata APIs.
- The demo gallery reads the validated, ordered
  `apps/moments/public/dev-fixtures/gallery.json` fixture. Its paths point to
  existing files available to the signed-in user. The fixture is bundled with
  the app and does not discover files from a directory listing.
- File Service owns authorized `/f/` reads, `/d/` downloads, and `/ls/`
  listings. File uploads use its established `/f/` contract. Neither browser
  demo nor Moments service scans NAS paths directly.
- Stream Service supplies the `/s/` media route, including original streaming,
  HEIC/HEIF JPEG previews, and a video transcode fallback. Nginx exposes the
  route on embedded and standalone hosts; Vite proxies it during development.
- The planned Moments metadata API will own album identities, tags,
  pagination, persisted settings, and cover descriptors. Its contract must be
  decided before the browser integrates it.

## Gallery demo

The gallery maps fixture records into a local collection model. The wall offers
mosaic, justified, masonry, and grid layouts. It virtualizes nearby tiles and
uses a bounded image queue. The modal viewer keeps the wall mounted and offers
navigation, slideshow, fullscreen, original downloads, and available EXIF
details. Videos use native playback with a transcode fallback. A recognized
same-folder HEIC/HEIF and MOV pair appears as a Live Photo; its motion can play
on the wall or in the viewer. Pairing in this demo is based on fixture paths
and filenames; a future metadata service must define durable media identity
and association rules.

The gallery does not imply album CRUD, recursive discovery, pagination, tags,
uploads, or persisted layout preferences. Its wall layout selection is local to
the current page. The [gallery design](../apps/gallery_taks.md) records the
broader experience and remaining validation work.

## Thumbnail maker demo

The authenticated browser app can build a thumbnail from up to 35 photos. It
uses the same gallery fixture for photo selection, limits random selection to
25, and also accepts local image uploads. Uploaded bytes stay in browser
memory as data URLs and can be included in exported JSON. Transparent source
pixels remain transparent. The editor supports scattered, ribbon, grid, and
featured layouts; image position, size, rotation, frame ratio, crop, and
effects; layer order; and whole-composition position and scale. A versioned
JSON recipe recreates the canvas. PNG or JPG export renders from that recipe at
a pixel area of at least 1920×1080. PNG can preserve a transparent canvas;
JPG fills it with white.

The picker loads one highlighted preview at a time, selected images decode in
a bounded queue, and export reads originals sequentially. The maker has no
backend storage, album cover association, or persistence beyond a downloaded
recipe or image.

## Remaining service integration

Define the Moments metadata API before replacing fixture records with
validated album/media responses. Settle ordering, pagination, stable media
identity, URL construction, dimensions, Live Photo association, and persisted
layout and cover settings in the service contract. Keep File Service as the
authorized file boundary and Stream Service as the media delivery boundary.
The current frontend recipe is a demo format; a persisted album cover
descriptor needs its own agreed contract and a policy for moved or missing
source files.

See [implemented requirements](../requirements/moments-skeleton.md), the
[app task](../apps/moments/task.md), the
[service task](../services/moments/task.md), and the
[app README](../apps/moments/README.md) for usage and deployment details.
