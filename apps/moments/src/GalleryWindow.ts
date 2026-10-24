import type { GalleryTile } from "./GalleryLayout";
import type { GalleryItem } from "./gallery-model";

/** Tiles retained for stable scrolling and tiles whose originals are actually visible. */
export type GalleryWindow = { rendered: GalleryTile[]; visible: GalleryTile[] };

/** Keeps DOM overscan separate from the smaller media-download viewport. */
export const galleryWindow = (
  tiles: readonly GalleryTile[],
  scrollTop: number,
  viewportHeight: number,
): GalleryWindow => {
  const bottom = scrollTop + viewportHeight;
  const rendered = tiles
    .filter((tile) => tile.y + tile.height >= scrollTop - 200 && tile.y <= bottom + 200)
    .sort((left, right) => Math.abs(left.y - scrollTop) - Math.abs(right.y - scrollTop))
    .slice(0, 60);
  const visible = rendered
    .filter((tile) => tile.y + tile.height > scrollTop && tile.y < bottom)
    .sort((left, right) => {
      const center = scrollTop + viewportHeight / 2;
      return (
        Math.abs(left.y + left.height / 2 - center) - Math.abs(right.y + right.height / 2 - center)
      );
    });
  return { rendered, visible };
};

/** Chooses a small unsettled batch from only the viewport's still images. */
export const wallImageRequests = (
  items: readonly GalleryItem[],
  visible: readonly GalleryTile[],
  status: (url: string) => string,
): string[] =>
  visible
    .filter((tile) => items[tile.index].kind === "image")
    .map((tile) => items[tile.index].url)
    .filter((url) => status(url) !== "ready" && status(url) !== "error")
    .slice(0, 8);
