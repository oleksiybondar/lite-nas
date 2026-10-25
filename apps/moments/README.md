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

## Thumbnail maker demo

Open **Thumbnail maker** from the Moments header after signing in. It uses the
same `public/dev-fixtures/gallery.json` photo paths as the gallery. You can add
photos through the searchable picker, generate a random selection of 1–25
photos, or upload PNG, JPEG, WebP, or GIF files. Uploaded images are embedded
as data URLs in browser memory and in downloaded JSON; they are not sent to a
backend. Each image may be up to 20 MB, and one thumbnail supports 1–35 images.
Transparent PNGs start without a backing frame and remain unframed when a
template is applied.

The editor offers scattered photos, a curved card ribbon, grid, and featured
photo layouts. New scattered layouts spread images across the canvas and fit
their outer corners. The ribbon starts with a transparent background. Drag the
selected image to move it; use the corner handle to resize and the top handle
to rotate. Numeric controls set position, size, frame ratio, rotation, zoom,
and crop focus. Whole-composition controls move, zoom, or fit all images
together; drag mode moves it with the pointer and uses the wheel to zoom. The layer
list controls selection and drawing order. Frames, color and grain effects,
shadow, canvas aspect ratio, and paper or solid backgrounds are saved in a
versioned JSON recipe. Older demo recipes without composition settings load
with a neutral transform. Paste a downloaded recipe into the import field to
render it again. PNG and JPG download from the same canvas renderer as the
preview, with pixel area at least equivalent to
1920×1080 and a maximum canvas dimension of 4096 pixels. JPG fills a
transparent background with white; PNG retains transparency.

The image picker lists at most twelve matches at a time and loads only the
highlighted photo preview after a short pause. Selected images decode two at a
time into bounded working previews. The layer list uses those decoded previews
without new image requests. Export loads originals one at a time so editing
does not retain all full-resolution images in memory.

The thumbnail maker uses bundled frontend libraries and browser canvas APIs;
it does not load scripts, fonts, or images from an internet CDN. Gallery photos
still need the local authenticated File and Stream Service routes. Uploaded
photos can be edited without those media routes after Moments has loaded.

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
The demo fixture stores `/f/` file URLs. `StreamContext` validates them and
derives the original `/s/`, transcoded `/s/transcode/`, and HEIC/HEIF JPEG
`/s/preview/` routes for the gallery. Components use standard HTML5 image and
video elements; the context owns the LiteNAS URL contract for future callers.
HEIC/HEIF wall, viewer, and thumbnail images use the JPEG preview. The gallery
combines an unambiguous same-folder HEIC/HEIF and MOV filename pair into one
Live Photo without waiting for media requests. It shows the still with a motion
control when the image loads, and shows the paired movie if the image fails.
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
