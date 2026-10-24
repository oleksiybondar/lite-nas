// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: URL decoding and file-kind selection are one validated adapter.
import { z } from "zod";
import { galleryLayoutModeSchema } from "./GalleryLayout";

/** Media supported by the fixture-backed gallery viewer. */
export type GalleryItem = {
  id: string;
  url: string;
  filename: string;
  kind: "image" | "video" | "unsupported";
};

export const galleryFileUrlSchema = z
  .string()
  .min(4)
  .refine((value) => {
    if (!value.startsWith("/f/") || value.includes("?") || value.includes("#")) return false;
    try {
      const path = decodeURIComponent(value.slice(3));
      return (
        path.startsWith("home/") && !path.split("/").some((part) => part === ".." || part === ".")
      );
    } catch {
      return false;
    }
  }, "Expected an original File Service path");

/** Validates the demo's file paths and its initial wall layout. */
export const galleryFixtureSchema = z.strictObject({
  defaultLayout: galleryLayoutModeSchema,
  items: z.array(galleryFileUrlSchema),
});

/** Derives the attachment route from a validated original File Service URL. */
export const galleryDownloadUrl = (url: string): string | null => {
  const result = galleryFileUrlSchema.safeParse(url);
  return result.success ? `/d/${result.data.slice(3)}` : null;
};

/** Adapts validated original-file paths into ordered presentation items. */
export const galleryItemsFromPaths = (paths: readonly string[]): GalleryItem[] =>
  paths.map((url, index) => {
    const filename = decodeURIComponent(url.slice(url.lastIndexOf("/") + 1));
    const extension = filename.split(".").pop()?.toLowerCase();
    let kind: GalleryItem["kind"] = "unsupported";
    if (extension && ["jpg", "jpeg", "png", "webp", "gif", "avif"].includes(extension))
      kind = "image";
    else if (extension && ["mp4", "mov", "webm", "m4v"].includes(extension)) kind = "video";
    return { id: `${index}:${url}`, url, filename, kind };
  });

/** Allowed time that a ready still image remains visible in slideshow mode. */
export const imageDelaySchema = z.union([z.literal(1000), z.literal(3000), z.literal(5000)]);

/** User choice for attempting video playback on selection. */
export const videoAutoplaySchema = z.boolean();
