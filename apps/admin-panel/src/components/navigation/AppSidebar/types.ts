import type { AppNavigationPageItem } from "@routes/navigation";
import type { ReactElement } from "react";

/**
 * Passed by AppSidebarTreeItem to AppSidebarTreeChildren so the two components
 * can reference each other without a circular module dependency.
 */
export type RenderTreeItemFn = (item: AppNavigationPageItem, depth: number) => ReactElement;
