# Moments browser measurements

Measured on 2026-10-24 against a local production build of Moments. This was
not a deployed LiteNAS instance.

## Setup

- Chromium 147.0.7727.0, headless, with its default viewport.
- The built app and its 139-item demo fixture were served on loopback with mock
  authentication. No backend gallery API or deployed Nginx/File Service was used.
- The mock server returned `Cache-Control: no-store`. One pass served 8 × 8 JPEGs
  to isolate wall lifecycle behavior; another served the fixture's real local
  original JPEGs. No network throttling was applied.
- Measurements were taken after 2.5 seconds per stop with tiny images and
  5 seconds per stop with real originals. JavaScript garbage collection was
  requested before each snapshot. Renderer memory is proportional set size
  (PSS) from Linux `smaps_rollup`, so it includes more than JavaScript objects.

## Scroll behavior with tiny images

| Wall position | Mounted tiles | Image elements | Cumulative media requests | DOM nodes | Renderer PSS |
| --- | ---: | ---: | ---: | ---: | ---: |
| Top | 7 | 5 | 5 | 145 | 109 MiB |
| Middle | 10 | 5 | 10 | 178 | 111 MiB |
| Bottom | 6 | 4 | 14 | 134 | 111 MiB |
| Back to top | 7 | 5 | 19 | 145 | 111 MiB |

Only the visible group of images was requested at each stop. Returning to the
top restored the original tile and DOM node counts. The five top images were
requested again because the mock server disabled caching.

## Scroll behavior with original JPEGs

| Wall position | Mounted tiles | Loaded originals | Cumulative bytes served | Renderer PSS | JS heap |
| --- | ---: | ---: | ---: | ---: | ---: |
| Top | 7 | 5 | 83.2 MiB | 634 MiB | 5.5 MiB |
| Bottom | 6 | 4 | 108.4 MiB | 496 MiB | 5.5 MiB |
| Back to top | 7 | 5 | 191.5 MiB | 658 MiB | 5.6 MiB |

The first five originals contain about 463 MiB of pixels at four bytes per
pixel before any browser-specific image overhead. Their compressed files total
83.2 MiB. Removing them from the wall coincided with a 138 MiB drop in renderer
PSS, but browser memory did not return to the tiny-image baseline. On returning
to the top, those five files were downloaded again and renderer PSS rose by
162 MiB. The browser can retain decoded or cached image memory after DOM removal;
these snapshots do not prove that every offscreen byte was freed.

## Layout changes and video requests

Switching the dropdown through justified rows, masonry, even grid, and back to
the same layouts gave 7–10 mounted tiles and 114–156 attached elements at the
top of the wall. Repeating the layouts produced the same attached-element counts
and approximately 81–85 MiB renderer PSS with tiny images. One or two video
elements appeared in some layouts, but none had a source URL before Play.

Chromium's broader DOM counter reached 502 nodes in masonry while only 156
elements were attached. It returned to the same value on a second masonry pass,
so this short test did not show continuing growth. A longer session would be
needed to classify the detached nodes.

## Findings and limits

The browser run supports that wall virtualization removes offscreen elements
and avoids requesting the whole collection. It also shows that full-resolution
originals can occupy hundreds of MiB even when displayed as small wall tiles.
Smaller media representations would be the main way to reduce that cost.

These measurements do not establish peak request concurrency, long-session
memory behavior, video playback cost, or deployed File Service performance.
The mock server's `no-store` response makes revisit downloads a controlled
worst case rather than a prediction of deployed browser caching.
