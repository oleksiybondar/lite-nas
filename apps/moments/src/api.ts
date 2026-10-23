import { saveLocalSharedUi, sharedUiSchema, writeProfileDocument } from "@lite-nas/profile";
import type { ThemeMode } from "@lite-nas/theme";
import { loginSchema, type Me, meSchema, sessionSchema } from "./schemas";

/** Sends a same-origin browser request with the existing HttpOnly auth cookies. */
export const request = async (path: string, init?: RequestInit): Promise<Response> => {
  const response = await fetch(path, { ...init, credentials: "include" });
  if (response.status !== 401 || path.startsWith("/api/auth/")) return response;
  const refreshed = await fetch("/api/auth/refresh", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
  if (!refreshed.ok) return response;
  sessionSchema.parse(await refreshed.json());
  return fetch(path, { ...init, credentials: "include" });
};

/** Loads the verified current user, rotating an expired session once. */
export const loadMe = async (): Promise<Me | null> => {
  let response = await request("/api/auth/me");
  if (response.status === 401) {
    const refreshed = await request("/api/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: "{}",
    });
    if (!refreshed.ok) return null;
    sessionSchema.parse(await refreshed.json());
    response = await request("/api/auth/me");
  }
  if (response.status === 401) return null;
  if (!response.ok) throw new Error("Could not load your session.");
  const body: unknown = await response.json();
  return meSchema.parse(body).data;
};

/** Authenticates a validated credential pair through Moments. */
export const login = async (credentials: { login: string; password: string }): Promise<void> => {
  const payload = loginSchema.parse(credentials);
  const response = await request("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!response.ok) throw new Error("Sign in failed.");
  sessionSchema.parse(await response.json());
};

/** Revokes the current browser refresh session. */
export const logout = async (): Promise<void> => {
  const response = await request("/api/auth/logout", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
  if (!response.ok) throw new Error("Sign out failed.");
};

/** Saves a validated shared appearance preference through File Service. */
export const saveSharedMode = async (value: ThemeMode): Promise<void> => {
  const preference = sharedUiSchema.parse({ preferredMode: value });
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
