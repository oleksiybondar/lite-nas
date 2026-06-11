import Box from "@mui/material/Box";
import type { ReactElement, ReactNode } from "react";

/**
 * Slots accepted by the shared application chrome layout.
 */
type AppChromeLayoutProps = {
  /**
   * Footer content shared by public and protected layouts.
   */
  footer: ReactNode;
  /**
   * Header content shared by public and protected layouts.
   */
  header: ReactNode;
  /**
   * Main route content.
   */
  main: ReactNode;
};

/**
 * Decomposed app shell that owns only global page chrome.
 *
 * Public pages can use this without a dashboard sidebar. Protected pages can
 * place a sidebar-aware dashboard frame into the `main` slot.
 */
export const AppChromeLayout = ({ footer, header, main }: AppChromeLayoutProps): ReactElement => {
  return (
    <Box
      data-testid="app-chrome-layout"
      display="flex"
      flexDirection="column"
      height="100dvh"
      overflow="hidden"
    >
      {header}
      <Box
        component="main"
        data-testid="app-main"
        display="flex"
        flex={1}
        minHeight={0}
        overflow="hidden"
      >
        {main}
      </Box>
      {footer}
    </Box>
  );
};
