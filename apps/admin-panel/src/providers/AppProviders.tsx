import { AppThemeProvider } from "@providers/AppThemeProvider";
import { QueryProvider } from "@providers/QueryProvider";
import { ThemeManagerProvider } from "@providers/ThemeManagerProvider";
import type { PropsWithChildren, ReactElement } from "react";

export const AppProviders = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <ThemeManagerProvider>
      <QueryProvider>
        <AppThemeProvider>{children}</AppThemeProvider>
      </QueryProvider>
    </ThemeManagerProvider>
  );
};
