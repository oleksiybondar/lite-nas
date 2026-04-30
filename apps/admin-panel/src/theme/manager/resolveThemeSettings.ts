import type { ResolvedThemeSettings, ThemeMode, ThemeSettings } from "@theme/types";

const defaultThemeMode: ThemeMode = "dark";
const defaultThemeTemplateName = "default";

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
