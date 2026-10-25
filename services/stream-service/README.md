# Stream Service

Stream Service serves an authorized source file through a loopback HTTP listener.
Nginx mounts it at `/s/` on both the admin and Moments HTTPS origins. The suffix
after `/s/` is the absolute host path: `/s/home/alice/movie.mov` addresses
`/home/alice/movie.mov`. The service also accepts a Bearer token for REST clients;
browsers use the existing Secure `lite-nas-at` cookie.

Every GET or HEAD request validates the access token through Auth, resolves its
username to a UID through RBAC, and asks RBAC to check read access to the resolved
path. There is no administrator bypass. A denied or missing path returns 404 after
authentication; a missing or invalid token returns 401. Failed Auth or RBAC
dependencies return 503. Neither a media URL nor a prior range grant carries
authorization into the next request.

## Original playback

`GET /s/<absolute-host-path>` and `HEAD /s/<absolute-host-path>` serve the stored
bytes without rewriting the source. GET supports one `Range: bytes=...` range,
including open-ended and suffix ranges. Valid ranges return 206 with
`Content-Range` and `Content-Length`; malformed or unsatisfiable ranges return
416. The response uses `Cache-Control: private, no-store`, `nosniff`, and a
fixed extension-to-content-type list for common passive media. Unknown formats
are delivered as `application/octet-stream`. HEIC/HEIF stills and MOV movies
remain independent original assets; no Live Photo pairing is asserted without
source association metadata.

## Browser fallback

Moments tries original video playback first. On a media error it requests
`GET /s/transcode/<absolute-host-path>?start=<seconds>`. `start` is an optional
whole number from 0 through 86400; the default is zero. The response is a
fragmented MP4 stream with H.264 video and AAC audio. The Moments controls
restart it 30 seconds earlier or later by requesting a fresh URL. Browser Range
headers on this route receive the sequential 200 response; byte ranges do not
apply to transformed output. Every restart repeats Auth, RBAC, and path
checks. The transcoder admits AVI, MOV/MP4, Matroska/WebM, MPEG, ASF, FLV, and
Ogg containers; it excludes playlist demuxers that could load another path.
A codec that the installed ffmpeg cannot decode still fails; original delivery
does not promise browser decoding.

HEIC/HEIF stills use `GET /s/preview/<absolute-host-path>` for a JPEG browser
preview. The same Auth, RBAC, path, and inode checks run for every request.
The original HEIC remains available through File Service for downloading.
The host `heif-dec` tool with its libde265 plugin composes tiled HEIC images;
conversion writes to a private 64 MiB tmpfs and removes the JPEG after the
response. The Debian package and local runtime installer include both packages.

Moments derives these routes from its `/f/` file URL in `StreamContext` and
loads them through standard HTML5 media elements. It displays HEIC/HEIF wall
and viewer images from the original `/s/` URL.
If the browser cannot render one, it switches that item to an HTML5 video using
the existing `/s/transcode/` URL. A play icon at the top left lets a user make
the same switch after the original image renders. The thumbnail strip continues
to use the JPEG preview. If video playback fails, the wall or viewer shows that
JPEG still. The video converter composes the HEIC still and emits
three one-second H.264 frames. This is a static video: these HEIC files contain
no motion frames, and the service cannot reconstruct Live Photo motion without
a separate movie component. A HEIC video conversion accepts `start=0` only.

The host `/usr/bin/ffmpeg` supplies codecs through the OS package manager.
The verified Unix user's helper opens the approved inode and gives ffmpeg a
seekable descriptor for source video. It writes its MP4 to the HTTP response
through memory pipes. Video sources do not create a temporary media file;
HEIC conversion uses the private JPEG temporary file described above. Each process uses at
most two codec threads, the service admits two conversions at once, and
systemd limits the service tree to 1 GiB memory and 200% CPU. Client disconnect
cancels the process. An ffmpeg error before output returns 422; a missing
ffmpeg executable returns 503.

## Host access

The dedicated `lite-nas-stream-service` account is non-root and joins
`lite-nas-fs-rd`. Deployment provisions read/search ACLs for `/home` and
currently mounted non-system ZFS trees. The handler discovers mounted ZFS
roots for each request and excludes OS-critical trees even if they are ZFS.
An unreadable private file is opened by a subprocess running as the verified
non-root Unix user through the fixed sudo rule. It checks the opened inode
against the inode seen before the RBAC decision. The group and sudo rule give
the process host access only; RBAC remains the caller-specific decision.
Run `scripts/deploy-stream-service.sh` after creating or mounting a new ZFS
dataset so its existing files receive the service read ACL. Default directory
ACLs cover new descendants until owners change their ACL masks; the per-user
reader handles private files whose masks remove group read access.
ACL provisioning is best effort and never prevents the service from starting.
Deployment skips ZFS mounts with ACLs disabled (`acltype=off`) because
`setfacl` cannot provision the read group there. New pools are discovered when
mounted; rerun deployment to grant the group access to existing files on an
ACL-capable pool. Existing Unix mode and traversal rights may still allow reads
without that grant; private paths on an ACL-disabled pool remain inaccessible
to this non-root service.

Local deployment runs `scripts/install-runtime-dependencies.sh`, which installs
ffmpeg if missing. The single `lite-nas` Debian package depends on ffmpeg.
