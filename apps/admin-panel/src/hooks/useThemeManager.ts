import { ThemeManagerContext } from "@contexts/theme-manager-context";
import type { ThemeManagerContextValue } from "@theme/index";
import { useContext } from "react";

export const useThemeManager = (): ThemeManagerContextValue => {
  return useContext(ThemeManagerContext);
};
