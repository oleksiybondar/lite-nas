import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import IconButton from "@mui/material/IconButton";
import Tooltip from "@mui/material/Tooltip";
import type { ReactElement } from "react";

type AppSidebarDrawerButtonProps = {
  /**
   * Called when the button is activated.
   */
  onOpen: () => void;
};

/**
 * Header control that opens mobile dashboard navigation.
 */
export const AppSidebarDrawerButton = ({ onOpen }: AppSidebarDrawerButtonProps): ReactElement => {
  return (
    <Tooltip title="Open navigation">
      <IconButton
        aria-label="Open navigation"
        color="inherit"
        data-testid="sidebar-drawer-open-button"
        onClick={onOpen}
        sx={{ display: { md: "none", xs: "inline-flex" } }}
      >
        <MenuRoundedIcon />
      </IconButton>
    </Tooltip>
  );
};
