import { createAppTheme } from "@lite-nas/theme";
import { resolveThemeSettings } from "@lite-nas/theme/manager/resolveThemeSettings";
import { loadThemeSettings, themeSettingsStorageKey } from "@lite-nas/theme/manager/storage";
import { CssBaseline, useMediaQuery } from "@mui/material";
import { ThemeProvider } from "@mui/material/styles";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";

import { FileExplorerPage } from "./FileExplorerPage";

/** Applies the shared LiteNAS theme around the standalone file browser. */
export function App(): ReactElement {
  const prefersDark = useMediaQuery("(prefers-color-scheme: dark)");
  const [settings, setSettings] = useState(loadThemeSettings);

  useEffect(() => {
    const onStorage = (event: StorageEvent): void => {
      if (event.key === themeSettingsStorageKey) {
        setSettings(loadThemeSettings());
      }
    };

    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  const { mode, templateName } = resolveThemeSettings(settings, prefersDark ? "dark" : "light");

  return (
    <ThemeProvider theme={createAppTheme(templateName, mode)}>
      <CssBaseline />
      <FileExplorerPage />
    </ThemeProvider>
  );
}
