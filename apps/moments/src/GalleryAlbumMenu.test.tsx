// @vitest-environment jsdom

import { cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { GalleryAlbumMenu } from "./GalleryAlbumMenu";
import {
  addCollectionDirectory,
  listGalleries,
  listReferences,
  removeCollectionDirectory,
} from "./moments-data";
import type { GalleryRecord } from "./schemas";
import { renderWithLanguage as render } from "./test-utils/language";

vi.mock("./moments-data", () => ({
  addCollectionDirectory: vi.fn(),
  listGalleries: vi.fn(),
  listReferences: vi.fn(),
  removeCollectionDirectory: vi.fn(),
}));

const album: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "ALBUM",
  name: "Photos",
  description: "",
  ownerId: "alice",
  visibility: "PUBLIC",
  uploadDirectory: "/NASPool/Photos",
  wallRenderStrategy: "masonry",
  thumbnailImageUrl: null,
  thumbnailPath: null,
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};
const collection: GalleryRecord = {
  ...album,
  id: "00000000-0000-4000-8000-000000000002",
  type: "COLLECTION",
  name: "Trips",
  uploadDirectory: null,
};

/** Opens an album's collection menu after the owned choices have loaded. */
async function openMenu(): Promise<void> {
  render(<GalleryAlbumMenu gallery={album} userId="alice" />);
  fireEvent.click(screen.getByRole("button", { name: "Collection options for Photos" }));
  await screen.findByRole("menu");
}

beforeEach(() => {
  vi.mocked(listGalleries).mockResolvedValue({
    items: [collection],
    metadata: { page: 1, size: 20, total_count: 1, total_pages: 1 },
  });
  vi.mocked(listReferences).mockResolvedValue([]);
  vi.mocked(addCollectionDirectory).mockResolvedValue(undefined);
  vi.mocked(removeCollectionDirectory).mockResolvedValue(undefined);
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

test("album menu offers owned collections beyond the first result page", async () => {
  vi.mocked(listGalleries)
    .mockResolvedValueOnce({
      items: [],
      metadata: { page: 1, size: 20, total_count: 21, total_pages: 2 },
    })
    .mockResolvedValueOnce({
      items: [collection],
      metadata: { page: 2, size: 20, total_count: 21, total_pages: 2 },
    });
  await openMenu();
  expect(await screen.findByRole("menuitem", { name: "Add to Trips" })).toBeTruthy();
});

test("album menu adds its source directory to an owned collection", async () => {
  await openMenu();
  fireEvent.click(await screen.findByRole("menuitem", { name: "Add to Trips" }));
  await waitFor(() =>
    expect(addCollectionDirectory).toHaveBeenCalledWith(collection.id, "/NASPool/Photos"),
  );
});

test("adding an album directory changes the menu action to remove", async () => {
  await openMenu();
  fireEvent.click(await screen.findByRole("menuitem", { name: "Add to Trips" }));
  expect(await screen.findByRole("menuitem", { name: "Remove from Trips" })).toBeTruthy();
});

test("album menu removes a directory already referenced by a collection", async () => {
  vi.mocked(listReferences).mockResolvedValue([
    {
      id: "00000000-0000-4000-8000-000000000003",
      type: "DIRECTORY",
      path: "/NASPool/Photos",
      createdAt: "2026-10-26T00:00:00Z",
    },
  ]);
  await openMenu();
  fireEvent.click(await screen.findByRole("menuitem", { name: "Remove from Trips" }));
  await waitFor(() =>
    expect(removeCollectionDirectory).toHaveBeenCalledWith(collection.id, "/NASPool/Photos"),
  );
});

test("album without a source directory has no collection menu", () => {
  render(<GalleryAlbumMenu gallery={{ ...album, uploadDirectory: null }} userId="alice" />);
  expect(screen.queryByRole("button", { name: "Collection options for Photos" })).toBeNull();
});

test("album menu reports collection update failures", async () => {
  vi.mocked(addCollectionDirectory).mockRejectedValue(new Error("unavailable"));
  await openMenu();
  fireEvent.click(await screen.findByRole("menuitem", { name: "Add to Trips" }));
  expect(await screen.findByText("Could not update the collection.")).toBeTruthy();
});
