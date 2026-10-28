# Moments Architecture

Moments is LiteNAS's personal photo and media application. Its browser gallery
loads persistent album and collection metadata from the Go backend and resolves
media through File Service. The backend provides collection references, sharing,
uploaded covers, and on-demand media rescans.

## Current boundaries

- The React app and loopback Go service provide authentication, shared profile
  and appearance settings, and embedded and standalone hosting. The Go service
  now exposes metadata routes under `/api/galleries`; its Swagger UI is at
  `/api/docs` on the standalone Moments host.
- The browser pages the accessible gallery feed through Moments. Albums resolve
  their source directory through File Service, and collections resolve their
  saved references through the same authorized file boundary.
- File Service owns authorized `/f/` reads, `/d/` downloads, and `/ls/`
  listings. File uploads use its established `/f/` contract. Moments does not
  scan NAS paths directly.
- Stream Service supplies the `/s/` media route, including original streaming,
  HEIC/HEIF JPEG previews, and a video transcode fallback. Nginx exposes the
  route on embedded and standalone hosts; Vite proxies it during development.
- The Moments metadata API owns gallery identities, shares, collection
  references, scan-derived tags, pagination, and optional cover data. The
  [backend API guide](../services/moments/README.md) documents its contract.
- Optional local BLIP and E5 workers generate captions and rank image search.
  The [local model guide](moments-models.md) covers their installation,
  authorization, restart behavior, and resource limits.

## Gallery

The gallery maps authorized originals into a local presentation model. The wall offers
mosaic, justified, masonry, and grid layouts. It virtualizes nearby tiles and
uses a bounded image queue. The modal viewer keeps the wall mounted and offers
navigation, slideshow, fullscreen, original downloads, and available EXIF
details. Videos use native playback with a transcode fallback. A recognized
same-folder HEIC/HEIF and MOV pair appears as a Live Photo; its motion can play
on the wall or in the viewer. Browser pairing uses file paths and filenames.

Owners create and edit gallery details in a modal, including descriptions,
visibility, wall layout, user tags, and private shares. The wall renders the
layout returned by Moments for that gallery. The feed pages gallery metadata;
search and filters use the same paged API. Owned collection actions use File
Service pickers to add file or folder references and can open the thumbnail
editor to save a generated cover through Moments.

## Remaining browser integration

File Service's recursive album listing is not paged. Large albums still need a
practical listing limit or a paged File Service contract. File Service remains
the authorized file boundary and Stream Service the media delivery boundary.
Uploaded cover images and optional generation JSON are persisted separately by
Moments.

See [browser requirements](../requirements/moments-skeleton.md),
[backend requirements](../requirements/moments-backend.md), and the
[app README](../apps/moments/README.md) for usage and deployment details.
