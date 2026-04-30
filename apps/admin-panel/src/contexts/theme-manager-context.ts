import type { ThemeManagerContextValue } from "@theme/index";
import { createContext } from "react";

const missingThemeManagerProvider = (): never => {
  throw new Error("ThemeManagerContext is missing its provider.");
};

export const ThemeManagerContext = createContext<ThemeManagerContextValue>({
  availableTemplates: ["default"],
  mode: "dark",
  resolvedMode: "dark",
  resolvedTemplateName: "default",
  setMode: missingThemeManagerProvider,
  setSettings: missingThemeManagerProvider,
  setSource: missingThemeManagerProvider,
  setTemplateName: missingThemeManagerProvider,
  source: "default",
  templateName: "default",
});
