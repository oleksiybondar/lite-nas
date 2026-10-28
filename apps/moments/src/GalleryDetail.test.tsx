// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { GalleryDetail } from "./GalleryDetail";
import {
  addCollectionDirectory,
  changeCollectionFile,
  createGallery,
  galleryFiles,
  getGallery,
  listGalleries,
  listReferences,
  scanUploadedMedia,
} from "./moments-data";
import type { GalleryRecord } from "./schemas";

vi.mock("@lite-nas/file-browser", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@lite-nas/file-browser")>()),
  fileServicePath: (_route: string, path: string) => `/f${path}`,
  FileSelectionDialog: ({
    open,
    onSelect,
    onClose,
  }: {
    open: boolean;
    onSelect: (path: string) => void;
    onClose: () => void;
  }) =>
    open && (
      <>
        <button type="button" onClick={() => onSelect("/home/Photos/photo.jpg")}>
          Choose file
        </button>
        <button type="button" onClick={onClose}>
          Close file picker
        </button>
      </>
    ),
  FolderSelectionDialog: ({
    open,
    onSelect,
    onClose,
  }: {
    open: boolean;
    onSelect: (path: string) => void;
    onClose: () => void;
  }) =>
    open && (
      <>
        <button type="button" onClick={() => onSelect("/home/Photos")}>
          Choose folder
        </button>
        <button type="button" onClick={onClose}>
          Close folder picker
        </button>
      </>
    ),
}));
vi.mock("@lite-nas/file-upload", () => ({
  UploadDialog: ({ open, onUploaded }: { open: boolean; onUploaded: (file: File) => void }) =>
    open && (
      <button type="button" onClick={() => onUploaded(new File(["photo"], "new.jpg"))}>
        Complete upload
      </button>
    ),
}));
vi.mock("./moments-data", () => ({
  addCollectionDirectory: vi.fn(),
  changeCollectionFile: vi.fn(),
  createGallery: vi.fn(),
  galleryFiles: vi.fn(),
  getGallery: vi.fn(),
  listGalleries: vi.fn(),
  listReferences: vi.fn(),
  scanUploadedMedia: vi.fn(),
}));
vi.mock("./GalleryScanControls", () => ({
  GalleryScanControls: () => <div>Media scan controls</div>,
}));
vi.mock("./Gallery", () => ({
  Gallery: ({
    collectionActions,
    layout,
  }: {
    layout: GalleryRecord["wallRenderStrategy"];
    collectionActions: {
      canRemove: (path: string) => boolean;
      add: (id: string, path: string) => Promise<void>;
      remove: (path: string) => Promise<void>;
      createAndAdd: (name: string, path: string) => Promise<void>;
    };
  }) => (
    <div>
      Gallery wall: {layout}
      <span>
        {collectionActions.canRemove("/home/Photos/photo.jpg")
          ? "Removal available"
          : "Removal unavailable"}
      </span>
      <button
        type="button"
        onClick={() => void collectionActions.add(collection.id, "/home/Photos/photo.jpg")}
      >
        Add via item menu
      </button>
      <button type="button" onClick={() => void collectionActions.remove("/home/Photos/photo.jpg")}>
        Remove via item menu
      </button>
      <button
        type="button"
        onClick={() => void collectionActions.createAndAdd("New", "/home/Photos/photo.jpg")}
      >
        Create via item menu
      </button>
      <button type="button" onClick={() => collectionActions.canRemove("/home/Photos/photo.jpg")}>
        Check removal
      </button>
    </div>
  ),
}));
vi.mock("./GalleryEditorDialog", () => ({
  GalleryEditorDialog: ({
    open,
    onClose,
    onSaved,
  }: {
    open: boolean;
    onClose: () => void;
    onSaved: (gallery: GalleryRecord) => void;
  }) =>
    open && (
      <div>
        <button type="button" onClick={onClose}>
          Close edit
        </button>
        <button type="button" onClick={() => onSaved({ ...collection, name: "Renamed" })}>
          Save edit
        </button>
        <button
          type="button"
          onClick={() => onSaved({ ...collection, wallRenderStrategy: "grid" })}
        >
          Save layout
        </button>
      </div>
    ),
}));

const collection: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "COLLECTION",
  name: "Saved",
  description: "",
  ownerId: "1001",
  visibility: "PRIVATE",
  uploadDirectory: null,
  wallRenderStrategy: "mosaic",
  thumbnailImageUrl: null,
  thumbnailPath: null,
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};

/** Prepares the collection's existing media and owned metadata. */
function renderCollection(
  onBack = vi.fn(),
  paths: string[] = [],
  fetched: GalleryRecord = collection,
): void {
  vi.mocked(getGallery).mockResolvedValue(fetched);
  vi.mocked(galleryFiles).mockResolvedValue(paths);
  vi.mocked(listReferences).mockResolvedValue([]);
  vi.mocked(listGalleries).mockResolvedValue({
    items: [collection],
    metadata: { page: 1, size: 20, total_count: 1, total_pages: 1 },
  });
  render(
    <GalleryDetail
      selected={collection}
      userId="1001"
      maxConcurrent={10}
      workingDir="/home"
      onBack={onBack}
    />,
  );
}

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

test("collection file picker adds its selected file", async () => {
  vi.mocked(changeCollectionFile).mockResolvedValue(undefined);
  renderCollection();
  fireEvent.click(screen.getByRole("button", { name: "Add file" }));
  fireEvent.click(screen.getByRole("button", { name: "Choose file" }));
  await waitFor(() =>
    expect(changeCollectionFile).toHaveBeenCalledWith(
      collection.id,
      "/home/Photos/photo.jpg",
      true,
    ),
  );
});

test("a completed album upload scans only the new file", async () => {
  vi.mocked(scanUploadedMedia).mockResolvedValue(undefined);
  renderCollection(undefined, [], {
    ...collection,
    type: "ALBUM",
    uploadDirectory: "/home/Photos",
  });
  fireEvent.click(await screen.findByRole("button", { name: "Upload media" }));
  fireEvent.click(screen.getByRole("button", { name: "Complete upload" }));
  await waitFor(() =>
    expect(scanUploadedMedia).toHaveBeenCalledWith(collection.id, "/home/Photos/new.jpg"),
  );
});

test("a scanned city is shown in human-readable form", async () => {
  renderCollection(undefined, [], {
    ...collection,
    tags: [
      {
        id: "00000000-0000-4000-8000-000000000002",
        type: "AUTO",
        category: "LOCATION",
        value: "Málaga, Spain",
        count: 1,
      },
    ],
  });
  expect(await screen.findByText("Locations: Málaga, Spain")).toBeTruthy();
});

test("the detail back action returns to the feed", () => {
  const onBack = vi.fn();
  renderCollection(onBack);
  fireEvent.click(screen.getByRole("button", { name: "Back to galleries" }));
  expect(onBack).toHaveBeenCalledOnce();
});

test("saving an edit updates the gallery heading", async () => {
  renderCollection();
  fireEvent.click(screen.getByRole("button", { name: "Edit" }));
  fireEvent.click(screen.getByRole("button", { name: "Save edit" }));
  expect(await screen.findByRole("heading", { name: "Renamed" })).toBeTruthy();
});

test("gallery wall uses the layout returned by the backend", async () => {
  renderCollection(undefined, ["/home/Photos/photo.jpg"], {
    ...collection,
    wallRenderStrategy: "masonry",
  });
  expect(await screen.findByText("Gallery wall: masonry")).toBeTruthy();
});

test("saving the owner's layout updates the open gallery wall", async () => {
  renderCollection(undefined, ["/home/Photos/photo.jpg"]);
  await screen.findByText("Gallery wall: mosaic");
  fireEvent.click(screen.getByRole("button", { name: "Edit" }));
  fireEvent.click(screen.getByRole("button", { name: "Save layout" }));
  expect(await screen.findByText("Gallery wall: grid")).toBeTruthy();
});

test("item menu adds a file to an owned collection", async () => {
  vi.mocked(changeCollectionFile).mockResolvedValue(undefined);
  renderCollection(undefined, ["/home/Photos/photo.jpg"]);
  fireEvent.click(await screen.findByRole("button", { name: "Add via item menu" }));
  await waitFor(() =>
    expect(changeCollectionFile).toHaveBeenCalledWith(
      collection.id,
      "/home/Photos/photo.jpg",
      true,
    ),
  );
});

test("item menu removes a file from an owned collection", async () => {
  vi.mocked(changeCollectionFile).mockResolvedValue(undefined);
  renderCollection(undefined, ["/home/Photos/photo.jpg"]);
  fireEvent.click(await screen.findByRole("button", { name: "Remove via item menu" }));
  await waitFor(() =>
    expect(changeCollectionFile).toHaveBeenCalledWith(
      collection.id,
      "/home/Photos/photo.jpg",
      false,
    ),
  );
});

test("item menu creates a collection for the selected file", async () => {
  vi.mocked(createGallery).mockResolvedValue(collection);
  vi.mocked(changeCollectionFile).mockResolvedValue(undefined);
  renderCollection(undefined, ["/home/Photos/photo.jpg"]);
  fireEvent.click(await screen.findByRole("button", { name: "Create via item menu" }));
  await waitFor(() =>
    expect(createGallery).toHaveBeenCalledWith({ type: "COLLECTION", name: "New" }),
  );
});

test("collection folder picker adds its selected directory", async () => {
  vi.mocked(addCollectionDirectory).mockResolvedValue(undefined);
  renderCollection();
  fireEvent.click(screen.getByRole("button", { name: "Add folder" }));
  fireEvent.click(screen.getByRole("button", { name: "Choose folder" }));
  await waitFor(() =>
    expect(addCollectionDirectory).toHaveBeenCalledWith(collection.id, "/home/Photos"),
  );
});

test("a file without explicit membership cannot be removed from a collection", async () => {
  renderCollection(undefined, ["/home/Photos/photo.jpg"]);
  expect(await screen.findByText("Removal unavailable")).toBeTruthy();
});

test("closing the editor returns to gallery detail", () => {
  renderCollection();
  fireEvent.click(screen.getByRole("button", { name: "Edit" }));
  fireEvent.click(screen.getByRole("button", { name: "Close edit" }));
  expect(screen.queryByRole("button", { name: "Close edit" })).toBeNull();
});

test("the file picker can close without choosing a source", () => {
  renderCollection();
  fireEvent.click(screen.getByRole("button", { name: "Add file" }));
  fireEvent.click(screen.getByRole("button", { name: "Close file picker" }));
  expect(screen.queryByRole("button", { name: "Close file picker" })).toBeNull();
});

test("the folder picker can close without choosing a source", () => {
  renderCollection();
  fireEvent.click(screen.getByRole("button", { name: "Add folder" }));
  fireEvent.click(screen.getByRole("button", { name: "Close folder picker" }));
  expect(screen.queryByRole("button", { name: "Close folder picker" })).toBeNull();
});
