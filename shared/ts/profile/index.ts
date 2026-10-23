import { z } from "zod";

/** Shared settings document stored as `.lite-nas/ui.json`. */
export const sharedUiSchema = z.object({ preferredMode: z.enum(["light", "dark"]) });

/** Browser upload contract for fixed PNG/JPEG avatars. */
export const avatarFileSchema = z
  .custom<File>((value) => typeof File !== "undefined" && value instanceof File)
  .refine(
    (file) =>
      file.type === "image/png" || file.type === "image/jpeg" || /\.(png|jpe?g)$/i.test(file.name),
    "Choose a PNG or JPEG image.",
  )
  .refine(
    (file) => file.size > 0 && file.size <= 5 * 1024 * 1024,
    "Image must be 5 MB or smaller.",
  );

/** Browser fallback key shared by LiteNAS applications on the same origin. */
export const sharedUiStorageKey = "lite-nas.ui.preferred-mode";

/** Validated shared appearance document. */
export type SharedUi = z.infer<typeof sharedUiSchema>;

/** Reads a validated shared preference from this browser or returns null. */
export const loadLocalSharedUi = (): SharedUi | null => {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(sharedUiStorageKey);
  if (raw === null) return null;
  try {
    const parsed: unknown = JSON.parse(raw);
    const result = sharedUiSchema.safeParse(parsed);
    return result.success ? result.data : null;
  } catch {
    return null;
  }
};

/** Replicates an authenticated shared preference into the browser fallback. */
export const saveLocalSharedUi = (value: SharedUi): void => {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(sharedUiStorageKey, JSON.stringify(sharedUiSchema.parse(value)));
};

/** Reads the authenticated UI file, caches it locally, and falls back to the browser copy. */
export const loadSharedUiPreference = async (
  request?: (url: string) => Promise<Response>,
): Promise<SharedUi | null> => {
  if (request) {
    try {
      const document = await readProfileDocument(request, "ui", sharedUiSchema);
      if (document) {
        saveLocalSharedUi(document);
        return document;
      }
    } catch {
      // The browser copy remains useful when the profile service is unavailable.
    }
  }
  return loadLocalSharedUi();
};

export { ProfilePreferencesPanel } from "./ProfilePreferencesPanel";

/** Reads a profile JSON file through an app-supplied authenticated request pipeline. */
export const readProfileDocument = async <T>(
  request: (url: string) => Promise<Response>,
  file: "ui" | "admin-panel",
  schema: z.ZodType<T>,
): Promise<T | null> => {
  const response = await request(`/f/profile/${file}`);
  if (response.status === 404) return null;
  if (!response.ok) throw new Error(`Profile read failed (${response.status}).`);
  const data: unknown = await response.json();
  return schema.parse(data);
};

/** Saves a validated profile JSON file through an app-supplied authenticated request pipeline. */
export const writeProfileDocument = async <T>(
  request: (url: string, body: { document: T }) => Promise<Response>,
  file: "ui" | "admin-panel",
  schema: z.ZodType<T>,
  document: T,
): Promise<void> => {
  const validated = schema.parse(document);
  const response = await request(`/f/profile/${file}`, { document: validated });
  if (!response.ok) throw new Error(`Profile save failed (${response.status}).`);
};

/** Reads the current account's validated avatar through an app's authenticated request pipeline. */
export const readProfileAvatar = async (
  request: (url: string) => Promise<Response>,
): Promise<Blob | null> => {
  const response = await request("/f/profile/avatar");
  if (response.status === 404) return null;
  if (!response.ok) throw new Error(`Avatar read failed (${response.status}).`);
  const blob = await response.blob();
  return blob.type === "image/png" || blob.type === "image/jpeg" ? blob : null;
};

/** Uploads a validated avatar to the fixed current-account route. */
export const writeProfileAvatar = async (
  request: (url: string, image: File) => Promise<Response>,
  file: File,
): Promise<void> => {
  const image = avatarFileSchema.parse(file);
  const response = await request("/f/profile/avatar", image);
  if (!response.ok) throw new Error(`Avatar upload failed (${response.status}).`);
};
