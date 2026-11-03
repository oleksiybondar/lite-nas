# Moments Skeleton and Demo Requirements

## Scope

The implemented increment establishes a browser session, shared profile
settings, a deployable standalone host, a fixture-backed gallery, and a
browser-only thumbnail maker. The gallery fixture references authorized media
files; it is not an album or metadata API. Album, tag, upload, and persistence
contracts still require separate API decisions.

## Functional requirements

- **moments/FR-001:** The Moments API rejects current-user requests
  without a valid LiteNAS access token. Its login, refresh, and logout routes
  use the existing Auth Service RPC contracts and browser cookie names.
- **moments/FR-002:** Browser authentication mutations from a foreign Origin
  are rejected before reaching Auth Service.
- **moments/FR-003:** The browser app shows a login screen for an anonymous
  user and a fixture-backed gallery home for an authenticated user. Its shared
  header offers the thumbnail maker, profile, appearance settings, and sign-out.
- **moments/FR-004:** The settings view reuses shared profile and appearance
  UI code, excludes admin-only settings, and accesses profile files through
  the authenticated File Service routes. The preferred light or dark mode
  applies to the whole Moments page, including its body background.
- **moments/FR-005:** The same built app can load under `/apps/moments/` on the
  admin host and `/` on the standalone Moments host. The admin panel offers a
  Moments application card and sidebar entry only while the Moments systemd
  service is running, and embeds the same-origin mount at
  `/applications/moments`.
- **moments/FR-006:** The gallery reads the validated, ordered demo fixture and
  offers virtualized mosaic, justified, masonry, and grid walls plus a modal
  viewer. Media requests remain bounded while scrolling and navigating.
- **moments/FR-007:** Gallery photos and videos open through authorized media
  routes. The viewer supports navigation, fullscreen, slideshow timing, video
  autoplay controls, original downloads, and available EXIF details. The wall
  can play a recognized Live Photo's motion from its control without opening
  the viewer.
- **moments/FR-008:** The thumbnail maker accepts 1–35 images from the same
  fixture or browser uploads, with random fixture selection limited to 25.
  Uploaded images remain in browser memory and can be embedded in exported
  versioned JSON recipes. The editor provides templates, per-image placement,
  crop and frame controls, composition transforms, layer ordering, and PNG/JPG
  export at a pixel area of at least 1920×1080. PNG export preserves a
  transparent canvas and transparent source pixels.

## Integration requirements

- **moments/IR-001:** Nginx exposes `/f/`, `/d/`, and `/ls/` on the standalone
  host through the existing File Service snippet. Uploads continue through
  `/f/` under the File Service contract.
- **moments/IR-002:** The standalone host proxies `/api/` to a loopback Moments
  listener. Its HTTP listener remains loopback-only for a Cloudflare tunnel.
- **moments/IR-003:** NATS authorizes the Moments service identity for the Auth
  Service subjects needed by the browser session lifecycle and the established
  RBAC subject-role lookup. File path authorization remains with File Service.
- **moments/IR-004:** Both Web Gateway and Moments accept the same optional,
  validated parent cookie domain. Empty values issue host-only cookies.
- **moments/IR-005:** The embedded and standalone hosts expose `/s/` through
  Stream Service for original playback, HEIC/HEIF previews, and video
  transcode fallback. Local Vite development proxies the same path.
