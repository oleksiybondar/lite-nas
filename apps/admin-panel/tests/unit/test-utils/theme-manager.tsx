import { ThemeManagerContext } from "@contexts/theme-manager-context";
import { render } from "@testing-library/react";
import type { ThemeManagerContextValue } from "@theme/index";
import type { ReactElement } from "react";

/**
 * Renders a component under a controlled theme manager context.
 */
export const renderWithThemeManager = (
  component: ReactElement,
  overrides: Partial<ThemeManagerContextValue> = {},
) => {
  return render(
    <ThemeManagerContext.Provider value={createThemeManagerValue(overrides)}>
      {component}
    </ThemeManagerContext.Provider>,
  );
};

/**
 * Creates a complete theme manager context value for tests.
 */
export const createThemeManagerValue = (
  overrides: Partial<ThemeManagerContextValue> = {},
): ThemeManagerContextValue => {
  return {
    availableTemplates: ["default"],
    mode: "dark",
    resolvedMode: "dark",
    resolvedTemplateName: "default",
    setMode: vi.fn(),
    setSettings: vi.fn(),
    setSource: vi.fn(),
    setTemplateName: vi.fn(),
    source: "user",
    templateName: "default",
    ...overrides,
  };
};
