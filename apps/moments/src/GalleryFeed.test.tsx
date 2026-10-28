// @vitest-environment jsdom
import { act, cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { GalleryFeed } from "./GalleryFeed";
import { listGalleries } from "./moments-data";
import type { GalleryRecord } from "./schemas";

vi.mock("./moments-data", () => ({
  listGalleries: vi.fn(),
  createGallery: vi.fn(),
  listReferences: vi.fn(),
  addCollectionDirectory: vi.fn(),
  removeCollectionDirectory: vi.fn(),
}));
vi.mock("./GalleryCover", () => ({ GalleryCover: () => <div aria-hidden="true" /> }));
vi.mock("./GalleryEditorDialog", () => ({
  GalleryEditorDialog: ({
    open,
    gallery,
    onClose,
    onSaved,
  }: {
    open: boolean;
    gallery?: GalleryRecord;
    onClose: () => void;
    onSaved: (gallery: GalleryRecord) => void;
  }) =>
    open && (
      <div>
        <button
          type="button"
          onClick={() => onSaved(gallery ? { ...gallery, name: "Edited album" } : album)}
        >
          {gallery ? "Save gallery changes" : "Create gallery"}
        </button>
        <button type="button" onClick={onClose}>
          Cancel creation
        </button>
      </div>
    ),
}));

const album: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "ALBUM",
  name: "First album",
  description: "",
  ownerId: "alice",
  visibility: "PRIVATE",
  uploadDirectory: "/NASPool/Photos",
  wallRenderStrategy: "masonry",
  thumbnailImageUrl: null,
  thumbnailPath: null,
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};

/** Prepares the first gallery page returned to feed tests. */
function mockFirstPage(items: GalleryRecord[]): void {
  vi.mocked(listGalleries).mockResolvedValue({
    items,
    metadata: { page: 1, size: 20, total_count: items.length, total_pages: items.length ? 1 : 0 },
  });
}

/** Renders the home feed with its standard test identity and directory. */
function renderHomeFeed(onOpen = vi.fn()): void {
  render(<GalleryFeed scope="home" userId="alice" workingDir="/home" onOpen={onOpen} />);
}

beforeEach(() => {
  mockFirstPage([]);
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
  vi.unstubAllGlobals();
  window.history.replaceState({}, "", "/");
});

test.each([
  { path: "/", href: `/galleries/${album.id}` },
  { path: "/apps/moments/", href: `/apps/moments/galleries/${album.id}` },
])("gallery cards expose a copyable link under $path", async ({ path, href }) => {
  window.history.replaceState({}, "", path);
  mockFirstPage([album]);
  renderHomeFeed();
  expect((await screen.findByText("First album")).closest("a")?.getAttribute("href")).toBe(href);
});

test("home feed loads the next backend page when its end becomes visible", async () => {
  let enter: IntersectionObserverCallback | undefined;
  vi.stubGlobal(
    "IntersectionObserver",
    class {
      constructor(callback: IntersectionObserverCallback) {
        enter = callback;
      }
      observe(): void {}
      disconnect(): void {}
    },
  );
  vi.mocked(listGalleries)
    .mockResolvedValueOnce({
      items: [album],
      metadata: { page: 1, size: 20, total_count: 21, total_pages: 2 },
    })
    .mockResolvedValueOnce({
      items: [{ ...album, id: "00000000-0000-4000-8000-000000000002", name: "Second album" }],
      metadata: { page: 2, size: 20, total_count: 21, total_pages: 2 },
    });
  render(<GalleryFeed scope="home" userId="alice" workingDir="/NASPool/Photos" onOpen={vi.fn()} />);
  await screen.findByText("First album");
  await waitFor(() => expect(enter).toBeDefined());
  const notify = enter;
  if (!notify) throw new Error("Feed pagination observer was not installed.");
  act(() =>
    notify([{ isIntersecting: true } as IntersectionObserverEntry], {} as IntersectionObserver),
  );
  expect(await screen.findByText("Second album", {}, { timeout: 10000 })).toBeTruthy();
}, 15000);

test("shared feed excludes the signed-in user's galleries", async () => {
  mockFirstPage([album]);
  render(
    <GalleryFeed scope="shared" userId="alice" workingDir="/NASPool/Photos" onOpen={vi.fn()} />,
  );
  expect(await screen.findByText("No galleries yet.", {}, { timeout: 10000 })).toBeTruthy();
}, 15000);

test("Load more continues pagination when intersection observation is unavailable", async () => {
  vi.stubGlobal("IntersectionObserver", undefined);
  vi.mocked(listGalleries)
    .mockResolvedValueOnce({
      items: [album],
      metadata: { page: 1, size: 20, total_count: 21, total_pages: 2 },
    })
    .mockResolvedValueOnce({
      items: [{ ...album, id: "00000000-0000-4000-8000-000000000002", name: "Second album" }],
      metadata: { page: 2, size: 20, total_count: 21, total_pages: 2 },
    });
  renderHomeFeed();
  fireEvent.click(await screen.findByRole("button", { name: "Load more" }));
  expect(await screen.findByText("Second album")).toBeTruthy();
});

test("feed search sends its query to the paged gallery API", async () => {
  renderHomeFeed();
  fireEvent.change(screen.getByLabelText("Search galleries"), { target: { value: "beach" } });
  await waitFor(() =>
    expect(listGalleries).toHaveBeenCalledWith(
      1,
      undefined,
      expect.objectContaining({ q: "beach" }),
    ),
  );
});

test.each([
  { scope: "home" as const, expected: [album.id, "00000000-0000-4000-8000-000000000002"] },
  { scope: "shared" as const, expected: ["00000000-0000-4000-8000-000000000002"] },
])("matching photo links follow the $scope gallery scope", async ({ scope, expected }) => {
  const collection = {
    ...album,
    id: "00000000-0000-4000-8000-000000000002",
    name: "City trip",
    type: "COLLECTION" as const,
    ownerId: "bob",
  };
  vi.mocked(listGalleries).mockResolvedValue({
    items: [album, collection],
    metadata: { page: 1, size: 20, total_count: 2, total_pages: 1 },
    matchingMedia: {
      items: [
        {
          galleryId: album.id,
          linkedGalleries: [
            { id: album.id, name: album.name, type: album.type, ownerId: album.ownerId },
            {
              id: collection.id,
              name: collection.name,
              type: collection.type,
              ownerId: collection.ownerId,
            },
          ],
          path: "/NASPool/Photos/dog.jpg",
          mediaType: "IMAGE",
          width: 640,
          height: 480,
          livePhotoPath: null,
          caption: "A dog on a sofa",
          tags: [{ category: "OBJECT", value: "Dog" }],
        },
      ],
      total: 1,
      limit: 50,
      offset: 0,
    },
  });
  render(<GalleryFeed scope={scope} userId="alice" workingDir="/home" onOpen={vi.fn()} />);
  fireEvent.change(screen.getByLabelText("Search galleries"), { target: { value: "dog" } });
  const card = await screen.findByTestId("moments-media-search-card");
  const links = within(card).getAllByRole("link");
  expect(links.map((link) => link.getAttribute("href"))).toEqual(
    expected.map((id) => `/galleries/${id}`),
  );
});

test.each([
  { label: "Type", option: "Collections", filter: { type: "COLLECTION" } },
  { label: "Visibility", option: "Public", filter: { visibility: "PUBLIC" } },
])("$label filter sends the selected value to the backend", async ({ label, option, filter }) => {
  renderHomeFeed();
  fireEvent.mouseDown(screen.getByLabelText(label));
  fireEvent.click(screen.getByRole("option", { name: option }));
  await waitFor(() =>
    expect(listGalleries).toHaveBeenCalledWith(1, undefined, expect.objectContaining(filter)),
  );
});

test("media filter sends multiple selected types to the backend", async () => {
  renderHomeFeed();
  fireEvent.mouseDown(screen.getByLabelText("Media type"));
  fireEvent.click(screen.getByRole("option", { name: "Photos" }));
  fireEvent.click(screen.getByRole("option", { name: "Videos" }));
  await waitFor(() =>
    expect(listGalleries).toHaveBeenCalledWith(
      1,
      undefined,
      expect.objectContaining({ mediaType: ["IMAGE", "VIDEO"] }),
    ),
  );
});

test("location filter passes entered place names to the API", async () => {
  renderHomeFeed();
  fireEvent.change(screen.getByLabelText("Location"), { target: { value: "Copenhagen" } });
  await waitFor(() =>
    expect(listGalleries).toHaveBeenCalledWith(
      1,
      undefined,
      expect.objectContaining({ location: "Copenhagen" }),
    ),
  );
});

test("clear filters restores the unfiltered feed", async () => {
  renderHomeFeed();
  fireEvent.mouseDown(screen.getByLabelText("Type"));
  fireEvent.click(screen.getByRole("option", { name: "Albums" }));
  fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
  await waitFor(() =>
    expect(listGalleries).toHaveBeenLastCalledWith(
      1,
      undefined,
      expect.objectContaining({ type: undefined }),
    ),
  );
});

test("opening a feed card selects its gallery", async () => {
  mockFirstPage([album]);
  const onOpen = vi.fn();
  renderHomeFeed(onOpen);
  fireEvent.click(await screen.findByText("First album"));
  expect(onOpen).toHaveBeenCalledWith(album);
});

test("ordinary gallery card clicks stay in the SPA", async () => {
  mockFirstPage([album]);
  renderHomeFeed();
  const card = (await screen.findByText("First album")).closest("a");
  if (!card) throw new Error("Gallery card link was not rendered.");
  expect(fireEvent.click(card)).toBe(false);
});

test("an owner can open the editor directly from a gallery card", async () => {
  mockFirstPage([album]);
  renderHomeFeed();
  fireEvent.click(await screen.findByRole("button", { name: "Edit album First album" }));
  expect(screen.getByRole("button", { name: "Save gallery changes" })).toBeTruthy();
});

test("an owned collection card also exposes an edit action", async () => {
  mockFirstPage([
    { ...album, type: "COLLECTION", name: "Family collection", uploadDirectory: null },
  ]);
  renderHomeFeed();
  expect(
    await screen.findByRole("button", { name: "Edit collection Family collection" }),
  ).toBeTruthy();
});

test("saving from the card editor updates the feed card", async () => {
  mockFirstPage([album]);
  renderHomeFeed();
  fireEvent.click(await screen.findByRole("button", { name: "Edit album First album" }));
  fireEvent.click(screen.getByRole("button", { name: "Save gallery changes" }));
  expect(await screen.findByText("Edited album")).toBeTruthy();
});

test("another user's gallery card has no edit action", async () => {
  mockFirstPage([{ ...album, ownerId: "bob" }]);
  renderHomeFeed();
  await screen.findByText("First album");
  expect(screen.queryByRole("button", { name: "Edit album First album" })).toBeNull();
});

test("creating a gallery opens the saved item", async () => {
  const onOpen = vi.fn();
  renderHomeFeed(onOpen);
  fireEvent.click(screen.getByRole("button", { name: "Create album or collection" }));
  fireEvent.click(screen.getByRole("button", { name: "Create gallery" }));
  expect(onOpen).toHaveBeenCalledWith(album);
});

test("cancel closes the gallery creation modal", async () => {
  renderHomeFeed();
  fireEvent.click(screen.getByRole("button", { name: "Create album or collection" }));
  fireEvent.click(screen.getByRole("button", { name: "Cancel creation" }));
  expect(screen.queryByRole("button", { name: "Cancel creation" })).toBeNull();
});
