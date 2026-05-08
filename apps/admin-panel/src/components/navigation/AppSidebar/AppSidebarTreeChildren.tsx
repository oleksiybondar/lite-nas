import Collapse from "@mui/material/Collapse";
import List from "@mui/material/List";
import type { AppNavigationPageItem } from "@routes/navigation";
import type { ReactElement } from "react";
import type { RenderTreeItemFn } from "./types";

/**
 * Collapsible child list for a sidebar tree item.
 */
export const AppSidebarTreeChildren = ({
  depth,
  isExpanded,
  item,
  renderItem,
}: {
  depth: number;
  isExpanded: boolean;
  item: AppNavigationPageItem;
  renderItem: RenderTreeItemFn;
}): ReactElement | null => {
  if (item.children === undefined || item.children.length === 0) {
    return null;
  }

  return (
    <Collapse
      data-test-class="sidebar-tree-children"
      data-test-name={item.title}
      data-test-path={item.path}
      in={isExpanded}
      timeout="auto"
      unmountOnExit
    >
      <List disablePadding>{item.children.map((child) => renderItem(child, depth + 1))}</List>
    </Collapse>
  );
};
