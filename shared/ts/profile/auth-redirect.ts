import { z } from "zod";

/** Browser return paths are bounded and cannot change hosts or escape through backslashes. */
const redirectPathSchema = z
  .string()
  .min(1)
  .max(2048)
  .refine((path) => path.startsWith("/") && !path.startsWith("//") && !path.includes("\\"))
  .refine((path) =>
    [...path].every((character) => {
      const code = character.charCodeAt(0);
      return code >= 32 && code !== 127;
    }),
  );

/** Resolves a login return target to a safe route within the current SPA. */
export const authDestination = (value: unknown): string => {
  const result = redirectPathSchema.safeParse(value);
  if (!result.success) return "/";
  try {
    const target = new URL(result.data, "https://lite-nas.invalid");
    const decodedPath = decodeURIComponent(target.pathname);
    if (
      target.origin !== "https://lite-nas.invalid" ||
      !redirectPathSchema.safeParse(decodedPath).success ||
      decodedPath.replace(/\/+$/, "") === "/login"
    )
      return "/";
    return `${target.pathname}${target.search}${target.hash}`;
  } catch {
    return "/";
  }
};

/** Builds a reload-safe login route with an encoded protected destination. */
export const loginRedirectPath = (path: string): string =>
  `/login?${new URLSearchParams({ redirect_to: authDestination(path) })}`;

/** Restores a query-persisted destination, accepting router state from older login links. */
export const loginDestination = (search: string, stateFrom?: unknown): string =>
  authDestination(new URLSearchParams(search).get("redirect_to") ?? stateFrom);
