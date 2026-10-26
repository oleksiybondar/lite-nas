import { directoryListingSchema, fileServicePath, hostPathSchema } from "@lite-nas/file-browser";
import { z } from "zod";
import { momentsApiPath, request } from "./api";
import {
  collectionReferencesSchema,
  createGallerySchema,
  type GalleryRecord,
  type GalleryScan,
  galleryListSchema,
  galleryScanSchema,
  galleryScansSchema,
  gallerySchema,
  gallerySearchSchema,
  shareUsernameSchema,
  updateGallerySchema,
  userTagInputSchema,
  userTagsSchema,
} from "./schemas";

/** Reads a protected JSON response through the shared session renewal path. */
async function readJson(path: string, init?: RequestInit): Promise<unknown> {
  const response = await request(path, init);
  if (!response.ok) throw new Error(`Moments request failed (HTTP ${response.status}).`);
  return response.json();
}

/** Encodes validated gallery filters using repeated values where supported. */
function appendGalleryFilters(
  query: URLSearchParams,
  filters: z.infer<typeof gallerySearchSchema>,
): void {
  for (const [key, value] of Object.entries(filters)) {
    if (typeof value === "string") query.set(key, value);
  }
  for (const mediaType of filters.mediaType ?? []) query.append("mediaType", mediaType);
}

/** Returns one page of galleries visible to the signed-in user. */
export async function listGalleries(
  page: number,
  ownerId?: string,
  filters: z.input<typeof gallerySearchSchema> = {},
): Promise<z.infer<typeof galleryListSchema>> {
  const parsedPage = z.number().int().positive().max(100000).parse(page);
  const query = new URLSearchParams({ page: String(parsedPage), size: "20" });
  if (ownerId) query.set("ownerId", z.string().min(1).max(160).parse(ownerId));
  appendGalleryFilters(query, gallerySearchSchema.parse(filters));
  return galleryListSchema.parse(await readJson(`${momentsApiPath("/api/galleries")}?${query}`));
}

/** Loads one gallery after validating its identifier. */
export async function getGallery(id: string): Promise<GalleryRecord> {
  return gallerySchema.parse(
    await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(id)}`)),
  );
}

/** Starts an owner-requested full media scan and returns its durable job. */
export async function startGalleryScan(galleryId: string): Promise<GalleryScan> {
  return galleryScanSchema.parse(
    await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(galleryId)}/scans`), {
      method: "POST",
    }),
  );
}

/** Reads the most recent owner-visible scan job for a gallery. */
export async function latestGalleryScan(galleryId: string): Promise<GalleryScan | null> {
  const result = galleryScansSchema.parse(
    await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(galleryId)}/scans?limit=1`)),
  );
  return result.items[0] ?? null;
}

/** Polls one scan, renewing the worker's access token through the backend. */
export async function getGalleryScan(scanId: string): Promise<GalleryScan> {
  return galleryScanSchema.parse(
    await readJson(momentsApiPath(`/api/scans/${z.uuid().parse(scanId)}`)),
  );
}

/** Requests cancellation of an owner-started scan. */
export async function cancelGalleryScan(scanId: string): Promise<GalleryScan> {
  return galleryScanSchema.parse(
    await readJson(momentsApiPath(`/api/scans/${z.uuid().parse(scanId)}`), {
      method: "DELETE",
    }),
  );
}

/** Creates an album or collection from a validated form payload. */
export async function createGallery(
  input: z.input<typeof createGallerySchema>,
): Promise<GalleryRecord> {
  const body = createGallerySchema.parse(input);
  return gallerySchema.parse(
    await readJson(momentsApiPath("/api/galleries"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
  );
}

/** Changes owner-editable metadata for one gallery. */
export async function updateGallery(
  id: string,
  input: z.input<typeof updateGallerySchema>,
): Promise<GalleryRecord> {
  return gallerySchema.parse(
    await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(id)}`), {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(updateGallerySchema.parse(input)),
    }),
  );
}

/** Loads only owner-editable tags, excluding scan-derived metadata. */
export async function listUserTags(id: string): Promise<z.infer<typeof userTagsSchema>["items"]> {
  return userTagsSchema.parse(
    await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(id)}/user-tags`)),
  ).items;
}

/** Adds a user tag to an owned gallery. */
export async function addUserTag(id: string, value: string): Promise<void> {
  await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(id)}/user-tags`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(userTagInputSchema.parse({ value })),
  });
}

/** Removes one owner-editable tag from a gallery. */
export async function removeUserTag(id: string, tagId: string): Promise<void> {
  const response = await request(
    momentsApiPath(`/api/galleries/${z.uuid().parse(id)}/user-tags/${z.uuid().parse(tagId)}`),
    { method: "DELETE" },
  );
  if (!response.ok) throw new Error(`Tag removal failed (HTTP ${response.status}).`);
}

/** Resolves a login name in RBAC and shares one owned gallery with its stable UID. */
export async function shareGalleryByUsername(id: string, username: string): Promise<void> {
  await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(id)}/shares`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(shareUsernameSchema.parse({ username })),
  });
}

/** Returns explicit collection file and directory memberships. */
export async function listReferences(
  id: string,
): Promise<z.infer<typeof collectionReferencesSchema>["items"]> {
  const body = await readJson(momentsApiPath(`/api/collections/${z.uuid().parse(id)}/references`));
  return collectionReferencesSchema.parse(body).items;
}

/** Adds or removes a file from an owned collection. */
export async function changeCollectionFile(id: string, path: string, add: boolean): Promise<void> {
  const response = await request(momentsApiPath(`/api/collections/${z.uuid().parse(id)}/files`), {
    method: add ? "POST" : "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path: hostPathSchema.parse(path) }),
  });
  if (!response.ok) throw new Error(`Collection update failed (HTTP ${response.status}).`);
}

/** Adds an authorized folder to an owned collection. */
export async function addCollectionDirectory(id: string, path: string): Promise<void> {
  await readJson(momentsApiPath(`/api/collections/${z.uuid().parse(id)}/directories`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path: hostPathSchema.parse(path) }),
  });
}

/** Removes one directory membership while leaving the source album untouched. */
export async function removeCollectionDirectory(id: string, path: string): Promise<void> {
  await readJson(momentsApiPath(`/api/collections/${z.uuid().parse(id)}/directories`), {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path: hostPathSchema.parse(path) }),
  });
}

/** Publishes a generated PNG or JPEG as an owned gallery's cover image. */
export async function saveGalleryThumbnail(id: string, image: Blob): Promise<GalleryRecord> {
  if (
    !(
      ["image/png", "image/jpeg"].includes(image.type) &&
      image.size > 0 &&
      image.size <= 16 * 1024 * 1024
    )
  ) {
    throw new Error("Thumbnail must be a PNG or JPEG up to 16 MB.");
  }
  return gallerySchema.parse(
    await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(id)}/thumbnail/image`), {
      method: "PUT",
      headers: { "Content-Type": image.type },
      body: image,
    }),
  );
}

/** Selects an authorized NAS image as a gallery's cover source. */
export async function chooseGalleryThumbnailPath(id: string, path: string): Promise<GalleryRecord> {
  return gallerySchema.parse(
    await readJson(momentsApiPath(`/api/galleries/${z.uuid().parse(id)}`), {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ thumbnailPath: hostPathSchema.parse(path) }),
    }),
  );
}

/** Resolves a gallery's source files with the caller's File Service permissions. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: album directories and collection references are resolved under one access path.
export async function galleryFiles(gallery: GalleryRecord, signal: AbortSignal): Promise<string[]> {
  const directories: string[] = [];
  const files: string[] = [];
  if (gallery.type === "ALBUM") {
    if (gallery.uploadDirectory) directories.push(gallery.uploadDirectory);
  } else {
    for (const reference of await listReferences(gallery.id)) {
      if (reference.type === "FILE") files.push(reference.path);
      else directories.push(reference.path);
    }
  }
  for (const directory of directories) {
    const body = await readJson(`${fileServicePath("/ls", directory)}?recursive&files-only`, {
      signal,
    });
    const listing = directoryListingSchema.parse(body);
    files.push(...listing.filter((entry) => entry.kind === "file").map((entry) => entry.path));
  }
  return [...new Set(files)].map((path) => hostPathSchema.parse(path));
}
