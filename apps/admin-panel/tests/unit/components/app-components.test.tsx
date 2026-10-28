import {
  AppLogo,
  AppPageLayout,
  AppTopBar,
  ThemeModeToggle,
  ThemeSourceSelector,
  ThemeTemplateSelector,
} from "@components/index";
import { fireEvent, screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";
import { selectMuiOption } from "@tests/unit/test-utils/mui";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";
import { renderWithThemeManager } from "@tests/unit/test-utils/theme-manager";

describe("app shell components", () => {
  test("renders the application logo", () => {
    renderWithLanguage(
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
    renderWithLanguage(
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

  test("keeps the language selector available in the app header", () => {
    renderWithLanguage(
      <TestMemoryRouter>
        <AppTopBar />
      </TestMemoryRouter>,
    );
    expect(screen.getByTestId("language-selector")).toBeInTheDocument();
  });
});

describe("shared theme controls", () => {
  test("changes the selected mode", () => {
    const onChange = vi.fn();
    renderWithLanguage(<ThemeModeToggle mode="light" onChange={onChange} source="user" />);
    fireEvent.click(screen.getByRole("switch", { name: "Light" }));
    expect(onChange).toHaveBeenCalledWith("dark");
  });

  test("disables mode selection for the default source", () => {
    renderWithLanguage(<ThemeModeToggle mode="dark" onChange={vi.fn()} source="default" />);
    expect(screen.getByRole("switch", { name: "Dark" })).toBeDisabled();
  });

  test("changes the selected source", () => {
    const onChange = vi.fn();
    renderWithLanguage(<ThemeSourceSelector onChange={onChange} value="default" />);
    selectMuiOption("Theme source", "User");
    expect(onChange).toHaveBeenCalledWith("user");
  });

  test("changes the selected template", () => {
    const onChange = vi.fn();
    renderWithLanguage(
      <ThemeTemplateSelector
        availableTemplates={["default", "backup" as never]}
        onChange={onChange}
        source="user"
        value="default"
      />,
    );
    selectMuiOption("Theme template", "Backup");
    expect(onChange).toHaveBeenCalledWith("backup");
  });

  test("disables template selection outside user source", () => {
    renderWithLanguage(
      <ThemeTemplateSelector
        availableTemplates={["default"]}
        onChange={vi.fn()}
        source="os"
        value="default"
      />,
    );
    expect(screen.getByRole("combobox", { name: "Theme template" })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
  });
});
