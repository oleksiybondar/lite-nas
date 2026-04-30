import { ThemeManagerContext } from "@contexts/theme-manager-context";
import { useMediaQuery } from "@mui/material";
import type { ThemeMode, ThemeSettings, ThemeSource, ThemeTemplateName } from "@theme/index";
import { resolveThemeSettings } from "@theme/manager/resolveThemeSettings";
import { loadThemeSettings } from "@theme/manager/storage";
import { themeRegistry } from "@theme/registry";
import type { PropsWithChildren, ReactElement } from "react";
import { useState } from "react";

export const ThemeManagerProvider = ({ children }: PropsWithChildren): ReactElement => {
  const prefersDarkMode = useMediaQuery("(prefers-color-scheme: dark)");
  const [settings, setSettings] = useState<ThemeSettings>(() => {
    return loadThemeSettings();
  });

  const osMode: ThemeMode = prefersDarkMode ? "dark" : "light";
  const resolvedThemeSettings = resolveThemeSettings(settings, osMode);

  return (
    <ThemeManagerContext.Provider
      value={{
        availableTemplates: Object.keys(themeRegistry) as ThemeTemplateName[],
        mode: settings.mode,
        resolvedMode: resolvedThemeSettings.mode,
        resolvedTemplateName: resolvedThemeSettings.templateName,
        setMode: (mode: ThemeMode) => {
          setSettings((currentSettings) => ({
            ...currentSettings,
            mode,
          }));
        },
        setSettings,
        setSource: (source: ThemeSource) => {
          setSettings((currentSettings) => ({
            ...currentSettings,
            source,
          }));
        },
        setTemplateName: (templateName: ThemeTemplateName) => {
          setSettings((currentSettings) => ({
            ...currentSettings,
            templateName,
          }));
        },
        source: settings.source,
        templateName: settings.templateName,
      }}
    >
      {children}
    </ThemeManagerContext.Provider>
  );
};
