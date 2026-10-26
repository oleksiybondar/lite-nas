// @vitest-environment jsdom
import { cleanup, fireEvent, render, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { GalleryCover } from "./GalleryCover";
import { galleryFiles } from "./moments-data";
import type { GalleryRecord } from "./schemas";

vi.mock("./moments-data", () => ({ galleryFiles: vi.fn() }));

const gallery: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "ALBUM",
  name: "Photos",
  description: "",
  ownerId: "1001",
  visibility: "PRIVATE",
  uploadDirectory: "/home/Photos",
  wallRenderStrategy: "mosaic",
  thumbnailImageUrl: null,
  thumbnailPath: null,
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

test("a gallery without a saved cover uses its first image", async () => {
  vi.mocked(galleryFiles).mockResolvedValue(["/home/Photos/movie.mp4", "/home/Photos/photo.jpg"]);
  const view = render(<GalleryCover gallery={gallery} />);
  await waitFor(() =>
    expect(view.container.querySelector("img")?.getAttribute("src")).toBe(
      "/f/home/Photos/photo.jpg",
    ),
  );
});

test("an uploaded gallery cover takes priority over source images", () => {
  const thumbnailImageUrl = `/api/galleries/${gallery.id}/thumbnail/image`;
  const view = render(<GalleryCover gallery={{ ...gallery, thumbnailImageUrl }} />);
  expect(view.container.querySelector("img")?.getAttribute("src")).toBe(thumbnailImageUrl);
});

test("a selected NAS image is used as the card cover", () => {
  const view = render(
    <GalleryCover gallery={{ ...gallery, thumbnailPath: "/home/Photos/cover.HEIC" }} />,
  );
  expect(view.container.querySelector("img")?.getAttribute("src")).toBe(
    "/s/preview/home/Photos/cover.HEIC",
  );
});

test("a failed NAS cover uses the previously uploaded cover", () => {
  const thumbnailImageUrl = `/api/galleries/${gallery.id}/thumbnail/image`;
  const view = render(
    <GalleryCover
      gallery={{ ...gallery, thumbnailPath: "/home/Photos/cover.jpg", thumbnailImageUrl }}
    />,
  );
  const image = view.container.querySelector("img");
  if (!image) throw new Error("Selected cover image was not rendered.");
  fireEvent.error(image);
  expect(view.container.querySelector("img")?.getAttribute("src")).toBe(thumbnailImageUrl);
});

test("a failed uploaded cover falls back to the first gallery image", async () => {
  vi.mocked(galleryFiles).mockResolvedValue(["/home/Photos/photo.jpg"]);
  const thumbnailImageUrl = `/api/galleries/${gallery.id}/thumbnail/image`;
  const view = render(<GalleryCover gallery={{ ...gallery, thumbnailImageUrl }} />);
  const image = view.container.querySelector("img");
  if (!image) throw new Error("Saved cover image was not rendered.");
  fireEvent.error(image);
  await waitFor(() =>
    expect(view.container.querySelector("img")?.getAttribute("src")).toBe(
      "/f/home/Photos/photo.jpg",
    ),
  );
});
