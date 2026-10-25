# Moments — Remaining React Application Work

The authenticated shell, embedded and standalone hosting, fixture-backed gallery,
and browser-only thumbnail maker are implemented. See the
[app README](README.md) and [implemented requirements](../../requirements/moments-skeleton.md)
for current behavior. The [architecture](../../docs/moments-architecture.md)
describes the service boundaries. The [gallery design](../gallery_taks.md) is a
reference for future integration and performance validation; its example view
models are not approved backend DTOs.

## Album and media workflow

- Add album creation, rename, and removal after the Moments service defines its
  API. Each album needs a readable name and a NAS source directory, including
  nested files where the authorized discovery contract permits it.
- Replace fixture-backed collection loading with validated album/media
  responses. Define stable media identity, ordering, pagination, dimensions,
  format and Live Photo association with the service. Preserve the existing
  bounded wall and viewer behavior.
- Add album filtering by name and tags, tag editing, and tag display in the
  gallery once those values exist in the metadata API.
- Add multi-file drag-and-drop uploads and directory selection where browsers
  support it. Upload flat into the selected album source through File Service;
  agree on collision behavior and show progress and failures.
- Persist gallery layout preferences using a documented user or album setting
  contract. Resolve Moments-specific theme precedence and persistence against
  shared profile settings before changing the current appearance behavior.
- Connect a versioned album cover descriptor to stored media identities and
  persist it through the Moments service. Define regeneration when selected
  media move or disappear. The existing thumbnail maker JSON is a browser-only
  recipe and must not be assumed to be the service DTO.

## Integration and acceptance

- Agree on Moments service routes and DTOs before implementation; validate
  responses and submitted payloads with reusable Zod schemas at the browser
  boundary.
- Keep file authorization with File Service, playback with Stream Service, and
  browser session handling with the existing shared cookie flow.
- Verify album CRUD, nested discovery, filtering, pagination, tags, uploads,
  persisted preferences, and cover regeneration through focused tests when
  these workflows are implemented.
