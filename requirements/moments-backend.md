# Moments Backend Requirements

These requirements define the implemented gallery metadata increment. The
existing browser-session requirements remain in
[moments-skeleton.md](moments-skeleton.md).

- **moments-backend/FR-001:** Authenticated users can create, list, read, edit,
  and delete gallery metadata. Albums have one File Service-authorized source
  directory; collections have explicit file and directory references. Metadata
  deletion never deletes an original file.
- **moments-backend/FR-002:** The owner can change membership and metadata.
  Shared members have read-only metadata access. Public metadata is visible to
  authenticated LiteNAS users. Hidden galleries and their tags, references,
  and jobs are not enumerable.
- **moments-backend/FR-003:** Source paths are resolved and authorized through
  File Service before album creation or collection reference addition. File
  Service provides per-file `stat` and source-backed media metadata routes with
  its normal Auth/RBAC checks.
- **moments-backend/FR-004:** An owner can start a bounded full rescan, poll
  durable progress, renew its volatile user JWT by same-user polling, and cancel
  work. A successful scan atomically reconciles changed path metadata and
  gallery links, removing a path only after File Service confirms deletion.
  Partial, failed, and cancelled scans preserve previously published records.
- **moments-backend/FR-005:** Gallery listing, published media, and tag
  aggregates are paged or bounded and apply metadata visibility before
  filtering or counting. Gallery listing uses page/size metadata; search
  includes albums and collections by default and matches name, description,
  user tags, and all published scan tags.
- **moments-backend/FR-007:** Owners can create, rename, and remove gallery-level
  user tags. Gallery reads include user and scan tags with a type discriminator.
  User tags survive rescans and source changes; scan-derived tags remain
  read-only to clients.
- **moments-backend/FR-008:** File Service extracts validated JPEG EXIF GPS
  coordinates. Moments resolves a nearby city from packaged offline GeoNames
  data and indexes the city, country, and alternate names for case- and
  accent-insensitive gallery search and location filtering. Sources without
  usable GPS or a nearby city remain unlabelled.
- **moments-backend/FR-009:** A completed browser upload in an owned Moments
  album triggers metadata extraction for that file only. Existing gallery
  records remain published. Shared source metadata is keyed by canonical file
  path, while gallery links retain membership and Live Photo pairing. An
  upload refreshes references in the owner's collections. Unchanged full-scan
  paths retain their gallery link identity and source tag associations.
- **moments-backend/FR-006:** A gallery may store an optional opaque JSON
  thumbnail recipe and a separate optional uploaded PNG/JPEG image. Asset
  reads follow gallery visibility; only the owner can upload or delete it.
- **moments-backend/IR-001:** Moments uses the installed nonroot service,
  SQLite state directory, configured loopback File Service origin, and narrow
  systemd/AppArmor write permissions. Restart marks interrupted scans failed.
- **moments-backend/IR-002:** Moments, File Service, and Stream Service expose
  same-origin Swagger UI and OpenAPI documents using local assets.
