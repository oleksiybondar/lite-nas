import { LanguageProvider, type LanguageTransport } from "@lite-nas/language";
import { loadLocalSharedUi, loadSharedUiPreference, sharedUiStorageKey } from "@lite-nas/profile";
import { createAppTheme } from "@lite-nas/theme";
import { resolveThemeSettings } from "@lite-nas/theme/manager/resolveThemeSettings";
import { defaultThemeSettings } from "@lite-nas/theme/manager/storage";
import { CssBaseline, useMediaQuery } from "@mui/material";
import { ThemeProvider } from "@mui/material/styles";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";

import { FileExplorerPage } from "./FileExplorerPage";

/** Uses the same-origin file service profile route used by the theme setting. */
const languageTransport: LanguageTransport = {
  read: (url) => fetch(url, { credentials: "same-origin" }),
  write: (url, body) =>
    fetch(url, {
      method: "PUT",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
};

/** Applies the authenticated shared preference or its local browser fallback. */
export function App(): ReactElement {
  const prefersDark = useMediaQuery("(prefers-color-scheme: dark)");
  const [preferredMode, setPreferredMode] = useState(
    () => loadLocalSharedUi()?.preferredMode ?? null,
  );

  useEffect(() => {
    let active = true;
    void loadSharedUiPreference((url) => fetch(url, { credentials: "same-origin" }))
      .then((document) => {
        if (!active || !document) return;
        setPreferredMode(document.preferredMode);
      })
      .catch(() => undefined);
    const onStorage = (event: StorageEvent): void => {
      if (event.key === sharedUiStorageKey) {
        setPreferredMode(loadLocalSharedUi()?.preferredMode ?? null);
      }
    };

    window.addEventListener("storage", onStorage);
    return () => {
      active = false;
      window.removeEventListener("storage", onStorage);
    };
  }, []);

  const settings = preferredMode
    ? { ...defaultThemeSettings, mode: preferredMode, source: "user" as const }
    : defaultThemeSettings;
  const { mode, templateName } = resolveThemeSettings(settings, prefersDark ? "dark" : "light");

  return (
    <ThemeProvider theme={createAppTheme(templateName, mode)}>
      <CssBaseline />
      <LanguageProvider userId="session" transport={languageTransport}>
        <FileExplorerPage />
      </LanguageProvider>
    </ThemeProvider>
  );
}
