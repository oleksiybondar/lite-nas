import type {
  AlertDomain,
  AlertSeverity,
  AlertsControlPanelContextValue,
  AlertsControlPanelOption,
} from "@dto/alerts/alerts";

const supportedSeverities: AlertSeverity[] = ["critical", "error", "info", "warning"];
const supportedPriorities = [1, 2, 3, 4, 5] as const;
const alertsCategoryOptionsByDomain: Record<AlertDomain, string[]> = {
  security: [],
  system: [],
};
const alertsSourceOptionsByDomain: Record<AlertDomain, string[]> = {
  security: [],
  system: ["resource-monitor"],
};

/**
 * Builds selectable category options from domain-specific alerts configuration.
 */
export const buildCategoryOptions = (domain: AlertDomain): AlertsControlPanelOption<string>[] => {
  return buildStringOptions(alertsCategoryOptionsByDomain[domain]);
};

/**
 * Builds selectable source options from domain-specific alerts configuration.
 */
export const buildSourceOptions = (domain: AlertDomain): AlertsControlPanelOption<string>[] => {
  return buildStringOptions(alertsSourceOptionsByDomain[domain]);
};

/**
 * Builds the fixed priority options currently supported by the alerts UI.
 */
export const buildPriorityOptions = (): AlertsControlPanelOption<number>[] => {
  return supportedPriorities.map((value) => ({
    label: String(value),
    value,
  }));
};

/**
 * Builds the supported severity options for alerts filtering.
 */
export const buildSeverityOptions = (): AlertsControlPanelOption<AlertSeverity>[] => {
  return supportedSeverities.map((severity) => ({
    label: formatAlertsFilterLabel(severity),
    value: severity,
  }));
};

/**
 * Formats one selected filter value for MUI multi-select rendering.
 */
export const formatFilterSelection = (values: Array<number | string>): string => {
  if (values.length === 0) {
    return "All";
  }

  return values.join(", ");
};

/**
 * Builds the focused control-panel context value from the shared alerts slice.
 */
export const buildAlertsControlPanelValue = ({
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
}: {
  categoryFilter: string[];
  clearFilters: () => void;
  domain: AlertDomain;
  page: number;
  priorityFilter: number[];
  search: string;
  setCategoryFilter: (value: string[]) => void;
  setPage: (page: number) => void;
  setPriorityFilter: (value: number[]) => void;
  setSearch: (value: string) => void;
  setSeverityFilter: (value: AlertSeverity[]) => void;
  setSourceFilter: (value: string[]) => void;
  severityFilter: AlertSeverity[];
  sourceFilter: string[];
  totalCount: number;
  totalPages: number;
}): AlertsControlPanelContextValue => {
  return {
    availableCategoryOptions: buildCategoryOptions(domain),
    availablePriorityOptions: buildPriorityOptions(),
    availableSeverityOptions: buildSeverityOptions(),
    availableSourceOptions: buildSourceOptions(domain),
    categoryFilter,
    clearFilters,
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
  };
};

/**
 * Converts one string list into sorted unique control-panel options.
 */
const buildStringOptions = (values: string[]): AlertsControlPanelOption<string>[] => {
  return [...new Set(values)]
    .sort((first, second) => first.localeCompare(second))
    .map((value) => ({
      label: value,
      value,
    }));
};

/**
 * Formats one alerts filter value for display in fixed option labels.
 */
const formatAlertsFilterLabel = (value: string): string => {
  return value.slice(0, 1).toUpperCase() + value.slice(1).replaceAll("-", " ");
};
