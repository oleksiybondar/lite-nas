import Box from "@mui/material/Box";
import IconButton from "@mui/material/IconButton";
import List from "@mui/material/List";
import Popover from "@mui/material/Popover";
import Tooltip from "@mui/material/Tooltip";
import type { AppNavigationItem, AppNavigationPageItem } from "@routes/navigation";
import type { MouseEvent, ReactElement } from "react";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import { AppSidebarFlyoutTree } from "./AppSidebarFlyoutTree";
import { isActiveNavigationItem } from "./helpers";

type AppSidebarFlyoutDisplay = {
  lg?: string;
  md?: string;
  xs?: string;
};

type AppSidebarFlyoutProps = {
  /**
   * Responsive display configuration for the flyout container.
   */
  display?: AppSidebarFlyoutDisplay;
  /**
   * Navigation tree rendered as icon rail entries.
   */
  items: AppNavigationItem[];
  /**
   * Currently selected page path.
   */
  selectedPath: string | null;
};

/**
 * Collapsed dashboard sidebar with flyout access to nested navigation.
 */
export const AppSidebarFlyout = ({
  display = { lg: "block", xs: "none" },
  items,
  selectedPath,
}: AppSidebarFlyoutProps): ReactElement => {
  const [anchorElement, setAnchorElement] = useState<HTMLElement | null>(null);
  const [activeItem, setActiveItem] = useState<AppNavigationPageItem | null>(null);

  return (
    <Box
      borderColor="divider"
      borderRight={1}
      component="nav"
      flexShrink={0}
      width={72}
      sx={{ display }}
    >
      <List
        disablePadding
        sx={{ alignItems: "center", display: "flex", flexDirection: "column", py: 1 }}
      >
        {items.map((item) => {
          return (
            <Tooltip key={item.path} placement="right" title={item.title}>
              <IconButton
                aria-label={item.title}
                color={isActiveNavigationItem(item, selectedPath) ? "primary" : "default"}
                component={RouterLink}
                onClick={(event: MouseEvent<HTMLAnchorElement>) => {
                  if (item.children !== undefined && item.children.length > 0) {
                    event.preventDefault();
                    setAnchorElement(event.currentTarget);
                    setActiveItem(item);
                  }
                }}
                sx={{ my: 0.5 }}
                to={item.path}
              >
                {item.icon}
              </IconButton>
            </Tooltip>
          );
        })}
      </List>
      <Popover
        anchorEl={anchorElement}
        anchorOrigin={{ horizontal: "right", vertical: "top" }}
        onClose={() => {
          setAnchorElement(null);
          setActiveItem(null);
        }}
        open={anchorElement !== null}
        transformOrigin={{ horizontal: "left", vertical: "top" }}
      >
        {activeItem !== null ? (
          <List disablePadding sx={{ minWidth: 260, py: 1 }}>
            <AppSidebarFlyoutTree item={activeItem} selectedPath={selectedPath} />
          </List>
        ) : null}
      </Popover>
    </Box>
  );
};
