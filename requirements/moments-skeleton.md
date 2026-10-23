# Moments Skeleton Requirements

## Scope

This first increment establishes a browser session, an empty home screen, shared
profile settings, and a deployable standalone host. Album, media, streaming,
and persistence contracts remain in the existing Moments task documents and
require separate API decisions before implementation.

## Functional requirements

- **moments/FR-001:** The Moments API rejects current-user requests
  without a valid LiteNAS access token. Its login, refresh, and logout routes
  use the existing Auth Service RPC contracts and browser cookie names.
- **moments/FR-002:** Browser authentication mutations from a foreign Origin
  are rejected before reaching Auth Service.
- **moments/FR-003:** The browser app shows a login screen for an anonymous
  user, an empty home screen for an authenticated user, and the shared header
  user menu with profile, application settings, and sign-out actions.
- **moments/FR-004:** The settings view reuses shared profile and appearance
  UI code, excludes admin-only settings, and accesses profile files through
  the authenticated File Service routes. The preferred light or dark mode
  applies to the whole Moments page, including its body background.
- **moments/FR-005:** The same built app can load under `/apps/moments/` on the
  admin host and `/` on the standalone Moments host. The admin panel offers a
  Moments application card and sidebar entry only while the Moments systemd
  service is running, and embeds the same-origin mount at
  `/applications/moments`.

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
