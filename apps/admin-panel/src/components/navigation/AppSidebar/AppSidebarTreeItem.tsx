import type { AppNavigationPageItem } from "@routes/navigation";
import type { MouseEvent, ReactElement } from "react";
import { useState } from "react";
import { AppSidebarTreeButton } from "./AppSidebarTreeButton";
import { AppSidebarTreeChildren } from "./AppSidebarTreeChildren";

/**
 * Recursive sidebar tree item — renders the row and its collapsible children.
 */
export const AppSidebarTreeItem = ({
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
    setIsUserExpanded((current) => !current);
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
        renderItem={(child, childDepth) => (
          <AppSidebarTreeItem
            depth={childDepth}
            item={child}
            key={child.path}
            onNavigate={onNavigate}
            selectedPath={selectedPath}
          />
        )}
      />
    </>
  );
};
