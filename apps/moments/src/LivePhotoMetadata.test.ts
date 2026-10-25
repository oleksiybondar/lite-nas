import { expect, test } from "vitest";
import fixture from "../public/dev-fixtures/gallery.json";
import { galleryItemsFromPaths } from "./gallery-model";
import { composeLivePhotoItems } from "./LivePhotoMetadata";

const paths = ["/f/home/test/IMG_2062.HEIC", "/f/home/test/IMG_2062.MOV"];

// Requirements: moments/FR-003. The requested pair leads the bundled demo gallery.
test("IMG_2062 still and movie are the first fixture paths", () => {
  expect(fixture.items.slice(0, 2)).toEqual([
    "/f/home/oleksiybondar/Pictures/IMG_2062.HEIC",
    "/f/home/oleksiybondar/Pictures/IMG_2062.MOV",
  ]);
});

// Requirements: moments/FR-003. Full fixture processing keeps the Live Photo first.
test("the full fixture composes IMG_2062 as its first media item", () => {
  const composed = composeLivePhotoItems(galleryItemsFromPaths(fixture.items));
  expect(composed[0]).toMatchObject({
    url: "/f/home/oleksiybondar/Pictures/IMG_2062.HEIC",
    motionUrl: "/f/home/oleksiybondar/Pictures/IMG_2062.MOV",
  });
});

// Requirements: moments/FR-003. Fixture processing keeps standalone HEIC files visible.
test("the full fixture keeps IMG_1864 as a standalone HEIC image", () => {
  const composed = composeLivePhotoItems(galleryItemsFromPaths(fixture.items));
  const still = composed.find((item) => item.filename === "IMG_1864.HEIC");
  expect(still?.kind === "image" && still.motionUrl === undefined).toBe(true);
});

// Requirements: moments/FR-003. Same-folder HEIC/MOV companions become one item.
test("matching filename companions compose one Live Photo", () => {
  const items = galleryItemsFromPaths(paths);
  expect(composeLivePhotoItems(items)).toEqual([{ ...items[0], motionUrl: items[1].url }]);
});

// Requirements: moments/FR-003. A HEIC without a matching movie remains an image.
test("a standalone HEIC remains an image", () => {
  const items = galleryItemsFromPaths([paths[0]]);
  expect(composeLivePhotoItems(items)).toEqual(items);
});

// Requirements: moments/FR-003. A MOV without a matching still remains a video.
test("a standalone MOV remains a video", () => {
  const items = galleryItemsFromPaths([paths[1]]);
  expect(composeLivePhotoItems(items)).toEqual(items);
});

// Requirements: moments/FR-003. Only same-folder filenames are paired.
test("different folders keep still and movie separate", () => {
  const items = galleryItemsFromPaths([paths[0], "/f/other/IMG_2062.MOV"]);
  expect(composeLivePhotoItems(items)).toEqual(items);
});

// Requirements: moments/FR-003. Distinct filenames are separate media.
test("different stems keep still and movie separate", () => {
  const items = galleryItemsFromPaths([paths[0], "/f/home/test/IMG_2063.MOV"]);
  expect(composeLivePhotoItems(items)).toEqual(items);
});

// Requirements: moments/FR-003. Duplicate still candidates do not hide either original.
test("ambiguous HEIC and HEIF stills remain separate from the movie", () => {
  const items = galleryItemsFromPaths([paths[0], "/f/home/test/IMG_2062.HEIF", paths[1]]);
  expect(composeLivePhotoItems(items)).toEqual(items);
});

// Requirements: moments/FR-003. Pair composition preserves the still's list position.
test("a pair retains its still's gallery position", () => {
  const items = galleryItemsFromPaths([
    "/f/home/test/first.jpg",
    ...paths,
    "/f/home/test/last.jpg",
  ]);
  expect(composeLivePhotoItems(items)).toEqual([
    items[0],
    { ...items[1], motionUrl: items[2].url },
    items[3],
  ]);
});

// Requirements: moments/FR-003. Unsupported files do not enter the media wall.
test("unsupported files are filtered without hiding supported media", () => {
  const items = galleryItemsFromPaths(["/f/home/test/notes.txt", ...paths]);
  expect(composeLivePhotoItems(items)).toEqual([{ ...items[1], motionUrl: items[2].url }]);
});
