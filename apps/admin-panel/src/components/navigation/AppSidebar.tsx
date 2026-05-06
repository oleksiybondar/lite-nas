import ExpandLessRoundedIcon from "@mui/icons-material/ExpandLessRounded";
import ExpandMoreRoundedIcon from "@mui/icons-material/ExpandMoreRounded";
import Box from "@mui/material/Box";
import Collapse from "@mui/material/Collapse";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import type { AppNavigationItem, AppNavigationPageItem } from "@routes/navigation";
import type { MouseEvent, ReactElement } from "react";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

/**
 * Props for the protected dashboard sidebar.
 */
type AppSidebarDisplay = {
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
 * Props for the protected dashboard sidebar.
 */
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

/**
 * Recursive sidebar tree item.
 */
const AppSidebarTreeItem = ({
  depth,
  item,
  onNavigate,
  selectedPath,
}: {
  depth: number;
  item: AppNavigationPageItem;
  onNavigate?: (() => void) | undefined;
  selectedPath: string | null;
}): ReactElement => {
  const hasChildren = item.children !== undefined && item.children.length > 0;
  const isSelected = selectedPath === item.path;
  const isActiveBranch = selectedPath?.startsWith(`${item.path}/`) ?? false;
  const [isUserExpanded, setIsUserExpanded] = useState(false);
  const isExpanded = hasChildren && (isActiveBranch || isUserExpanded);
  const toggleExpanded = (event: MouseEvent<HTMLSpanElement>): void => {
    event.preventDefault();
    event.stopPropagation();
    setIsUserExpanded((currentValue) => !currentValue);
  };

  return (
    <>
      <AppSidebarTreeButton
        depth={depth}
        hasChildren={hasChildren}
        isExpanded={isExpanded}
        isSelected={isSelected}
        item={item}
        onNavigate={onNavigate}
        onToggleExpanded={toggleExpanded}
      />
      <AppSidebarTreeChildren
        depth={depth}
        isExpanded={isExpanded}
        item={item}
        onNavigate={onNavigate}
        selectedPath={selectedPath}
      />
    </>
  );
};

/**
 * Clickable row for a sidebar tree item.
 */
const AppSidebarTreeButton = ({
  depth,
  hasChildren,
  isExpanded,
  isSelected,
  item,
  onNavigate,
  onToggleExpanded,
}: {
  depth: number;
  hasChildren: boolean;
  isExpanded: boolean;
  isSelected: boolean;
  item: AppNavigationPageItem;
  onNavigate?: (() => void) | undefined;
  onToggleExpanded: (event: MouseEvent<HTMLSpanElement>) => void;
}): ReactElement => {
  return (
    <ListItemButton
      component={RouterLink}
      onClick={onNavigate}
      selected={isSelected}
      sx={{ minHeight: 44, pl: 2 + depth * 2.5, pr: 1.5 }}
      to={item.path}
    >
      <AppSidebarTreeIcon item={item} />
      <ListItemText primary={item.title} />
      <AppSidebarTreeExpandControl
        hasChildren={hasChildren}
        isExpanded={isExpanded}
        item={item}
        onToggleExpanded={onToggleExpanded}
      />
    </ListItemButton>
  );
};

/**
 * Optional icon cell for a sidebar tree row.
 */
const AppSidebarTreeIcon = ({ item }: { item: AppNavigationPageItem }): ReactElement | null => {
  if (item.icon === undefined) {
    return null;
  }

  return <ListItemIcon sx={{ minWidth: 36 }}>{item.icon}</ListItemIcon>;
};

/**
 * Expand/collapse affordance for sidebar rows that own child routes.
 */
const AppSidebarTreeExpandControl = ({
  hasChildren,
  isExpanded,
  item,
  onToggleExpanded,
}: {
  hasChildren: boolean;
  isExpanded: boolean;
  item: AppNavigationPageItem;
  onToggleExpanded: (event: MouseEvent<HTMLSpanElement>) => void;
}): ReactElement | null => {
  if (!hasChildren) {
    return null;
  }

  return (
    <Box
      aria-label={`${isExpanded ? "Collapse" : "Expand"} ${item.title}`}
      component="span"
      onClick={onToggleExpanded}
    >
      {isExpanded ? <ExpandLessRoundedIcon /> : <ExpandMoreRoundedIcon />}
    </Box>
  );
};

/**
 * Collapsible child list for a sidebar tree item.
 */
const AppSidebarTreeChildren = ({
  depth,
  isExpanded,
  item,
  onNavigate,
  selectedPath,
}: {
  depth: number;
  isExpanded: boolean;
  item: AppNavigationPageItem;
  onNavigate?: (() => void) | undefined;
  selectedPath: string | null;
}): ReactElement | null => {
  if (item.children === undefined || item.children.length === 0) {
    return null;
  }

  return (
    <Collapse in={isExpanded} timeout="auto" unmountOnExit>
      <List disablePadding>
        {item.children.map((child) => {
          return (
            <AppSidebarTreeItem
              depth={depth + 1}
              item={child}
              key={child.path}
              onNavigate={onNavigate}
              selectedPath={selectedPath}
            />
          );
        })}
      </List>
    </Collapse>
  );
};
