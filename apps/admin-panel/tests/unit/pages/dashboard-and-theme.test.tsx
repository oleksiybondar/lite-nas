import { App } from "@app/App";
import { DashboardPage } from "@pages/DashboardPage";
import { AppProviders } from "@providers/AppProviders";
import { AppThemeProvider } from "@providers/AppThemeProvider";
import { act, render, screen } from "@testing-library/react";
import { responseWithJson, responseWithStatus } from "@tests/unit/test-utils/responses";
import { renderWithThemeManager } from "@tests/unit/test-utils/theme-manager";
import { createAppTheme, themeRegistry } from "@theme/index";
import { getComponentOverrides } from "@theme/mui/components";

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

  test("renders login from the app shell for anonymous sessions", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(responseWithStatus(401)));

    await act(async () => {
      render(<App />);
    });

    expect(await screen.findByRole("heading", { name: "Sign in" })).toBeInTheDocument();
  });

  test("renders dashboard from the app shell for authenticated sessions", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(responseWithJson(200, authenticatedMeBody)));

    await act(async () => {
      render(<App />);
    });

    expect(await screen.findByRole("heading", { name: "LiteNAS operations" })).toBeInTheDocument();
  });
});

/**
 * Current-user response for an authenticated app-shell render.
 */
const authenticatedMeBody = {
  data: {
    authenticated: true,
    auth_type: "password",
    roles: ["admin"],
    scopes: ["admin"],
    user: {
      id: "admin-id",
      login: "admin",
    },
  },
};
