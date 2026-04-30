import type { ThemeOptions } from "@mui/material/styles";

export type ThemeSource = "default" | "os" | "user";
export type ThemeMode = "light" | "dark";

export interface ThemeTemplate {
  dark: ThemeOptions;
  light: ThemeOptions;
}

export type ThemeTemplateName = "default";

export type ThemeRegistry = Record<ThemeTemplateName, ThemeTemplate>;

export interface ThemeSettings {
  mode: ThemeMode;
  source: ThemeSource;
  templateName: ThemeTemplateName;
}

export interface ResolvedThemeSettings {
  mode: ThemeMode;
  templateName: ThemeTemplateName;
}

export interface ThemeManagerContextValue extends ThemeSettings {
  availableTemplates: ThemeTemplateName[];
  resolvedMode: ThemeMode;
  resolvedTemplateName: ThemeTemplateName;
  setMode: (mode: ThemeMode) => void;
  setSettings: (settings: ThemeSettings) => void;
  setSource: (source: ThemeSource) => void;
  setTemplateName: (themeName: ThemeTemplateName) => void;
}
