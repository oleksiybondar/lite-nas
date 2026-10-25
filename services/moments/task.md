# Moments Service — Remaining Metadata Backend Work

`services/moments` already provides the loopback authentication service used by
the embedded and standalone Moments app. Its login, refresh, logout, shared
cookie behavior, service deployment, and Nginx mounts are in place. The
remaining service work is an album, tag, and cover metadata API. See the
[architecture](../../docs/moments-architecture.md) for the current boundaries.

## Responsibilities

- Persist album names and source directories, tags and their relationships,
  gallery preferences, and versioned cover descriptors in SQLite at a
  configurable pool location. Define provisioning, migration, concurrency,
  backup, and restore behavior.
- Discover album contents recursively through authorized File Service
  operations when read. Avoid an eager private media copy and avoid loading an
  entire large album into memory. File Service retains path authorization,
  reads, and uploads; Stream Service retains playback.
- Derive ownership from the validated browser session. Authorize every album
  operation and source selection, including access changes after creation.
- Keep the API focused on metadata. Original media bytes and uploaded files
  must not pass through a second Moments media stack.

## Contract decisions before implementation

1. Define routes, DTOs, error envelope, album/media identity, sort order,
   filtering, and pagination. Document these in source-qualified requirements
   before writing adapters. Use Huma DTO validation for HTTP inputs.
2. Define source-directory selection, recursive File Service discovery,
   authorization failures, and pagination over large directories. Confirm the
   required File Service contract before relying on it.
3. Define SQLite schema and migrations, database placement and provisioning,
   and behavior when the configured pool is unavailable.
4. Define durable Live Photo association and metadata extraction. The current
   frontend demo's same-folder filename pairing is not a persistence contract.
5. Define the cover descriptor's version, media references, and missing-file
   policy. The browser thumbnail maker recipe may inform the design but is not
   an approved service DTO and may contain uploaded data URLs.
6. Define how album upload and persisted user or album preferences connect to
   existing File Service and profile settings, including authorization and
   collision behavior.

## Implementation and acceptance

- Implement validated HTTP adapters, service-owned request models, SQLite
  persistence, and album/tag/preference/cover operations after the contracts
  above are settled.
- Add only configuration and deployment changes needed for metadata storage or
  new integrations; retain the existing service, Nginx, cookie, build, and
  package wiring.
- Add focused tests for CRUD, nested discovery, authorization changes,
  filtering/pagination, persistence and migration, cover regeneration, upload
  delegation, and dependency failures.
- Complete when albums and metadata persist on the selected pool, authorized
  nested media can be discovered, and the documented API supports the app's
  remaining workflows without bypassing File or Stream Service.
