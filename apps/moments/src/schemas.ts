import { hostPathSchema } from "@lite-nas/file-browser";
import { z } from "zod";
import { galleryLayoutModeSchema } from "./GalleryLayout";

/** Credentials accepted by the shared LiteNAS login contract. */
export const loginSchema = z.object({
  login: z
    .string()
    .min(1)
    .regex(/^(?:[A-Za-z0-9._-]+|[^@\s]+@[^@\s]+\.[^@\s]+)$/),
  password: z.string().min(4),
});

/** Verified current-user response from the Moments auth API. */
export const meSchema = z.object({
  data: z.object({
    authenticated: z.literal(true),
    auth_type: z.string(),
    user: z.object({
      id: z.string(),
      login: z.string(),
      full_name: z.string().optional(),
      avatar_url: z.string().optional(),
    }),
    roles: z.array(z.string()),
    scopes: z.array(z.string()),
  }),
});

/** Session envelope returned after login or refresh. */
export const sessionSchema = z.object({
  data: z.object({
    authenticated: z.literal(true),
    user: z.object({ id: z.string() }),
    redirect_to: z.string().max(2048).optional(),
  }),
});

/** Current user's validated browser profile. */
export type Me = z.infer<typeof meSchema>["data"];

/** Runtime contract for the authenticated Moments picker configuration. */
export const momentsConfigSchema = z.object({ workingDir: hostPathSchema });

/** Gallery metadata returned by list, detail, and create routes. */
export const gallerySchema = z.object({
  id: z.uuid(),
  type: z.enum(["ALBUM", "COLLECTION"]),
  name: z.string().min(1),
  description: z.string(),
  ownerId: z.string().min(1),
  visibility: z.enum(["PRIVATE", "PUBLIC"]),
  uploadDirectory: hostPathSchema.nullable(),
  wallRenderStrategy: galleryLayoutModeSchema,
  thumbnailImageUrl: z.string().startsWith("/api/galleries/").nullable(),
  thumbnailPath: hostPathSchema.nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
  tags: z
    .array(
      z.object({
        id: z.uuid(),
        type: z.enum(["AUTO", "USER"]),
        category: z.string(),
        value: z.string(),
        count: z.number().int().nonnegative(),
      }),
    )
    .optional(),
});

/** Validated gallery returned from Moments. */
export type GalleryRecord = z.infer<typeof gallerySchema>;

/** Returns the city names published for a gallery's scanned media. */
export const galleryLocationNames = (gallery: GalleryRecord): string[] =>
  gallery.tags?.filter((tag) => tag.category === "LOCATION").map((tag) => tag.value) ?? [];

/** Durable owner-visible media scan progress returned by Moments. */
export const galleryScanSchema = z.object({
  id: z.uuid(),
  galleryId: z.uuid(),
  status: z.enum([
    "PENDING",
    "DISCOVERING",
    "RUNNING",
    "COMPLETED",
    "PARTIAL",
    "FAILED",
    "CANCELLED",
  ]),
  totalItems: z.number().int().nonnegative().nullable(),
  processedItems: z.number().int().nonnegative(),
  successfulItems: z.number().int().nonnegative(),
  failedItems: z.number().int().nonnegative(),
  createdAt: z.string(),
  startedAt: z.string().nullable(),
  finishedAt: z.string().nullable(),
  errorSummary: z.string().nullable(),
});

/** Validated media scan job shown in gallery management. */
export type GalleryScan = z.infer<typeof galleryScanSchema>;

/** Owner's most recent media scan jobs for one gallery. */
export const galleryScansSchema = z.object({ items: z.array(galleryScanSchema) });

/** Paged gallery search response. */
export const galleryListSchema = z.object({
  items: z.array(gallerySchema),
  metadata: z.object({
    page: z.number().int().positive(),
    size: z.number().int().positive(),
    total_count: z.number().int().nonnegative(),
    total_pages: z.number().int().nonnegative(),
  }),
});

/** Media types accepted by the paged gallery search API. */
export const galleryMediaTypeSchema = z.enum(["IMAGE", "VIDEO", "AUDIO", "OTHER", "LIVE_PHOTO"]);

/** Selected media types sent as repeated query values. */
export const galleryMediaTypeSelectionSchema = z
  .array(galleryMediaTypeSchema)
  .max(5)
  .refine((values) => new Set(values).size === values.length);

/** Media type value recognized by the gallery search API. */
export type GalleryMediaType = z.infer<typeof galleryMediaTypeSchema>;

/** Supported search and filter fields for the paged gallery API. */
export const gallerySearchSchema = z.object({
  q: z.string().trim().max(160).optional(),
  type: z.enum(["ALBUM", "COLLECTION"]).optional(),
  visibility: z.enum(["PRIVATE", "PUBLIC"]).optional(),
  mediaType: galleryMediaTypeSelectionSchema.optional(),
  location: z.string().trim().max(160).optional(),
});

/** One explicit collection member returned by Moments. */
export const collectionReferenceSchema = z.object({
  id: z.uuid(),
  type: z.enum(["FILE", "DIRECTORY"]),
  path: hostPathSchema,
  createdAt: z.string(),
});

/** Collection membership list response. */
export const collectionReferencesSchema = z.object({ items: z.array(collectionReferenceSchema) });

/** Validates gallery creation before sending it to Moments. */
export const createGallerySchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("ALBUM"),
    name: z.string().trim().min(1).max(160),
    description: z.string().max(4000).default(""),
    visibility: z.enum(["PRIVATE", "PUBLIC"]).default("PUBLIC"),
    uploadDirectory: hostPathSchema,
    wallRenderStrategy: galleryLayoutModeSchema.optional(),
  }),
  z.object({
    type: z.literal("COLLECTION"),
    name: z.string().trim().min(1).max(160),
    description: z.string().max(4000).default(""),
    visibility: z.enum(["PRIVATE", "PUBLIC"]).default("PUBLIC"),
    wallRenderStrategy: galleryLayoutModeSchema.optional(),
  }),
]);

/** Fields accepted by the owner-only gallery metadata update route. */
export const updateGallerySchema = z.object({
  name: z.string().trim().min(1).max(160),
  description: z.string().max(4000),
  visibility: z.enum(["PRIVATE", "PUBLIC"]),
  wallRenderStrategy: galleryLayoutModeSchema.optional(),
});

/** Editable user tags returned separately from scan-derived tags. */
export const userTagsSchema = z.object({
  items: z.array(z.object({ id: z.uuid(), value: z.string() })),
});

/** One new user tag accepted by the gallery tag route. */
export const userTagInputSchema = z.object({ value: z.string().trim().min(1).max(160) });

/** One login name to resolve to a stable user ID before sharing. */
export const shareUsernameSchema = z.object({
  username: z
    .string()
    .trim()
    .min(1)
    .max(128)
    .regex(/^[A-Za-z0-9_][A-Za-z0-9._-]*$/),
});

/** Operator choice for parallel lazy image downloads. */
export const imageConcurrencySchema = z.number().int().min(1).max(20);

/** App-specific user settings grouped for the Moments profile JSON file. */
export const momentsSettingsSchema = z.object({
  gallery: z.object({ maxConcurrentImageLoads: imageConcurrencySchema }),
});

/** Persisted Moments settings after runtime validation. */
export type MomentsSettings = z.infer<typeof momentsSettingsSchema>;
