import {
  createMonitoringScrollableTableSx,
  monitoringDetailColumnSx,
} from "@components/monitoring/table-panel-shared";
import type {
  DiskFilesystemPanelRow,
  DiskMountPanelRow,
  DiskMountsPanelData,
} from "@helpers/disk-metric-panel";
import type { MetricChartLabel } from "@helpers/system-metric-chart";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

const diskMountScrollableTableSx = createMonitoringScrollableTableSx(280);

type DiskMountsPanelProps = {
  /**
   * Browser-facing mounts and filesystem data rendered by the storage panel.
   */
  mounts: DiskMountsPanelData;
};

/**
 * Renders the active storage mounts and filesystem summary panel.
 */
export const DiskMountsPanel = ({ mounts }: DiskMountsPanelProps): ReactElement => {
  return (
    <Paper data-testid="disk-mounts-panel" sx={{ p: 2, width: "100%" }}>
      <Stack spacing={1.25}>
        <Typography data-testid="disk-mounts-panel-title" variant="h4">
          Mounts & filesystems
        </Typography>
        {renderSummaryRow(mounts.summaryLabels)}
        <Stack
          data-testid="disk-mounts-detail-columns"
          direction="row"
          flexWrap="wrap"
          gap={1.5}
          useFlexGap
        >
          {renderMountsTable(mounts.mountRows)}
          {renderFilesystemsTable(mounts.filesystemRows)}
        </Stack>
      </Stack>
    </Paper>
  );
};

/**
 * Renders the summary metadata row above the storage detail tables.
 */
const renderSummaryRow = (summaryLabels: MetricChartLabel[]): ReactElement => {
  return (
    <Stack data-testid="disk-mounts-summary-row" direction="row" flexWrap="wrap" gap={1} useFlexGap>
      {summaryLabels.map((label) => {
        return (
          <Typography
            data-test-class="disk-mounts-summary-label"
            data-test-name={label.key}
            key={label.key}
            variant="body2"
          >
            {label.key}: {label.value}
          </Typography>
        );
      })}
    </Stack>
  );
};

/**
 * Renders the active storage mounts table.
 */
const renderMountsTable = (rows: DiskMountPanelRow[]): ReactElement => {
  return (
    <Paper sx={monitoringDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="disk-mounts-table-title"
          data-test-name="Storage mounts"
          variant="body2"
        >
          Storage mounts
        </Typography>
        <TableContainer data-testid="disk-mounts-table" sx={diskMountScrollableTableSx}>
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>Mount</TableCell>
                <TableCell>Device</TableCell>
                <TableCell>FS</TableCell>
                <TableCell align="right">Usage</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows.length > 0 ? (
                rows.map((row) => {
                  return (
                    <TableRow
                      data-test-class="disk-mount-row"
                      data-test-name={row.mountpoint}
                      key={row.mountpoint}
                    >
                      <TableCell>{row.mountpoint}</TableCell>
                      <TableCell>{row.deviceLabel}</TableCell>
                      <TableCell>{row.filesystem}</TableCell>
                      <TableCell align="right">{formatDiskMountUsage(row)}</TableCell>
                    </TableRow>
                  );
                })
              ) : (
                <TableRow>
                  <TableCell colSpan={4}>No storage mounts observed</TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Stack>
    </Paper>
  );
};

/**
 * Renders the filesystem distribution summary table.
 */
const renderFilesystemsTable = (rows: DiskFilesystemPanelRow[]): ReactElement => {
  return (
    <Paper sx={monitoringDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="disk-mounts-table-title"
          data-test-name="Filesystem mix"
          variant="body2"
        >
          Filesystem mix
        </Typography>
        <TableContainer data-testid="disk-filesystems-table" sx={diskMountScrollableTableSx}>
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>Filesystem</TableCell>
                <TableCell>Class</TableCell>
                <TableCell align="right">Mounts</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows.length > 0 ? (
                rows.map((row) => {
                  return (
                    <TableRow
                      data-test-class="disk-filesystem-row"
                      data-test-name={row.filesystem}
                      key={row.filesystem}
                    >
                      <TableCell>{row.filesystem}</TableCell>
                      <TableCell>{row.type}</TableCell>
                      <TableCell align="right">{row.mountCount}</TableCell>
                    </TableRow>
                  );
                })
              ) : (
                <TableRow>
                  <TableCell colSpan={3}>No filesystem summaries observed</TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Stack>
    </Paper>
  );
};

/**
 * Formats the compact usage summary shown in each disk mount row.
 */
const formatDiskMountUsage = (row: DiskMountPanelRow): string => {
  return `${row.usedLabel} / ${row.totalLabel} (${row.usageLabel})`;
};
