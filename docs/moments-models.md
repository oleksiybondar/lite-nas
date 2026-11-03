# Moments local models

Moments uses two optional local models. BLIP generates captions for still images;
multilingual E5 ranks photo search results from text already attached to those
images. Both run on demand in separate Python processes. Missing dependencies or
weights do not prevent Moments from serving galleries.

| Model | Used when | Input | Saved result |
| --- | --- | --- | --- |
| BLIP base (`Salesforce/blip-image-captioning-base`) | An owner starts a caption scan | One authorized still image at a time | A caption keyed by source path |
| E5 small (`intfloat/multilingual-e5-small`) | A visible gallery search needs semantic ranking | Search text, captions, source tags, and person names | A 384-value vector keyed by source path and text hash |

## Installation and offline use

The local Moments deployment attempts to prepare both Python environments and
download pinned, checksum-checked model files. The BLIP files are installed at
`/usr/local/lib/lite-nas/moments/blip-model`; E5 files are installed at
`/usr/local/lib/lite-nas/moments/e5-model`. The Python environments are separate
`caption-venv` and `embedding-venv` directories under the same parent. Model
files are owned by root and readable by the Moments service group. A later
deployment reuses verified files. BLIP can also relocate an older verified copy
from `/var/lib/lite-nas/moments/blip-model`; a cross-filesystem relocation needs
one local copy.

The service sets `HF_HUB_OFFLINE=1` and `TRANSFORMERS_OFFLINE=1`. Both workers
load only local files at runtime and have no network access in their AppArmor
child profiles. Downloads happen during model setup, not during a caption scan
or search. Model installation is optional: without BLIP, caption jobs fail and
existing captions remain; without E5, search continues with text matching.

The configured worker scripts are in `/etc/lite-nas/moments.conf`:

```ini
caption_script=/usr/share/lite-nas/moments/caption_image.py
embedding_script=/usr/share/lite-nas/moments/embed_text.py
```

An empty `caption_script` disables BLIP; an empty `embedding_script` disables
semantic search. The normal local deployment entry point is
`./scripts/deploy-moments.sh` (run with the repository's deployment privileges).
Its model setup failure is reported as a warning and does not stop deployment.

## Caption jobs

The owner starts `POST /api/galleries/{galleryId}/caption-scans` with
`{"force":false}` to skip files that already have captions, or `{"force":true}`
to regenerate them. `GET /api/scans/{scanId}` reports progress and
`DELETE /api/scans/{scanId}` cancels the job. Captions are shared by galleries
that link the same physical file. BLIP is prompted to name the main visible
subject; species or other detailed identification is best effort.

At job start, Moments fixes an ordered list of up to 10,000 linked image paths
and asks Auth Service over NATS for an opaque grant limited to those paths and
File Service reads. File Service validates the grant with Auth Service for each
source read, then applies the user's current RBAC and filesystem permissions.
The grant lasts at most seven days. Browser logout does not stop a running job;
the grant cannot authorize other APIs or file mutations. Auth Service stores
only a hash of the grant. Moments stores the grant and source list in its SQLite
database so a service restart can resume the job. The grant is revoked when the
job finishes or is cancelled, with cleanup retried after a restart.

SQLite has a `resume_pending` flag for active caption jobs. The worker saves
each successful caption immediately, but checkpoints progress counters after
ten images or thirty minutes and at job end. A restart resumes from the last
counter checkpoint and can repeat images processed after it. Old active jobs
without a saved grant cannot resume. Media scans have a separate 20-minute
deadline and still use an in-memory user token; their interrupted jobs fail on
restart.

BLIP reads a source through File Service into bounded process memory and sends
it to Python through a pipe. The compressed source is capped at 64 MiB, the
input line at 96 MiB, and decoded images at 100 million pixels. Python reduces
the image to a 1024-pixel thumbnail before inference. The worker loads its
weights once per caption job and exits afterward. One model worker runs at a
time; the caption job deadline is seven days.

## Semantic search

For multiword or non-Latin `q`, Moments tries E5 ranking of image matches. A
single Latin word uses text matching first and tries semantic ranking only
when text matching finds no photos. Gallery result ordering and pagination
remain lexical. The worker receives bounded text, never original image bytes,
paths, tokens, or database handles. It processes up to 10,000 visible image
candidates in batches of 16 and has a 60-second request deadline. Moments
stores vectors for the model revision and the hash of the source text, so
unchanged text reuses its vector. Search falls back to text matching if the
model is unavailable, busy, or fails.

## Resource and storage limits

The systemd unit limits the combined Go service and model worker to
`MemoryHigh=2500M`, `MemoryMax=3200M`, `MemorySwapMax=0`, and `CPUQuota=200%`.
The model workers share one lock, preventing BLIP and E5 from loading together.
Moments already has a private, size-limited 64 MiB tmpfs at `/tmp`. Model
weights are read from protected storage, and image bytes pass through memory
and a pipe; neither is staged in `/tmp`. SQLite stores derived captions,
vectors, job state, and the fixed caption source list on persistent storage.

For a failed job, inspect its scan status and the Moments service journal:

```sh
systemctl status lite-nas-moments
journalctl -u lite-nas-moments -b
```

An unavailable BLIP worker produces a failed caption job while existing
captions remain. A rejected or expired grant fails the job. E5 failures leave
normal text search available. Device-level memory peaks and SQLite write
volume should be measured with a real caption job; unit tests do not establish
those hardware limits.
