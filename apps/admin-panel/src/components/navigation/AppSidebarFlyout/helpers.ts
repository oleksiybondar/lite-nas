import type { AppNavigationPageItem } from "@routes/navigation";

/**
 * Reports whether an item or one of its descendants owns the selected route.
 */
export const isActiveNavigationItem = (
  item: AppNavigationPageItem,
  selectedPath: string | null,
): boolean => {
  if (selectedPath === null) {
    return false;
  }

  return selectedPath === item.path || selectedPath.startsWith(`${item.path}/`);
};
