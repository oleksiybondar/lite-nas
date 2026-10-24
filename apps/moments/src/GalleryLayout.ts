// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: the mosaic placement loop keeps its geometry invariants explicit.
// biome-ignore-all lint/complexity/useMaxParams: related placement coordinates are grouped in local option objects where reused.
import { z } from "zod";
import type { GalleryItem } from "./gallery-model";

/** Stable wall geometry, calculated without requesting original media. */
export type GalleryTile = { index: number; x: number; y: number; width: number; height: number };
/** Layout choices accepted by the fixture and the wall selector. */
export const galleryLayoutModeSchema = z.enum(["mosaic", "justified", "masonry", "grid"]);
export type GalleryLayoutMode = z.infer<typeof galleryLayoutModeSchema>;
type GalleryPlan = { tiles: GalleryTile[]; height: number };
const gap = 8;

/** Plans a responsive wall using deterministic ratios while source dimensions are unavailable. */
export const planGallery = (
  items: readonly GalleryItem[],
  width: number,
  mode: GalleryLayoutMode,
): GalleryPlan => {
  if (width <= 0 || items.length === 0) return { tiles: [], height: 0 };
  const columns = Math.max(1, Math.min(5, Math.floor((width + gap) / 210)));
  const cell = (width - gap * (columns - 1)) / columns;
  switch (mode) {
    case "mosaic":
      return planMosaic(items.length, columns, cell);
    case "justified":
      return planJustified(items, width, cell);
    case "masonry":
      return planMasonry(items, columns, cell);
    case "grid":
      return planGrid(items.length, columns, cell);
  }
};

/** Places equal-size cells in reading order. */
const planGrid = (count: number, columns: number, cell: number): GalleryPlan => {
  const height = Math.max(150, Math.min(230, cell));
  const tiles = Array.from({ length: count }, (_, index) => ({
    index,
    x: (index % columns) * (cell + gap),
    y: Math.floor(index / columns) * (height + gap),
    width: cell,
    height,
  }));
  return { tiles, height: Math.ceil(count / columns) * (height + gap) - gap };
};

/** Provides deterministic demo proportions until item dimensions come from metadata. */
const fallbackRatio = (item: GalleryItem, index: number): number =>
  item.kind === "video" ? 16 / 9 : [1, 1.35, 0.85, 1.6, 0.75][index % 5];

/** Places each item into the currently shortest column. */
const planMasonry = (items: readonly GalleryItem[], columns: number, cell: number): GalleryPlan => {
  const heights = Array<number>(columns).fill(0);
  const tiles = items.map((item, index) => {
    const column = heights.indexOf(Math.min(...heights));
    const height = Math.max(140, Math.min(400, cell / fallbackRatio(item, index)));
    const tile = { index, x: column * (cell + gap), y: heights[column], width: cell, height };
    heights[column] += height + gap;
    return tile;
  });
  return { tiles, height: Math.max(...heights) - gap };
};

/** Finds the next usable mosaic position without rescanning completed rows. */
const findSpot = (
  occupied: boolean[][],
  spot: { startRow: number; columns: number; colSpan: number; rowSpan: number },
): { row: number; col: number } => {
  const { startRow, columns, colSpan, rowSpan } = spot;
  for (let row = startRow; ; row += 1) {
    for (let col = 0; col <= columns - colSpan; col += 1) {
      if (isSpotFree(occupied, { row, col, colSpan, rowSpan })) return { row, col };
    }
  }
};

/** Checks every cell covered by a proposed mosaic tile. */
const isSpotFree = (
  occupied: boolean[][],
  spot: { row: number; col: number; colSpan: number; rowSpan: number },
): boolean => {
  const { row, col, colSpan, rowSpan } = spot;
  for (let r = row; r < row + rowSpan; r += 1)
    for (let c = col; c < col + colSpan; c += 1) if (occupied[r]?.[c]) return false;
  return true;
};

/** Marks each cell covered by a completed mosaic tile. */
const occupySpot = (
  occupied: boolean[][],
  spot: { row: number; col: number; colSpan: number; rowSpan: number; columns: number },
): void => {
  const { row, col, colSpan, rowSpan, columns } = spot;
  for (let r = row; r < row + rowSpan; r += 1) {
    occupied[r] ??= Array(columns).fill(false);
    for (let c = col; c < col + colSpan; c += 1) occupied[r][c] = true;
  }
};

/** Places mosaic tiles in stable input order with a bounded number of live rows. */
const planMosaic = (count: number, columns: number, cell: number): GalleryPlan => {
  const rowHeight = Math.max(150, Math.min(230, cell));
  const occupied: boolean[][] = [];
  const tiles: GalleryTile[] = [];
  let firstOpenRow = 0;
  for (let index = 0; index < count; index += 1) {
    const colSpan = columns > 2 && index % 9 === 0 ? 2 : 1;
    const rowSpan = columns > 2 && index % 13 === 4 ? 2 : 1;
    const { row, col } = findSpot(occupied, { startRow: firstOpenRow, columns, colSpan, rowSpan });
    occupySpot(occupied, { row, col, colSpan, rowSpan, columns });
    while (occupied[firstOpenRow]?.every(Boolean)) firstOpenRow += 1;
    tiles.push({
      index,
      x: col * (cell + gap),
      y: row * (rowHeight + gap),
      width: colSpan * cell + (colSpan - 1) * gap,
      height: rowSpan * rowHeight + (rowSpan - 1) * gap,
    });
  }
  return { tiles, height: occupied.length * (rowHeight + gap) - gap };
};

/** Adds one completed justified row without changing any earlier geometry. */
const appendRow = (
  tiles: GalleryTile[],
  indices: readonly number[],
  ratios: readonly number[],
  geometry: { width: number; target: number; y: number; last: boolean },
): number => {
  const { width, target, y, last } = geometry;
  const available = width - gap * (indices.length - 1);
  const sum = indices.reduce((total, index) => total + ratios[index], 0);
  const height = last ? target : Math.max(130, Math.min(280, available / sum));
  let x = 0;
  for (const index of indices) {
    const tileWidth = height * ratios[index];
    tiles.push({ index, x, y, width: tileWidth, height });
    x += tileWidth + gap;
  }
  return y + height + gap;
};

/** Packs original-media placeholders into rows with a stable fallback aspect ratio. */
const planJustified = (items: readonly GalleryItem[], width: number, cell: number): GalleryPlan => {
  const target = Math.max(150, Math.min(250, cell));
  const ratios = items.map(fallbackRatio);
  const tiles: GalleryTile[] = [];
  let row: number[] = [];
  let ratioSum = 0;
  let y = 0;
  for (let index = 0; index < items.length; index += 1) {
    row.push(index);
    ratioSum += ratios[index];
    if (ratioSum * target + gap * (row.length - 1) < width) continue;
    y = appendRow(tiles, row, ratios, { width, target, y, last: false });
    row = [];
    ratioSum = 0;
  }
  if (row.length > 0) y = appendRow(tiles, row, ratios, { width, target, y, last: true });
  return { tiles, height: Math.max(0, y - gap) };
};
