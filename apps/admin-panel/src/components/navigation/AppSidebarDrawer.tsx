import { AppSidebar } from "@components/navigation/AppSidebar";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import Drawer from "@mui/material/Drawer";
import IconButton from "@mui/material/IconButton";
import Tooltip from "@mui/material/Tooltip";
import type { AppNavigationItem } from "@routes/navigation";
import type { ReactElement } from "react";

/**
 * Props accepted by the mobile dashboard sidebar drawer.
 */
type AppSidebarDrawerProps = {
  /**
   * Navigation entries rendered in drawer order.
   */
  items: AppNavigationItem[];
  /**
   * Called when the drawer should close.
   */
  onClose: () => void;
  /**
   * Whether the drawer is currently open.
   */
  open: boolean;
  /**
   * Currently selected page path.
   */
  selectedPath: string | null;
};

/**
 * Tap-driven mobile navigation drawer for the protected dashboard layout.
 */
export const AppSidebarDrawer = ({
  items,
  onClose,
  open,
  selectedPath,
}: AppSidebarDrawerProps): ReactElement => {
  return (
    <Drawer
      onClose={onClose}
      open={open}
      sx={{ display: { md: "none", xs: "block" } }}
      variant="temporary"
    >
      <AppSidebar
        display={{ xs: "block" }}
        items={items}
        onNavigate={onClose}
        selectedPath={selectedPath}
      />
    </Drawer>
  );
};

/**
 * Props accepted by the mobile sidebar drawer button.
 */
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
        onClick={onOpen}
        sx={{ display: { md: "none", xs: "inline-flex" } }}
      >
        <MenuRoundedIcon />
      </IconButton>
    </Tooltip>
  );
};
