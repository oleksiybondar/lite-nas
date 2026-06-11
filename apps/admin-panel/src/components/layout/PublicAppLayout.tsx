import { AppChromeLayout } from "@components/layout/AppChromeLayout";
import { AppFooter } from "@components/navigation/AppFooter";
import { AppTopBar } from "@components/navigation/AppTopBar";
import Box from "@mui/material/Box";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Layout for unauthenticated pages.
 *
 * It reuses global header and footer chrome without mounting the dashboard
 * sidebar or dashboard main-area constraints.
 */
export const PublicAppLayout = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <AppChromeLayout
      footer={<AppFooter />}
      header={<AppTopBar />}
      main={
        <Box data-testid="public-app-content" flex={1} minHeight={0} overflow="auto">
          {children}
        </Box>
      }
    />
  );
};
