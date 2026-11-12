import { expect, test, vi } from "vitest";
import {
  arrangeThumbnail,
  exportDimensions,
  fitThumbnailComposition,
  newThumbnailLayer,
  thumbnailDocumentSchema,
} from "./thumbnail-model";

const source = "/f/home/alice/Pictures/photo.png";

// Requirements: moments/FR-003. A portable recipe accepts 35 independently placed images.
test("thumbnail JSON accepts 35 images", () => {
  const layers = Array.from({ length: 35 }, (_, index) => ({
    ...newThumbnailLayer(source, "Photo"),
    id: String(index),
  }));
  expect(
    thumbnailDocumentSchema.safeParse({
      version: 1,
      aspect: "16:9",
      background: "transparent",
      layers,
    }).success,
  ).toBe(true);
});

// Requirements: moments/FR-003. A recipe rejects more than the editor limit.
test("thumbnail JSON rejects 36 images", () => {
  const layers = Array.from({ length: 36 }, (_, index) => ({
    ...newThumbnailLayer(source, "Photo"),
    id: String(index),
  }));
  expect(
    thumbnailDocumentSchema.safeParse({ version: 1, aspect: "16:9", background: "white", layers })
      .success,
  ).toBe(false);
});

// Requirements: moments/FR-003. Imported recipes cannot pull images from arbitrary internet hosts.
test("thumbnail JSON rejects external URLs", () => {
  const layers = [{ ...newThumbnailLayer("https://example.com/photo.png", "Photo"), id: "one" }];
  expect(
    thumbnailDocumentSchema.safeParse({ version: 1, aspect: "16:9", background: "white", layers })
      .success,
  ).toBe(false);
});

// Requirements: moments/FR-003. Embedded transparent uploads remain unframed after arranging.
test("templates leave uploaded PNGs without a backing frame", () => {
  const layer = newThumbnailLayer("data:image/png;base64,aGVsbG8=", "Sticker");
  expect(arrangeThumbnail([layer], "scatter")[0].frame).toBe("none");
});

// Requirements: moments/FR-003. Portrait exports preserve their ratio at the canvas size cap.
test("portrait export keeps its aspect ratio", () => {
  const [width, height] = exportDimensions("9:16", 10000);
  expect(width / height).toBe(9 / 16);
});

// Requirements: moments/FR-003. Export resolution is never below Full HD area.
test("minimum export area is Full HD equivalent", () => {
  const [width, height] = exportDimensions("1:1");
  expect(width * height).toBeGreaterThanOrEqual(1920 * 1080);
});

// Requirements: moments/FR-003. Old demo recipes load with an unchanged composition.
test("thumbnail JSON defaults a missing composition", () => {
  const parsed = thumbnailDocumentSchema.parse({
    version: 1,
    aspect: "16:9",
    background: "white",
    layers: [newThumbnailLayer(source, "Photo")],
  });
  expect(parsed.composition).toEqual({ x: 0, y: 0, scale: 1 });
});

// Requirements: moments/FR-003. Fitting a narrow arrangement can fill more canvas area.
test("fit composition enlarges a small arrangement", () => {
  const layer = { ...newThumbnailLayer(source, "Photo"), x: 40, y: 40, width: 20, height: 20 };
  expect(fitThumbnailComposition([layer], "16:9").scale).toBeGreaterThan(1);
});

// Requirements: moments/FR-003. A single scattered photo spans most of the canvas.
test("scattered single photo uses most of the canvas", () => {
  const layer = arrangeThumbnail([newThumbnailLayer(source, "Photo")], "scatter")[0];
  expect(layer.width * layer.height).toBeGreaterThan(5500);
});

// Requirements: moments/FR-003. Typical random output spreads across the canvas.
test("five scattered photos span most of the canvas", () => {
  const random = vi.spyOn(Math, "random").mockReturnValue(0.5);
  const layers = arrangeThumbnail(
    Array.from({ length: 5 }, () => newThumbnailLayer(source, "Photo")),
    "scatter",
  );
  random.mockRestore();
  const left = Math.min(...layers.map((layer) => layer.x));
  const right = Math.max(...layers.map((layer) => layer.x + layer.width));
  const top = Math.min(...layers.map((layer) => layer.y));
  const bottom = Math.max(...layers.map((layer) => layer.y + layer.height));
  expect((right - left) * (bottom - top)).toBeGreaterThan(6500);
});
