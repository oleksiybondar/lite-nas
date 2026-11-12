import { useApi } from "@hooks/useApi";
import { readProfileAvatar } from "@lite-nas/profile";
import { useEffect, useState } from "react";

/** Loads the current user's avatar through the shared authenticated API pipeline. */
export const useProfileAvatar = (): string | null => {
  const { get } = useApi();
  const [url, setUrl] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    let objectUrl: string | null = null;
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The request and object URL lifecycle must stay together to revoke stale URLs safely.
    const load = async (): Promise<void> => {
      try {
        const blob = await readProfileAvatar((path) => get(path).execute());
        if (!blob) {
          if (active) {
            if (objectUrl) URL.revokeObjectURL(objectUrl);
            objectUrl = null;
            setUrl(null);
          }
          return;
        }
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
  }, [get]);

  return url;
};
