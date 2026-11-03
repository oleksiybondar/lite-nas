import { expect, test } from "vitest";
import { planGallery } from "./GalleryLayout";
import type { GalleryItem } from "./gallery-model";

const items: GalleryItem[] = Array.from({ length: 1000 }, (_, index) => ({
  id: `${index}`,
  url: `/f/home/test/image-${index}.jpg`,
  filename: `image-${index}.jpg`,
  kind: "image",
}));

test.each([
  "mosaic",
  "justified",
  "masonry",
  "grid",
] as const)("%s layout keeps every tile within wall width", (mode) => {
  const { tiles } = planGallery(items, 960, mode);
  expect(tiles.every((tile) => tile.x >= 0 && tile.x + tile.width <= 960.01)).toBe(true);
});

test("narrow walls retain one tile for every media item", () => {
  expect(planGallery(items, 320, "mosaic").tiles).toHaveLength(items.length);
});

test.each(["masonry", "grid"] as const)("%s layout retains one tile per item", (mode) => {
  expect(planGallery(items, 960, mode).tiles).toHaveLength(items.length);
});

test("masonry places successive items in the shortest column", () => {
  const { tiles } = planGallery(items.slice(0, 5), 660, "masonry");
  expect(tiles.map((tile) => tile.x)).toEqual([
    tiles[0].x,
    tiles[1].x,
    tiles[2].x,
    tiles[1].x,
    tiles[0].x,
  ]);
});

test("even grid gives every item the same tile size", () => {
  const { tiles } = planGallery(items.slice(0, 6), 660, "grid");
  expect(new Set(tiles.map((tile) => `${tile.width}:${tile.height}`)).size).toBe(1);
});

test.each(["masonry", "grid"] as const)("%s tiles do not overlap", (mode) => {
  const { tiles } = planGallery(items.slice(0, 30), 660, mode);
  expect(
    tiles.every((left, index) =>
      tiles
        .slice(index + 1)
        .every(
          (right) =>
            left.x + left.width <= right.x ||
            right.x + right.width <= left.x ||
            left.y + left.height <= right.y ||
            right.y + right.height <= left.y,
        ),
    ),
  ).toBe(true);
});
