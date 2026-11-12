import { AlertsContext } from "@contexts/alerts-context";
import type { AlertsContextValue } from "@dto/alerts/alerts";
import { AlertsDashboardPage } from "@pages/AlertsDashboardPage";
import { AlertsLandingPage } from "@pages/AlertsLandingPage";
import { AlertsSecurityLandingPage } from "@pages/AlertsSecurityLandingPage";
import { AlertsSystemLandingPage } from "@pages/AlertsSystemLandingPage";
import { AlertsControlPanelProvider } from "@providers/AlertsControlPanelProvider";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { createAlertListItem, createAlertsContextValue } from "@tests/unit/test-utils/alerts";
import { renderWithLanguage as render } from "@tests/unit/test-utils/language";
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
  test("acknowledges only unacknowledged alerts visible on the page", async () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [
      createAlertListItem({ EventID: "first" }),
      createAlertListItem({ EventID: "done", Acknowledged: true }),
      createAlertListItem({ EventID: "second" }),
    ];
    vi.mocked(value.acknowledgeMany).mockResolvedValue({ acknowledged: 2, failed: 0 });
    renderAlertsDashboardPage(value);

    fireEvent.click(screen.getByTestId("alerts-acknowledge-all-button"));

    await waitFor(() => expect(value.acknowledgeMany).toHaveBeenCalledWith(["first", "second"]));
  });

  test("disables bulk acknowledgement when the current page has no unacknowledged alerts", () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [createAlertListItem({ Acknowledged: true })];
    renderAlertsDashboardPage(value);

    expect(screen.getByTestId("alerts-acknowledge-all-button")).toBeDisabled();
  });

  test("reports partial bulk acknowledgement in a notification", async () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [
      createAlertListItem({ EventID: "first" }),
      createAlertListItem({ EventID: "second" }),
    ];
    vi.mocked(value.acknowledgeMany).mockResolvedValue({ acknowledged: 1, failed: 1 });
    renderAlertsDashboardPage(value);

    fireEvent.click(screen.getByTestId("alerts-acknowledge-all-button"));

    expect(await screen.findByText("Acknowledged 1 alerts; 1 failed.")).toBeInTheDocument();
  });
});

const alertRouteCases = [
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
];

describe("AlertsDashboardPage route content", () => {
  test.each(alertRouteCases)(
    "renders $domain heading from params",
    ({ category, domain, heading }) => {
      renderAlertsDashboardPage(createAlertsContextValue(domain, category));
      expect(screen.getByRole("heading", { name: heading })).toBeInTheDocument();
    },
  );

  test.each(alertRouteCases)("renders $domain label from params", ({ category, domain, label }) => {
    renderAlertsDashboardPage(createAlertsContextValue(domain, category));
    expect(screen.getByText(label)).toBeInTheDocument();
  });

  test.each(alertRouteCases)("renders $domain controls", ({ category, domain }) => {
    renderAlertsDashboardPage(createAlertsContextValue(domain, category));
    expect(screen.getByTestId("alerts-control-panel")).toBeInTheDocument();
  });

  test.each(alertRouteCases)("renders $domain pagination", ({ category, domain }) => {
    renderAlertsDashboardPage(createAlertsContextValue(domain, category));
    expect(screen.getAllByTestId("alerts-pagination-control")).toHaveLength(2);
  });

  test.each(alertRouteCases)("renders $domain table", ({ category, domain }) => {
    renderAlertsDashboardPage(createAlertsContextValue(domain, category));
    expect(screen.getByTestId("alerts-table")).toBeInTheDocument();
  });
});

/**
 * Renders the alerts dashboard page under a minimal alerts context.
 */
const renderAlertsDashboardPage = (value: AlertsContextValue): void => {
  render(
    <TestMemoryRouter>
      <AlertsContext.Provider value={value}>
        <AlertsControlPanelProvider>
          <AlertsDashboardPage />
        </AlertsControlPanelProvider>
      </AlertsContext.Provider>
    </TestMemoryRouter>,
  );
};
