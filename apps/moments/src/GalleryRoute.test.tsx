// @vitest-environment jsdom

import { cleanup, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, expect, test, vi } from "vitest";
import { GalleryRoute } from "./GalleryRoute";
import { getGallery } from "./moments-data";
import type { GalleryRecord } from "./schemas";
import { renderWithLanguage as render } from "./test-utils/language";

vi.mock("./moments-data", () => ({ getGallery: vi.fn() }));
vi.mock("./GalleryDetail", () => ({
  GalleryDetail: ({ selected }: { selected: GalleryRecord }) => <h1>{selected.name}</h1>,
}));

const gallery: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "ALBUM",
  name: "Shared album",
  description: "",
  ownerId: "alice",
  visibility: "PUBLIC",
  uploadDirectory: "/home/photos",
  wallRenderStrategy: "mosaic",
  thumbnailImageUrl: null,
  thumbnailPath: null,
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};

/** Opens one browser route with the authenticated gallery loader. */
function renderRoute(path: string): void {
  render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route
          path="/galleries/:id"
          element={<GalleryRoute userId="alice" maxConcurrent={10} workingDir="/home" />}
        />
      </Routes>
    </MemoryRouter>,
  );
}

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

test("a direct gallery URL loads its saved gallery", async () => {
  vi.mocked(getGallery).mockResolvedValue(gallery);
  renderRoute(`/galleries/${gallery.id}`);
  expect(await screen.findByRole("heading", { name: "Shared album" })).toBeTruthy();
});

test("an invalid gallery URL is rejected before the API request", async () => {
  renderRoute("/galleries/not-an-id");
  await screen.findByText("Gallery link is invalid.");
  expect(getGallery).not.toHaveBeenCalled();
});

test("an inaccessible gallery URL shows an error", async () => {
  vi.mocked(getGallery).mockRejectedValue(new Error("forbidden"));
  renderRoute(`/galleries/${gallery.id}`);
  expect(await screen.findByText("Could not open this gallery.")).toBeTruthy();
});
