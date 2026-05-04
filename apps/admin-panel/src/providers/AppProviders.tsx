import { ApiProvider } from "@providers/ApiProvider";
import { AppThemeProvider } from "@providers/AppThemeProvider";
import { AuthProvider } from "@providers/AuthProvider";
import { ThemeManagerProvider } from "@providers/ThemeManagerProvider";
import type { PropsWithChildren, ReactElement } from "react";

export const AppProviders = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <ThemeManagerProvider>
      <AppThemeProvider>
        <ApiProvider>
          <AuthProvider>{children}</AuthProvider>
        </ApiProvider>
      </AppThemeProvider>
    </ThemeManagerProvider>
  );
};
