import Box from "@mui/material/Box";
import List from "@mui/material/List";
import type { AppNavigationItem } from "@routes/navigation";
import type { ReactElement } from "react";
import { AppSidebarTreeItem } from "./AppSidebarTreeItem";

type AppSidebarDisplay = {
  lg?: string;
  md?: string;
  xs?: string;
};

type AppSidebarProps = {
  /**
   * Responsive display configuration for the sidebar container.
   */
  display?: AppSidebarDisplay;
  /**
   * Navigation entries rendered in sidebar order.
   */
  items: AppNavigationItem[];
  /**
   * Called after a route item is selected.
   */
  onNavigate?: (() => void) | undefined;
  /**
   * Currently selected page path.
   */
  selectedPath: string | null;
};

/**
 * Sidebar navigation for authenticated dashboard routes.
 */
export const AppSidebar = ({
  display = { lg: "none", md: "block", xs: "none" },
  items,
  onNavigate,
  selectedPath,
}: AppSidebarProps): ReactElement => {
  return (
    <Box
      borderColor="divider"
      borderRight={1}
      component="nav"
      flexShrink={0}
      width={280}
      sx={{ display }}
    >
      <List disablePadding sx={{ py: 2 }}>
        {items.map((item) => {
          return (
            <AppSidebarTreeItem
              depth={0}
              item={item}
              key={item.path}
              onNavigate={onNavigate}
              selectedPath={selectedPath}
            />
          );
        })}
      </List>
    </Box>
  );
};
