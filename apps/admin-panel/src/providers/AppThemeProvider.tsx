import { useThemeManager } from "@hooks/useThemeManager";

import CssBaseline from "@mui/material/CssBaseline";
import { ThemeProvider } from "@mui/material/styles";
import { createAppTheme } from "@theme/index";
import type { PropsWithChildren, ReactElement } from "react";

export const AppThemeProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { resolvedMode, resolvedTemplateName } = useThemeManager();
  const theme = createAppTheme(resolvedTemplateName, resolvedMode);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
};
