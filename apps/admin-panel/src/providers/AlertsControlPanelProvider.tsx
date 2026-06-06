import { buildAlertsControlPanelValue } from "@components/alerts/AlertsControlPanel/helpers";
import { AlertsControlPanelContext } from "@contexts/alerts-control-panel-context";
import { useAlerts } from "@hooks/useAlerts";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Provides focused alerts control-panel state derived from the shared alerts route slice.
 */
export const AlertsControlPanelProvider = ({ children }: PropsWithChildren): ReactElement => {
  const {
    categoryFilter,
    clearFilters,
    domain,
    page,
    priorityFilter,
    search,
    setCategoryFilter,
    setPage,
    setPriorityFilter,
    setSearch,
    setSeverityFilter,
    setSourceFilter,
    severityFilter,
    sourceFilter,
    totalCount,
    totalPages,
  } = useAlerts();

  const value = buildAlertsControlPanelValue({
    categoryFilter,
    clearFilters,
    domain,
    page,
    priorityFilter,
    search,
    setCategoryFilter,
    setPage,
    setPriorityFilter,
    setSearch,
    setSeverityFilter,
    setSourceFilter,
    severityFilter,
    sourceFilter,
    totalCount,
    totalPages,
  });

  return (
    <AlertsControlPanelContext.Provider value={value}>
      {children}
    </AlertsControlPanelContext.Provider>
  );
};
