import { AppLogo } from "@components/branding/AppLogo";
import { useThemeManager } from "@hooks/useThemeManager";
import DarkModeRoundedIcon from "@mui/icons-material/DarkModeRounded";
import LightModeRoundedIcon from "@mui/icons-material/LightModeRounded";
import AppBar from "@mui/material/AppBar";
import IconButton from "@mui/material/IconButton";
import Toolbar from "@mui/material/Toolbar";
import Tooltip from "@mui/material/Tooltip";
import type { ReactElement } from "react";

export const AppTopBar = (): ReactElement => {
  const { mode, setMode, setSource } = useThemeManager();
  const nextMode = mode === "dark" ? "light" : "dark";

  return (
    <AppBar color="transparent" elevation={0} position="sticky">
      <Toolbar sx={{ borderBottom: 1, borderColor: "divider", gap: 2 }}>
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
