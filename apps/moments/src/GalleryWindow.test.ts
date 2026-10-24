import { expect, test } from "vitest";
import type { GalleryTile } from "./GalleryLayout";
import { galleryWindow, wallImageRequests } from "./GalleryWindow";

const tiles: GalleryTile[] = Array.from({ length: 1000 }, (_, index) => ({
  index,
  x: 0,
  y: index * 200,
  width: 200,
  height: 190,
}));

test("an initial viewport does not request originals below the viewport", () => {
  expect(galleryWindow(tiles, 0, 600).visible.map((tile) => tile.index)).toEqual([1, 2, 0]);
});

test("the materialized wall remains bounded after a long scroll", () => {
  expect(galleryWindow(tiles, 120000, 600).rendered.length).toBeLessThanOrEqual(60);
});

test("wall demand admits only eight viewport originals at a time", () => {
  const items = tiles.map((tile) => ({
    id: `${tile.index}`,
    url: `/f/home/test/${tile.index}.jpg`,
    filename: `${tile.index}.jpg`,
    kind: "image" as const,
  }));
  expect(wallImageRequests(items, tiles.slice(0, 40), () => "queued")).toHaveLength(8);
});
