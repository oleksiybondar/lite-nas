import { ServiceMetricContext } from "@contexts/service-metric-context";
import { buildPaginationContext } from "@domain/pagination/helpers/buildPaginationContext";
import { useClientPagination } from "@domain/pagination/hooks/useClientPagination";
import { usePaginationState } from "@domain/pagination/hooks/usePaginationState";
import type {
  ServiceMetricActionMethods,
  ServiceMetricContextValue,
  ServiceMetricSnapshotDTO,
  ServiceMetricStartupBehaviour,
  ServiceMetricStateAction,
  ServiceMetricUnitDTO,
} from "@dto/monitoring/service-metric";
import { useFilteredRecords } from "@hooks/useFilteredRecords";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { usePollingResource } from "@hooks/usePollingResource";
import { PollingResourceProvider } from "@providers/PollingResourceProvider";
import { parseServiceMetricSnapshotResponse } from "@schemas/monitoring/service-metric";
import type { PropsWithChildren, ReactElement } from "react";
import { useMemo } from "react";

type ServiceMetricFilterKey = "activeState" | "enabledState";

type ServiceMetricFilterState = ReturnType<typeof useServiceMetricFilterState>;
type ServiceMetricFilterOptions = ReturnType<typeof useServiceMetricFilterOptions>;

const defaultServicesPage = 1;
const defaultServicesPageSize = 50;

/**
 * Snapshot polling provider configured for gateway-backed service inspection.
 */
export const ServiceMetricProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { snapshotIntervalMs } = useMonitoringPollingSettings();
  const actions = useMemo<ServiceMetricActionMethods>(() => {
    return {
      restart: async (serviceName: string): Promise<void> => {
        void serviceName;
      },
      toggleStartupBehaviour: async (
        serviceName: string,
        behaviour: ServiceMetricStartupBehaviour,
      ): Promise<void> => {
        void serviceName;
        void behaviour;
      },
      toggleState: async (serviceName: string, action: ServiceMetricStateAction): Promise<void> => {
        void serviceName;
        void action;
      },
    };
  }, []);

  return (
    <PollingResourceProvider
      actions={actions}
      errorMessage="Failed to load service metrics snapshot."
      parseResponse={parseServiceMetricSnapshotResponse}
      path="/api/service-metrics/snapshot"
      queryKey={["polling-resource", "service-metrics", "snapshot"]}
      refetchIntervalMs={snapshotIntervalMs}
    >
      <ServiceMetricStateProvider>{children}</ServiceMetricStateProvider>
    </PollingResourceProvider>
  );
};

/**
 * Adds service-specific record search, filter, and local pagination state on top of the polling slice.
 */
const ServiceMetricStateProvider = ({ children }: PropsWithChildren): ReactElement => {
  const base = usePollingResource<ServiceMetricSnapshotDTO, ServiceMetricActionMethods>();
  const value = useServiceMetricContextValue(base);

  return <ServiceMetricContext.Provider value={value}>{children}</ServiceMetricContext.Provider>;
};

/**
 * Builds the derived service-metric context value layered on top of the polling resource.
 */
const useServiceMetricContextValue = (
  base: ReturnType<typeof usePollingResource<ServiceMetricSnapshotDTO, ServiceMetricActionMethods>>,
): ServiceMetricContextValue => {
  const allServices = base.snapshot?.services ?? [];
  const filterState = useServiceMetricFilterState(allServices);
  const filterOptions = useServiceMetricFilterOptions(allServices);

  return useMemoServiceMetricContextValue(base, allServices, filterState, filterOptions);
};

/**
 * Builds the filtered record and pagination state derived from the full service snapshot.
 */
const useServiceMetricFilterState = (allServices: ServiceMetricUnitDTO[]) => {
  const paginationState = usePaginationState({
    initialPage: defaultServicesPage,
    initialPageSize: defaultServicesPageSize,
  });
  const filtered = useFilteredRecords<ServiceMetricUnitDTO, ServiceMetricFilterKey>({
    getSearchText: buildServiceSearchText,
    initialFilters: {
      activeState: "",
      enabledState: "",
    },
    matchesFilter: matchServiceFilter,
    records: allServices,
  });
  const pagination = useClientPagination({
    pagination: paginationState,
    records: filtered.records,
  });

  return { filtered, pagination };
};

/**
 * Builds the selectable filter option lists exposed by the service snapshot view.
 */
const useServiceMetricFilterOptions = (allServices: ServiceMetricUnitDTO[]) => {
  const availableActiveStates = useMemo(() => {
    return buildServiceFilterOptions(allServices, (service) => service.active_state);
  }, [allServices]);
  const availableEnabledStates = useMemo(() => {
    return buildServiceFilterOptions(allServices, (service) => service.enabled_state);
  }, [allServices]);

  return { availableActiveStates, availableEnabledStates };
};

/**
 * Memoizes the final context value exposed by the service metric provider.
 */
const useMemoServiceMetricContextValue = (
  base: ReturnType<typeof usePollingResource<ServiceMetricSnapshotDTO, ServiceMetricActionMethods>>,
  allServices: ServiceMetricUnitDTO[],
  filterState: ServiceMetricFilterState,
  filterOptions: ServiceMetricFilterOptions,
): ServiceMetricContextValue => {
  const { filtered, pagination } = filterState;

  return useMemo(() => {
    return {
      ...base,
      ...buildPaginationContext(pagination),
      activeStateFilter: filtered.filters.activeState,
      availableActiveStates: filterOptions.availableActiveStates,
      availableEnabledStates: filterOptions.availableEnabledStates,
      clearFilters: () => {
        filtered.resetFilters();
        pagination.resetPage();
      },
      enabledStateFilter: filtered.filters.enabledState,
      getServiceByName: (serviceName: string): ServiceMetricUnitDTO | null => {
        return allServices.find((service) => service.name === serviceName) ?? null;
      },
      search: filtered.search,
      services: pagination.records,
      setActiveStateFilter: (value: string) => {
        pagination.resetPage();
        filtered.setFilterValue("activeState", value);
      },
      setEnabledStateFilter: (value: string) => {
        pagination.resetPage();
        filtered.setFilterValue("enabledState", value);
      },
      setSearch: (value: string) => {
        pagination.resetPage();
        filtered.setSearch(value);
      },
      totalServices: filtered.totalCount,
      visibleServicesCount: filtered.filteredCount,
    };
  }, [allServices, base, filterOptions, filtered, pagination]);
};

/**
 * Builds the free-text search blob used by the generic filtered-records hook.
 */
const buildServiceSearchText = (service: ServiceMetricUnitDTO): string => {
  return [
    service.name,
    service.description,
    service.active_state,
    service.sub_state,
    service.enabled_state,
    service.load_state,
    service.manager,
    service.unit_type,
  ]
    .filter((value): value is string => typeof value === "string" && value.trim().length > 0)
    .join(" ");
};

/**
 * Matches one service record against one active string-valued filter.
 */
const matchServiceFilter = (
  service: ServiceMetricUnitDTO,
  key: ServiceMetricFilterKey,
  value: string,
): boolean => {
  if (key === "activeState") {
    return service.active_state === value;
  }

  return service.enabled_state === value;
};

/**
 * Builds one sorted unique option list from the current service snapshot values.
 */
const buildServiceFilterOptions = (
  services: ServiceMetricUnitDTO[],
  selectValue: (service: ServiceMetricUnitDTO) => string | null | undefined,
): string[] => {
  return [
    ...new Set(
      services.map(selectValue).filter((value): value is string => Boolean(value?.trim())),
    ),
  ].sort((first, second) => first.localeCompare(second));
};
