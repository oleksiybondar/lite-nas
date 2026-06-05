import { useRbac } from "@hooks/useRbac";
import { AlertsDashboardPage } from "@pages/AlertsDashboardPage";
import { AlertsLandingPage } from "@pages/AlertsLandingPage";
import { AlertsSecurityLandingPage } from "@pages/AlertsSecurityLandingPage";
import { AlertsSystemLandingPage } from "@pages/AlertsSystemLandingPage";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";

vi.mock("@hooks/useRbac", () => ({
  useRbac: vi.fn(),
}));

describe("AlertsLandingPage", () => {
  test("renders only the system card for operator-only access", () => {
    mockRbac({ requireOperator: () => true, requireSecurity: () => false });

    render(
      <MemoryRouter>
        <AlertsLandingPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "System" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Security" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /System/ })).toHaveAttribute("href", "/alerts/system");
  });

  test("renders only the security card for security-only access", () => {
    mockRbac({ requireOperator: () => false, requireSecurity: () => true });

    render(
      <MemoryRouter>
        <AlertsLandingPage />
      </MemoryRouter>,
    );

    expect(screen.queryByRole("heading", { name: "System" })).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Security" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Security/ })).toHaveAttribute(
      "href",
      "/alerts/security",
    );
  });

  test("renders both domain cards when both guards pass", () => {
    mockRbac({ requireOperator: () => true, requireSecurity: () => true });

    render(
      <MemoryRouter>
        <AlertsLandingPage />
      </MemoryRouter>,
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
    render(<MemoryRouter>{page}</MemoryRouter>);

    for (const cardName of cardNames) {
      expect(screen.getByRole("heading", { name: cardName })).toBeInTheDocument();
    }
  });

  test("renders system alert cards as full-card links", () => {
    render(
      <MemoryRouter>
        <AlertsSystemLandingPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: /Unacknowledged alerts/ })).toHaveAttribute(
      "href",
      "/alerts/system/unacknowledged",
    );
  });

  test("renders security alert cards as full-card links", () => {
    render(
      <MemoryRouter>
        <AlertsSecurityLandingPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: /All alerts/ })).toHaveAttribute(
      "href",
      "/alerts/security/all",
    );
  });
});

describe("AlertsDashboardPage", () => {
  test("renders system route labels from params", () => {
    renderAlertsDashboardRoute("/alerts/system/unacknowledged");

    expect(screen.getByRole("heading", { name: "Unacknowledged alerts" })).toBeInTheDocument();
    expect(screen.getByText("System")).toBeInTheDocument();
  });

  test("renders security route labels from params", () => {
    renderAlertsDashboardRoute("/alerts/security/all");

    expect(screen.getByRole("heading", { name: "All alerts" })).toBeInTheDocument();
    expect(screen.getByText("Security")).toBeInTheDocument();
  });
});

/**
 * Mocks the RBAC hook contract used by the alerts landing page.
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
 * Renders the alerts dashboard page under the route shape it expects.
 */
const renderAlertsDashboardRoute = (initialEntry: string): void => {
  render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route element={<AlertsDashboardPage />} path="/alerts/:group/:category" />
      </Routes>
    </MemoryRouter>,
  );
};
