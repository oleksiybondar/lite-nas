# Moments

Moments contains browser authentication, a paged gallery feed on its home page,
owned and shared gallery views, the shared user dropdown, and a settings sidebar
with Profile, Appearance, and Gallery tabs. Gallery cards and collection
membership come from the Moments API. The settings heading and tabs use the
shared `@lite-nas/settings` preferences components also used by Admin Panel.
Opening a gallery changes the browser URL to `/galleries/<id>` on the standalone
host or `/apps/moments/galleries/<id>` in the embedded app. These links can be
shared with people who have access to the gallery. The route reads the gallery
from the API after sign-in, so opening a saved link or using browser Back/Forward
returns to the same view. Vite's built entry point sets its asset base from the
current mount, and both Nginx hosts send browser routes to `index.html`.
When sign-in is needed, Moments keeps the requested route in
`/login?redirect_to=...`; the login API echoes a validated destination and the
SPA returns there after authentication. Admin Panel uses the same return-path
rule for its protected routes.
Album media and referenced collection
directories are read through the caller's File Service permissions. The
configured `workingDir` from `/api/config` starts both file and folder pickers.
Owners can create albums and collections with a description, visibility, and
user tags, then edit those fields in a modal from the gallery detail view.
The edit modal also starts an on-demand media scan, shows its progress and
result, and lets the owner cancel an active job. Scan status polling continues
while the gallery view remains open so the backend can renew the worker token.
An album card's collection menu adds or removes its source directory in an
owned collection without changing the album's files.
New albums and collections start as public; owners can select private in the
create or edit modal.
Private galleries accept login names for sharing; Moments resolves each login
to a stable user ID before granting access.
On the admin host, the embedded app reaches the same Moments API through
`/apps/moments/api/`; the admin gateway retains its own `/api/` routes.
Collection, wall, and viewer providers share one image queue. The Gallery
settings tab controls its parallel image load limit, which defaults to ten and
is saved in the signed-in user's `~/.lite-nas/moments.json` under `gallery`.
The Profile, Appearance, and Gallery tabs separate the settings groups;
appearance remains shared through `~/.lite-nas/ui.json`.
The feed has search and filters for type, visibility, media type, and location;
the media type filter accepts multiple values using the backend's repeated
`mediaType` query parameter. Any selected media type may match a gallery, and
changing a filter restarts pagination at the first page.

Feed cards show a selected NAS cover through File Service or an uploaded cover
stored by Moments. Without either, a card lazily uses the first accessible
image from that album or collection, including HEIC previews through Stream
Service. Empty galleries show a placeholder.

The gallery wall uses its saved render strategy from Moments. Owners choose
mosaic, justified rows, masonry, or an even grid when creating or editing an
album or collection. A saved change updates the open wall and resets its scroll
position. Without item dimensions, masonry and justified rows use deterministic
placeholder proportions.

The wall renders only nearby tiles, requests only images intersecting its
viewport, pauses new image requests during scrolling, and limits active
originals to the configured amount. Loaded image tiles rise and fade in,
with reduced motion respected. Each tile is a `GalleryWallItem` component that
owns its position, hover and focus state, and an optional hover action area;
photo and video components handle their own media. Wall video tiles use native HTML5 video without
an initial source or autoplay; pressing Play starts the original. Opening the
viewer pauses wall playback. The viewer has previous/next controls, a position counter, a browser fullscreen
button, slideshow Play/Pause with 1/3/5-second image intervals, and an
independent video autoplay switch. Videos advance a running slideshow when
playback ends; a blocked autoplay attempt leaves native controls available.
A failed image can be retried from the viewer.
Hovering or focusing a wall photo reveals its filename and available EXIF capture
time, camera, and GPS coordinates. JPEG metadata is read from a bounded File
Service range only after interest in that tile; the gallery cancels abandoned
reads and caches small parsed results. Photos with GPS coordinates offer a
**Show on map** link that opens Google Maps when selected. The carousel also
shows available EXIF details beside the selected file name and shares that
cache. Its download icon requests the selected original through File Service's
`/d/` attachment route. Wall and carousel items share a three-dot menu for
adding files to owned collections, creating a new collection, and removing an
explicit file from an owned collection.
Appearance preferences apply to both embedded and standalone views.

Owned collections offer **Add file** and **Add folder** pickers backed by File
Service. In the edit modal, owners of albums and collections can choose an
existing NAS image, upload a PNG/JPEG cover, or generate one with the reusable
thumbnail editor.
Generated images are uploaded to the Moments thumbnail route. The editor also
retains image and recipe downloads. These controls live in gallery detail
rather than standalone navigation pages.

## Local build and deployment

Build the two artifacts with `./scripts/build-moments-binary.sh` and
`./scripts/build-moments.sh`. `./scripts/deploy-moments.sh` builds and installs
both, the loopback service, and the Nginx configuration. The script requires
root privileges and accepts `--binary`, `--assets-source`, `--no-start`, and
`--skip-bootstrap` for an existing prepared host.

After deployment, open `https://moments.localhost/`,
`https://localhost/apps/moments/`, or the admin panel's **Applications → Moments**
page at `/applications/moments`. The admin panel embeds the same-origin
`/apps/moments/` Nginx mount in an iframe. A manual command such as
`curl --resolve moments.localhost:443:127.0.0.1` can resolve the local host
without DNS changes. The local self-signed certificate includes
`moments.localhost` after rotation. Moments uses its own `moments.crt` and
`moments.key`; adding Moments no longer replaces the admin host certificate.
Browsers must trust each self-signed certificate separately for direct local
HTTPS access.

With empty `cookie_domain` settings, each host has a host-only session and
requires its own sign-in. To share a session across real subdomains, set the
same parent DNS domain in `/etc/lite-nas/web-gateway.conf` and
`/etc/lite-nas/moments.conf`, then restart both services. For example, an
admin host `nas.example.com` and Moments host `moments.example.com` use
`cookie_domain=example.com`. Use HTTPS on both browser origins.

## Cloudflare tunnel

Publish a Moments hostname through a Cloudflare tunnel to
`http://127.0.0.1:80`, preserving the external hostname in the Host header
and forwarding `X-Forwarded-Proto: https`.
The standalone Nginx server matches `moments.<domain>` and keeps the origin
listener on loopback. Add the corresponding DNS record to the tunnel. The
browser reaches `/api/`, `/f/`, `/s/`, `/d/`, and `/ls/` through the same origin, so
no cross-origin token transfer is needed. For direct HTTPS without a tunnel,
install a certificate covering the production hostname and adjust the Nginx
TLS certificate paths.

The existing File Service snippet handles uploads under `/f/`, along with
source files under `/f/`, downloads under `/d/`, and listings under `/ls/`.
The gallery source provides `/f/` file URLs. `StreamContext` validates them and
derives the original `/s/`, transcoded `/s/transcode/`, and HEIC/HEIF JPEG
`/s/preview/` routes for the gallery. Components use standard HTML5 image and
video elements; the context owns the LiteNAS URL contract for future callers.
HEIC/HEIF wall, viewer, and thumbnail images use the JPEG preview. The gallery
combines an unambiguous same-folder HEIC/HEIF and MOV filename pair into one
Live Photo without waiting for media requests. Its wall chip plays and stops
motion inline without opening the viewer; that control handles its own click
so it does not activate the tile. The viewer can also play motion. If the
still image fails, the paired movie remains available.
Standalone HEIC files remain images; standalone MOV files remain videos and can
fall back to transcoding when native playback fails.
Downloads continue to use the original File Service bytes.

## Local frontend development

Run `npm run dev --workspace apps/moments` and open the Vite URL. The dev server
proxies `/api/` to the Moments API at `127.0.0.1:9091` and `/f/`, `/d/`, and
`/ls/` to File Service at `127.0.0.1:8082`. Set
`LITE_NAS_MOMENTS_API_ORIGIN` or `LITE_NAS_FILE_SERVICE_ORIGIN` before starting
Vite when those services use different loopback origins. File requests keep
their browser-facing paths: `/f/home/...` maps to `/home/...`, and
`/d/home/...` maps to `/download/home/...` on File Service.

Vite proxies `/s/` to Stream Service at `127.0.0.1:8083` by default, matching
the Nginx mount. Set `LITE_NAS_STREAM_SERVICE_ORIGIN` before starting Vite when
the service uses another loopback origin. `/s/home/...` maps to `/home/...` on
Stream Service. The Vite server reloads frontend changes without rebuilding or
redeploying Moments; its backend services must still be running.
