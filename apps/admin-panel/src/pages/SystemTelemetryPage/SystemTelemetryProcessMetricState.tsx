import { createMonitoringScrollableTableSx } from "@components/monitoring/table-panel-shared";
import type { ProcessMetricProcessDTO, ProcessMetricSortKey } from "@dto/monitoring/process-metric";
import { formatMetricBytes } from "@helpers/metric-display";
import { useProcessMetric } from "@hooks/useProcessMetric";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import type { SxProps, Theme } from "@mui/material/styles";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableSortLabel from "@mui/material/TableSortLabel";
import TextField from "@mui/material/TextField";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { SystemTelemetrySnapshotScaffold } from "./SystemTelemetrySnapshotScaffold";

const processTableScrollableSx = createMonitoringScrollableTableSx(720);
const processCompactColumnSx = { whiteSpace: "nowrap", width: 1 } as const;
const processNameColumnSx = { minWidth: 360 } as const;
const processStartedColumnSx = { minWidth: 180, whiteSpace: "nowrap" } as const;

type SortableHeaderCellOptions = {
  align?: "left" | "right";
  label: string;
  sortKey: ProcessMetricSortKey;
  sx?: SxProps<Theme>;
};

/**
 * Task-manager-style processes state rendered from the snapshot polling provider.
 */
export const SystemTelemetryProcessMetricState = (): ReactElement => {
  const state = useProcessMetric();

  return (
    <SystemTelemetrySnapshotScaffold
      content={renderProcessMetricTable(state)}
      controls={renderProcessMetricControls(state)}
      hasNextPage={state.hasNextPage}
      hasPreviousPage={state.hasPreviousPage}
      itemLabel="processes"
      pageSize={state.pageSize}
      pagination={{
        page: state.page,
        setPage: state.setPage,
        totalCount: state.visibleProcessesCount,
        totalPages: state.totalPages,
      }}
      summary={buildProcessMetricSummary(state)}
      testIdPrefix="process-metric"
      title="Processes snapshot"
      totalSummary={`${state.visibleProcessesCount} of ${state.totalProcesses} processes match the current search.`}
    />
  );
};

/**
 * Renders the shared search and reset controls for the processes snapshot view.
 */
const renderProcessMetricControls = (state: ReturnType<typeof useProcessMetric>): ReactElement => {
  return (
    <Stack direction={{ md: "row", xs: "column" }} spacing={1.5}>
      <TextField
        data-testid="process-metric-search-control"
        label="Search processes"
        name="processMetricSearch"
        onChange={(event) => {
          state.setSearch(event.target.value);
        }}
        size="small"
        sx={{ minWidth: 280 }}
        value={state.search}
      />
      <Button
        data-testid="process-metric-clear-filters-button"
        onClick={state.clearFilters}
        variant="text"
      >
        Clear search
      </Button>
    </Stack>
  );
};

/**
 * Renders the current page of process rows inside the task-manager-style table.
 */
const renderProcessMetricTable = (state: ReturnType<typeof useProcessMetric>): ReactElement => {
  return (
    <TableContainer data-testid="process-metric-table" sx={processTableScrollableSx}>
      <Table size="small" stickyHeader sx={{ minWidth: 1200 }}>
        <TableHead>
          <TableRow>
            {renderSortableProcessHeaderCell(state, {
              align: "right",
              label: "PID",
              sortKey: "pid",
              sx: processCompactColumnSx,
            })}
            {renderSortableProcessHeaderCell(state, {
              label: "Name",
              sortKey: "name",
              sx: processNameColumnSx,
            })}
            {renderSortableProcessHeaderCell(state, {
              align: "right",
              label: "UID",
              sortKey: "uid",
              sx: processCompactColumnSx,
            })}
            {renderSortableProcessHeaderCell(state, {
              align: "right",
              label: "GID",
              sortKey: "gid",
              sx: processCompactColumnSx,
            })}
            {renderSortableProcessHeaderCell(state, {
              label: "User",
              sortKey: "username",
              sx: processCompactColumnSx,
            })}
            <TableCell sx={processCompactColumnSx}>State</TableCell>
            {renderSortableProcessHeaderCell(state, {
              align: "right",
              label: "CPU",
              sortKey: "cpu",
              sx: processCompactColumnSx,
            })}
            {renderSortableProcessHeaderCell(state, {
              align: "right",
              label: "RAM",
              sortKey: "ram",
              sx: processCompactColumnSx,
            })}
            <TableCell align="right" sx={processCompactColumnSx}>
              Threads
            </TableCell>
            <TableCell align="right" sx={processCompactColumnSx}>
              FDs
            </TableCell>
            <TableCell align="right" sx={processCompactColumnSx}>
              Action
            </TableCell>
            <TableCell sx={processStartedColumnSx}>Started</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>{renderProcessMetricRows(state)}</TableBody>
      </Table>
    </TableContainer>
  );
};

/**
 * Renders one sortable process header cell with tri-state toggle behavior.
 */
const renderSortableProcessHeaderCell = (
  state: ReturnType<typeof useProcessMetric>,
  { align = "left", label, sortKey, sx }: SortableHeaderCellOptions,
): ReactElement => {
  const isActive = state.sortKey === sortKey;

  return (
    <TableCell align={align} sx={sx}>
      <TableSortLabel
        active={isActive}
        data-testid={`process-metric-sort-${sortKey}`}
        direction={isActive ? state.sortDirection : "asc"}
        onClick={() => {
          state.setSort(sortKey);
        }}
      >
        {label}
      </TableSortLabel>
    </TableCell>
  );
};

/**
 * Renders process table rows or the current empty state row.
 */
const renderProcessMetricRows = (
  state: ReturnType<typeof useProcessMetric>,
): ReactElement | ReactElement[] => {
  if (state.processes.length === 0) {
    return (
      <TableRow data-testid="process-metric-empty-row">
        <TableCell colSpan={12}>No processes match the current search.</TableCell>
      </TableRow>
    );
  }

  return state.processes.map((process) => {
    return (
      <TableRow
        data-test-class="process-metric-row"
        data-test-name={process.name}
        data-test-pid={String(process.pid)}
        key={process.pid}
      >
        <TableCell align="right" sx={processCompactColumnSx}>
          {process.pid}
        </TableCell>
        <TableCell sx={processNameColumnSx}>
          <Stack spacing={0.25}>
            <Typography
              data-testid={`process-metric-name-${process.pid}`}
              variant="body2"
              sx={{
                fontWeight: 600,
                overflowWrap: "anywhere",
              }}
            >
              {process.name}
            </Typography>
            {renderProcessCommandLine(process)}
          </Stack>
        </TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          {process.uid}
        </TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          {process.gid}
        </TableCell>
        <TableCell sx={processCompactColumnSx}>{process.username ?? "unknown"}</TableCell>
        <TableCell sx={processCompactColumnSx}>{process.state}</TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          {process.cpu.total_ticks}
        </TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          {formatMetricBytes(process.memory.rss_bytes)}
        </TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          {process.threads}
        </TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          {process.open_fds}
        </TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          <Button
            color="error"
            data-testid={`process-metric-terminate-${process.pid}`}
            onClick={() => {
              void state.terminate(process.pid);
            }}
            sx={{ minWidth: 96 }}
            title={`Terminate process ${process.pid}`}
            variant="outlined"
            size="small"
          >
            Terminate
          </Button>
        </TableCell>
        <TableCell sx={processStartedColumnSx}>{process.start_time}</TableCell>
      </TableRow>
    );
  });
};

/**
 * Renders the secondary command-line preview for one process row when available.
 */
const renderProcessCommandLine = (process: ProcessMetricProcessDTO): ReactElement | null => {
  if (typeof process.cmdline !== "string" || process.cmdline.trim().length === 0) {
    return null;
  }

  return (
    <Tooltip arrow placement="top-start" title={process.cmdline}>
      <Typography
        color="text.secondary"
        data-testid={`process-metric-cmdline-${process.pid}`}
        noWrap
        variant="caption"
      >
        {process.cmdline}
      </Typography>
    </Tooltip>
  );
};

/**
 * Resolves the short summary copy shown above the process snapshot table.
 */
const buildProcessMetricSummary = (state: ReturnType<typeof useProcessMetric>): string => {
  if (state.isLoading) {
    return "Loading the latest process snapshot from the gateway.";
  }

  if (state.isError) {
    return state.error?.message ?? "Failed to load the latest process snapshot.";
  }

  const refreshState = state.isFetching ? " A refresh is currently in progress." : "";
  const sortState =
    state.sortKey === null
      ? " No client-side sorting is applied."
      : ` Sorted by ${state.sortKey} (${state.sortDirection}).`;

  return `${state.visibleProcessesCount} of ${state.totalProcesses} processes match the current search.${sortState}${refreshState}`;
};
