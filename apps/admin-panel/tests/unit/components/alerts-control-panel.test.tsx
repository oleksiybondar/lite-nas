import { AlertsControlPanel } from "@components/alerts/AlertsControlPanel";
import { AlertsControlPanelPagination } from "@components/alerts/AlertsControlPanel/AlertsControlPanelPagination";
import { fireEvent, render, screen } from "@testing-library/react";
import {
  AlertsProvidersTestHarness,
  createAlertsContextValue,
} from "@tests/unit/test-utils/alerts";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";

describe("AlertsControlPanel", () => {
  test("renders separate filters and search panels", () => {
    const value = createAlertsContextValue("system", "unacknowledged");

    renderAlertsControlPanel(value);

    expect(screen.getByTestId("alerts-control-panel")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-filters-panel")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-search-panel")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-filters-control")).toBeInTheDocument();
    expect(screen.getByTestId("alerts-search-control")).toBeInTheDocument();
  });

  test("updates search through the dedicated control-panel context", () => {
    const value = createAlertsContextValue("system", "unacknowledged");

    renderAlertsControlPanel(value);
    fireEvent.change(screen.getByRole("textbox", { name: "Search current page" }), {
      target: { value: "disk" },
    });

    expect(value.setSearch).toHaveBeenCalledWith("disk");
  });
});

describe("AlertsControlPanel pagination", () => {
  test("updates pagination through the dedicated control-panel context", () => {
    const value = createAlertsContextValue("system", "unacknowledged");
    value.page = 1;
    value.totalCount = 45;
    value.totalPages = 3;

    renderAlertsPagination(value);
    fireEvent.click(screen.getByRole("button", { name: "Go to page 2" }));

    expect(value.setPage).toHaveBeenCalledWith(2);
  });
});

describe("AlertsControlPanel filters", () => {
  test("renders fixed and configured filter options and clears active filters", () => {
    const value = createAlertsControlPanelValueWithConfiguredFilters();

    renderAlertsControlPanel(value);

    expect(screen.getByRole("combobox", { name: "Category" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Source" })).toBeInTheDocument();
    expect(screen.getByTestId("alertsPriorityFilter-select")).toHaveTextContent("2");
    expect(screen.getByTestId("alertsSeverityFilter-select")).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("alerts-clear-filters-button"));

    expect(value.clearFilters).toHaveBeenCalled();
  });
});

/**
 * Renders the alerts control panel under the shared route and control-panel providers.
 */
const renderAlertsControlPanel = (value: ReturnType<typeof createAlertsContextValue>): void => {
  render(
    <TestMemoryRouter>
      <AlertsProvidersTestHarness value={value}>
        <AlertsControlPanel />
      </AlertsProvidersTestHarness>
    </TestMemoryRouter>,
  );
};

/**
 * Renders the dedicated alerts pagination panel under the shared route and control-panel providers.
 */
const renderAlertsPagination = (value: ReturnType<typeof createAlertsContextValue>): void => {
  render(
    <TestMemoryRouter>
      <AlertsProvidersTestHarness value={value}>
        <AlertsControlPanelPagination />
      </AlertsProvidersTestHarness>
    </TestMemoryRouter>,
  );
};

/**
 * Creates one alerts-context fixture with configured filterable values.
 */
const createAlertsControlPanelValueWithConfiguredFilters = (): ReturnType<
  typeof createAlertsContextValue
> => {
  const value = createAlertsContextValue("system", "all");

  value.priorityFilter = [2];
  value.severityFilter = ["warning"];
  value.sourceFilter = ["resource-monitor"];

  return value;
};
