import type { ThemeSettings, ThemeSource, ThemeTemplateName } from "@theme/index";

const themeSettingsStorageKey = "lite-nas.admin-panel.theme.settings";

const isThemeSource = (value: string): value is ThemeSource => {
  return value === "default" || value === "os" || value === "user";
};

const isThemeMode = (value: string): value is ThemeSettings["mode"] => {
  return value === "light" || value === "dark";
};

const isThemeTemplateName = (value: string): value is ThemeTemplateName => {
  return value === "default";
};

export const defaultThemeSettings: ThemeSettings = {
  mode: "dark",
  source: "default",
  templateName: "default",
};

const parseThemeSettings = (value: unknown): ThemeSettings | null => {
  if (
    typeof value !== "object" ||
    value === null ||
    !("source" in value) ||
    !("mode" in value) ||
    !("templateName" in value)
  ) {
    return null;
  }

  const { mode, source, templateName } = value;

  if (
    typeof source !== "string" ||
    typeof mode !== "string" ||
    typeof templateName !== "string" ||
    !isThemeSource(source) ||
    !isThemeMode(mode) ||
    !isThemeTemplateName(templateName)
  ) {
    return null;
  }

  return {
    mode,
    source,
    templateName,
  };
};

export const loadThemeSettings = (): ThemeSettings => {
  if (typeof window === "undefined") {
    return defaultThemeSettings;
  }

  const rawSettings = window.localStorage.getItem(themeSettingsStorageKey);

  if (rawSettings === null) {
    return defaultThemeSettings;
  }

  try {
    const parsedSettings: unknown = JSON.parse(rawSettings);
    return parseThemeSettings(parsedSettings) ?? defaultThemeSettings;
  } catch {
    return defaultThemeSettings;
  }
};

export const saveThemeSettings = (settings: ThemeSettings): void => {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(themeSettingsStorageKey, JSON.stringify(settings));
};
