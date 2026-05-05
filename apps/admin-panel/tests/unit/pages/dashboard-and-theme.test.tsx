import { App } from "@app/App";
import { ThemeManagerContext } from "@contexts/theme-manager-context";
import { DashboardPage } from "@pages/DashboardPage";
import { AppProviders } from "@providers/AppProviders";
import { AppThemeProvider } from "@providers/AppThemeProvider";
import { act, render, screen } from "@testing-library/react";
import type { ThemeManagerContextValue } from "@theme/index";
import { createAppTheme, themeRegistry } from "@theme/index";
import { getComponentOverrides } from "@theme/mui/components";
import type { ReactElement } from "react";

describe("dashboard page", () => {
  test("renders the initial dashboard sections", () => {
    renderWithThemeManager(<DashboardPage />);

    expect(screen.getByRole("heading", { name: "LiteNAS operations" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "System Metrics" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Access" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Services" })).toBeInTheDocument();
  });
});

describe("theme creation", () => {
  test("creates light and dark app themes from the registry", () => {
    const lightTheme = createAppTheme("default", "light");
    const darkTheme = createAppTheme("default", "dark");

    expect(themeRegistry.default.light.palette?.mode).toBe("light");
    expect(lightTheme.palette.mode).toBe("light");
    expect(darkTheme.palette.mode).toBe("dark");
    expect(lightTheme.components?.MuiButton?.defaultProps?.disableElevation).toBe(true);
  });

  test("provides component overrides for Material UI", () => {
    const overrides = getComponentOverrides();

    expect(overrides?.MuiContainer?.defaultProps?.maxWidth).toBe("lg");
    expect(overrides?.MuiPaper?.defaultProps?.variant).toBe("outlined");
  });
});

describe("application providers", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  test("renders children inside the full provider stack", () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(responseWithStatus(401)));

    render(
      <AppProviders>
        <span>Provider child</span>
      </AppProviders>,
    );

    expect(screen.getByText("Provider child")).toBeInTheDocument();
  });

  test("renders children inside the MUI theme provider", () => {
    renderWithThemeManager(
      <AppThemeProvider>
        <span>Themed child</span>
      </AppThemeProvider>,
    );

    expect(screen.getByText("Themed child")).toBeInTheDocument();
  });

  test("renders the app shell", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(responseWithStatus(401)));

    await act(async () => {
      render(<App />);
    });

    expect(await screen.findByRole("heading", { name: "LiteNAS operations" })).toBeInTheDocument();
  });
});

/**
 * Renders a component with a stable theme manager context.
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
 * Creates a complete theme manager context value for page and theme tests.
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
 * Creates a minimal response with the supplied status.
 */
const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};
