import { AppSidebar } from "@components/navigation/AppSidebar";
import Drawer from "@mui/material/Drawer";
import type { AppNavigationItem } from "@routes/navigation";
import type { ReactElement } from "react";

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
      data-testid="sidebar-drawer"
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
