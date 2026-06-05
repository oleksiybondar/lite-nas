import { AppAlertsControl } from "@components/navigation/AppAlertsControl";
import { useAlertsCount } from "@domain/alerts/hooks/useAlertsCount";
import { useAuth } from "@hooks/useAuth";
import { useRbac } from "@hooks/useRbac";
import type { UseQueryResult } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";

vi.mock("@hooks/useAuth", () => ({
  useAuth: vi.fn(),
}));

vi.mock("@hooks/useRbac", () => ({
  useRbac: vi.fn(),
}));

vi.mock("@domain/alerts/hooks/useAlertsCount", () => ({
  useAlertsCount: vi.fn(),
}));

describe("AppAlertsControl visibility", () => {
  test("hides the control for anonymous sessions", () => {
    mockAuth({ isAuthenticated: false });
    mockRbac({ requireOperator: () => false, requireSecurity: () => false });
    mockAlertsCount();

    render(<AppAlertsControl />);

    expect(screen.queryByTestId("alerts-control")).not.toBeInTheDocument();
  });

  test("hides the control when the user has no alerts permissions", () => {
    mockAuth({ isAuthenticated: true });
    mockRbac({ requireOperator: () => false, requireSecurity: () => false });
    mockAlertsCount({
      system: 4,
    });

    render(<AppAlertsControl />);

    expect(screen.queryByTestId("alerts-control")).not.toBeInTheDocument();
  });
});

describe("AppAlertsControl rendering", () => {
  test("renders only the system indicator when operator access is available", () => {
    mockAuth({ isAuthenticated: true });
    mockRbac({ requireOperator: () => true, requireSecurity: () => false });
    mockAlertsCount({
      system: 4,
    });

    render(<AppAlertsControl />);

    expect(screen.getByTestId("alerts-control")).toBeInTheDocument();
    expect(findIndicator("system")).toHaveTextContent("4");
    expect(findIndicator("security")).toBeNull();
    expect(screen.getByTestId("alerts-control-icon-svg")).toHaveStyle({ fontSize: "38px" });
  });

  test("renders both security and system indicators for security-capable users", () => {
    mockAuth({ isAuthenticated: true });
    mockRbac({ requireOperator: () => true, requireSecurity: () => true });
    mockAlertsCount({
      security: 9,
      system: 2,
    });

    render(<AppAlertsControl />);

    expect(findIndicator("security")).toHaveTextContent("9");
    expect(findIndicator("system")).toHaveTextContent("2");
    expect(findIndicator("security")).toHaveAttribute("data-test-position", "top");
    expect(findIndicator("system")).toHaveAttribute("data-test-position", "bottom");
  });

  test("caps visible indicator values at 20+", () => {
    mockAuth({ isAuthenticated: true });
    mockRbac({ requireOperator: () => true, requireSecurity: () => true });
    mockAlertsCount({
      security: 21,
      system: 34,
    });

    render(<AppAlertsControl />);

    expect(findIndicator("security")).toHaveTextContent("20+");
    expect(findIndicator("system")).toHaveTextContent("20+");
  });
});

/**
 * Returns one rendered alerts indicator selected by domain name.
 */
const findIndicator = (domain: "security" | "system"): HTMLElement | null => {
  return document.querySelector(`[data-test-class="alerts-indicator"][data-test-name="${domain}"]`);
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
 * Mocks the public RBAC-hook contract used by the alerts control.
 */
const mockRbac = ({
  requireOperator,
  requireSecurity,
}: {
  requireOperator: () => boolean;
  requireSecurity: () => boolean;
}): void => {
  vi.mocked(useRbac).mockReturnValue({
    requireAdmin: () => false,
    requireOperator,
    requireSecurity,
    roles: [],
    scopes: [],
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
