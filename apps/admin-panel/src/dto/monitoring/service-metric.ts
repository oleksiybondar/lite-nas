import type { PollingResourceContextValue } from "@dto/monitoring/polling-resource";
import type { PaginationActions, PaginationMeta, PaginationState } from "@dto/pagination";

/**
 * Cgroup-scoped service resource values returned by the transport when available.
 */
export type ServiceMetricResourcesDTO = {
  cpu_usage_usec?: number | null;
  memory_bytes?: number | null;
};

/**
 * One systemd-managed service unit snapshot returned by the transport.
 */
export type ServiceMetricUnitDTO = {
  active_state?: string | null;
  cgroup?: string | null;
  control_pid?: number | null;
  description?: string | null;
  enabled_state?: string | null;
  fragment_path?: string | null;
  load_state?: string | null;
  main_pid?: number | null;
  manager: string;
  name: string;
  pids?: number[] | null;
  resources?: ServiceMetricResourcesDTO | null;
  result?: string | null;
  slice?: string | null;
  started_at?: string | null;
  sub_state?: string | null;
  unit_type: string;
};

/**
 * One timestamped service metrics snapshot item.
 */
export type ServiceMetricSnapshotDTO = {
  services: ServiceMetricUnitDTO[];
  timestamp: string;
};

/**
 * Snapshot payload shape returned by the transport before frontend normalization.
 */
export type ServiceMetricSnapshotTransportDTO = {
  services?: ServiceMetricUnitDTO[] | null;
  timestamp: string;
};

/**
 * Response envelope returned by service metrics snapshot endpoints.
 */
export type ServiceMetricSnapshotResponseDTO = {
  code?: string;
  data: ServiceMetricSnapshotTransportDTO;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Startup-state action exposed by the services provider.
 */
export type ServiceMetricStartupBehaviour = "disable" | "enable";

/**
 * Runtime-state action exposed by the services provider.
 */
export type ServiceMetricStateAction = "start" | "stop";

/**
 * Service-level action methods exposed by the services provider.
 */
export type ServiceMetricActionMethods = {
  restart: (serviceName: string) => Promise<void>;
  toggleStartupBehaviour: (
    serviceName: string,
    behaviour: ServiceMetricStartupBehaviour,
  ) => Promise<void>;
  toggleState: (serviceName: string, action: ServiceMetricStateAction) => Promise<void>;
};

/**
 * Shared pagination contract exposed by the services provider.
 */
export type ServiceMetricPaginationState = Pick<PaginationState, "page" | "pageSize"> &
  Pick<PaginationMeta, "hasNextPage" | "hasPreviousPage" | "totalPages"> &
  Pick<
    PaginationActions,
    "nextPage" | "previousPage" | "resetPage" | "resetPagination" | "setPage" | "setPageSize"
  >;

/**
 * Service-specific local search and filter state exposed by the wrapper provider.
 */
export type ServiceMetricFilterState = ServiceMetricPaginationState & {
  activeStateFilter: string;
  availableActiveStates: string[];
  availableEnabledStates: string[];
  clearFilters: () => void;
  enabledStateFilter: string;
  getServiceByName: (serviceName: string) => ServiceMetricUnitDTO | null;
  search: string;
  services: ServiceMetricUnitDTO[];
  setActiveStateFilter: (value: string) => void;
  setEnabledStateFilter: (value: string) => void;
  setSearch: (value: string) => void;
  totalServices: number;
  visibleServicesCount: number;
};

/**
 * Runtime state exposed by the services polling provider.
 */
export type ServiceMetricContextValue = PollingResourceContextValue<
  ServiceMetricSnapshotDTO,
  ServiceMetricActionMethods & ServiceMetricFilterState
>;
