# Moments Service — Go Backend Implementation Task

## Goal

Create the Go backend for Moments. The service owns Moments configuration,
album and tag CRUD, SQLite metadata, media discovery metadata, and collage
layout descriptors. It is one new Moments Go service; use the existing Stream
Service for media byte delivery and File Service for file access and uploads.

The installed service name is `lite-nas-moments`. The application path is
`services/moments`; its browser UI is the separate `apps/moments` React app.
The [Moments architecture](../../docs/moments-architecture.md) distinguishes
the initial File Service gallery demo from this later metadata API work.

## Responsibilities and boundaries

- Store the SQLite database at a configurable location on the NAS pool so it
  can be moved with the pool. Validate configuration, create required parent
  directories only through an explicitly authorized provisioning path, and
  document backup/restore behavior.
- Persist album metadata separately from source directories. An album has a
  user-readable name and a source directory; its contents are discovered by
  recursively traversing the source through File Service when the album is
  read. Do not make an eager private copy or rely on a one-time import scan.
- Persist tags and their relationships to albums/media, and support album
  search/filter, pagination, and cover-layout retrieval/regeneration.
- Use File Service's authenticated listing/read/upload capabilities and
  Stream Service's playback contract. Do not access pool paths directly from
  this service or bypass File Service/RBAC authorization.
- Authenticate browser requests with the shared `lite-nas-at` session and the
  same login/refresh/logout behavior as Web Gateway. Reuse the shared browser
  auth module described in
  [shared browser auth task](../../shared/go/browserauth/task.md).
- Require album ownership/authorization checks for every operation. Never
  accept a caller-supplied username as the identity; derive it from validated
  authentication and authorize the selected source through File Service.

## Decisions required before implementation

1. Define the API routes, DTOs, error envelope, pagination semantics, and
   album/media identities. Follow existing HTTP route and Huma validation
   patterns; do not implement an undocumented contract.
2. Define safe source-directory selection and the authorization behavior when
   a user's access to a source changes after album creation. Directory listing
   must use File Service and fail closed on authorization or transport errors.
3. Define how recursive File Service results are requested and paginated, and
   how the service avoids holding an entire large album in memory.
4. Define SQLite schema/migrations and concurrency behavior. Keep the database
   location configurable and ensure it is on the selected pool when configured
   that way.
5. Define the serializable, versioned collage-layout descriptor and how a
   regenerated cover handles files that have moved or disappeared. The
   descriptor must identify selected media and placement values without
   embedding media bytes.
6. Define Live Photo association and metadata extraction from existing media
   assets using the Stream Service/File Service contracts. Do not infer a pair
   solely from similar filenames.
7. Coordinate cookie domain, refresh/logout behavior, CSRF protection, and
   same-origin versus standalone-host routing with Web Gateway and the shared
   auth module.

## Implementation actions

1. Inspect File Service, Stream Service, Web Gateway, shared Go configuration,
   NATS/certificate, service packaging, deployment, AppArmor, and logrotate
   conventions.
2. Write `requirements.md` with source-qualified requirement IDs and settle the
   decisions above before introducing API or storage contracts.
3. Implement validated HTTP adapters, service-owned request models, SQLite
   persistence, and album/tag/cover operations. Keep validation at transport
   boundaries and use Huma DTO validation for HTTP inputs.
4. Add service configuration for database path, HTTP listener, auth cookie
   domain, and enabled state using existing LiteNAS config patterns.
5. Add the loopback service, Nginx mounts at `/apps/moments/` and the configured
   `moments.<host>` virtual host, certificate/runtime wiring, and deployment
   documentation. Record the required Cloudflare DNS/proxy configuration for
   the standalone hostname; do not require a Cloudflare API integration.
6. Integrate the service in local build/deploy flows and the single Debian
   package. Add the service account, AppArmor policy, logs, NATS authorization,
   and certificates only where the selected integrations require them.
7. Add focused service tests for auth, album CRUD, nested source discovery,
   authorization changes, tags/filtering/pagination, SQLite persistence and
   relocation, collage regeneration, upload delegation, and dependency
   failures.

## Acceptance criteria

- Moments metadata is stored in configurable SQLite on the selected pool.
- Albums point to existing NAS directories and discover nested files through
  File Service when read.
- Album names, tags, filters, pagination, and cover descriptors persist and
  behave as documented.
- File upload is delegated to File Service; playback is delegated to Stream
  Service; this service does not bypass their authorization boundaries.
- The service uses the shared Web Gateway login/session behavior at both the
  embedded path and standalone host.
- `lite-nas-moments` is independently enableable and integrated with local
  build/deploy and single-package install flows.
