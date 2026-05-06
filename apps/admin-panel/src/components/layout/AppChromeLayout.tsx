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
    <Box display="flex" flexDirection="column" minHeight="100vh">
      {header}
      <Box component="main" flex={1}>
        {main}
      </Box>
      {footer}
    </Box>
  );
};
