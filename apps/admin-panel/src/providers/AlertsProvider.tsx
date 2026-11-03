import { AlertsContext } from "@contexts/alerts-context";
import {
  type AcknowledgeAlertInput,
  useAcknowledgeAlert,
} from "@domain/alerts/hooks/useAcknowledgeAlert";
import { useAlertsFiltersState } from "@domain/alerts/hooks/useAlertsFiltersState";
import { useAlertsList } from "@domain/alerts/hooks/useAlertsList";
import { useAlertsPaginationState } from "@domain/alerts/hooks/useAlertsPaginationState";
import { useAlertsSearchState } from "@domain/alerts/hooks/useAlertsSearchState";
import { useServerPagination } from "@domain/pagination/hooks/useServerPagination";
import type {
  AlertCategory,
  AlertDomain,
  AlertListItemDTO,
  AlertsContextValue,
} from "@dto/alerts/alerts";
import type { ServerPaginationController } from "@dto/pagination";
import { buildAlertsListPath, buildAlertsQueryKey, buildAlertsRoutePath } from "@helpers/alerts";
import {
  type QueryClient,
  type QueryObserverResult,
  type RefetchOptions,
  type UseMutationResult,
  useQueryClient,
} from "@tanstack/react-query";
import type { PropsWithChildren, ReactElement } from "react";

type AlertsProviderProps = PropsWithChildren<{
  /**
   * Current alerts route category.
   */
  category: AlertCategory;
  /**
   * Current alerts route domain.
   */
  domain: AlertDomain;
}>;

/**
 * Provides shared alerts state and commands for one route slice.
 */
export const AlertsProvider = ({
  category,
  children,
  domain,
}: AlertsProviderProps): ReactElement => {
  const queryClient = useQueryClient();
  const value = useAlertsProviderValue(domain, category, queryClient);

  return <AlertsContext.Provider value={value}>{children}</AlertsContext.Provider>;
};

/**
 * Builds the shared alerts context value for one route slice.
 */
const useAlertsProviderValue = (
  domain: AlertDomain,
  category: AlertCategory,
  queryClient: QueryClient,
): AlertsContextValue => {
  const paginationState = useAlertsPaginationState();
  const filters = useAlertsFiltersState();
  const searchState = useAlertsSearchState();
  const alertsQuery = useAlertsList({
    category,
    categoryFilter: filters.categoryFilter,
    domain,
    page: paginationState.page,
    pageSize: paginationState.pageSize,
    priorityFilter: filters.priorityFilter,
    severityFilter: filters.severityFilter,
    sourceFilter: filters.sourceFilter,
  });
  const acknowledgeMutation = useAcknowledgeAlert(domain);
  const pagination = useServerPagination<AlertListItemDTO>({
    data: alertsQuery.data
      ? {
          records: alertsQuery.data.items,
          totalCount: alertsQuery.data.metadata.total_count,
          totalPages: alertsQuery.data.metadata.total_pages,
        }
      : null,
    pagination: paginationState,
  });
  const queryState = {
    category,
    categoryFilter: filters.categoryFilter,
    domain,
    page: pagination.page,
    pageSize: pagination.pageSize,
    priorityFilter: filters.priorityFilter,
    severityFilter: filters.severityFilter,
    sourceFilter: filters.sourceFilter,
  };

  return buildAlertsContextValue({
    acknowledgeMutation,
    alertsQuery,
    category,
    domain,
    filters,
    pagination,
    queryClient,
    queryState,
    searchState,
  });
};

type BuildAlertsContextValueOptions = {
  acknowledgeMutation: UseMutationResult<void, Error, AcknowledgeAlertInput>;
  alertsQuery: ReturnType<typeof useAlertsList>;
  category: AlertCategory;
  domain: AlertDomain;
  filters: ReturnType<typeof useAlertsFiltersState>;
  pagination: ServerPaginationController<AlertListItemDTO>;
  queryClient: QueryClient;
  queryState: Parameters<typeof buildAlertsListPath>[0];
  searchState: ReturnType<typeof useAlertsSearchState>;
};

const buildAlertsContextValue = ({
  acknowledgeMutation,
  alertsQuery,
  category,
  domain,
  filters,
  pagination,
  queryClient,
  queryState,
  searchState,
}: BuildAlertsContextValueOptions): AlertsContextValue => {
  const refetch = createAlertsRefetch(alertsQuery.refetch);

  return {
    acknowledge: createAcknowledge(domain, acknowledgeMutation.mutateAsync, queryClient),
    acknowledgeMany: createAcknowledgeMany(domain, acknowledgeMutation.mutateAsync, queryClient),
    apiPath: buildAlertsListPath(queryState),
    category,
    categoryFilter: filters.categoryFilter,
    clearFilters: createClearFilters(filters.clearFilters, pagination.resetPage),
    domain,
    error: alertsQuery.error ?? null,
    hasNextPage: pagination.hasNextPage,
    hasPreviousPage: pagination.hasPreviousPage,
    isAcknowledging: acknowledgeMutation.isPending,
    isError: alertsQuery.isError,
    isFetching: alertsQuery.isFetching,
    isLoading: alertsQuery.isLoading,
    items: pagination.records,
    nextPage: pagination.nextPage,
    page: pagination.page,
    pageSize: pagination.pageSize,
    previousPage: pagination.previousPage,
    priorityFilter: filters.priorityFilter,
    queryKey: buildAlertsQueryKey(queryState),
    refetch,
    resetPage: pagination.resetPage,
    resetPagination: pagination.resetPagination,
    routePath: buildAlertsRoutePath(domain, category),
    search: searchState.search,
    setCategoryFilter: createFilterSetter(filters.setCategoryFilter, pagination.resetPage),
    setPage: pagination.setPage,
    setPageSize: pagination.setPageSize,
    setPriorityFilter: createFilterSetter(filters.setPriorityFilter, pagination.resetPage),
    setSearch: searchState.setSearch,
    setSeverityFilter: createFilterSetter(filters.setSeverityFilter, pagination.resetPage),
    setSourceFilter: createFilterSetter(filters.setSourceFilter, pagination.resetPage),
    severityFilter: filters.severityFilter,
    sourceFilter: filters.sourceFilter,
    totalCount: pagination.totalCount,
    totalPages: pagination.totalPages,
  };
};

/**
 * Shared primitive used by single and bulk acknowledge flows.
 */
const acknowledgeSingle = async (
  input: AcknowledgeAlertInput,
  mutateAsync: (input: AcknowledgeAlertInput) => Promise<void>,
): Promise<void> => {
  await mutateAsync(input);
};

const createAcknowledge = (
  domain: AlertDomain,
  mutateAsync: (input: AcknowledgeAlertInput) => Promise<void>,
  queryClient: QueryClient,
): ((id: string) => Promise<void>) => {
  return async (id: string): Promise<void> => {
    await acknowledgeSingle({ id }, mutateAsync);
    await queryClient.invalidateQueries({ queryKey: ["alerts", domain] });
  };
};

const createAcknowledgeMany = (
  domain: AlertDomain,
  mutateAsync: (input: AcknowledgeAlertInput) => Promise<void>,
  queryClient: QueryClient,
): ((ids: string[]) => Promise<void>) => {
  return async (ids: string[]): Promise<void> => {
    const hasSucceeded = await acknowledgeManyUntilFailure(ids, mutateAsync, domain);

    if (hasSucceeded) {
      await queryClient.invalidateQueries({ queryKey: ["alerts", domain] });
    }
  };
};

const acknowledgeManyUntilFailure = async (
  ids: string[],
  mutateAsync: (input: AcknowledgeAlertInput) => Promise<void>,
  domain: AlertDomain,
): Promise<boolean> => {
  let hasSucceeded = false;

  for (const id of ids) {
    const itemSucceeded = await acknowledgeOneUntilFailure(id, mutateAsync, domain);

    if (!itemSucceeded) {
      return hasSucceeded;
    }

    hasSucceeded = true;
  }

  return hasSucceeded;
};

const createAlertsRefetch = (
  refetch: (
    options?: RefetchOptions,
  ) => Promise<QueryObserverResult<ReturnType<typeof useAlertsList>["data"], Error>>,
): (() => Promise<unknown>) => {
  return async (): Promise<unknown> => {
    return refetch();
  };
};

const createClearFilters = (clearFilters: () => void, resetPage: () => void): (() => void) => {
  return (): void => {
    clearFilters();
    resetPage();
  };
};

const acknowledgeOneUntilFailure = async (
  id: string,
  mutateAsync: (input: AcknowledgeAlertInput) => Promise<void>,
  domain: AlertDomain,
): Promise<boolean> => {
  try {
    await acknowledgeSingle({ id }, mutateAsync);
    return true;
  } catch (error) {
    notifyError(error instanceof Error ? error.message : `Failed to acknowledge ${domain} alerts.`);
    return false;
  }
};

const createFilterSetter = <T,>(
  setter: (value: T) => void,
  resetPage: () => void,
): ((value: T) => void) => {
  return (value: T): void => {
    setter(value);
    resetPage();
  };
};

/**
 * Temporary centralized error reporting stub until snackbar context is added.
 */
const notifyError = (message: string): void => {
  void message;
};
