import type { Theme } from "@mui/material/styles";
import { createTheme } from "@mui/material/styles";

import { getComponentOverrides } from "@theme/mui/components";
import { themeRegistry } from "@theme/registry";
import type { ThemeMode, ThemeTemplateName } from "@theme/types";

export const createAppTheme = (templateName: ThemeTemplateName, mode: ThemeMode): Theme => {
  const selectedTemplate = themeRegistry[templateName];
  const templateOptions = mode === "dark" ? selectedTemplate.dark : selectedTemplate.light;

  return createTheme({
    ...templateOptions,
    components: getComponentOverrides(),
  });
};
