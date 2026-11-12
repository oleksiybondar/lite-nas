import { loadProfileSettings, localPreferredMode } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import { useAuth } from "@hooks/useAuth";
import { useThemeManager } from "@hooks/useThemeManager";
import { themeSettingsStorageKey } from "@theme/manager/storage";
import type { PropsWithChildren, ReactElement } from "react";
import { useEffect, useState } from "react";

/** Applies the file-backed theme or the shared browser preference when no app theme exists. */
const applyProfileTheme = (
  admin: Awaited<ReturnType<typeof loadProfileSettings>>,
  setSettings: ReturnType<typeof useThemeManager>["setSettings"],
): void => {
  if (admin) {
    setSettings(admin.theme);
    return;
  }
  if (window.localStorage.getItem(themeSettingsStorageKey) !== null) return;
  const mode = localPreferredMode();
  if (mode) setSettings({ mode, source: "user", templateName: "default" });
};

/** Loads authenticated profile settings before mounting protected app content. */
export const ProfileSettingsLoader = ({ children }: PropsWithChildren): ReactElement | null => {
  const { isAuthenticated, me } = useAuth();
  const { get } = useApi();
  const { setSettings } = useThemeManager();
  const [loadedUser, setLoadedUser] = useState<string | null>(null);
  const userId = isAuthenticated ? (me?.user.id ?? null) : null;

  useEffect(() => {
    if (userId === null) setLoadedUser(null);
  }, [userId]);

  useEffect(() => {
    if (userId === null) return;
    let active = true;
    void loadProfileSettings({ get })
      .then((admin) => {
        if (!active) return;
        applyProfileTheme(admin, setSettings);
        setLoadedUser(userId);
      })
      .catch(() => {
        if (active) setLoadedUser(userId);
      });
    return () => {
      active = false;
    };
  }, [get, setSettings, userId]);

  if (userId !== null && loadedUser !== userId) return null;
  return <>{children}</>;
};
