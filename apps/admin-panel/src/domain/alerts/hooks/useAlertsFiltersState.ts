import type { AlertSeverity, AlertsFilterState } from "@dto/alerts/alerts";
import { createEmptyAlertsFilterState } from "@helpers/alerts";
import { useState } from "react";

/**
 * Shared filter state owned by one alerts route slice.
 */
export type AlertsFiltersStateResult = AlertsFilterState & {
  clearFilters: () => void;
  setCategoryFilter: (value: string[]) => void;
  setPriorityFilter: (value: number[]) => void;
  setSeverityFilter: (value: AlertSeverity[]) => void;
  setSourceFilter: (value: string[]) => void;
};

/**
 * Creates the shared filter state used by the alerts provider.
 */
export const useAlertsFiltersState = (): AlertsFiltersStateResult => {
  const [categoryFilter, setCategoryFilter] = useState<string[]>([]);
  const [sourceFilter, setSourceFilter] = useState<string[]>([]);
  const [priorityFilter, setPriorityFilter] = useState<number[]>([]);
  const [severityFilter, setSeverityFilter] = useState<AlertSeverity[]>([]);

  return {
    categoryFilter,
    clearFilters: () => {
      const emptyFilters = createEmptyAlertsFilterState();

      setCategoryFilter(emptyFilters.categoryFilter);
      setSourceFilter(emptyFilters.sourceFilter);
      setPriorityFilter(emptyFilters.priorityFilter);
      setSeverityFilter(emptyFilters.severityFilter);
    },
    priorityFilter,
    setCategoryFilter,
    setPriorityFilter,
    setSeverityFilter,
    setSourceFilter,
    severityFilter,
    sourceFilter,
  };
};
