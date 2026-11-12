// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: range validation and stream bounds are security checks.
import { request } from "./api";
import { type PhotoMetadata, parseJpegExif } from "./GalleryExif";
import { galleryFileUrlSchema } from "./gallery-model";

const prefixBytes = 64 * 1024;

/** Reads and bounds one chunk before appending it to the requested prefix. */
const appendChunk = (bytes: Uint8Array, value: Uint8Array, offset: number): number => {
  if (offset + value.byteLength > bytes.byteLength)
    throw new Error("JPEG metadata range exceeded its limit");
  bytes.set(value, offset);
  return offset + value.byteLength;
};

/** Collects at most the requested prefix even if a response misstates its size. */
const boundedBody = async (response: Response, expectedLength: number): Promise<ArrayBuffer> => {
  if (!response.body) throw new Error("JPEG range body is missing");
  const reader = response.body.getReader();
  const bytes = new Uint8Array(expectedLength);
  let offset = 0;
  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      offset = appendChunk(bytes, value, offset);
    }
  } catch (error) {
    await reader.cancel().catch(() => {});
    throw error;
  } finally {
    reader.releaseLock();
  }
  if (offset !== expectedLength) throw new Error("JPEG metadata range was incomplete");
  return bytes.buffer;
};

/** Reads only the leading EXIF portion of an authorized original JPEG. */
const readJpegMetadata = async (
  url: string,
  signal: AbortSignal,
): Promise<PhotoMetadata | null> => {
  const response = await request(url, {
    headers: { Range: `bytes=0-${prefixBytes - 1}` },
    signal,
  });
  const contentRange = response.headers.get("Content-Range")?.match(/^bytes 0-(\d+)\/(\d+)$/);
  const contentLength = Number(response.headers.get("Content-Length"));
  if (
    response.status !== 206 ||
    !/^image\/jpeg(?:;|$)/i.test(response.headers.get("Content-Type") ?? "") ||
    !contentRange ||
    !Number.isSafeInteger(contentLength) ||
    contentLength < 1 ||
    contentLength > prefixBytes ||
    Number(contentRange[1]) + 1 !== contentLength ||
    Number(contentRange[2]) < contentLength
  ) {
    await response.body?.cancel();
    throw new Error("File Service did not return a bounded JPEG range");
  }
  return parseJpegExif(await boundedBody(response, contentLength));
};

/** Shares one on-demand EXIF request and caches only small parsed metadata. */
export class GalleryMetadataReader {
  private readonly cache = new Map<string, PhotoMetadata | null>();
  private active: {
    url: string;
    controller: AbortController;
    promise: Promise<PhotoMetadata | null>;
  } | null = null;

  /** Returns cached metadata or undefined when a file has not been inspected. */
  cached = (url: string): PhotoMetadata | null | undefined => this.cache.get(url);

  /** Reads a JPEG header after explicit interest, canceling a previous read. */
  load = (url: string): Promise<PhotoMetadata | null> => {
    if (this.cache.has(url)) return Promise.resolve(this.cache.get(url) ?? null);
    if (!galleryFileUrlSchema.safeParse(url).success || !/\.jpe?g$/i.test(url))
      return Promise.resolve(null);
    if (this.active?.url === url) return this.active.promise;
    this.active?.controller.abort();
    const controller = new AbortController();
    const promise = readJpegMetadata(url, controller.signal)
      .then((metadata) => {
        if (!controller.signal.aborted) this.cache.set(url, metadata);
        return metadata;
      })
      .catch(() => null)
      .finally(() => {
        if (this.active?.controller === controller) this.active = null;
      });
    this.active = { url, controller, promise };
    return promise;
  };

  /** Stops an active read when its wall tile or viewer item changes. */
  cancel = (url: string): void => {
    if (this.active?.url === url) {
      this.active.controller.abort();
      this.active = null;
    }
  };

  /** Releases the active range request when the gallery unmounts. */
  dispose = (): void => {
    this.active?.controller.abort();
    this.cache.clear();
  };
}
