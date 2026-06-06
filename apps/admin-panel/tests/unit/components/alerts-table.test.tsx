import { AlertsTable } from "@components/alerts/AlertsTable";
import { fireEvent, render, screen } from "@testing-library/react";
import {
  AlertsProvidersTestHarness,
  createAlertListItem,
  createAlertsContextValue,
} from "@tests/unit/test-utils/alerts";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";

describe("AlertsTable", () => {
  test("renders the reordered starter table headers with compact severity and priority labels first", () => {
    renderAlertsTable(createAlertsContextValue("system", "all"));

    expect(screen.getAllByRole("columnheader").map((header) => header.textContent)).toEqual([
      "S",
      "P",
      "Event ID",
      "Source",
      "Category",
      "Status",
      "Created at",
      "Acknowledged at",
      "Acknowledged",
      "Value",
      "Message",
      "Acknowledge",
    ]);
    expect(screen.getByTestId("alerts-table-footer-row")).toBeInTheDocument();
  });

  test("renders the empty-state row when there are no items", () => {
    renderAlertsTable(createAlertsContextValue("system", "all"));

    expect(screen.getByTestId("alerts-table-empty-row")).toBeInTheDocument();
    expect(screen.getByText("No alerts found on this page.")).toBeInTheDocument();
  });
});

describe("AlertsTable rows", () => {
  test("renders iconized severity plus starter row values from alert items", () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [
      createAlertListItem({
        Acknowledged: true,
        AcknowledgedAt: "2026-06-06T09:00:00Z",
        Category: "Power",
        CreatedAt: "2026-06-06T08:00:00Z",
        EventID: "evt-42",
        EventRecID: 42,
        LastValueNum: 82,
        LastValueUnit: "C",
        Message: "Fan threshold exceeded",
        Priority: 5,
        Severity: "critical",
        Source: "resource-monitor",
        Status: "open",
      }),
    ];

    renderAlertsTable(value);

    expect(screen.getByTestId("alerts-severity-cell-evt-42")).toBeInTheDocument();
    expect(screen.getByText("evt-42")).toBeInTheDocument();
    expect(screen.getByText("resource-monitor")).toBeInTheDocument();
    expect(screen.getByText("Power")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
    expect(screen.getByText("open")).toBeInTheDocument();
    expect(screen.getByText("2026-06-06T08:00:00Z")).toBeInTheDocument();
    expect(screen.getByText("2026-06-06T09:00:00Z")).toBeInTheDocument();
    expect(screen.getByText("Yes")).toBeInTheDocument();
    expect(screen.getByText("82 C")).toBeInTheDocument();
    expect(screen.getByText("Fan threshold exceeded")).toBeInTheDocument();
  });
});

describe("AlertsTable acknowledge", () => {
  test("disables the acknowledge action when the alert is already acknowledged", () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [
      createAlertListItem({
        Acknowledged: true,
        EventID: "evt-acknowledged",
      }),
    ];

    renderAlertsTable(value);

    expect(screen.getByTestId("alerts-acknowledge-button-evt-acknowledged")).toBeDisabled();
  });

  test("disables the acknowledge action while another acknowledge request is pending", () => {
    const value = createAlertsContextValue("system", "all");
    value.isAcknowledging = true;
    value.items = [
      createAlertListItem({
        EventID: "evt-pending",
      }),
    ];

    renderAlertsTable(value);

    expect(screen.getByTestId("alerts-acknowledge-button-evt-pending")).toBeDisabled();
  });

  test("calls acknowledge for the current event id when the action button is clicked", () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [
      createAlertListItem({
        EventID: "evt-click",
      }),
    ];

    renderAlertsTable(value);
    fireEvent.click(screen.getByTestId("alerts-acknowledge-button-evt-click"));

    expect(value.acknowledge).toHaveBeenCalledWith("evt-click");
  });
});

describe("AlertsTable security columns", () => {
  test("renders the mitigate column only for security alerts", () => {
    const value = createAlertsContextValue("security", "all");
    value.items = [
      createAlertListItem({
        EventID: "evt-security",
        Meta: {
          mitigate: "Rotate credentials",
        },
      }),
    ];

    renderAlertsTable(value);

    expect(screen.getByRole("columnheader", { name: "Mitigate" })).toBeInTheDocument();
    expect(screen.getByText("Rotate credentials")).toBeInTheDocument();
  });

  test("renders the security header set with mitigate at the end", () => {
    renderAlertsTable(createAlertsContextValue("security", "all"));

    expect(screen.getAllByRole("columnheader").map((header) => header.textContent)).toEqual([
      "S",
      "P",
      "Event ID",
      "Source",
      "Category",
      "Status",
      "Created at",
      "Acknowledged at",
      "Acknowledged",
      "Value",
      "Message",
      "Acknowledge",
      "Mitigate",
    ]);
  });
});

/**
 * Renders the starter alerts table under the shared route and control-panel providers.
 */
const renderAlertsTable = (value: ReturnType<typeof createAlertsContextValue>): void => {
  render(
    <TestMemoryRouter>
      <AlertsProvidersTestHarness value={value}>
        <AlertsTable />
      </AlertsProvidersTestHarness>
    </TestMemoryRouter>,
  );
};
