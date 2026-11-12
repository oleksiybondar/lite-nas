import {
  localizeMonitoringMetricLabel,
  localizeMonitoringMetricValue,
} from "@components/monitoring/metric-labels";
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
import { useLanguage } from "@lite-nas/language";
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
  const { getLabel } = useLanguage();
  return (
    <Paper data-testid="disk-mounts-panel" sx={{ p: 2, width: "100%" }}>
      <Stack spacing={1.25}>
        <Typography data-testid="disk-mounts-panel-title" variant="h4">
          {getLabel("admin.monitoringMetrics.mountsPanelTitle")}
        </Typography>
        {renderSummaryRow(mounts.summaryLabels, getLabel)}
        <Stack
          data-testid="disk-mounts-detail-columns"
          direction="row"
          flexWrap="wrap"
          gap={1.5}
          useFlexGap
        >
          {renderMountsTable(mounts.mountRows, getLabel)}
          {renderFilesystemsTable(mounts.filesystemRows, getLabel)}
        </Stack>
      </Stack>
    </Paper>
  );
};

/**
 * Renders the summary metadata row above the storage detail tables.
 */
const renderSummaryRow = (
  summaryLabels: MetricChartLabel[],
  getLabel: (key: string) => string,
): ReactElement => {
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
            {localizeMonitoringMetricLabel(label.key, getLabel)}:{" "}
            {localizeMonitoringMetricValue(label.value, getLabel)}
          </Typography>
        );
      })}
    </Stack>
  );
};

/**
 * Renders the active storage mounts table.
 */
const renderMountsTable = (
  rows: DiskMountPanelRow[],
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Paper sx={monitoringDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="disk-mounts-table-title"
          data-test-name="Storage mounts"
          variant="body2"
        >
          {getLabel("admin.monitoringMetrics.storageMounts")}
        </Typography>
        <TableContainer data-testid="disk-mounts-table" sx={diskMountScrollableTableSx}>
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>{getLabel("admin.monitoringMetrics.mount")}</TableCell>
                <TableCell>{getLabel("admin.monitoringMetrics.device")}</TableCell>
                <TableCell>{getLabel("admin.monitoringMetrics.filesystemShort")}</TableCell>
                <TableCell align="right">{getLabel("admin.monitoringMetrics.usage")}</TableCell>
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
                  <TableCell colSpan={4}>
                    {getLabel("admin.monitoringMetrics.noStorageMounts")}
                  </TableCell>
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
const renderFilesystemsTable = (
  rows: DiskFilesystemPanelRow[],
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Paper sx={monitoringDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="disk-mounts-table-title"
          data-test-name="Filesystem mix"
          variant="body2"
        >
          {getLabel("admin.monitoringMetrics.filesystemMix")}
        </Typography>
        <TableContainer data-testid="disk-filesystems-table" sx={diskMountScrollableTableSx}>
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>{getLabel("admin.monitoringMetrics.filesystems")}</TableCell>
                <TableCell>{getLabel("admin.monitoringMetrics.class")}</TableCell>
                <TableCell align="right">{getLabel("admin.monitoringMetrics.mounts")}</TableCell>
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
                      <TableCell>
                        {row.type === "unknown"
                          ? getLabel("admin.monitoringMetrics.unknown")
                          : row.type}
                      </TableCell>
                      <TableCell align="right">{row.mountCount}</TableCell>
                    </TableRow>
                  );
                })
              ) : (
                <TableRow>
                  <TableCell colSpan={3}>
                    {getLabel("admin.monitoringMetrics.noFilesystemSummaries")}
                  </TableCell>
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
