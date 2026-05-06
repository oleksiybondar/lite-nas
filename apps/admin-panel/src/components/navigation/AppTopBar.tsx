import { AppLogo } from "@components/branding/AppLogo";
import { useThemeManager } from "@hooks/useThemeManager";
import DarkModeRoundedIcon from "@mui/icons-material/DarkModeRounded";
import LightModeRoundedIcon from "@mui/icons-material/LightModeRounded";
import AppBar from "@mui/material/AppBar";
import IconButton from "@mui/material/IconButton";
import Toolbar from "@mui/material/Toolbar";
import Tooltip from "@mui/material/Tooltip";
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
};

/**
 * Shared application top bar used by public and protected layouts.
 */
export const AppTopBar = ({ leadingAction }: AppTopBarProps = {}): ReactElement => {
  const { mode, setMode, setSource } = useThemeManager();
  const nextMode = mode === "dark" ? "light" : "dark";

  return (
    <AppBar color="transparent" elevation={0} position="sticky">
      <Toolbar sx={{ borderBottom: 1, borderColor: "divider", gap: 2 }}>
        {leadingAction}
        <AppLogo />
        <Tooltip title={`Switch to ${nextMode} mode`}>
          <IconButton
            aria-label={`Switch to ${nextMode} mode`}
            color="inherit"
            onClick={() => {
              setSource("user");
              setMode(nextMode);
            }}
            sx={{ ml: "auto" }}
          >
            {mode === "dark" ? <LightModeRoundedIcon /> : <DarkModeRoundedIcon />}
          </IconButton>
        </Tooltip>
      </Toolbar>
    </AppBar>
  );
};
