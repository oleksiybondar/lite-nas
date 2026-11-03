// @vitest-environment jsdom

import { cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import type { GalleryItem } from "./gallery-model";
import { ThumbnailMaker } from "./ThumbnailMaker";
import { renderWithLanguage as render } from "./test-utils/language";

vi.mock("./ThumbnailCanvas", () => ({
  beginThumbnail: vi.fn(() => ({})),
  drawThumbnailLayer: vi.fn(),
  loadOriginalThumbnailImage: vi.fn(async () => ({
    source: document.createElement("canvas"),
    width: 800,
    height: 600,
    originalWidth: 800,
    originalHeight: 600,
  })),
  loadThumbnailPreviews: vi.fn(
    async (recipe: { layers: { id: string }[] }) =>
      new Map(
        recipe.layers.map((layer) => [
          layer.id,
          {
            source: document.createElement("canvas"),
            width: 800,
            height: 600,
            originalWidth: 800,
            originalHeight: 600,
          },
        ]),
      ),
  ),
  renderThumbnail: vi.fn(),
  thumbnailExportSize: vi.fn(() => [800, 450]),
}));

const photo: GalleryItem = {
  id: "photo-1",
  url: "/f/home/Photos/photo.jpg",
  filename: "photo.jpg",
  kind: "image",
};

beforeEach(() => {
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({
    clearRect: vi.fn(),
    drawImage: vi.fn(),
  } as unknown as CanvasRenderingContext2D);
  vi.spyOn(HTMLCanvasElement.prototype, "toBlob").mockImplementation((callback) => {
    callback(new Blob(["cover"], { type: "image/png" }));
  });
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

test("random gallery photo can be generated and uploaded as a cover", async () => {
  const onSave = vi.fn().mockResolvedValue(undefined);
  render(<ThumbnailMaker items={[photo]} onSave={onSave} />);
  fireEvent.click(screen.getByRole("button", { name: "Random" }));
  fireEvent.click(await screen.findByRole("button", { name: "Save as gallery cover" }));
  await waitFor(() => expect(onSave).toHaveBeenCalledWith(expect.any(Blob)));
});

test("a gallery photo can be chosen from the searchable image picker", async () => {
  render(<ThumbnailMaker items={[photo]} onSave={vi.fn()} />);
  fireEvent.change(screen.getByLabelText("Add gallery photo"), { target: { value: "photo" } });
  fireEvent.click(await screen.findByRole("option", { name: "photo.jpg" }));
  expect(await screen.findByText("Selected image: photo.jpg")).toBeTruthy();
});

test("a generated cover keeps the selected template and canvas controls usable", async () => {
  render(<ThumbnailMaker items={[photo]} onSave={vi.fn()} />);
  fireEvent.click(screen.getByRole("button", { name: "Random" }));
  fireEvent.mouseDown(screen.getByLabelText("Template"));
  fireEvent.click(screen.getByRole("option", { name: "Neat grid" }));
  fireEvent.click(screen.getByRole("button", { name: "Apply template to selected images" }));
  fireEvent.mouseDown(screen.getByLabelText("Aspect ratio"));
  fireEvent.click(screen.getByRole("option", { name: "1:1" }));
  fireEvent.mouseDown(screen.getByLabelText("Background"));
  fireEvent.click(screen.getByRole("option", { name: "White" }));
  fireEvent.click(screen.getByRole("button", { name: "Zoom in" }));
  fireEvent.click(screen.getByRole("button", { name: "Fit composition" }));
  expect(await screen.findByText("Selected image: photo.jpg")).toBeTruthy();
});

test("a browser image can be uploaded and saved as a cover", async () => {
  const onSave = vi.fn().mockResolvedValue(undefined);
  render(<ThumbnailMaker items={[]} onSave={onSave} />);
  const input = document.querySelector('input[type="file"]');
  if (!input) throw new Error("Upload input was not rendered.");
  fireEvent.change(input, {
    target: { files: [new File(["image"], "local.png", { type: "image/png" })] },
  });
  await screen.findByText("Selected image: local.png");
  fireEvent.click(screen.getByRole("button", { name: "Save as gallery cover" }));
  await waitFor(() => expect(onSave).toHaveBeenCalledWith(expect.any(Blob)));
});

test("a selected photo can be styled and positioned", async () => {
  render(<ThumbnailMaker items={[photo]} onSave={vi.fn()} />);
  fireEvent.click(screen.getByRole("button", { name: "Random" }));
  fireEvent.mouseDown(screen.getByLabelText("Frame"));
  fireEvent.click(screen.getByRole("option", { name: "Polaroid" }));
  fireEvent.mouseDown(screen.getByLabelText("Effect"));
  fireEvent.click(screen.getByRole("option", { name: "Sepia" }));
  fireEvent.click(screen.getByRole("switch", { name: "Shadow" }));
  const position = screen.getAllByLabelText("X %")[1];
  fireEvent.focus(position);
  fireEvent.change(position, { target: { value: "25" } });
  fireEvent.keyDown(position, { key: "Enter" });
  expect((screen.getAllByLabelText("X %")[1] as HTMLInputElement).value).toBe("25");
});

test("selected gallery photos can be reordered and removed", async () => {
  const second = {
    ...photo,
    id: "photo-2",
    url: "/f/home/Photos/second.jpg",
    filename: "second.jpg",
  };
  render(<ThumbnailMaker items={[photo, second]} onSave={vi.fn()} />);
  fireEvent.click(screen.getByRole("button", { name: "Random" }));
  fireEvent.click(screen.getByRole("button", { name: "Send back" }));
  fireEvent.click(screen.getByRole("button", { name: "Bring forward" }));
  fireEvent.click(screen.getByRole("button", { name: "Remove" }));
  expect(await screen.findByText(/Images \(1\//)).toBeTruthy();
});

test("a generated recipe can be downloaded as JSON", async () => {
  const objectUrl = vi.fn(() => "blob:thumbnail");
  /** Supplies browser blob URL methods for the download interaction. */
  class TestURL extends URL {
    static createObjectURL = objectUrl;
    static revokeObjectURL = vi.fn();
  }
  vi.stubGlobal("URL", TestURL);
  vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => undefined);
  render(<ThumbnailMaker items={[photo]} onSave={vi.fn()} />);
  fireEvent.click(screen.getByRole("button", { name: "Random" }));
  fireEvent.click(screen.getByRole("button", { name: "Download JSON" }));
  await vi.waitFor(() => expect(TestURL.revokeObjectURL).toHaveBeenCalledWith("blob:thumbnail"), {
    timeout: 2000,
  });
});

test("unsupported browser uploads report their accepted image formats", async () => {
  render(<ThumbnailMaker items={[]} onSave={vi.fn()} />);
  const input = document.querySelector('input[type="file"]');
  if (!input) throw new Error("Upload input was not rendered.");
  fireEvent.change(input, {
    target: { files: [new File(["text"], "notes.txt", { type: "text/plain" })] },
  });
  expect(await screen.findByText("Use PNG, JPEG, WebP, or GIF images.")).toBeTruthy();
});

test("invalid imported thumbnail JSON is rejected", async () => {
  render(<ThumbnailMaker items={[]} onSave={vi.fn()} />);
  fireEvent.change(screen.getByLabelText("Paste thumbnail JSON to load"), {
    target: { value: "{invalid" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Load JSON" }));
  expect(await screen.findByRole("alert")).toBeTruthy();
});
