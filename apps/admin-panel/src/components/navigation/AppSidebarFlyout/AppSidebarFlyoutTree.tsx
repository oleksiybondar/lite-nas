import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import type { AppNavigationPageItem } from "@routes/navigation";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

/**
 * Recursive flyout content for nested page items.
 */
export const AppSidebarFlyoutTree = ({
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
        data-test-class="sidebar-flyout-tree-item"
        data-test-name={item.title}
        data-test-path={item.path}
        selected={selectedPath === item.path}
        sx={{ minHeight: 42, pl: 2 + depth * 2 }}
        to={item.path}
      >
        {item.icon !== undefined ? (
          <ListItemIcon
            data-test-class="sidebar-flyout-tree-item-icon"
            data-test-name={item.title}
            sx={{ minWidth: 36 }}
          >
            {item.icon}
          </ListItemIcon>
        ) : null}
        <ListItemText
          data-test-class="sidebar-flyout-tree-item-label"
          data-test-name={item.title}
          primary={item.title}
        />
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
