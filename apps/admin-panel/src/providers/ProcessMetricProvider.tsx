import { ProcessMetricContext } from "@contexts/process-metric-context";
import { buildPaginationContext } from "@domain/pagination/helpers/buildPaginationContext";
import { useClientPagination } from "@domain/pagination/hooks/useClientPagination";
import { usePaginationState } from "@domain/pagination/hooks/usePaginationState";
import type {
  ProcessMetricActionMethods,
  ProcessMetricContextValue,
  ProcessMetricProcessDTO,
  ProcessMetricSnapshotDTO,
  ProcessMetricSortDirection,
  ProcessMetricSortKey,
} from "@dto/monitoring/process-metric";
import { useFilteredRecords } from "@hooks/useFilteredRecords";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { usePollingResource } from "@hooks/usePollingResource";
import { PollingResourceProvider } from "@providers/PollingResourceProvider";
import { parseProcessMetricSnapshotResponse } from "@schemas/monitoring/process-metric";
import type { PropsWithChildren, ReactElement } from "react";
import { useMemo, useState } from "react";

const defaultProcessesPage = 1;
const defaultProcessesPageSize = 100;
const defaultSortDirection: ProcessMetricSortDirection = "asc";

/**
 * Snapshot polling provider configured for gateway-backed process inspection.
 */
export const ProcessMetricProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { snapshotIntervalMs } = useMonitoringPollingSettings();
  const actions = useMemo<ProcessMetricActionMethods>(() => {
    return {
      terminate: async (pid: number): Promise<void> => {
        void pid;
      },
    };
  }, []);

  return (
    <PollingResourceProvider
      actions={actions}
      errorMessage="Failed to load process metrics snapshot."
      parseResponse={parseProcessMetricSnapshotResponse}
      path="/api/process-metrics/snapshot"
      queryKey={["polling-resource", "process-metrics", "snapshot"]}
      refetchIntervalMs={snapshotIntervalMs}
    >
      <ProcessMetricStateProvider>{children}</ProcessMetricStateProvider>
    </PollingResourceProvider>
  );
};

/**
 * Adds process-specific search, sort, and local pagination state on top of the polling slice.
 */
const ProcessMetricStateProvider = ({ children }: PropsWithChildren): ReactElement => {
  const base = usePollingResource<ProcessMetricSnapshotDTO, ProcessMetricActionMethods>();
  const value = useProcessMetricContextValue(base);

  return <ProcessMetricContext.Provider value={value}>{children}</ProcessMetricContext.Provider>;
};

/**
 * Builds the derived process-metric context value layered on top of the polling resource.
 */
const useProcessMetricContextValue = (
  base: ReturnType<typeof usePollingResource<ProcessMetricSnapshotDTO, ProcessMetricActionMethods>>,
): ProcessMetricContextValue => {
  const allProcesses = base.snapshot?.processes ?? [];
  const filterState = useProcessMetricFilterState(allProcesses);

  return useMemoProcessMetricContextValue(base, allProcesses, filterState);
};

/**
 * Builds the filtered, sorted, and paginated process state derived from the full snapshot.
 */
const useProcessMetricFilterState = (allProcesses: ProcessMetricProcessDTO[]) => {
  const paginationState = usePaginationState({
    initialPage: defaultProcessesPage,
    initialPageSize: defaultProcessesPageSize,
  });
  const filtered = useFilteredRecords({
    getSearchText: buildProcessSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: allProcesses,
  });
  const [sortKey, setSortKey] = useState<ProcessMetricSortKey | null>(null);
  const [sortDirection, setSortDirection] =
    useState<ProcessMetricSortDirection>(defaultSortDirection);
  const sortedRecords = useMemo(() => {
    if (sortKey === null) {
      return filtered.records;
    }

    return [...filtered.records].sort((first, second) => {
      return compareProcesses(first, second, sortKey, sortDirection);
    });
  }, [filtered.records, sortDirection, sortKey]);
  const pagination = useClientPagination({
    pagination: paginationState,
    records: sortedRecords,
  });

  return {
    filtered,
    pagination,
    setSort: (nextSortKey: ProcessMetricSortKey) => {
      pagination.resetPage();

      if (sortKey !== nextSortKey) {
        setSortKey(nextSortKey);
        setSortDirection(defaultSortDirection);
        return;
      }

      if (sortDirection === "asc") {
        setSortDirection("desc");
        return;
      }

      setSortKey(null);
      setSortDirection(defaultSortDirection);
    },
    sortDirection,
    sortKey,
  };
};

/**
 * Memoizes the final context value exposed by the process metric provider.
 */
const useMemoProcessMetricContextValue = (
  base: ReturnType<typeof usePollingResource<ProcessMetricSnapshotDTO, ProcessMetricActionMethods>>,
  allProcesses: ProcessMetricProcessDTO[],
  filterState: ReturnType<typeof useProcessMetricFilterState>,
): ProcessMetricContextValue => {
  const { filtered, pagination, setSort, sortDirection, sortKey } = filterState;

  return useMemo(() => {
    return {
      ...base,
      ...buildPaginationContext(pagination),
      clearFilters: () => {
        filtered.setSearch("");
        pagination.resetPage();
      },
      processes: pagination.records,
      search: filtered.search,
      setSearch: (value: string) => {
        pagination.resetPage();
        filtered.setSearch(value);
      },
      setSort,
      sortDirection,
      sortKey,
      totalProcesses: allProcesses.length,
      visibleProcessesCount: filtered.filteredCount,
    };
  }, [allProcesses.length, base, filtered, pagination, setSort, sortDirection, sortKey]);
};

/**
 * Builds the free-text search blob used by the generic filtered-records hook.
 */
const buildProcessSearchText = (process: ProcessMetricProcessDTO): string => {
  return [
    process.name,
    process.username,
    process.cmdline,
    process.state,
    String(process.pid),
    String(process.ppid),
    String(process.uid),
    String(process.gid),
  ]
    .filter((value): value is string => typeof value === "string" && value.trim().length > 0)
    .join(" ");
};

/**
 * Compares two process rows for one active client-side sort rule.
 */
const compareProcesses = (
  first: ProcessMetricProcessDTO,
  second: ProcessMetricProcessDTO,
  sortKey: ProcessMetricSortKey,
  sortDirection: ProcessMetricSortDirection,
): number => {
  const primaryComparison = resolvePrimaryProcessComparison(first, second, sortKey, sortDirection);

  if (primaryComparison !== 0) {
    return primaryComparison;
  }

  return buildProcessTieBreakComparison(first, second);
};

/**
 * Resolves the primary comparison result for the active process sort rule.
 */
const resolvePrimaryProcessComparison = (
  first: ProcessMetricProcessDTO,
  second: ProcessMetricProcessDTO,
  sortKey: ProcessMetricSortKey,
  sortDirection: ProcessMetricSortDirection,
): number => {
  const direction = sortDirection === "asc" ? 1 : -1;

  if (sortKey === "name" || sortKey === "username") {
    return compareProcessTextValues(first, second, sortKey, direction);
  }

  return compareProcessNumericValues(first, second, sortKey, direction);
};

/**
 * Compares two process rows by a text-based sort key.
 */
const compareProcessTextValues = (
  first: ProcessMetricProcessDTO,
  second: ProcessMetricProcessDTO,
  sortKey: Extract<ProcessMetricSortKey, "name" | "username">,
  direction: number,
): number => {
  return (
    normalizeSortableText(selectSortableProcessText(first, sortKey)).localeCompare(
      normalizeSortableText(selectSortableProcessText(second, sortKey)),
    ) * direction
  );
};

/**
 * Compares two process rows by a numeric sort key.
 */
const compareProcessNumericValues = (
  first: ProcessMetricProcessDTO,
  second: ProcessMetricProcessDTO,
  sortKey: Exclude<ProcessMetricSortKey, "name" | "username">,
  direction: number,
): number => {
  return (
    (normalizeSortableNumber(selectSortableProcessValue(first, sortKey)) -
      normalizeSortableNumber(selectSortableProcessValue(second, sortKey))) *
    direction
  );
};

/**
 * Resolves the numeric sort value used by one sortable process column.
 */
const selectSortableProcessValue = (
  process: ProcessMetricProcessDTO,
  sortKey: Exclude<ProcessMetricSortKey, "name" | "username">,
): number => {
  if (sortKey === "cpu") {
    return process.cpu.total_ticks;
  }

  if (sortKey === "ram") {
    return process.memory.rss_bytes;
  }

  return process[sortKey];
};

/**
 * Resolves the text sort value used by one sortable process text column.
 */
const selectSortableProcessText = (
  process: ProcessMetricProcessDTO,
  sortKey: Extract<ProcessMetricSortKey, "name" | "username">,
): string => {
  if (sortKey === "username") {
    return process.username ?? "";
  }

  return process.name;
};

/**
 * Builds a deterministic tie-breaker so equal primary sort values do not reshuffle rows.
 */
const buildProcessTieBreakComparison = (
  first: ProcessMetricProcessDTO,
  second: ProcessMetricProcessDTO,
): number => {
  const nameComparison = normalizeSortableText(first.name).localeCompare(
    normalizeSortableText(second.name),
  );

  if (nameComparison !== 0) {
    return nameComparison;
  }

  return first.pid - second.pid;
};

/**
 * Normalizes numeric sort values before comparison.
 */
const normalizeSortableNumber = (value: number): number => {
  return Number.isFinite(value) ? value : Number.NEGATIVE_INFINITY;
};

/**
 * Normalizes text sort values into a stable case-insensitive form.
 */
const normalizeSortableText = (value: string): string => {
  return value.trim().toLocaleLowerCase();
};
