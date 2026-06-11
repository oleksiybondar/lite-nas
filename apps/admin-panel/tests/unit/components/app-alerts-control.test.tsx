import { AppAlertsControl } from "@components/navigation/AppAlertsControl";
import { useAlertsCount } from "@domain/alerts/hooks/useAlertsCount";
import { useAuth } from "@hooks/useAuth";
import type { UseQueryResult } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { mockRbacAccess } from "@tests/unit/test-utils/rbac";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";
import type { ReactNode } from "react";

const navigateMock = vi.fn();

vi.mock("@hooks/useAuth", () => ({
  useAuth: vi.fn(),
}));

vi.mock("@hooks/useRbac", () => ({
  useRbac: vi.fn(),
}));

vi.mock("@domain/alerts/hooks/useAlertsCount", () => ({
  useAlertsCount: vi.fn(),
}));

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");

  return {
    ...actual,
    useNavigate: () => navigateMock,
  };
});

describe("AppAlertsControl visibility", () => {
  test("hides the control for anonymous sessions", () => {
    mockAuth({ isAuthenticated: false });
    mockRbacAccess({ requireOperator: false, requireSecurity: false });
    mockAlertsCount();

    renderWithRouter(<AppAlertsControl />);

    expect(screen.queryByTestId("alerts-control")).not.toBeInTheDocument();
  });

  test("hides the control when the user has no alerts permissions", () => {
    mockAuth({ isAuthenticated: true });
    mockRbacAccess({ requireOperator: false, requireSecurity: false });
    mockAlertsCount({
      system: 4,
    });

    renderWithRouter(<AppAlertsControl />);

    expect(screen.queryByTestId("alerts-control")).not.toBeInTheDocument();
  });
});

describe("AppAlertsControl rendering", () => {
  test.each([
    {
      counts: {
        system: 4,
      },
      expectedIndicators: {
        security: null,
        system: "4",
      },
      name: "renders only the system indicator when operator access is available",
      permissions: {
        requireOperator: true,
        requireSecurity: false,
      },
    },
    {
      counts: {
        security: 9,
        system: 2,
      },
      expectedIndicators: {
        security: "9",
        system: "2",
      },
      name: "renders both security and system indicators for security-capable users",
      permissions: {
        requireOperator: true,
        requireSecurity: true,
      },
    },
  ])("$name", ({ counts, expectedIndicators, permissions }) => {
    renderAlertsControlForAuthenticatedUser({
      counts,
      permissions,
    });

    expect(screen.getByTestId("alerts-control")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-control-icon-svg")).toHaveStyle({ fontSize: "38px" });

    expectIndicatorValue("system", expectedIndicators.system);
    expectIndicatorValue("security", expectedIndicators.security);
  });

  test("caps visible indicator values at 20+", () => {
    renderAlertsControlForAuthenticatedUser({
      counts: {
        security: 21,
        system: 34,
      },
      permissions: {
        requireOperator: true,
        requireSecurity: true,
      },
    });

    expect(findIndicator("security")).toHaveTextContent("20+");
    expect(findIndicator("system")).toHaveTextContent("20+");
    expect(findIndicator("security")).toHaveAttribute("data-test-position", "top");
    expect(findIndicator("system")).toHaveAttribute("data-test-position", "bottom");
  });
});

describe("AppAlertsControl menu", () => {
  test.each([
    {
      counts: {
        system: 4,
      },
      hiddenLabels: ["0 security alerts"],
      name: "opens a dropdown with the system alerts action for operator access",
      permissions: {
        requireOperator: true,
        requireSecurity: false,
      },
      visibleLabels: ["4 system alerts"],
    },
    {
      counts: {
        security: 9,
        system: 2,
      },
      hiddenLabels: [],
      name: "opens a dropdown with both domain actions when both guards pass",
      permissions: {
        requireOperator: true,
        requireSecurity: true,
      },
      visibleLabels: ["2 system alerts", "9 security alerts"],
    },
  ])("$name", ({ counts, hiddenLabels, permissions, visibleLabels }) => {
    openAlertsMenu({
      counts,
      permissions,
    });

    for (const label of visibleLabels) {
      expect(screen.getByRole("menuitem", { name: label })).toBeInTheDocument();
    }

    for (const label of hiddenLabels) {
      expect(screen.queryByRole("menuitem", { name: label })).not.toBeInTheDocument();
    }
  });
});

describe("AppAlertsControl navigation", () => {
  test.each([
    {
      counts: {
        system: 4,
      },
      menuItemLabel: "4 system alerts",
      name: "navigates to the system unacknowledged alerts page from the dropdown",
      path: "/alerts/system/unacknowledged",
      permissions: {
        requireOperator: true,
        requireSecurity: false,
      },
    },
    {
      counts: {
        security: 9,
        system: 2,
      },
      menuItemLabel: "9 security alerts",
      name: "navigates to the security unacknowledged alerts page from the dropdown",
      path: "/alerts/security/unacknowledged",
      permissions: {
        requireOperator: true,
        requireSecurity: true,
      },
    },
  ])("$name", ({ counts, menuItemLabel, path, permissions }) => {
    openAlertsMenu({
      counts,
      permissions,
    });
    fireEvent.click(screen.getByRole("menuitem", { name: menuItemLabel }));

    expect(navigateMock).toHaveBeenCalledWith(path);
  });
});

/**
 * Returns one rendered alerts indicator selected by domain name.
 */
const findIndicator = (domain: "security" | "system"): HTMLElement | null => {
  return document.querySelector(`[data-test-class="alerts-indicator"][data-test-name="${domain}"]`);
};

/**
 * Renders a component under a memory router for navigation-aware tests.
 */
const renderWithRouter = (component: ReactNode) => {
  navigateMock.mockReset();

  return render(<TestMemoryRouter>{component}</TestMemoryRouter>);
};

/**
 * Renders the alerts control for one authenticated session with explicit alert permissions.
 */
const renderAlertsControlForAuthenticatedUser = ({
  counts,
  permissions,
}: {
  counts: Partial<Record<"security" | "system", number>>;
  permissions: {
    requireOperator: boolean;
    requireSecurity: boolean;
  };
}): void => {
  mockAuth({ isAuthenticated: true });
  mockRbacAccess(permissions);
  mockAlertsCount(counts);

  renderWithRouter(<AppAlertsControl />);
};

/**
 * Opens the alerts dropdown for one authenticated session with explicit alert permissions.
 */
const openAlertsMenu = ({
  counts,
  permissions,
}: {
  counts: Partial<Record<"security" | "system", number>>;
  permissions: {
    requireOperator: boolean;
    requireSecurity: boolean;
  };
}): void => {
  renderAlertsControlForAuthenticatedUser({
    counts,
    permissions,
  });
  fireEvent.click(screen.getByRole("button", { name: "Alerts menu" }));
};

/**
 * Asserts one rendered alerts indicator value or its absence by domain.
 */
const expectIndicatorValue = (
  domain: "security" | "system",
  expectedValue: string | null,
): void => {
  const indicator = findIndicator(domain);

  if (expectedValue === null) {
    expect(indicator).toBeNull();
    return;
  }

  expect(indicator).toHaveTextContent(expectedValue);
};

/**
 * Mocks the public auth-hook contract used by the alerts control.
 */
const mockAuth = ({ isAuthenticated }: { isAuthenticated: boolean }): void => {
  vi.mocked(useAuth).mockReturnValue({
    isAuthInited: true,
    isAuthenticated,
    login: vi.fn(),
    logout: vi.fn(),
    me: null,
  });
};

/**
 * Mocks the alerts-count query hook for both supported alert domains.
 */
const mockAlertsCount = ({
  security = 0,
  system = 0,
}: Partial<Record<"security" | "system", number>> = {}): void => {
  vi.mocked(useAlertsCount).mockImplementation((domain) => {
    const count = domain === "security" ? security : system;

    return {
      data: { count },
    } as UseQueryResult<{ count: number }>;
  });
};
