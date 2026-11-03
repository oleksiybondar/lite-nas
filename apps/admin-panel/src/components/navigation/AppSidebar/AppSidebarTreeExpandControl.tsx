import { useLanguage } from "@lite-nas/language";
import ExpandLessRoundedIcon from "@mui/icons-material/ExpandLessRounded";
import ExpandMoreRoundedIcon from "@mui/icons-material/ExpandMoreRounded";
import Box from "@mui/material/Box";
import type { AppNavigationPageItem } from "@routes/navigation";
import type { MouseEvent, ReactElement } from "react";

/**
 * Expand/collapse affordance for sidebar rows that own child routes.
 */
export const AppSidebarTreeExpandControl = ({
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
  const { getLabel } = useLanguage();
  if (!hasChildren) {
    return null;
  }

  return (
    <Box
      aria-label={`${getLabel(isExpanded ? "navigation.collapse" : "navigation.expand")} ${getLabel(`navigation.${item.title}`)}`}
      component="span"
      data-test-class="sidebar-tree-expand-control"
      data-test-name={item.title}
      data-test-path={item.path}
      onClick={onToggleExpanded}
    >
      {isExpanded ? <ExpandLessRoundedIcon /> : <ExpandMoreRoundedIcon />}
    </Box>
  );
};
