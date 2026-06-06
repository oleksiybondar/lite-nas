import { AlertsContext } from "@contexts/alerts-context";
import type { AlertsContextValue } from "@dto/alerts/alerts";
import { AlertsControlPanelProvider } from "@providers/AlertsControlPanelProvider";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Creates one minimal shared alerts context value for page and component tests.
 */
export const createAlertsContextValue = (
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

type AlertsProvidersTestHarnessProps = PropsWithChildren<{
  value: AlertsContextValue;
}>;

/**
 * Nests the shared alerts and focused control-panel providers for unit tests.
 */
export const AlertsProvidersTestHarness = ({
  children,
  value,
}: AlertsProvidersTestHarnessProps): ReactElement => {
  return (
    <AlertsContext.Provider value={value}>
      <AlertsControlPanelProvider>{children}</AlertsControlPanelProvider>
    </AlertsContext.Provider>
  );
};
