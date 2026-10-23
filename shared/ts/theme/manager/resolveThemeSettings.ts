import type { ResolvedThemeSettings, ThemeMode, ThemeSettings } from "../types";

const defaultThemeMode: ThemeMode = "dark";
const defaultThemeTemplateName = "default";

/** Resolves default, OS, or saved user preferences into active theme options. */
export const resolveThemeSettings = (
  settings: ThemeSettings,
  osMode: ThemeMode,
): ResolvedThemeSettings => {
  if (settings.source === "default") {
    return {
      mode: defaultThemeMode,
      templateName: defaultThemeTemplateName,
    };
  }

  if (settings.source === "os") {
    return {
      mode: osMode,
      templateName: defaultThemeTemplateName,
    };
  }

  return {
    mode: settings.mode,
    templateName: settings.templateName,
  };
};
