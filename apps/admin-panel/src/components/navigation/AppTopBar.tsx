import { AppLogo } from "@components/branding/AppLogo";
import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Toolbar from "@mui/material/Toolbar";
import type { ReactElement, ReactNode } from "react";

/**
 * Props accepted by the shared application top bar.
 */
type AppTopBarProps = {
  /**
   * Optional leading control rendered before the logo.
   *
   * Protected layouts use this for dashboard navigation entry points while
   * public layouts keep the shared header free of dashboard controls.
   */
  leadingAction?: ReactNode;
  /**
   * Optional trailing control rendered on the right side of the header.
   *
   * Protected layouts use this for authenticated user actions. Public layouts
   * leave it empty so anonymous screens do not expose session controls.
   */
  trailingAction?: ReactNode;
};

/**
 * Shared application top bar used by public and protected layouts.
 */
export const AppTopBar = ({ leadingAction, trailingAction }: AppTopBarProps = {}): ReactElement => {
  return (
    <AppBar color="transparent" elevation={0} position="sticky">
      <Toolbar sx={{ borderBottom: 1, borderColor: "divider", gap: 2 }}>
        {leadingAction}
        <AppLogo />
        <Box alignItems="center" display="flex" gap={1} sx={{ ml: "auto" }}>
          {trailingAction}
        </Box>
      </Toolbar>
    </AppBar>
  );
};
