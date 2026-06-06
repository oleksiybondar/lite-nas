import { AlertsContext } from "@contexts/alerts-context";
import type { AlertsContextValue } from "@dto/alerts/alerts";
import { AlertsDashboardPage } from "@pages/AlertsDashboardPage";
import { AlertsLandingPage } from "@pages/AlertsLandingPage";
import { AlertsSecurityLandingPage } from "@pages/AlertsSecurityLandingPage";
import { AlertsSystemLandingPage } from "@pages/AlertsSystemLandingPage";
import { render, screen } from "@testing-library/react";
import { mockRbacAccess } from "@tests/unit/test-utils/rbac";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";

vi.mock("@hooks/useRbac", () => ({
  useRbac: vi.fn(),
}));

describe("AlertsLandingPage", () => {
  test("renders only the system card for operator-only access", () => {
    mockRbacAccess({ requireOperator: true, requireSecurity: false });

    render(
      <TestMemoryRouter>
        <AlertsLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "System" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Security" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /System/ })).toHaveAttribute("href", "/alerts/system");
  });

  test("renders only the security card for security-only access", () => {
    mockRbacAccess({ requireOperator: false, requireSecurity: true });

    render(
      <TestMemoryRouter>
        <AlertsLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.queryByRole("heading", { name: "System" })).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Security" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Security/ })).toHaveAttribute(
      "href",
      "/alerts/security",
    );
  });

  test("renders both domain cards when both guards pass", () => {
    mockRbacAccess({ requireOperator: true, requireSecurity: true });

    render(
      <TestMemoryRouter>
        <AlertsLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "System" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Security" })).toBeInTheDocument();
  });
});

describe("alerts domain landing pages", () => {
  test.each([
    {
      cardNames: ["Unacknowledged alerts", "Active alerts", "All alerts"],
      page: <AlertsSystemLandingPage />,
      title: "System",
    },
    {
      cardNames: ["Unacknowledged alerts", "Active alerts", "All alerts"],
      page: <AlertsSecurityLandingPage />,
      title: "Security",
    },
  ])("renders $title alert cards", ({ cardNames, page }) => {
    render(<TestMemoryRouter>{page}</TestMemoryRouter>);

    for (const cardName of cardNames) {
      expect(screen.getByRole("heading", { name: cardName })).toBeInTheDocument();
    }
  });

  test("renders system alert cards as full-card links", () => {
    render(
      <TestMemoryRouter>
        <AlertsSystemLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.getByRole("link", { name: /Unacknowledged alerts/ })).toHaveAttribute(
      "href",
      "/alerts/system/unacknowledged",
    );
  });

  test("renders security alert cards as full-card links", () => {
    render(
      <TestMemoryRouter>
        <AlertsSecurityLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.getByRole("link", { name: /All alerts/ })).toHaveAttribute(
      "href",
      "/alerts/security/all",
    );
  });
});

describe("AlertsDashboardPage", () => {
  test.each([
    {
      category: "unacknowledged" as const,
      domain: "system" as const,
      heading: "Unacknowledged alerts",
      label: "System",
    },
    {
      category: "all" as const,
      domain: "security" as const,
      heading: "All alerts",
      label: "Security",
    },
  ])("renders $domain route labels from params", ({ category, domain, heading, label }) => {
    renderAlertsDashboardPage(createAlertsContextValue(domain, category));

    expect(screen.getByRole("heading", { name: heading })).toBeInTheDocument();
    expect(screen.getByText(label)).toBeInTheDocument();
  });
});

/**
 * Renders the alerts dashboard page under a minimal alerts context.
 */
const renderAlertsDashboardPage = (value: AlertsContextValue): void => {
  render(
    <TestMemoryRouter>
      <AlertsContext.Provider value={value}>
        <AlertsDashboardPage />
      </AlertsContext.Provider>
    </TestMemoryRouter>,
  );
};

/**
 * Creates one minimal alerts context value for page rendering tests.
 */
const createAlertsContextValue = (
  domain: AlertsContextValue["domain"],
  category: AlertsContextValue["category"],
): AlertsContextValue => {
  return {
    acknowledge: vi.fn(),
    acknowledgeMany: vi.fn(),
    apiPath: `/api/alerts/${domain}/${category}`,
    category,
    categoryFilter: [],
    clearFilters: vi.fn(),
    domain,
    error: null,
    isAcknowledging: false,
    isError: false,
    isFetching: false,
    isLoading: false,
    items: [],
    nextPage: vi.fn(),
    page: 1,
    pageSize: 20,
    previousPage: vi.fn(),
    priorityFilter: [],
    queryKey: ["alerts", domain, category],
    refetch: vi.fn(),
    routePath: `/alerts/${domain}/${category}`,
    search: "",
    setCategoryFilter: vi.fn(),
    setPage: vi.fn(),
    setPageSize: vi.fn(),
    setPriorityFilter: vi.fn(),
    setSearch: vi.fn(),
    setSeverityFilter: vi.fn(),
    setSourceFilter: vi.fn(),
    severityFilter: [],
    sourceFilter: [],
    totalCount: 0,
    totalPages: 0,
  };
};
