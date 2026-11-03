// biome-ignore-all lint/complexity/useMaxParams: canvas drawing keeps the geometry operands explicit.
// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: each frame branch follows a visible drawing operation.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: drawing each layer in one save/restore scope preserves canvas state.
import { streamProtocol } from "./StreamContext";
import {
  aspectValue,
  exportDimensions,
  type ThumbnailDocument,
  type ThumbnailLayer,
} from "./thumbnail-model";

/** Resolves a validated layer source to a browser display URL. */
const displaySource = (src: string): string =>
  src.startsWith("data:") ? src : (streamProtocol.image(src) ?? "");

/** A bounded preview source with the original dimensions retained for export sizing. */
export type ThumbnailImage = {
  source: CanvasImageSource;
  width: number;
  height: number;
  originalWidth: number;
  originalHeight: number;
};

/** Decodes one original only when export or preview preparation needs it. */
export const loadOriginalThumbnailImage = async (
  src: string,
  signal?: AbortSignal,
): Promise<ThumbnailImage> => {
  if (signal?.aborted) throw new Error("Image load canceled");
  const image = new Image();
  const abort = (): void => {
    image.removeAttribute("src");
  };
  signal?.addEventListener("abort", abort, { once: true });
  image.src = displaySource(src);
  try {
    await image.decode();
  } finally {
    signal?.removeEventListener("abort", abort);
  }
  if (signal?.aborted) throw new Error("Image load canceled");
  return {
    source: image,
    width: image.naturalWidth,
    height: image.naturalHeight,
    originalWidth: image.naturalWidth,
    originalHeight: image.naturalHeight,
  };
};

/** Downsamples one selected original so editing does not retain its full decode. */
const loadPreviewImage = async (src: string, signal: AbortSignal): Promise<ThumbnailImage> => {
  const original = await loadOriginalThumbnailImage(src, signal);
  if (signal.aborted) throw new Error("Image load canceled");
  const longest = Math.max(original.width, original.height);
  const scale = Math.min(1, 1200 / longest);
  const canvas = globalThis.document.createElement("canvas");
  canvas.width = Math.max(1, Math.round(original.width * scale));
  canvas.height = Math.max(1, Math.round(original.height * scale));
  const context = canvas.getContext("2d");
  if (!context) throw new Error("Canvas is unavailable");
  context.drawImage(original.source, 0, 0, canvas.width, canvas.height);
  if (original.source instanceof HTMLImageElement) original.source.removeAttribute("src");
  return {
    source: canvas,
    width: canvas.width,
    height: canvas.height,
    originalWidth: original.originalWidth,
    originalHeight: original.originalHeight,
  };
};

/** Loads only new selected sources, with at most two original decodes in flight. */
export const loadThumbnailPreviews = async (
  document: ThumbnailDocument,
  cache: Map<string, ThumbnailImage>,
  onProgress: () => void,
  signal: AbortSignal,
): Promise<Map<string, ThumbnailImage>> => {
  const pending = [...new Set(document.layers.map((layer) => layer.src))].filter(
    (src) => !cache.has(src),
  );
  let next = 0;
  const worker = async (): Promise<void> => {
    while (next < pending.length && !signal.aborted) {
      const src = pending[next++];
      try {
        const image = await loadPreviewImage(src, signal);
        if (signal.aborted) return;
        cache.set(src, image);
        onProgress();
      } catch {
        /* Keep other selected images usable. */
      }
    }
  };
  await Promise.all([worker(), worker()]);
  return new Map(
    document.layers.flatMap((layer) => {
      const image = cache.get(layer.src);
      return image ? [[layer.id, image] as const] : [];
    }),
  );
};

/** Calculates export pixels from aspect and the best available original width. */
export const thumbnailExportSize = (
  document: ThumbnailDocument,
  images: Map<string, ThumbnailImage>,
): [number, number] => {
  const sourceScale = Math.max(
    0,
    ...document.layers.map((layer) => {
      const image = images.get(layer.id);
      return image ? image.originalWidth / (layer.width / 100) : 0;
    }),
  );
  return exportDimensions(document.aspect, sourceScale);
};

/** Draws a cover crop with user-controlled zoom and focal point. */
const drawPhoto = (
  context: CanvasRenderingContext2D,
  image: ThumbnailImage,
  layer: ThumbnailLayer,
  width: number,
  height: number,
): void => {
  const scale = Math.max(width / image.width, height / image.height) * layer.zoom;
  const cropWidth = width / scale;
  const cropHeight = height / scale;
  const sourceX = ((image.width - cropWidth) * layer.focusX) / 100;
  const sourceY = ((image.height - cropHeight) * layer.focusY) / 100;
  context.drawImage(image.source, sourceX, sourceY, cropWidth, cropHeight, 0, 0, width, height);
};

/** Adds deterministic speckles so the preview and exported recipe render alike. */
const drawGrain = (context: CanvasRenderingContext2D, width: number, height: number): void => {
  context.save();
  context.filter = "none";
  context.globalCompositeOperation = "source-atop";
  const dot = Math.max(1, Math.min(width, height) / 200);
  const count = Math.min(260, Math.ceil(Math.min(width, height) / 4));
  for (let index = 0; index < count; index += 1) {
    context.fillStyle = index % 2 === 0 ? "rgba(255,255,255,0.13)" : "rgba(0,0,0,0.13)";
    context.fillRect(
      (((index * 7919) % 997) / 997) * width,
      (((index * 1543) % 991) / 991) * height,
      dot,
      dot,
    );
  }
  context.restore();
};

/** Draws one positioned layer, preserving transparency when no backing frame is selected. */
const drawLayer = (
  context: CanvasRenderingContext2D,
  image: ThumbnailImage,
  layer: ThumbnailLayer,
  canvasWidth: number,
  canvasHeight: number,
): void => {
  const x = (layer.x / 100) * canvasWidth;
  const y = (layer.y / 100) * canvasHeight;
  const width = (layer.width / 100) * canvasWidth;
  const height = (layer.height / 100) * canvasHeight;
  const border = layer.frame === "polaroid" ? Math.min(width, height) * 0.06 : 0;
  const footer = layer.frame === "polaroid" ? height * 0.13 : 0;
  context.save();
  context.translate(x + width / 2, y + height / 2);
  context.rotate((layer.angle * Math.PI) / 180);
  context.translate(-width / 2, -height / 2);
  if (layer.shadow) {
    context.shadowColor = "rgba(0, 0, 0, 0.35)";
    context.shadowBlur = Math.max(8, canvasWidth * 0.012);
    context.shadowOffsetY = Math.max(3, canvasWidth * 0.006);
  }
  if (layer.frame !== "none") {
    context.fillStyle = layer.frame === "polaroid" ? "#fffdf8" : "#ffffff";
    context.beginPath();
    context.roundRect(
      0,
      0,
      width,
      height,
      layer.frame === "rounded" ? Math.min(width, height) * 0.12 : 2,
    );
    context.fill();
  }
  if (layer.frame !== "none") {
    context.shadowColor = "transparent";
    context.shadowBlur = 0;
    context.shadowOffsetY = 0;
  }
  context.beginPath();
  context.roundRect(
    border,
    border,
    width - border * 2,
    height - border * 2 - footer,
    layer.frame === "rounded" ? Math.min(width, height) * 0.1 : 0,
  );
  context.clip();
  context.translate(border, border);
  const filters = {
    none: "none",
    grayscale: "grayscale(1)",
    sepia: "sepia(0.85)",
    warm: "sepia(0.2) saturate(1.25)",
    grain: "none",
  };
  context.filter = filters[layer.effect];
  drawPhoto(context, image, layer, width - border * 2, height - border * 2 - footer);
  if (layer.effect === "grain")
    drawGrain(context, width - border * 2, height - border * 2 - footer);
  context.restore();
};

/** Prepares the canvas background before any image layers are drawn. */
export const beginThumbnail = (
  canvas: HTMLCanvasElement,
  document: ThumbnailDocument,
  width: number,
  height = Math.round(width / aspectValue(document.aspect)),
  jpeg = false,
): CanvasRenderingContext2D => {
  canvas.width = width;
  canvas.height = height;
  const context = canvas.getContext("2d");
  if (!context) throw new Error("Canvas is unavailable");
  const backgrounds = {
    transparent: "transparent",
    white: "#ffffff",
    cream: "#f4ecdc",
    paper: "#f4ecdc",
    charcoal: "#24232a",
  };
  if (document.background !== "transparent" || jpeg) {
    context.fillStyle =
      document.background === "transparent" ? "#ffffff" : backgrounds[document.background];
    context.fillRect(0, 0, width, height);
    if (document.background === "paper") {
      context.fillStyle = "rgba(114, 91, 64, 0.1)";
      const dot = Math.max(1, width / 1200);
      for (let index = 0; index < Math.min(1200, Math.ceil(width / 3)); index += 1)
        context.fillRect(
          (((index * 7919) % 997) / 997) * width,
          (((index * 1543) % 991) / 991) * height,
          dot,
          dot,
        );
    }
  }
  return context;
};

/** Draws one layer under the document's shared composition transform. */
export const drawThumbnailLayer = (
  context: CanvasRenderingContext2D,
  document: ThumbnailDocument,
  layer: ThumbnailLayer,
  image: ThumbnailImage,
  width: number,
  height: number,
): void => {
  context.save();
  context.translate(
    width * (0.5 + document.composition.x / 100),
    height * (0.5 + document.composition.y / 100),
  );
  context.scale(document.composition.scale, document.composition.scale);
  context.translate(-width / 2, -height / 2);
  drawLayer(context, image, layer, width, height);
  context.restore();
};

/** Renders the same document for editor preview and high-resolution download. */
export const renderThumbnail = (
  canvas: HTMLCanvasElement,
  document: ThumbnailDocument,
  images: Map<string, ThumbnailImage>,
  width: number,
  height = Math.round(width / aspectValue(document.aspect)),
  jpeg = false,
): void => {
  const context = beginThumbnail(canvas, document, width, height, jpeg);
  for (const layer of document.layers) {
    const image = images.get(layer.id);
    if (image) drawThumbnailLayer(context, document, layer, image, width, height);
  }
};
