# Moments — React Application Task

## Goal

Build Moments as a self-hosted photo and personal-media library UI. It is a
separately deployable React application that works embedded in the admin panel
at `/apps/moments` and directly through the Moments host configured for the
deployment.

The [LiteNAS Moments architecture](../../docs/moments-architecture.md) defines
the current service boundaries and first gallery demo. The
[gallery design task](../gallery_taks.md) describes the longer-term wall and
viewer experience. Start with the contract-backed File Service demo before
implementing the full album workflow below; do not treat the gallery task's
example TypeScript models as backend DTOs.

## User behavior

- Let users create, rename, and remove albums. Each album has a readable name
  and a source directory on the NAS; one album includes that directory and its
  subdirectories.
- Show photos, videos, GIFs, and supported lossless media in a responsive media
  wall. Select and document a layout algorithm before implementation; the wall
  should use available space well and support vertical scrolling on narrow and
  wide screens.
- Open a selected item in a carousel/modal with previous/next navigation and
  optional autoplay. Display tags and available metadata such as date/time and
  geolocation on hover and in the viewer.
- Support album filtering by name and tags, plus pagination.
- Allow users to add, edit, and remove tags and to upload media to an album.
  Upload multiple files with drag and drop; support whole-directory selection
  where the browser permits it. Upload files flat into the selected album's
  source directory. Define collision behavior before implementation.
- Display Apple Live Photos as a still with its paired motion component when
  the source contains a valid pair. Ordinary videos remain playable as media;
  Moments is not a general streaming catalog.
- Generate a collage-style album cover from up to 35 selected photos, with
  randomized fixture selection limited to 25 photos. Keep a
  serializable layout descriptor containing the chosen items and their
  positions, angles, and sizes so the cover can be regenerated consistently.
- Use a Moments-specific theme setting when one exists. Otherwise inherit the
  authenticated user's preferred UI settings from `.lite-nas` when available,
  then use the app default. Save changes made in Moments under its own
  preference key.

## Integration boundaries

- Use the Moments Go service for album, tag, pagination, and cover-layout
  metadata operations.
- Use File Service for authorized directory listings, source file access, and
  uploads. Album traversal must include subdirectories when reading an album;
  do not scan NAS paths directly from the browser.
- Use Stream Service for byte-range media delivery and the established Live
  Photo/original-format playback contract. Do not implement a second media
  streaming stack in the Moments API service. Use the existing `/s/` route on
  both the admin-panel origin and the standalone Moments host.
- Reuse the `lite-nas-at` cookie and the shared login flow. Do not pass tokens
  through URLs or iframe messages.

## Implementation instructions

1. Review the existing admin-panel React conventions, File Service `/ls/`,
   `/f/`, and upload contracts, Stream Service playback contract, and Moments
   API contract before building the UI.
2. Define the browser-facing DTOs from the Moments service contract and validate
   every API response and submitted payload with reusable Zod schemas.
3. Implement album management, media wall, filtering/pagination, carousel,
   metadata/tag display, upload progress/errors, and cover regeneration.
4. Implement same-origin routing at `/apps/moments` and direct-host routing
   without relying on admin-panel local storage being accessible cross-origin.
   Read shared user preferences through an authenticated service/API boundary.
5. Add app build, deployment, and packaging integration. Preserve test
   selectors and the repository's frontend component organization.

## Acceptance criteria

- Albums use readable names and NAS directory sources, including nested
  directories.
- The media wall, carousel, filtering, pagination, metadata, tags, uploads,
  theme precedence, and repeatable collage cover behavior meet the requirements
  above.
- Browser-side requests are validated and use the established auth and file or
  media service contracts.
- The app works at `/apps/moments` and at its configured standalone host.
- Focused UI tests cover album workflows, media viewing, filtering, upload
  behavior, and preference precedence.
