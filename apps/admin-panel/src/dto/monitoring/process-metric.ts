import type { PollingResourceContextValue } from "@dto/monitoring/polling-resource";
import type { PaginationActions, PaginationMeta, PaginationState } from "@dto/pagination";

/**
 * CPU usage block returned for one process snapshot item.
 */
export type ProcessMetricCpuDTO = {
  system_ticks: number;
  total_ticks: number;
  user_ticks: number;
};

/**
 * Memory usage block returned for one process snapshot item.
 */
export type ProcessMetricMemoryDTO = {
  rss_bytes: number;
  vms_bytes: number;
};

/**
 * One process snapshot row returned by the transport.
 */
export type ProcessMetricProcessDTO = {
  cmdline?: string | null;
  cpu: ProcessMetricCpuDTO;
  cwd?: string | null;
  exe?: string | null;
  gid: number;
  memory: ProcessMetricMemoryDTO;
  name: string;
  open_fds: number;
  pid: number;
  ppid: number;
  start_time: string;
  state: string;
  threads: number;
  uid: number;
  username?: string | null;
};

/**
 * One timestamped process metrics snapshot item.
 */
export type ProcessMetricSnapshotDTO = {
  processes: ProcessMetricProcessDTO[];
  timestamp: string;
};

/**
 * Snapshot payload shape returned by the transport before frontend normalization.
 */
export type ProcessMetricSnapshotTransportDTO = {
  processes?: ProcessMetricProcessDTO[] | null;
  timestamp: string;
};

/**
 * Response envelope returned by process metrics snapshot endpoints.
 */
export type ProcessMetricSnapshotResponseDTO = {
  code?: string;
  data: ProcessMetricSnapshotTransportDTO;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Sortable process columns exposed by the task-manager view.
 */
export type ProcessMetricSortKey = "cpu" | "gid" | "name" | "pid" | "ram" | "uid" | "username";

/**
 * Sort direction for one active process sort column.
 */
export type ProcessMetricSortDirection = "asc" | "desc";

/**
 * Client-side sort state exposed by the processes provider.
 */
export type ProcessMetricSortState = {
  setSort: (key: ProcessMetricSortKey) => void;
  sortDirection: ProcessMetricSortDirection;
  sortKey: ProcessMetricSortKey | null;
};

/**
 * Runtime action methods exposed by the processes provider.
 */
export type ProcessMetricActionMethods = {
  terminate: (pid: number) => Promise<void>;
};

/**
 * Shared pagination contract exposed by the processes provider.
 */
export type ProcessMetricPaginationState = Pick<PaginationState, "page" | "pageSize"> &
  Pick<PaginationMeta, "hasNextPage" | "hasPreviousPage" | "totalPages"> &
  Pick<
    PaginationActions,
    "nextPage" | "previousPage" | "resetPage" | "resetPagination" | "setPage" | "setPageSize"
  >;

/**
 * Process-specific local search, sort, and pagination state exposed by the wrapper provider.
 */
export type ProcessMetricFilterState = ProcessMetricPaginationState &
  ProcessMetricSortState & {
    clearFilters: () => void;
    processes: ProcessMetricProcessDTO[];
    search: string;
    setSearch: (value: string) => void;
    totalProcesses: number;
    visibleProcessesCount: number;
  };

/**
 * Runtime state exposed by the processes polling provider.
 */
export type ProcessMetricContextValue = PollingResourceContextValue<
  ProcessMetricSnapshotDTO,
  ProcessMetricActionMethods & ProcessMetricFilterState
>;
