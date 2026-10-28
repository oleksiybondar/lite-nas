import { loadLocalLanguage } from "@lite-nas/language";
import {
  authDestination,
  readProfileDocument,
  saveLocalSharedUi,
  sharedUiSchema,
  writeProfileDocument,
} from "@lite-nas/profile";
import type { ThemeMode } from "@lite-nas/theme";
import {
  imageConcurrencySchema,
  loginSchema,
  type Me,
  meSchema,
  momentsConfigSchema,
  momentsSettingsSchema,
  sessionSchema,
} from "./schemas";

/** Event raised when a browser session cannot authorize protected requests. */
export const sessionExpiredEvent = "lite-nas-moments-session-expired";

/** Routes embedded Moments API calls around the admin host's gateway API. */
export const momentsApiPath = (path: string): string => {
  if (!path.startsWith("/api/")) throw new Error("Expected a Moments API path.");
  return typeof window !== "undefined" && window.location.pathname.startsWith("/apps/moments/")
    ? `/apps/moments${path}`
    : path;
};

/** Notifies the mounted shell that credentials can no longer be renewed. */
const notifySessionExpired = (): void => {
  if (typeof window !== "undefined") window.dispatchEvent(new Event(sessionExpiredEvent));
};

/** Rotates cookies once and rejects unavailable or malformed renewal responses. */
let pendingRefresh: Promise<boolean> | null = null;

const refreshSession = (): Promise<boolean> => {
  if (pendingRefresh) return pendingRefresh;
  pendingRefresh = (async () => {
    try {
      const response = await fetch(momentsApiPath("/api/auth/refresh"), {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      });
      if (!response.ok) return false;
      sessionSchema.parse(await response.json());
      return true;
    } catch {
      return false;
    } finally {
      pendingRefresh = null;
    }
  })();
  return pendingRefresh;
};

/** Sends a same-origin browser request with the existing HttpOnly auth cookies. */
export const request = async (path: string, init?: RequestInit): Promise<Response> => {
  const response = await fetch(path, { ...init, credentials: "include" });
  if (response.status !== 401 || path.startsWith(momentsApiPath("/api/auth/"))) return response;
  if (!(await refreshSession())) {
    notifySessionExpired();
    return response;
  }
  const retried = await fetch(path, { ...init, credentials: "include" });
  if (retried.status === 401) notifySessionExpired();
  return retried;
};

/** Loads the verified current user, rotating an expired session once. */
export const loadMe = async (): Promise<Me | null> => {
  let response = await request(momentsApiPath("/api/auth/me"));
  if (response.status === 401) {
    if (!(await refreshSession())) {
      notifySessionExpired();
      return null;
    }
    response = await request(momentsApiPath("/api/auth/me"));
  }
  if (response.status === 401) {
    notifySessionExpired();
    return null;
  }
  if (!response.ok) throw new Error("Could not load your session.");
  const body: unknown = await response.json();
  return meSchema.parse(body).data;
};

/** Authenticates a validated credential pair through Moments. */
export const login = async (
  credentials: { login: string; password: string },
  redirectTo = "/",
): Promise<string> => {
  const payload = loginSchema.parse(credentials);
  const destination = authDestination(redirectTo);
  const path = `${momentsApiPath("/api/auth/login")}?${new URLSearchParams({ redirect_to: destination })}`;
  const response = await request(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!response.ok) throw new Error("Sign in failed.");
  return authDestination(
    sessionSchema.parse(await response.json()).data.redirect_to ?? destination,
  );
};

/** Revokes the current browser refresh session. */
export const logout = async (): Promise<void> => {
  const response = await request(momentsApiPath("/api/auth/logout"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
  if (!response.ok) throw new Error("Sign out failed.");
};

/** Returns the server configured initial path for Moments file pickers. */
export const loadMomentsConfig = async (): Promise<string> => {
  const response = await request(momentsApiPath("/api/config"));
  if (!response.ok) throw new Error("Could not load Moments settings.");
  const body: unknown = await response.json();
  return momentsConfigSchema.parse(body).workingDir;
};

/** Saves a validated shared appearance preference through File Service. */
export const saveSharedMode = async (value: ThemeMode): Promise<void> => {
  const preference = sharedUiSchema.parse({
    preferredMode: value,
    preferredLanguage: loadLocalLanguage(),
  });
  saveLocalSharedUi(preference);
  await writeProfileDocument(
    (path, body) =>
      request(path, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    "ui",
    sharedUiSchema,
    preference,
  );
};

/** Reads the signed-in user's gallery setting, then the browser fallback. */
export const loadGalleryConcurrency = async (): Promise<number> => {
  try {
    const settings = await readProfileDocument(request, "moments", momentsSettingsSchema);
    if (settings) {
      return settings.gallery.maxConcurrentImageLoads;
    }
  } catch {
    // Continue with a validated browser value when profile storage is unavailable.
  }
  const stored = Number(
    typeof window === "undefined" ? null : window.localStorage.getItem("moments-image-concurrency"),
  );
  const parsed = imageConcurrencySchema.safeParse(stored);
  return parsed.success ? parsed.data : 10;
};

let gallerySaveQueue = Promise.resolve();

/** Saves the gallery group to this user's profile file in change order. */
export const saveGalleryConcurrency = async (value: number): Promise<void> => {
  const settings = momentsSettingsSchema.parse({ gallery: { maxConcurrentImageLoads: value } });
  if (typeof window !== "undefined")
    window.localStorage.setItem("moments-image-concurrency", String(value));
  const pending = gallerySaveQueue.then(() =>
    writeProfileDocument(
      (path, body) =>
        request(path, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        }),
      "moments",
      momentsSettingsSchema,
      settings,
    ),
  );
  gallerySaveQueue = pending.catch(() => undefined);
  return pending;
};
