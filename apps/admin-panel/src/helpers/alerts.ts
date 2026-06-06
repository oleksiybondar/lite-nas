import type {
  AlertCategory,
  AlertDomain,
  AlertFilterDTO,
  AlertSeverity,
  AlertsFilterState,
} from "@dto/alerts/alerts";

const defaultPage = 1;
const defaultPageSize = 20;

/**
 * Browser-facing query state that uniquely identifies one alerts list slice.
 */
export type AlertsQueryState = AlertsFilterState & {
  category: AlertCategory;
  domain: AlertDomain;
  page: number;
  pageSize: number;
};

/**
 * Returns the default alerts page number used by the provider state.
 */
export const getDefaultAlertsPage = (): number => {
  return defaultPage;
};

/**
 * Returns the default alerts page size used by the provider state.
 */
export const getDefaultAlertsPageSize = (): number => {
  return defaultPageSize;
};

/**
 * Builds the browser route path for one alerts domain/category page.
 */
export const buildAlertsRoutePath = (domain: AlertDomain, category: AlertCategory): string => {
  return `/alerts/${domain}/${category}`;
};

/**
 * Builds the gateway list path for one alerts domain/category page and filter state.
 */
export const buildAlertsListPath = ({
  category,
  categoryFilter,
  domain,
  page,
  pageSize,
  priorityFilter,
  severityFilter,
  sourceFilter,
}: AlertsQueryState): string => {
  const searchParams = new URLSearchParams({
    page: String(page),
    size: String(pageSize),
  });

  for (const filter of buildAlertsFilters({
    categoryFilter,
    priorityFilter,
    severityFilter,
    sourceFilter,
  })) {
    searchParams.append("filters", JSON.stringify(filter));
  }

  return `${buildAlertsCategoryBasePath(domain, category)}?${searchParams.toString()}`;
};

/**
 * Builds one stable TanStack query key for an alerts list slice.
 */
export const buildAlertsQueryKey = ({
  category,
  categoryFilter,
  domain,
  page,
  pageSize,
  priorityFilter,
  severityFilter,
  sourceFilter,
}: AlertsQueryState): readonly unknown[] => {
  return [
    "alerts",
    domain,
    category,
    page,
    pageSize,
    {
      categoryFilter: sortStrings(categoryFilter),
      priorityFilter: sortNumbers(priorityFilter),
      severityFilter: sortSeverities(severityFilter),
      sourceFilter: sortStrings(sourceFilter),
    },
  ] as const;
};

/**
 * Builds the gateway filter objects for the current alerts UI state.
 */
export const buildAlertsFilters = ({
  categoryFilter,
  priorityFilter,
  severityFilter,
  sourceFilter,
}: AlertsFilterState): AlertFilterDTO[] => {
  const filters: AlertFilterDTO[] = [];

  if (categoryFilter.length > 0) {
    filters.push({
      condition: "in",
      key: "category",
      values: sortStrings(categoryFilter),
    });
  }

  if (sourceFilter.length > 0) {
    filters.push({
      condition: "in",
      key: "source",
      values: sortStrings(sourceFilter),
    });
  }

  if (priorityFilter.length > 0) {
    filters.push({
      condition: "in",
      key: "priority",
      values: sortNumbers(priorityFilter).map(String),
    });
  }

  if (severityFilter.length > 0) {
    filters.push({
      condition: "in",
      key: "severity",
      values: sortSeverities(severityFilter),
    });
  }

  return filters;
};

/**
 * Builds the browser-facing title for one alerts route slice.
 */
export const buildAlertsPageTitle = (category: AlertCategory): string => {
  return `${formatAlertsLabel(category)} alerts`;
};

/**
 * Builds the browser-facing summary for one alerts route slice.
 */
export const buildAlertsPageSummary = (domain: AlertDomain, category: AlertCategory): string => {
  return `${formatAlertsLabel(domain)} ${category} alert panels for this route will be wired to gateway-backed alert queries.`;
};

/**
 * Formats one route-safe alerts segment for display.
 */
export const formatAlertsLabel = (value: string): string => {
  return value.slice(0, 1).toUpperCase() + value.slice(1).replaceAll("-", " ");
};

/**
 * Returns the initial empty filter state for shared alerts pages.
 */
export const createEmptyAlertsFilterState = (): AlertsFilterState => {
  return {
    categoryFilter: [],
    priorityFilter: [],
    severityFilter: [],
    sourceFilter: [],
  };
};

const buildAlertsCategoryBasePath = (domain: AlertDomain, category: AlertCategory): string => {
  if (category === "all") {
    return `/api/alerts/${domain}`;
  }

  return `/api/alerts/${domain}/${category}`;
};

const sortStrings = (values: string[]): string[] => {
  return [...values].sort((first, second) => first.localeCompare(second));
};

const sortNumbers = (values: number[]): number[] => {
  return [...values].sort((first, second) => first - second);
};

const sortSeverities = (values: AlertSeverity[]): AlertSeverity[] => {
  return [...values].sort((first, second) => first.localeCompare(second)) as AlertSeverity[];
};
