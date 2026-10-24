// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: bounds checks are kept next to each untrusted TIFF/JPEG read.
/** Descriptive fields found in an original JPEG's EXIF header. */
export type PhotoMetadata = {
  capturedAt?: string;
  camera?: string;
  location?: { latitude: number; longitude: number };
};

type ExifField = { type: number; count: number; offset: number };
type Tiff = { view: DataView; base: number; limit: number; little: boolean };
const fieldSizes: Record<number, number> = { 1: 1, 2: 1, 3: 2, 4: 4, 5: 8 };

/** Checks whether a TIFF field points entirely inside its containing payload. */
const fieldOffset = (tiff: Tiff, entry: number, length: number): number | null => {
  if (length <= 4) return entry + 8;
  const offset = tiff.base + tiff.view.getUint32(entry + 8, tiff.little);
  return offset >= tiff.base && offset + length <= tiff.limit ? offset : null;
};

/** Reads a bounded TIFF field without trusting file-supplied offsets. */
const readField = (tiff: Tiff, entry: number): { tag: number; field: ExifField } | null => {
  const { view, base, limit, little } = tiff;
  if (entry < base || entry + 12 > limit) return null;
  const tag = view.getUint16(entry, little);
  const type = view.getUint16(entry + 2, little);
  const count = view.getUint32(entry + 4, little);
  const size = fieldSizes[type];
  if (!size || count === 0 || count > 4096) return null;
  const length = count * size;
  const offset = fieldOffset(tiff, entry, length);
  if (offset === null) return null;
  return { tag, field: { type, count, offset } };
};

/** Reads one TIFF directory, rejecting truncated or excessive entry tables. */
const readIfd = (tiff: Tiff, relativeOffset: number): Map<number, ExifField> | null => {
  const start = tiff.base + relativeOffset;
  const count = directoryCount(tiff, start);
  if (count === null) return null;
  const fields = new Map<number, ExifField>();
  for (let index = 0; index < count; index += 1) {
    const entry = readField(tiff, start + 2 + index * 12);
    if (entry) fields.set(entry.tag, entry.field);
  }
  return fields;
};

/** Validates and returns the number of entries in a bounded TIFF directory. */
const directoryCount = (tiff: Tiff, start: number): number | null => {
  if (start < tiff.base || start + 2 > tiff.limit) return null;
  const count = tiff.view.getUint16(start, tiff.little);
  return count <= 128 && start + 2 + count * 12 <= tiff.limit ? count : null;
};

/** Reads an ASCII EXIF field while discarding its terminating NUL. */
const readAscii = (tiff: Tiff, field?: ExifField): string | null => {
  if (!field || field.type !== 2 || field.count > 128) return null;
  const bytes = new Uint8Array(tiff.view.buffer, tiff.view.byteOffset + field.offset, field.count);
  return (
    String.fromCharCode(...bytes)
      .split("\0", 1)[0]
      .trim() || null
  );
};

/** Reads a single LONG pointer into another TIFF directory. */
const readPointer = (tiff: Tiff, field?: ExifField): number | null =>
  field?.type === 4 && field.count === 1 ? tiff.view.getUint32(field.offset, tiff.little) : null;

/** Converts three unsigned EXIF rationals from degrees, minutes, and seconds. */
const readRational = (tiff: Tiff, offset: number): number | null => {
  const numerator = tiff.view.getUint32(offset, tiff.little);
  const denominator = tiff.view.getUint32(offset + 4, tiff.little);
  return denominator === 0 ? null : numerator / denominator;
};

/** Converts three unsigned EXIF rationals from degrees, minutes, and seconds. */
const readDegrees = (tiff: Tiff, field?: ExifField): number | null => {
  if (!field || field.type !== 5 || field.count !== 3) return null;
  const values = Array.from({ length: 3 }, (_, index) =>
    readRational(tiff, field.offset + index * 8),
  );
  if (values.some((value) => value === null)) return null;
  const [degrees, minutes, seconds] = values as [number, number, number];
  if (minutes >= 60 || seconds >= 60) return null;
  return degrees + minutes / 60 + seconds / 3600;
};

/** Accepts only supported GPS direction references. */
const validGpsReferences = (lat: string | null, lon: string | null): lat is "N" | "S" =>
  (lat === "N" || lat === "S") && (lon === "E" || lon === "W");

/** Reads valid latitude, longitude, and direction fields from the GPS directory. */
const gpsCoordinates = (tiff: Tiff, gps: Map<number, ExifField>): PhotoMetadata["location"] => {
  const latRef = readAscii(tiff, gps.get(1));
  const lonRef = readAscii(tiff, gps.get(3));
  const lat = readDegrees(tiff, gps.get(2));
  const lon = readDegrees(tiff, gps.get(4));
  if (!validGpsReferences(latRef, lonRef) || lat === null || lon === null) return undefined;
  if (lat > 90 || lon > 180) return undefined;
  return { latitude: latRef === "S" ? -lat : lat, longitude: lonRef === "W" ? -lon : lon };
};

/** Extracts a validated coordinate pair only when both GPS directions exist. */
const readLocation = (tiff: Tiff, ifd0: Map<number, ExifField>): PhotoMetadata["location"] => {
  const pointer = readPointer(tiff, ifd0.get(0x8825));
  if (pointer === null) return undefined;
  const gps = readIfd(tiff, pointer);
  return gps ? gpsCoordinates(tiff, gps) : undefined;
};

/** Reads the capture time without inventing a timezone absent from EXIF. */
const readCapturedAt = (tiff: Tiff, ifd0: Map<number, ExifField>): string | undefined => {
  const pointer = readPointer(tiff, ifd0.get(0x8769));
  if (pointer === null) return undefined;
  const exif = readIfd(tiff, pointer);
  const value = exif ? readAscii(tiff, exif.get(0x9003)) : null;
  const match = value?.match(/^(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})$/);
  return match ? `${match[1]}-${match[2]}-${match[3]} ${match[4]}:${match[5]}` : undefined;
};

/** Creates a TIFF reader when the header has a supported byte order and marker. */
const createTiff = (view: DataView, base: number, limit: number): Tiff | null => {
  if (base + 8 > limit) return null;
  const marker = view.getUint16(base, false);
  if (marker !== 0x4949 && marker !== 0x4d4d) return null;
  const tiff: Tiff = { view, base, limit, little: marker === 0x4949 };
  return view.getUint16(base + 2, tiff.little) === 42 ? tiff : null;
};

/** Parses the TIFF payload from an EXIF APP1 segment. */
const parseTiff = (view: DataView, base: number, limit: number): PhotoMetadata | null => {
  const tiff = createTiff(view, base, limit);
  if (!tiff) return null;
  const ifd0 = readIfd(tiff, view.getUint32(base + 4, tiff.little));
  if (!ifd0) return null;
  const camera = readAscii(tiff, ifd0.get(0x0110)) ?? undefined;
  const capturedAt = readCapturedAt(tiff, ifd0);
  const location = readLocation(tiff, ifd0);
  return camera || capturedAt || location ? { camera, capturedAt, location } : null;
};

/** Extracts metadata from the prefix of an original JPEG, without decoding image pixels. */
const isExifSegment = (view: DataView, offset: number): boolean =>
  view.getUint8(offset) === 0xff &&
  view.getUint8(offset + 1) === 0xe1 &&
  offset + 10 <= view.byteLength &&
  view.getUint32(offset + 4, false) === 0x45786966 &&
  view.getUint16(offset + 8, false) === 0;

/** Scans JPEG segments until EXIF metadata or image data begins. */
const findExif = (view: DataView): PhotoMetadata | null => {
  let offset = 2;
  while (offset + 4 <= view.byteLength) {
    if (view.getUint8(offset) !== 0xff) return null;
    const marker = view.getUint8(offset + 1);
    if (marker === 0xda || marker === 0xd9) return null;
    const length = view.getUint16(offset + 2, false);
    if (length < 2) return null;
    const end = offset + 2 + length;
    if (isExifSegment(view, offset))
      return parseTiff(view, offset + 10, Math.min(end, view.byteLength));
    if (end > view.byteLength) return null;
    offset = end;
  }
  return null;
};

/** Extracts metadata from the prefix of an original JPEG, without decoding image pixels. */
export const parseJpegExif = (buffer: ArrayBuffer): PhotoMetadata | null => {
  const view = new DataView(buffer);
  if (view.byteLength < 4 || view.getUint16(0, false) !== 0xffd8) return null;
  return findExif(view);
};
