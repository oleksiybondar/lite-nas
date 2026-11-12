import { readProfileAvatar } from "@lite-nas/profile";
import { useEffect, useState } from "react";
import { request } from "./api";

/** Loads and refreshes the current user's avatar for the shared header and profile panel. */
export const useProfileAvatar = (): string | null => {
  const [url, setUrl] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    let objectUrl: string | null = null;
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The object URL lifecycle must revoke stale values during reloads and unmount.
    const load = async (): Promise<void> => {
      try {
        const blob = await readProfileAvatar(request);
        if (!blob) return;
        const nextUrl = URL.createObjectURL(blob);
        if (!active) {
          URL.revokeObjectURL(nextUrl);
          return;
        }
        if (objectUrl) URL.revokeObjectURL(objectUrl);
        objectUrl = nextUrl;
        setUrl(nextUrl);
      } catch {
        if (active) setUrl(null);
      }
    };
    const refresh = (): void => {
      void load();
    };
    void load();
    window.addEventListener("lite-nas-avatar-updated", refresh);
    return () => {
      active = false;
      window.removeEventListener("lite-nas-avatar-updated", refresh);
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, []);

  return url;
};
