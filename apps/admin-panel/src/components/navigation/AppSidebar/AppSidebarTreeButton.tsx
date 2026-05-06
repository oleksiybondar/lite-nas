import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";
import type { AppNavigationPageItem } from "@routes/navigation";
import type { MouseEvent, ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";
import { AppSidebarTreeExpandControl } from "./AppSidebarTreeExpandControl";
import { AppSidebarTreeIcon } from "./AppSidebarTreeIcon";

/**
 * Clickable row for a sidebar tree item.
 */
export const AppSidebarTreeButton = ({
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
