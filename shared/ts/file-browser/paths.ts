import { hostPathSchema } from "./schemas";

/** Returns the last display segment of one validated absolute host path. */
export function displayName(path: string): string {
  hostPathSchema.parse(path);
  return path === "/" ? "/" : path.slice(path.lastIndexOf("/") + 1);
}

/** Returns the parent directory, or null at the host root. */
export function parentPath(path: string): string | null {
  hostPathSchema.parse(path);
  if (path === "/") {
    return null;
  }
  return path.slice(0, path.lastIndexOf("/")) || "/";
}

/** Encodes each host path segment without introducing encoded slashes. */
export function fileServicePath(mount: "/ls" | "/f", path: string): string {
  hostPathSchema.parse(path);
  return `${mount}${path.split("/").map(encodeURIComponent).join("/")}`;
}
