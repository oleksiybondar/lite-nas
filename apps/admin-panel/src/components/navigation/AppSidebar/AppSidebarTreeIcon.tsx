import ListItemIcon from "@mui/material/ListItemIcon";
import type { AppNavigationPageItem } from "@routes/navigation";
import type { ReactElement } from "react";

/**
 * Optional icon cell for a sidebar tree row.
 */
export const AppSidebarTreeIcon = ({
  item,
}: {
  item: AppNavigationPageItem;
}): ReactElement | null => {
  if (item.icon === undefined) {
    return null;
  }

  return <ListItemIcon sx={{ minWidth: 36 }}>{item.icon}</ListItemIcon>;
};
