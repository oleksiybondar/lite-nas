// biome-ignore-all lint/complexity/useMaxParams: the compact fixture helper mirrors TIFF entry fields.
import { afterEach, expect, test, vi } from "vitest";
import { parseJpegExif } from "./GalleryExif";
import { googleMapsUrl } from "./GalleryMaps";
import { GalleryMetadataReader } from "./GalleryMetadataReader";

/** Builds a small, valid JPEG EXIF prefix with camera and GPS fields for boundary tests. */
const jpegWithGps = (): ArrayBuffer => {
  const buffer = new ArrayBuffer(304);
  const view = new DataView(buffer);
  const bytes = new Uint8Array(buffer);
  bytes.set([0xff, 0xd8, 0xff, 0xe1, 0x01, 0x2c, 0x45, 0x78, 0x69, 0x66, 0, 0]);
  const base = 12;
  const short = (offset: number, value: number): void => view.setUint16(base + offset, value, true);
  const long = (offset: number, value: number): void => view.setUint32(base + offset, value, true);
  const field = (offset: number, tag: number, type: number, count: number, value: number): void => {
    short(offset, tag);
    short(offset + 2, type);
    long(offset + 4, count);
    long(offset + 8, value);
  };
  bytes.set([0x49, 0x49], base);
  short(2, 42);
  long(4, 8);
  short(8, 2);
  field(10, 0x0110, 2, 6, 100);
  field(22, 0x8825, 4, 1, 120);
  bytes.set([0x43, 0x61, 0x6d, 0x65, 0x72, 0], base + 100);
  short(120, 4);
  field(122, 1, 2, 2, 0x4e);
  field(134, 2, 5, 3, 180);
  field(146, 3, 2, 2, 0x45);
  field(158, 4, 5, 3, 204);
  for (const [offset, parts] of [
    [180, [55, 45, 0]],
    [204, [12, 30, 0]],
  ] as const) {
    parts.forEach((part, index) => {
      long(offset + index * 8, part);
      long(offset + index * 8 + 4, 1);
    });
  }
  return buffer;
};

afterEach(() => vi.unstubAllGlobals());

test("extracts GPS coordinates from a JPEG EXIF segment", () => {
  expect(parseJpegExif(jpegWithGps())?.location).toEqual({ latitude: 55.75, longitude: 12.5 });
});

test("ignores an incomplete GPS direction", () => {
  const buffer = jpegWithGps();
  new DataView(buffer).setUint8(12 + 122 + 8, 0x58);
  expect(parseJpegExif(buffer)?.location).toBeUndefined();
});

test("rejects an incomplete JPEG prefix", () => {
  expect(parseJpegExif(new Uint8Array([0xff, 0xd8, 0xff]).buffer)).toBeNull();
});

test("builds a Google Maps link for a GPS pair", () => {
  expect(googleMapsUrl({ latitude: 55.75, longitude: 12.5 })).toBe(
    "https://www.google.com/maps/search/?api=1&query=55.750000%2C12.500000",
  );
});

test("rejects coordinates outside Earth bounds", () => {
  expect(googleMapsUrl({ latitude: 91, longitude: 12.5 })).toBeNull();
});

test("reads one bounded range from File Service after interest", async () => {
  const buffer = jpegWithGps();
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(buffer, {
      status: 206,
      headers: {
        "Content-Type": "image/jpeg",
        "Content-Range": "bytes 0-303/1000",
        "Content-Length": "304",
      },
    }),
  );
  vi.stubGlobal("fetch", fetchMock);
  const reader = new GalleryMetadataReader();
  await reader.load("/f/home/test/photo.jpeg");
  expect(fetchMock.mock.calls[0][1].headers.Range).toBe("bytes=0-65535");
  reader.dispose();
});

test("does not parse an unbounded full-file response", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(jpegWithGps())));
  const reader = new GalleryMetadataReader();
  expect(await reader.load("/f/home/test/photo.jpeg")).toBeNull();
  reader.dispose();
});

test("stops reading when a range body exceeds its declared limit", async () => {
  const oversized = new Uint8Array(305);
  oversized.set(new Uint8Array(jpegWithGps()));
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(oversized, {
        status: 206,
        headers: {
          "Content-Type": "image/jpeg",
          "Content-Range": "bytes 0-303/1000",
          "Content-Length": "304",
        },
      }),
    ),
  );
  const reader = new GalleryMetadataReader();
  expect(await reader.load("/f/home/test/photo.jpeg")).toBeNull();
  reader.dispose();
});

test("does not request metadata for an invalid File Service URL", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  const reader = new GalleryMetadataReader();
  await reader.load("https://example.com/photo.jpeg");
  expect(fetchMock).not.toHaveBeenCalled();
  reader.dispose();
});

test("a canceled viewer read can restart for the same file", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 206 }));
  vi.stubGlobal("fetch", fetchMock);
  const reader = new GalleryMetadataReader();
  void reader.load("/f/home/test/photo.jpeg");
  reader.cancel("/f/home/test/photo.jpeg");
  void reader.load("/f/home/test/photo.jpeg");
  expect(fetchMock).toHaveBeenCalledTimes(2);
  reader.dispose();
});
