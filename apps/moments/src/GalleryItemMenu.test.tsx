// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { GalleryProvider, type GalleryState } from "./GalleryContext";
import { GalleryItemMenu } from "./GalleryItemMenu";
import type { GalleryItem } from "./gallery-model";
import type { GalleryRecord } from "./schemas";

const item: GalleryItem = {
  id: "photo-1",
  url: "/f/home/Photos/photo.jpg",
  filename: "photo.jpg",
  kind: "image",
};
const collection: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "COLLECTION",
  name: "Favorites",
  description: "",
  ownerId: "alice",
  visibility: "PUBLIC",
  uploadDirectory: null,
  wallRenderStrategy: "mosaic",
  thumbnailImageUrl: null,
  thumbnailPath: null,
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};

/** Renders the media menu with owner collection actions. */
function renderMenu(overrides: Partial<NonNullable<GalleryState["collectionActions"]>> = {}) {
  const actions = {
    collections: [collection],
    canRemove: vi.fn(() => true),
    add: vi.fn(async () => undefined),
    remove: vi.fn(async () => undefined),
    createAndAdd: vi.fn(async () => undefined),
    ...overrides,
  };
  render(
    <GalleryProvider items={[item]} collectionActions={actions}>
      <GalleryItemMenu item={item} />
    </GalleryProvider>,
  );
  fireEvent.click(screen.getByRole("button", { name: "More options for photo.jpg" }));
  return actions;
}

afterEach(() => cleanup());

test("a gallery photo can be added to an owned collection", async () => {
  const actions = renderMenu();
  fireEvent.click(screen.getByRole("menuitem", { name: "Add to Favorites" }));
  await waitFor(() =>
    expect(actions.add).toHaveBeenCalledWith(collection.id, "/home/Photos/photo.jpg"),
  );
});

test("a gallery photo can create and join a new collection", async () => {
  const actions = renderMenu();
  fireEvent.click(screen.getByRole("menuitem", { name: "Add to new collection" }));
  fireEvent.change(screen.getByLabelText("Collection name"), { target: { value: "Trips" } });
  fireEvent.click(screen.getByRole("button", { name: "Create and add" }));
  await waitFor(() =>
    expect(actions.createAndAdd).toHaveBeenCalledWith("Trips", "/home/Photos/photo.jpg"),
  );
});

test("the new collection dialog closes from its header", async () => {
  renderMenu();
  fireEvent.click(screen.getByRole("menuitem", { name: "Add to new collection" }));
  fireEvent.click(screen.getByRole("button", { name: "Close new collection" }));
  await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
});

test("an explicit collection member can be removed", async () => {
  const actions = renderMenu();
  fireEvent.click(screen.getByRole("menuitem", { name: "Remove from collection" }));
  await waitFor(() => expect(actions.remove).toHaveBeenCalledWith("/home/Photos/photo.jpg"));
});

test("a failed collection update reports an error", async () => {
  renderMenu({ add: vi.fn().mockRejectedValue(new Error("unavailable")) });
  fireEvent.click(screen.getByRole("menuitem", { name: "Add to Favorites" }));
  expect((await screen.findByRole("alert")).textContent).toBe("Collection update failed.");
});
