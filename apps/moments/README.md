# Moments

Moments contains browser authentication, a fixture-backed demo gallery on its
home page, the shared user dropdown, and shared profile/appearance settings.
The gallery reads `public/dev-fixtures/gallery.json` and uses the existing
`/f/` File Service route. The fixture paths must exist and be authorized on the
running host. Collection, wall, and viewer providers share one headless image
queue through the collection provider.

The fixture stores `defaultLayout` and an `items` array. The demo's Layout
dropdown starts with that default and offers mosaic, justified rows, masonry,
and an even grid. Selecting a layout resets the wall scroll position; it does
not rewrite the bundled fixture. Without item dimensions, masonry and justified
rows use deterministic placeholder proportions.

The wall renders only nearby tiles, requests only images intersecting its
viewport, pauses new image requests during scrolling, and limits loading to
eight pending and two active originals. Loaded image tiles rise and fade in,
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
`/d/` attachment route. Wall and carousel items share a three-dot menu with a
disabled Add to collection entry until collection support is available. The
path-only demo fixture has no tag values to display.
Album and gallery metadata APIs are not implemented yet. Appearance preferences
apply to both embedded and standalone views.

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
browser reaches `/api/`, `/f/`, `/d/`, and `/ls/` through the same origin, so
no cross-origin token transfer is needed. For direct HTTPS without a tunnel,
install a certificate covering the production hostname and adjust the Nginx
TLS certificate paths.

The existing File Service snippet handles uploads under `/f/`, along with
source files under `/f/`, downloads under `/d/`, and listings under `/ls/`.
The demo gallery reads original files through `/f/`; uploads and directory
listings are not exposed by this demo.

## Local frontend development

Run `npm run dev --workspace apps/moments` and open the Vite URL. The dev server
proxies `/api/` to the Moments API at `127.0.0.1:9091` and `/f/`, `/d/`, and
`/ls/` to File Service at `127.0.0.1:8082`. Set
`LITE_NAS_MOMENTS_API_ORIGIN` or `LITE_NAS_FILE_SERVICE_ORIGIN` before starting
Vite when those services use different loopback origins. File requests keep
their browser-facing paths: `/f/home/...` maps to `/home/...`, and
`/d/home/...` maps to `/download/home/...` on File Service.

When Stream Service is available, set `LITE_NAS_STREAM_SERVICE_ORIGIN` to its
loopback HTTP origin before starting Vite. The `/s/home/...` proxy then maps to
`/home/...` on Stream Service. This repository currently specifies Stream
Service but does not implement its listener, so `/s/` has no target until that
origin is configured; Vite returns a clear 503 for `/s/` in the meantime. The
Vite server reloads frontend changes without rebuilding or redeploying Moments;
its backend services must still be running.
