// @vitest-environment jsdom
import { afterEach, expect, test, vi } from "vitest";
import { loadThumbnailPreviews, renderThumbnail, thumbnailExportSize } from "./ThumbnailCanvas";
import { newThumbnailLayer, thumbnailDocumentSchema } from "./thumbnail-model";

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

// Requirements: moments/FR-003. Preparing selected previews never decodes more than two originals together.
test("preview loading limits concurrent original decodes", async () => {
  let active = 0;
  let peak = 0;
  /** Simulates a slow browser decode so overlapping requests can be counted. */
  class DeferredImage {
    src = "";
    naturalWidth = 2400;
    naturalHeight = 1200;
    /** Completes one controlled asynchronous image decode. */
    async decode(): Promise<void> {
      active += 1;
      peak = Math.max(peak, active);
      await new Promise((resolve) => setTimeout(resolve, 0));
      active -= 1;
    }
  }
  vi.stubGlobal("Image", DeferredImage);
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({
    drawImage: vi.fn(),
  } as unknown as CanvasRenderingContext2D);
  const layers = Array.from({ length: 6 }, (_, index) => ({
    ...newThumbnailLayer(`/f/home/alice/Pictures/${index}.jpg`, "Photo"),
    id: String(index),
  }));
  const document = thumbnailDocumentSchema.parse({
    version: 1,
    aspect: "16:9",
    background: "white",
    layers,
  });
  await loadThumbnailPreviews(document, new Map(), () => {}, new AbortController().signal);
  expect(peak).toBe(2);
});

test("gallery cover rendering draws a styled photo onto the output canvas", () => {
  const drawImage = vi.fn();
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({
    beginPath: vi.fn(),
    clip: vi.fn(),
    drawImage,
    fill: vi.fn(),
    fillRect: vi.fn(),
    restore: vi.fn(),
    rotate: vi.fn(),
    roundRect: vi.fn(),
    save: vi.fn(),
    scale: vi.fn(),
    translate: vi.fn(),
  } as unknown as CanvasRenderingContext2D);
  const layer = {
    ...newThumbnailLayer("/f/home/Photos/photo.jpg", "photo.jpg"),
    frame: "polaroid" as const,
    effect: "grain" as const,
    shadow: true,
  };
  const recipe = thumbnailDocumentSchema.parse({
    version: 1,
    aspect: "16:9",
    background: "paper",
    layers: [layer],
  });
  const image = {
    source: document.createElement("canvas"),
    width: 800,
    height: 600,
    originalWidth: 800,
    originalHeight: 600,
  };
  renderThumbnail(document.createElement("canvas"), recipe, new Map([[layer.id, image]]), 800);
  expect(drawImage).toHaveBeenCalledOnce();
});

test("export size uses original photo resolution", () => {
  const layer = newThumbnailLayer("/f/home/Photos/photo.jpg", "photo.jpg");
  const recipe = thumbnailDocumentSchema.parse({
    version: 1,
    aspect: "16:9",
    background: "white",
    layers: [layer],
  });
  const image = {
    source: document.createElement("canvas"),
    width: 800,
    height: 600,
    originalWidth: 2400,
    originalHeight: 1800,
  };
  const [width] = thumbnailExportSize(recipe, new Map([[layer.id, image]]));
  expect(width).toBeGreaterThan(800);
});
