import {
  AppLogo,
  AppPageLayout,
  AppTopBar,
  ThemeModeToggle,
  ThemeSourceSelector,
  ThemeTemplateSelector,
} from "@components/index";
import { fireEvent, render, screen } from "@testing-library/react";
import { selectMuiOption } from "@tests/unit/test-utils/mui";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";
import { renderWithThemeManager } from "@tests/unit/test-utils/theme-manager";

describe("app shell components", () => {
  test("renders the application logo", () => {
    render(
      <TestMemoryRouter>
        <AppLogo />
      </TestMemoryRouter>,
    );

    expect(screen.getByText("LiteNAS")).toBeInTheDocument();
  });

  test("renders page content inside the application layout", () => {
    renderWithThemeManager(
      <TestMemoryRouter>
        <AppPageLayout>
          <h1>Storage overview</h1>
        </AppPageLayout>
      </TestMemoryRouter>,
    );

    expect(screen.getByText("LiteNAS")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Storage overview" })).toBeInTheDocument();
  });

  test("renders supplied top bar action slots", () => {
    render(
      <TestMemoryRouter>
        <AppTopBar
          leadingAction={<button type="button">Open nav</button>}
          trailingAction={<span>User</span>}
        />
      </TestMemoryRouter>,
    );

    expect(screen.getByRole("button", { name: "Open nav" })).toBeInTheDocument();
    expect(screen.getByText("User")).toBeInTheDocument();
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
