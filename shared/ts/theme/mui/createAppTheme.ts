import type { Theme } from "@mui/material/styles";
import { createTheme } from "@mui/material/styles";
import { themeRegistry } from "../registry";
import type { ThemeMode, ThemeTemplateName } from "../types";
import { getComponentOverrides } from "./components";

/** Creates the MUI theme used by each LiteNAS browser application. */
export const createAppTheme = (templateName: ThemeTemplateName, mode: ThemeMode): Theme => {
  const selectedTemplate = themeRegistry[templateName];
  const templateOptions = mode === "dark" ? selectedTemplate.dark : selectedTemplate.light;

  return createTheme({
    ...templateOptions,
    components: getComponentOverrides(),
  });
};
