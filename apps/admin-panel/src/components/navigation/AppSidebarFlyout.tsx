import Box from "@mui/material/Box";
import IconButton from "@mui/material/IconButton";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import Popover from "@mui/material/Popover";
import Tooltip from "@mui/material/Tooltip";
import type { AppNavigationItem, AppNavigationPageItem } from "@routes/navigation";
import type { MouseEvent, ReactElement } from "react";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

/**
 * Props for the collapsed sidebar flyout navigation.
 */
type AppSidebarFlyoutDisplay = {
  /**
   * Display value for large desktop viewports.
   */
  lg?: string;
  /**
   * Display value for tablet and compact desktop viewports.
   */
  md?: string;
  /**
   * Display value for mobile viewports.
   */
  xs?: string;
};

/**
 * Props for the collapsed sidebar flyout navigation.
 */
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

/**
 * Recursive flyout content for nested page items.
 */
const AppSidebarFlyoutTree = ({
  depth = 0,
  item,
  selectedPath,
}: {
  depth?: number;
  item: AppNavigationPageItem;
  selectedPath: string | null;
}): ReactElement => {
  return (
    <>
      <ListItemButton
        component={RouterLink}
        selected={selectedPath === item.path}
        sx={{ minHeight: 42, pl: 2 + depth * 2 }}
        to={item.path}
      >
        {item.icon !== undefined ? (
          <ListItemIcon sx={{ minWidth: 36 }}>{item.icon}</ListItemIcon>
        ) : null}
        <ListItemText primary={item.title} />
      </ListItemButton>
      {item.children?.map((child) => {
        return (
          <AppSidebarFlyoutTree
            depth={depth + 1}
            item={child}
            key={child.path}
            selectedPath={selectedPath}
          />
        );
      })}
    </>
  );
};

/**
 * Reports whether an item or one of its descendants owns the selected route.
 */
const isActiveNavigationItem = (
  item: AppNavigationPageItem,
  selectedPath: string | null,
): boolean => {
  if (selectedPath === null) {
    return false;
  }

  return selectedPath === item.path || selectedPath.startsWith(`${item.path}/`);
};
