// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: each template maps directly to its placement formula.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: keep template geometry together for initial review.
import { z } from "zod";
import { galleryFileUrlSchema } from "./gallery-model";

/** Maximum editable image layers in one in-memory thumbnail recipe. */
export const MAX_THUMBNAIL_LAYERS = 35;
/** Maximum fixture images included by one random generation action. */
export const MAX_RANDOM_THUMBNAIL_LAYERS = 25;

const uploadedImageSchema = z
  .string()
  .max(28 * 1024 * 1024)
  .regex(/^data:image\/(?:png|jpeg|webp|gif);base64,[A-Za-z0-9+/]+={0,2}$/);

/** A validated original gallery image or an image embedded by the browser. */
export const thumbnailSourceSchema = z.union([galleryFileUrlSchema, uploadedImageSchema]);

/** One independently positioned image in a serializable collage. */
export const thumbnailLayerSchema = z.strictObject({
  id: z.string().min(1),
  src: thumbnailSourceSchema,
  name: z.string().min(1),
  x: z.number().finite().min(-100).max(200),
  y: z.number().finite().min(-100).max(200),
  width: z.number().finite().min(2).max(200),
  height: z.number().finite().min(2).max(200),
  angle: z.number().finite().min(-180).max(180),
  zoom: z.number().finite().min(1).max(4),
  focusX: z.number().finite().min(0).max(100),
  focusY: z.number().finite().min(0).max(100),
  frame: z.enum(["none", "polaroid", "rounded"]),
  effect: z.enum(["none", "grayscale", "sepia", "warm", "grain"]),
  shadow: z.boolean(),
});

/** Moves and scales all image layers together while leaving the canvas fixed. */
export const thumbnailCompositionSchema = z.strictObject({
  x: z.number().finite().min(-200).max(200),
  y: z.number().finite().min(-200).max(200),
  scale: z.number().finite().min(0.25).max(4),
});

export type ThumbnailComposition = z.infer<typeof thumbnailCompositionSchema>;
export const defaultThumbnailComposition: ThumbnailComposition = { x: 0, y: 0, scale: 1 };

/** Complete portable thumbnail recipe; array order is back-to-front order. */
export const thumbnailDocumentSchema = z.strictObject({
  version: z.literal(1),
  aspect: z.enum(["16:9", "4:3", "1:1", "3:4", "9:16"]),
  background: z.enum(["transparent", "white", "cream", "paper", "charcoal"]),
  composition: thumbnailCompositionSchema.default(defaultThumbnailComposition),
  layers: z
    .array(thumbnailLayerSchema)
    .min(1)
    .max(MAX_THUMBNAIL_LAYERS)
    .refine(
      (layers) => new Set(layers.map((layer) => layer.id)).size === layers.length,
      "Layer IDs must be unique",
    ),
});

export type ThumbnailLayer = z.infer<typeof thumbnailLayerSchema>;
export type ThumbnailDocument = z.infer<typeof thumbnailDocumentSchema>;
export type ThumbnailAspect = ThumbnailDocument["aspect"];

/** Aspect ratio used by both interactive preview and image export. */
export const aspectValue = (aspect: ThumbnailAspect): number => {
  const [width, height] = aspect.split(":").map(Number);
  return width / height;
};

/** Chooses at least a Full HD pixel area while avoiding very large browser canvases. */
export const exportDimensions = (aspect: ThumbnailAspect, sourceScale = 0): [number, number] => {
  const ratio = aspectValue(aspect);
  const width = Math.max(Math.sqrt(1920 * 1080 * ratio), sourceScale);
  const boundedWidth = Math.min(4096, 4096 * ratio, Math.ceil(width));
  return [Math.ceil(boundedWidth), Math.ceil(boundedWidth / ratio)];
};

/** Creates an editable layer without changing the original image bytes. */
export const newThumbnailLayer = (src: string, name: string): ThumbnailLayer => ({
  id: crypto.randomUUID(),
  src,
  name,
  x: 25,
  y: 25,
  width: 50,
  height: 50,
  angle: 0,
  zoom: 1,
  focusX: 50,
  focusY: 50,
  frame: "none",
  effect: "none",
  shadow: true,
});

/** Places selected images using the requested starting arrangement. */
export const arrangeThumbnail = (
  layers: readonly ThumbnailLayer[],
  template: "scatter" | "ribbon" | "grid" | "featured",
  canvasAspect = 16 / 9,
): ThumbnailLayer[] =>
  layers.map((layer, index) => {
    const count = layers.length;
    const columns = Math.ceil(Math.sqrt(count));
    const rows = Math.ceil(count / columns);
    if (template === "grid") {
      const width = 90 / columns;
      const height = 86 / rows;
      return {
        ...layer,
        x: 5 + (index % columns) * width,
        y: 7 + Math.floor(index / columns) * height,
        width: width - 2,
        height: height - 2,
        angle: 0,
        frame: layer.src.startsWith("data:image/png") ? "none" : "rounded",
      };
    }
    if (template === "ribbon") {
      const fraction = count === 1 ? 0.5 : index / (count - 1);
      return {
        ...layer,
        x: 8 + fraction * 62,
        y: 30 + 14 * Math.sin(fraction * Math.PI),
        width: 30,
        height: 42,
        angle: -22 + fraction * 44,
        frame: layer.src.startsWith("data:image/png") ? "none" : "rounded",
      };
    }
    if (template === "featured") {
      if (index === 0)
        return {
          ...layer,
          x: 8,
          y: 10,
          width: 62,
          height: 80,
          angle: -3,
          frame: layer.src.startsWith("data:image/png") ? "none" : "rounded",
        };
      const slot = index - 1;
      return {
        ...layer,
        x: 65 + (slot % 2) * 14,
        y: 7 + (Math.floor(slot / 2) % 4) * 22,
        width: 24,
        height: 24,
        angle: slot % 2 ? 8 : -8,
        frame: layer.src.startsWith("data:image/png") ? "none" : "polaroid",
      };
    }
    const scatteredColumns = Math.min(
      count,
      Math.max(1, Math.round(Math.sqrt((count * canvasAspect) / 1.3))),
    );
    const scatteredRows = Math.ceil(count / scatteredColumns);
    const cellWidth = 90 / scatteredColumns;
    const cellHeight = 86 / scatteredRows;
    return {
      ...layer,
      x:
        5 +
        (index % scatteredColumns) * cellWidth -
        cellWidth * 0.06 +
        Math.random() * cellWidth * 0.12,
      y:
        7 +
        Math.floor(index / scatteredColumns) * cellHeight -
        cellHeight * 0.06 +
        Math.random() * cellHeight * 0.12,
      width: cellWidth * (0.88 + Math.random() * 0.17),
      height: cellHeight * (0.88 + Math.random() * 0.17),
      angle: -12 + Math.random() * 24,
      frame: layer.src.startsWith("data:image/png") ? "none" : "polaroid",
    };
  });

/** Fits the outer corners of an arrangement near the canvas edges. */
export const fitThumbnailComposition = (
  layers: readonly ThumbnailLayer[],
  aspect: ThumbnailAspect,
): ThumbnailComposition => {
  if (layers.length === 0) return defaultThumbnailComposition;
  const canvasAspect = aspectValue(aspect);
  const bounds = layers.map((layer) => {
    const radians = (layer.angle * Math.PI) / 180;
    const cosine = Math.abs(Math.cos(radians));
    const sine = Math.abs(Math.sin(radians));
    const halfWidth = (layer.width * cosine + (layer.height * sine) / canvasAspect) / 2;
    const halfHeight = (layer.width * sine * canvasAspect + layer.height * cosine) / 2;
    const centerX = layer.x + layer.width / 2;
    const centerY = layer.y + layer.height / 2;
    return {
      left: centerX - halfWidth,
      right: centerX + halfWidth,
      top: centerY - halfHeight,
      bottom: centerY + halfHeight,
    };
  });
  const left = Math.min(...bounds.map((bound) => bound.left));
  const right = Math.max(...bounds.map((bound) => bound.right));
  const top = Math.min(...bounds.map((bound) => bound.top));
  const bottom = Math.max(...bounds.map((bound) => bound.bottom));
  const scale = Math.max(
    0.25,
    Math.min(4, 92 / Math.max(1, right - left), 92 / Math.max(1, bottom - top)),
  );
  return {
    scale,
    x: Math.max(-200, Math.min(200, (50 - (left + right) / 2) * scale)),
    y: Math.max(-200, Math.min(200, (50 - (top + bottom) / 2) * scale)),
  };
};
