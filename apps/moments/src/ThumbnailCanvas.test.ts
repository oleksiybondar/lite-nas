// @vitest-environment jsdom
import { afterEach, expect, test, vi } from "vitest";
import { loadThumbnailPreviews } from "./ThumbnailCanvas";
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
