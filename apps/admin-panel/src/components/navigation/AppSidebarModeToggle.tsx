import MenuOpenRoundedIcon from "@mui/icons-material/MenuOpenRounded";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import IconButton from "@mui/material/IconButton";
import Tooltip from "@mui/material/Tooltip";
import type { ReactElement } from "react";

/**
 * Props accepted by the desktop sidebar mode toggle.
 */
type AppSidebarModeToggleProps = {
  /**
   * Whether the desktop sidebar is currently collapsed into icon rail mode.
   */
  isCollapsed: boolean;
  /**
   * Called when the user switches between full and collapsed sidebar modes.
   */
  onToggle: () => void;
};

/**
 * Header control that switches desktop navigation between tree and flyout modes.
 */
export const AppSidebarModeToggle = ({
  isCollapsed,
  onToggle,
}: AppSidebarModeToggleProps): ReactElement => {
  const label = isCollapsed ? "Expand sidebar" : "Collapse sidebar";

  return (
    <Tooltip title={label}>
      <IconButton
        aria-label={label}
        color="inherit"
        onClick={onToggle}
        sx={{ display: { md: "inline-flex", xs: "none" } }}
      >
        {isCollapsed ? <MenuOpenRoundedIcon /> : <MenuRoundedIcon />}
      </IconButton>
    </Tooltip>
  );
};
