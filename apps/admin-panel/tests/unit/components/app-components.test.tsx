import {
  AppLogo,
  AppPageLayout,
  AppTopBar,
  ThemeModeToggle,
  ThemeSourceSelector,
  ThemeTemplateSelector,
} from "@components/index";
import { ThemeManagerContext } from "@contexts/theme-manager-context";
import { fireEvent, render, screen, within } from "@testing-library/react";
import type { ThemeManagerContextValue } from "@theme/index";
import type { ReactElement } from "react";

describe("app shell components", () => {
  test("renders the application logo", () => {
    render(<AppLogo />);

    expect(screen.getByText("LiteNAS")).toBeInTheDocument();
  });

  test("renders page content inside the application layout", () => {
    renderWithThemeManager(
      <AppPageLayout>
        <h1>Storage overview</h1>
      </AppPageLayout>,
    );

    expect(screen.getByText("LiteNAS")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Storage overview" })).toBeInTheDocument();
  });

  test("switches the top bar theme mode through context handlers", () => {
    const setMode = vi.fn();
    const setSource = vi.fn();

    renderWithThemeManager(<AppTopBar />, {
      mode: "dark",
      setMode,
      setSource,
    });

    fireEvent.click(screen.getByRole("button", { name: "Switch to light mode" }));

    expect(setSource).toHaveBeenCalledWith("user");
    expect(setMode).toHaveBeenCalledWith("light");
  });
});

describe("theme mode toggle", () => {
  test("uses supplied props before context values", () => {
    const onChange = vi.fn();

    renderWithThemeManager(<ThemeModeToggle mode="light" onChange={onChange} source="user" />, {
      mode: "dark",
      resolvedMode: "dark",
    });

    const modeSwitch = screen.getByRole("switch", { name: "Light" });
    fireEvent.click(modeSwitch);

    expect(onChange).toHaveBeenCalledWith("dark");
  });

  test("uses context state when props are not supplied", () => {
    const setMode = vi.fn();

    renderWithThemeManager(<ThemeModeToggle />, {
      mode: "dark",
      setMode,
      source: "user",
    });

    fireEvent.click(screen.getByRole("switch", { name: "Dark" }));

    expect(setMode).toHaveBeenCalledWith("light");
  });

  test("disables user selection when the selected source is not user", () => {
    renderWithThemeManager(<ThemeModeToggle source="default" />);

    expect(screen.getByRole("switch", { name: "Dark" })).toBeDisabled();
  });
});

describe("theme source selector", () => {
  test("calls the supplied change handler when controlled", () => {
    const onChange = vi.fn();

    renderWithThemeManager(<ThemeSourceSelector onChange={onChange} value="default" />);
    selectMuiOption("Theme source", "User");

    expect(onChange).toHaveBeenCalledWith("user");
  });

  test("updates context when uncontrolled", () => {
    const setSource = vi.fn();

    renderWithThemeManager(<ThemeSourceSelector />, {
      setSource,
      source: "os",
    });
    selectMuiOption("Theme source", "Default");

    expect(setSource).toHaveBeenCalledWith("default");
  });
});

describe("theme template selector", () => {
  test("calls the supplied change handler when controlled", () => {
    const onChange = vi.fn();

    renderWithThemeManager(<ThemeTemplateSelector onChange={onChange} source="user" />, {
      availableTemplates: ["default", "backup" as never],
    });
    selectMuiOption("Theme template", "Backup");

    expect(onChange).toHaveBeenCalledWith("backup");
  });

  test("updates context when uncontrolled", () => {
    const setTemplateName = vi.fn();

    renderWithThemeManager(<ThemeTemplateSelector source="user" />, {
      availableTemplates: ["default", "backup" as never],
      setTemplateName,
    });
    selectMuiOption("Theme template", "Backup");

    expect(setTemplateName).toHaveBeenCalledWith("backup");
  });

  test("honors an explicit disabled override", () => {
    renderWithThemeManager(<ThemeTemplateSelector disabled={false} source="default" />);

    expect(screen.getByRole("combobox", { name: "Theme template" })).not.toHaveAttribute(
      "aria-disabled",
      "true",
    );
  });
});

/**
 * Renders a component under a predictable theme manager context.
 */
const renderWithThemeManager = (
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
 * Creates a complete theme manager value for component tests.
 */
const createThemeManagerValue = (
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

/**
 * Selects an option from a Material UI select rendered as a combobox.
 */
const selectMuiOption = (selectName: string, optionName: string): void => {
  fireEvent.mouseDown(screen.getByRole("combobox", { name: selectName }));
  fireEvent.click(within(screen.getByRole("listbox")).getByRole("option", { name: optionName }));
};
