import { createMonitoringScrollableTableSx } from "@components/monitoring/table-panel-shared";
import type { ProcessMetricProcessDTO, ProcessMetricSortKey } from "@dto/monitoring/process-metric";
import { formatMetricBytes } from "@helpers/metric-display";
import { useProcessMetric } from "@hooks/useProcessMetric";
import { useLanguage } from "@lite-nas/language";
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
  const { getLabel } = useLanguage();

  return (
    <SystemTelemetrySnapshotScaffold
      content={renderProcessMetricTable(state, getLabel)}
      controls={renderProcessMetricControls(state, getLabel)}
      hasNextPage={state.hasNextPage}
      hasPreviousPage={state.hasPreviousPage}
      itemLabel={getLabel("admin.telemetry.process.itemLabel")}
      pageSize={state.pageSize}
      pagination={{
        page: state.page,
        setPage: state.setPage,
        totalCount: state.visibleProcessesCount,
        totalPages: state.totalPages,
      }}
      summary={buildProcessMetricSummary(state, getLabel)}
      testIdPrefix="process-metric"
      title={getLabel("admin.telemetry.process.title")}
      totalSummary={getLabel("admin.telemetry.process.totalSummary")
        .replace("{visible}", String(state.visibleProcessesCount))
        .replace("{total}", String(state.totalProcesses))}
    />
  );
};

/**
 * Renders the shared search and reset controls for the processes snapshot view.
 */
const renderProcessMetricControls = (
  state: ReturnType<typeof useProcessMetric>,
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack direction={{ md: "row", xs: "column" }} spacing={1.5}>
      <TextField
        data-testid="process-metric-search-control"
        label={getLabel("admin.telemetry.process.search")}
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
        {getLabel("admin.telemetry.process.clear")}
      </Button>
    </Stack>
  );
};

/**
 * Renders the current page of process rows inside the task-manager-style table.
 */
const renderProcessMetricTable = (
  state: ReturnType<typeof useProcessMetric>,
  getLabel: (key: string) => string,
): ReactElement => {
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
              label: getLabel("admin.telemetry.process.columnName"),
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
              label: getLabel("admin.telemetry.process.columnUser"),
              sortKey: "username",
              sx: processCompactColumnSx,
            })}
            <TableCell sx={processCompactColumnSx}>
              {getLabel("admin.telemetry.process.columnState")}
            </TableCell>
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
              {getLabel("admin.telemetry.process.columnThreads")}
            </TableCell>
            <TableCell align="right" sx={processCompactColumnSx}>
              {getLabel("admin.telemetry.process.columnFileDescriptors")}
            </TableCell>
            <TableCell align="right" sx={processCompactColumnSx}>
              {getLabel("admin.telemetry.process.columnAction")}
            </TableCell>
            <TableCell sx={processStartedColumnSx}>
              {getLabel("admin.telemetry.process.columnStarted")}
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>{renderProcessMetricRows(state, getLabel)}</TableBody>
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
  getLabel: (key: string) => string,
): ReactElement | ReactElement[] => {
  if (state.processes.length === 0) {
    return (
      <TableRow data-testid="process-metric-empty-row">
        <TableCell colSpan={12}>{getLabel("admin.telemetry.process.empty")}</TableCell>
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
        <TableCell sx={processCompactColumnSx}>
          {process.username ?? getLabel("admin.telemetry.process.unknown")}
        </TableCell>
        <TableCell sx={processCompactColumnSx}>{process.state}</TableCell>
        <TableCell align="right" sx={processCompactColumnSx}>
          {formatProcessCPUPct(process.cpu.cpu_pct)}
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
            title={getLabel("admin.telemetry.process.terminateTitle").replace(
              "{pid}",
              String(process.pid),
            )}
            variant="outlined"
            size="small"
          >
            {getLabel("admin.telemetry.process.terminate")}
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
const buildProcessMetricSummary = (
  state: ReturnType<typeof useProcessMetric>,
  getLabel: (key: string) => string,
): string => {
  if (state.isLoading) {
    return getLabel("admin.telemetry.process.loading");
  }

  if (state.isError) {
    return state.error?.message ?? getLabel("admin.telemetry.process.loadFailed");
  }

  const refreshState = state.isFetching ? getLabel("admin.telemetry.process.refreshing") : "";
  const sortState =
    state.sortKey === null
      ? getLabel("admin.telemetry.process.unsorted")
      : getLabel("admin.telemetry.process.sorted")
          .replace("{key}", getLabel(`admin.telemetry.process.sort.${state.sortKey}`))
          .replace(
            "{direction}",
            getLabel(`admin.telemetry.process.direction.${state.sortDirection}`),
          );

  return getLabel("admin.telemetry.process.summary")
    .replace("{visible}", String(state.visibleProcessesCount))
    .replace("{total}", String(state.totalProcesses))
    .replace("{sortState}", sortState)
    .replace("{refreshState}", refreshState);
};

/**
 * Formats one process CPU percent value for compact task-manager table output.
 */
const formatProcessCPUPct = (value: number): string => {
  return `${value.toFixed(1)}%`;
};
