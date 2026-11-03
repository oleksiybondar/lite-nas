// @vitest-environment jsdom
import { cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { GalleryEditorDialog } from "./GalleryEditorDialog";
import {
  addUserTag,
  chooseGalleryThumbnailPath,
  createGallery,
  listFriends,
  listUserTags,
  removeUserTag,
  saveGalleryThumbnail,
  shareGalleryByUsername,
  updateGallery,
} from "./moments-data";
import type { GalleryRecord } from "./schemas";
import { renderWithLanguage as render } from "./test-utils/language";

vi.mock("@lite-nas/file-browser", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@lite-nas/file-browser")>()),
  FileSelectionDialog: ({ open, onSelect }: { open: boolean; onSelect: (path: string) => void }) =>
    open && (
      <button type="button" onClick={() => onSelect("/home/photo.jpg")}>
        Choose file
      </button>
    ),
}));
vi.mock("./ThumbnailMaker", () => ({
  ThumbnailMaker: ({ onSave }: { onSave: (image: Blob) => Promise<void> }) => (
    <button type="button" onClick={() => void onSave(new Blob(["png"], { type: "image/png" }))}>
      Save generated cover
    </button>
  ),
}));
vi.mock("./GalleryScanControls", () => ({
  GalleryScanControls: () => <div>Media scan controls</div>,
}));
vi.mock("./moments-data", () => ({
  addUserTag: vi.fn(),
  chooseGalleryThumbnailPath: vi.fn(),
  createGallery: vi.fn(),
  listFriends: vi.fn().mockResolvedValue([]),
  listUserTags: vi.fn(),
  removeUserTag: vi.fn(),
  saveGalleryThumbnail: vi.fn(),
  shareGalleryByUsername: vi.fn(),
  updateGallery: vi.fn(),
}));

const gallery: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "COLLECTION",
  name: "Saved",
  description: "Old description",
  ownerId: "1001",
  visibility: "PRIVATE",
  uploadDirectory: null,
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

/** Prepares successful tag requests for one modal submission. */
function mockTags(items: { id: string; value: string }[] = []): void {
  vi.mocked(listUserTags).mockResolvedValue(items);
  vi.mocked(listFriends).mockResolvedValue([]);
  vi.mocked(addUserTag).mockResolvedValue(undefined);
  vi.mocked(removeUserTag).mockResolvedValue(undefined);
}

/** Renders an open editor with standard callback and directory fixtures. */
function renderEditor(selectedGallery?: GalleryRecord): void {
  render(
    <GalleryEditorDialog
      open
      gallery={selectedGallery}
      workingDir="/home"
      onClose={vi.fn()}
      onSaved={vi.fn()}
    />,
  );
}

test.each([
  { label: "Close gallery creator", gallery: undefined },
  { label: "Close gallery editor", gallery },
])("$label calls the supplied close action", ({ label, gallery: selectedGallery }) => {
  mockTags();
  const onClose = vi.fn();
  render(
    <GalleryEditorDialog
      open
      gallery={selectedGallery}
      workingDir="/home"
      onClose={onClose}
      onSaved={vi.fn()}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: label }));
  expect(onClose).toHaveBeenCalledOnce();
});

test("the thumbnail generator closes from its header", () => {
  mockTags();
  renderEditor(gallery);
  fireEvent.click(screen.getByRole("button", { name: "Generate cover" }));
  fireEvent.click(screen.getByRole("button", { name: "Close thumbnail generator" }));
  expect(screen.queryByRole("button", { name: "Save generated cover" })).toBeNull();
});

test("creation sends description and visibility to the gallery API", async () => {
  mockTags();
  vi.mocked(createGallery).mockResolvedValue(gallery);
  renderEditor();
  fireEvent.change(screen.getByLabelText("Name"), { target: { value: "New album" } });
  fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Holiday" } });
  fireEvent.mouseDown(screen.getByLabelText("Gallery layout"));
  fireEvent.click(screen.getByRole("option", { name: "Even grid" }));
  fireEvent.click(screen.getByRole("button", { name: "Create" }));
  await waitFor(() =>
    expect(createGallery).toHaveBeenCalledWith({
      type: "ALBUM",
      name: "New album",
      description: "Holiday",
      visibility: "PUBLIC",
      wallRenderStrategy: "grid",
      uploadDirectory: "/home",
    }),
  );
});

test("a new collection also defaults to public", async () => {
  mockTags();
  vi.mocked(createGallery).mockResolvedValue(gallery);
  renderEditor();
  fireEvent.mouseDown(screen.getByLabelText("Type"));
  fireEvent.click(screen.getByRole("option", { name: "Collection" }));
  fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Saved" } });
  fireEvent.click(screen.getByRole("button", { name: "Create" }));
  await waitFor(() =>
    expect(createGallery).toHaveBeenCalledWith({
      type: "COLLECTION",
      name: "Saved",
      description: "",
      visibility: "PUBLIC",
      wallRenderStrategy: "mosaic",
    }),
  );
});

test("editing saves changed gallery metadata", async () => {
  mockTags();
  vi.mocked(updateGallery).mockResolvedValue({ ...gallery, description: "Updated" });
  renderEditor(gallery);
  fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Updated" } });
  fireEvent.mouseDown(screen.getByLabelText("Gallery layout"));
  fireEvent.click(screen.getByRole("option", { name: "Masonry" }));
  fireEvent.click(await screen.findByRole("button", { name: "Save" }));
  await waitFor(() =>
    expect(updateGallery).toHaveBeenCalledWith(gallery.id, {
      name: "Saved",
      description: "Updated",
      visibility: "PRIVATE",
      wallRenderStrategy: "masonry",
    }),
  );
});

test("editing removes a deleted user tag", async () => {
  const tag = { id: "00000000-0000-4000-8000-000000000002", value: "Old" };
  mockTags([tag]);
  vi.mocked(updateGallery).mockResolvedValue(gallery);
  renderEditor(gallery);
  await screen.findByDisplayValue("Old");
  fireEvent.change(screen.getByLabelText("Tags (comma separated)"), { target: { value: "" } });
  fireEvent.click(screen.getByRole("button", { name: "Save" }));
  await waitFor(() => expect(removeUserTag).toHaveBeenCalledWith(gallery.id, tag.id));
});

test("private gallery editor sends a login name for sharing", async () => {
  mockTags();
  vi.mocked(listFriends).mockResolvedValue([
    {
      username: "test-friend",
      firstName: "",
      lastName: "",
      nickname: "",
      createdAt: "2026-10-26T00:00:00Z",
      updatedAt: "2026-10-26T00:00:00Z",
    },
  ]);
  vi.mocked(updateGallery).mockResolvedValue(gallery);
  vi.mocked(shareGalleryByUsername).mockResolvedValue(undefined);
  renderEditor(gallery);
  fireEvent.mouseDown(await screen.findByRole("combobox", { name: "Share with username" }));
  fireEvent.click(await screen.findByRole("option", { name: "test-friend" }));
  fireEvent.click(await screen.findByRole("button", { name: "Save" }));
  await waitFor(() =>
    expect(shareGalleryByUsername).toHaveBeenCalledWith(gallery.id, "test-friend"),
  );
});

test("the edit modal saves a NAS image as its cover path", async () => {
  mockTags();
  vi.mocked(chooseGalleryThumbnailPath).mockResolvedValue({
    ...gallery,
    thumbnailPath: "/home/photo.jpg",
  });
  const onCoverSaved = vi.fn();
  render(
    <GalleryEditorDialog
      open
      gallery={gallery}
      workingDir="/home"
      onClose={vi.fn()}
      onSaved={vi.fn()}
      onCoverSaved={onCoverSaved}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Choose NAS cover file" }));
  fireEvent.click(screen.getByText("Choose file"));
  await waitFor(() =>
    expect(chooseGalleryThumbnailPath).toHaveBeenCalledWith(gallery.id, "/home/photo.jpg"),
  );
});

test("the edit modal uploads a browser cover", async () => {
  mockTags();
  vi.mocked(saveGalleryThumbnail).mockResolvedValue(gallery);
  renderEditor(gallery);
  const input = document.querySelector('input[type="file"]');
  if (!input) throw new Error("Cover upload input was not rendered.");
  const file = new File(["image"], "cover.png", { type: "image/png" });
  fireEvent.change(input, { target: { files: [file] } });
  await waitFor(() => expect(saveGalleryThumbnail).toHaveBeenCalledWith(gallery.id, file));
});

test("the edit modal uploads a generated cover", async () => {
  mockTags();
  vi.mocked(saveGalleryThumbnail).mockResolvedValue(gallery);
  renderEditor(gallery);
  fireEvent.click(screen.getByRole("button", { name: "Generate cover" }));
  fireEvent.click(screen.getByRole("button", { name: "Save generated cover" }));
  await waitFor(() =>
    expect(saveGalleryThumbnail).toHaveBeenCalledWith(gallery.id, expect.any(Blob)),
  );
});

test("updating the saved cover keeps unsaved description edits", () => {
  mockTags();
  const props = { open: true, workingDir: "/home", onClose: vi.fn(), onSaved: vi.fn() };
  const view = render(<GalleryEditorDialog {...props} gallery={gallery} />);
  fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Draft changes" } });
  view.rerender(
    <GalleryEditorDialog {...props} gallery={{ ...gallery, thumbnailPath: "/home/photo.jpg" }} />,
  );
  expect((screen.getByLabelText("Description") as HTMLTextAreaElement).value).toBe("Draft changes");
});
