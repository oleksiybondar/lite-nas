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
      data-testid="app-sidebar-flyout"
      flexShrink={0}
      width={72}
      sx={{ display }}
    >
      <List
        data-testid="app-sidebar-flyout-list"
        disablePadding
        sx={{ alignItems: "center", display: "flex", flexDirection: "column", py: 1 }}
      >
        {items.map((item) => {
          return renderFlyoutButton({
            item,
            key: item.path,
            onOpenFlyout: (event) => {
              setAnchorElement(event.currentTarget);
              setActiveItem(item);
            },
            selectedPath,
          });
        })}
      </List>
      {renderFlyoutPopover({
        activeItem,
        anchorElement,
        onClose: () => {
          setAnchorElement(null);
          setActiveItem(null);
        },
        selectedPath,
      })}
    </Box>
  );
};

/**
 * State and commands for rendering one collapsed sidebar rail button.
 */
type FlyoutButtonRenderOptions = {
  item: AppNavigationPageItem;
  key: string;
  onOpenFlyout: (event: MouseEvent<HTMLAnchorElement>) => void;
  selectedPath: string | null;
};

/**
 * Builds a collapsed sidebar rail item and opens the flyout for parent routes.
 */
const renderFlyoutButton = ({
  item,
  key,
  onOpenFlyout,
  selectedPath,
}: FlyoutButtonRenderOptions): ReactElement => {
  return (
    <Tooltip key={key} placement="right" title={item.title}>
      <IconButton
        aria-label={item.title}
        color={isActiveNavigationItem(item, selectedPath) ? "primary" : "default"}
        component={RouterLink}
        data-test-class="sidebar-flyout-button"
        data-test-name={item.title}
        data-test-path={item.path}
        onClick={(event: MouseEvent<HTMLAnchorElement>) => {
          if (item.children !== undefined && item.children.length > 0) {
            event.preventDefault();
            onOpenFlyout(event);
          }
        }}
        sx={{ my: 0.5 }}
        to={item.path}
      >
        {item.icon}
      </IconButton>
    </Tooltip>
  );
};

/**
 * State and commands for rendering the collapsed sidebar popover.
 */
type FlyoutPopoverRenderOptions = {
  activeItem: AppNavigationPageItem | null;
  anchorElement: HTMLElement | null;
  onClose: () => void;
  selectedPath: string | null;
};

/**
 * Builds the flyout popover that exposes nested navigation items.
 */
const renderFlyoutPopover = ({
  activeItem,
  anchorElement,
  onClose,
  selectedPath,
}: FlyoutPopoverRenderOptions): ReactElement => {
  return (
    <Popover
      anchorEl={anchorElement}
      anchorOrigin={{ horizontal: "right", vertical: "top" }}
      data-testid="sidebar-flyout-popover"
      onClose={onClose}
      open={anchorElement !== null}
      transformOrigin={{ horizontal: "left", vertical: "top" }}
    >
      {activeItem !== null ? (
        <List
          data-testid="sidebar-flyout-popover-list"
          disablePadding
          sx={{ minWidth: 260, py: 1 }}
        >
          <AppSidebarFlyoutTree item={activeItem} selectedPath={selectedPath} />
        </List>
      ) : null}
    </Popover>
  );
};
