import { App } from "@app/App";
import { ThemeManagerContext } from "@contexts/theme-manager-context";
import { DashboardPage } from "@pages/DashboardPage";
import { PreferencesApplicationSettingsPage } from "@pages/PreferencesApplicationSettingsPage";
import { PreferencesLandingPage } from "@pages/PreferencesLandingPage";
import { PreferencesProfilePage } from "@pages/PreferencesProfilePage";
import { SystemLandingPage } from "@pages/SystemLandingPage";
import { SystemPerformanceLandingPage } from "@pages/SystemPerformanceLandingPage";
import { SystemSensorsLandingPage } from "@pages/SystemSensorsLandingPage";
import { AppProviders } from "@providers/AppProviders";
import { AppThemeProvider } from "@providers/AppThemeProvider";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { selectMuiOption } from "@tests/unit/test-utils/mui";
import { responseWithJson, responseWithStatus } from "@tests/unit/test-utils/responses";
import { renderWithThemeManager } from "@tests/unit/test-utils/theme-manager";
import { createAppTheme, type ThemeSettings, themeRegistry } from "@theme/index";
import { loadThemeSettings } from "@theme/manager/storage";
import { getComponentOverrides } from "@theme/mui/components";
import type { PropsWithChildren, ReactElement } from "react";
import { useState } from "react";
import { MemoryRouter } from "react-router-dom";

describe("dashboard page", () => {
  test("renders the initial dashboard sections", () => {
    renderWithThemeManager(<DashboardPage />);

    expect(screen.getByRole("heading", { name: "LiteNAS operations" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "System Metrics" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Access" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Services" })).toBeInTheDocument();
  });
});

describe("category landing pages", () => {
  test.each([
    {
      cardName: "Performance",
      page: <SystemLandingPage />,
    },
    {
      cardName: "Sensors",
      page: <SystemLandingPage />,
    },
    {
      cardName: "System",
      page: <SystemPerformanceLandingPage />,
    },
    {
      cardName: "Network",
      page: <SystemPerformanceLandingPage />,
    },
    {
      cardName: "Disk",
      page: <SystemPerformanceLandingPage />,
    },
    {
      cardName: "ZFS",
      page: <SystemPerformanceLandingPage />,
    },
    {
      cardName: "Temperature",
      page: <SystemSensorsLandingPage />,
    },
    {
      cardName: "Voltage",
      page: <SystemSensorsLandingPage />,
    },
    {
      cardName: "Clock",
      page: <SystemSensorsLandingPage />,
    },
    {
      cardName: "Throttling",
      page: <SystemSensorsLandingPage />,
    },
    {
      cardName: "Fan",
      page: <SystemSensorsLandingPage />,
    },
    {
      cardName: "User profile",
      page: <PreferencesLandingPage />,
    },
    {
      cardName: "Application settings",
      page: <PreferencesLandingPage />,
    },
  ])("renders $cardName category card", ({ cardName, page }) => {
    render(<MemoryRouter>{page}</MemoryRouter>);

    expect(screen.getByRole("heading", { name: cardName })).toBeInTheDocument();
  });

  test("renders category cards as full-card links", () => {
    render(
      <MemoryRouter>
        <SystemLandingPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: /Performance/ })).toHaveAttribute(
      "href",
      "/system/performance",
    );
  });
});

describe("preferences pages", () => {
  test("renders the profile preferences page", () => {
    renderWithThemeManager(<PreferencesProfilePage />);

    expect(screen.getByRole("heading", { name: "User profile" })).toBeInTheDocument();
  });

  test("renders the application settings preferences page", () => {
    renderWithThemeManager(<PreferencesApplicationSettingsPage />);

    expect(screen.getByRole("heading", { name: "Application settings" })).toBeInTheDocument();
  });

  test("persists changed application settings on apply", () => {
    window.localStorage.clear();
    render(
      <StatefulThemeManagerProvider>
        <PreferencesApplicationSettingsPage />
      </StatefulThemeManagerProvider>,
    );

    selectMuiOption("Theme source", "OS");
    fireEvent.click(screen.getByRole("button", { name: "Apply" }));

    expect(loadThemeSettings().source).toBe("os");
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

  test("renders authenticated user menu from the app shell", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(responseWithJson(200, authenticatedMeBody)));

    await act(async () => {
      render(<App />);
    });

    fireEvent.click(await screen.findByRole("button", { name: "User menu" }));

    expect(screen.getByRole("menuitem", { name: "User profile" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Application settings" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Logout" })).toBeInTheDocument();
  });

  test("switches desktop sidebar mode from the app shell", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(responseWithJson(200, authenticatedMeBody)));

    await act(async () => {
      render(<App />);
    });

    fireEvent.click(await screen.findByRole("button", { name: "Collapse sidebar" }));

    expect(screen.getByRole("button", { name: "Expand sidebar" })).toBeInTheDocument();
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
      full_name: "Admin User",
    },
  },
};

/**
 * Stateful theme manager provider used by preferences persistence tests.
 */
const StatefulThemeManagerProvider = ({ children }: PropsWithChildren): ReactElement => {
  const [settings, setSettings] = useState<ThemeSettings>({
    mode: "dark",
    source: "user",
    templateName: "default",
  });

  return (
    <ThemeManagerContext.Provider
      value={{
        availableTemplates: ["default"],
        mode: settings.mode,
        resolvedMode: settings.mode,
        resolvedTemplateName: settings.templateName,
        setMode: (mode) => {
          setSettings((currentSettings) => ({ ...currentSettings, mode }));
        },
        setSettings,
        setSource: (source) => {
          setSettings((currentSettings) => ({ ...currentSettings, source }));
        },
        setTemplateName: (templateName) => {
          setSettings((currentSettings) => ({ ...currentSettings, templateName }));
        },
        source: settings.source,
        templateName: settings.templateName,
      }}
    >
      {children}
    </ThemeManagerContext.Provider>
  );
};
