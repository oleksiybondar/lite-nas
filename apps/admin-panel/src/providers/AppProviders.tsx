import { ApiProvider } from "@providers/ApiProvider";
import { AppThemeProvider } from "@providers/AppThemeProvider";
import { AuthProvider } from "@providers/AuthProvider";
import { ProfileSettingsLoader } from "@providers/ProfileSettingsLoader";
import { QueryProvider } from "@providers/QueryProvider";
import { RbacProvider } from "@providers/RbacProvider";
import { ThemeManagerProvider } from "@providers/ThemeManagerProvider";
import type { PropsWithChildren, ReactElement } from "react";

export const AppProviders = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <ThemeManagerProvider>
      <AppThemeProvider>
        <ApiProvider>
          <QueryProvider>
            <AuthProvider>
              <RbacProvider>
                <ProfileSettingsLoader>{children}</ProfileSettingsLoader>
              </RbacProvider>
            </AuthProvider>
          </QueryProvider>
        </ApiProvider>
      </AppThemeProvider>
    </ThemeManagerProvider>
  );
};
