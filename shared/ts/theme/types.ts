import type { ThemeOptions } from "@mui/material/styles";

/** Selects how an app resolves its active LiteNAS theme. */
export type ThemeSource = "default" | "os" | "user";
/** Supported MUI palette modes. */
export type ThemeMode = "light" | "dark";

/** MUI options for both modes of one visual template. */
export interface ThemeTemplate {
  dark: ThemeOptions;
  light: ThemeOptions;
}

/** Registered LiteNAS visual template names. */
export type ThemeTemplateName = "default";

/** Complete set of templates available to LiteNAS apps. */
export type ThemeRegistry = Record<ThemeTemplateName, ThemeTemplate>;

/** Persisted preference shared by apps on the same browser origin. */
export interface ThemeSettings {
  mode: ThemeMode;
  source: ThemeSource;
  templateName: ThemeTemplateName;
}

/** Mode and template after applying the selected preference source. */
export interface ResolvedThemeSettings {
  mode: ThemeMode;
  templateName: ThemeTemplateName;
}

/** Admin panel controls for reading and editing the shared theme preference. */
export interface ThemeManagerContextValue extends ThemeSettings {
  availableTemplates: ThemeTemplateName[];
  resolvedMode: ThemeMode;
  resolvedTemplateName: ThemeTemplateName;
  setMode: (mode: ThemeMode) => void;
  setSettings: (settings: ThemeSettings) => void;
  setSource: (source: ThemeSource) => void;
  setTemplateName: (themeName: ThemeTemplateName) => void;
}
