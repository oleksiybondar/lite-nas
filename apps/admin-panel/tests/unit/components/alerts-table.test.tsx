import { AlertsTable } from "@components/alerts/AlertsTable";
import { fireEvent, render, screen } from "@testing-library/react";
import {
  AlertsProvidersTestHarness,
  createAlertListItem,
  createAlertsContextValue,
} from "@tests/unit/test-utils/alerts";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";

describe("AlertsTable", () => {
  test("renders the reordered table headers with compact severity and priority labels first", () => {
    renderAlertsTable(createAlertsContextValue("system", "all"));

    expect(screen.getAllByRole("columnheader").map((header) => header.textContent)).toEqual([
      "S",
      "P",
      "Event ID",
      "Message",
      "Value",
      "Source",
      "Status",
      "Acknowledged",
      "Category",
      "Created at",
      "Acknowledged at",
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
  test("renders iconized severity plus row values from alert items", () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [createSystemAlertRow()];

    renderAlertsTable(value);

    expect(screen.getByTestId("alerts-severity-cell-evt-42")).toBeInTheDocument();
    expect(screen.getByText("evt-42")).toBeInTheDocument();
    expect(screen.getByText("Fan threshold exceeded")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
    expect(screen.getByText("2026-06-06T08:00:00Z")).toBeInTheDocument();
    expect(screen.getByText("2026-06-06T09:00:00Z")).toBeInTheDocument();
    expectCellTone("82 C", "primary");
    expectCellTone("resource-monitor", "warning");
    expectCellTone("Power", "warning");
    expectCellTone("open", "warning");
    expectCellTone("testoperator", "primary");
  });
});

describe("AlertsTable acknowledge", () => {
  test("renders acknowledged by instead of an action button when the alert is already acknowledged", () => {
    const value = createAlertsContextValue("system", "all");
    value.items = [
      createAlertListItem({
        Acknowledged: true,
        AcknowledgedBy: "testoperator",
        EventID: "evt-acknowledged",
      }),
    ];

    renderAlertsTable(value);

    expect(
      screen.queryByTestId("alerts-acknowledge-button-evt-acknowledged"),
    ).not.toBeInTheDocument();
    expect(screen.getByText("testoperator")).toBeInTheDocument();
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
    expectCellTone("Rotate credentials", "primary");
  });

  test("renders the security header set with mitigate before category and dates", () => {
    renderAlertsTable(createAlertsContextValue("security", "all"));

    expect(screen.getAllByRole("columnheader").map((header) => header.textContent)).toEqual([
      "S",
      "P",
      "Event ID",
      "Message",
      "Value",
      "Source",
      "Status",
      "Mitigate",
      "Acknowledged",
      "Category",
      "Created at",
      "Acknowledged at",
    ]);
  });
});

/**
 * Renders the alerts table under the shared route and control-panel providers.
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

/**
 * Creates one representative system alert row used by row rendering assertions.
 */
const createSystemAlertRow = () => {
  return createAlertListItem({
    Acknowledged: true,
    AcknowledgedAt: "2026-06-06T09:00:00Z",
    AcknowledgedBy: "testoperator",
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
  });
};

/**
 * Asserts that one rendered text cell uses the expected semantic emphasis tone.
 */
const expectCellTone = (text: string, tone: string): void => {
  expect(screen.getByText(text)).toBeInTheDocument();
  expect(screen.getByText(text).closest("td")).toHaveAttribute("data-test-tone", tone);
};
