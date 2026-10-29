import { localizeMonitoringMetricLabel } from "@components/monitoring/metric-labels";
import {
  createMonitoringScrollableTableSx,
  monitoringDetailColumnSx,
} from "@components/monitoring/table-panel-shared";
import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { formatMetricValue } from "@helpers/metric-display";
import type {
  NetworkProtocolsPanelData,
  NetworkProtocolsPanelPortRow,
  NetworkProtocolsPanelRemoteIPRow,
} from "@helpers/network-protocols-panel";
import type { MetricChartLabel } from "@helpers/system-metric-chart";
import { useLanguage } from "@lite-nas/language";
import Divider from "@mui/material/Divider";
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
import type { NetworkProtocolsPanelProps } from "./types";

const networkProtocolMetaColumnSx = {
  flex: "1 1 320px",
  minWidth: 0,
  p: 1,
} as const;

const networkProtocolChartColumnSx = {
  flex: "2 1 640px",
  minWidth: 0,
  p: 1,
} as const;

const networkProtocolScrollableTableSx = createMonitoringScrollableTableSx(260);

/**
 * Renders the protocols-and-sockets panel with a wider totals chart and scrollable detail tables.
 */
export const NetworkProtocolsPanel = ({
  capacity,
  protocols,
}: NetworkProtocolsPanelProps): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Paper sx={{ p: 2, width: "100%" }}>
      <Stack spacing={1.25}>
        <Typography data-testid="network-protocols-panel-title" variant="h4">
          {getLabel("admin.monitoringMetrics.protocolsAndSockets")}
        </Typography>
        {renderSummaryRow(protocols.summaryLabels, getLabel)}
        <Divider />
        {renderWorkspaceRow(capacity, protocols, getLabel)}
        {renderDetailTables(protocols, getLabel)}
      </Stack>
    </Paper>
  );
};

/**
 * Renders the summary metadata row above the main protocols workspace.
 */
const renderSummaryRow = (
  summaryLabels: MetricChartLabel[],
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack
      data-testid="network-protocols-summary-row"
      direction="row"
      flexWrap="wrap"
      gap={1}
      useFlexGap
    >
      {summaryLabels.map((label) => {
        return renderLabel(label, "network-protocols-summary-label", getLabel);
      })}
    </Stack>
  );
};

/**
 * Renders the metadata and chart workspace row with a wider totals chart.
 */
const renderWorkspaceRow = (
  capacity: number,
  protocols: NetworkProtocolsPanelData,
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack
      data-testid="network-protocols-workspace-row"
      direction="row"
      flexWrap="wrap"
      gap={1.5}
      useFlexGap
    >
      <Paper sx={networkProtocolMetaColumnSx} variant="outlined">
        <Stack spacing={1}>
          {renderMetadataSection({
            title: getLabel("admin.monitoringMetrics.connectionState"),
            testId: "network-protocols-state-row",
            testClass: "network-protocols-state-label",
            labels: protocols.socketStateLabels,
            getLabel,
          })}
          <Divider />
          {renderMetadataSection({
            title: getLabel("admin.monitoringMetrics.protocolDistribution"),
            testId: "network-protocols-distribution-row",
            testClass: "network-protocols-distribution-label",
            labels: protocols.protocolTotalsLabels,
            getLabel,
          })}
        </Stack>
      </Paper>
      <Paper sx={networkProtocolChartColumnSx} variant="outlined">
        <Stack spacing={0.625}>
          <Typography
            data-test-class="network-protocols-chart-title"
            data-test-name="Protocol socket totals"
            variant="body2"
          >
            {getLabel("admin.monitoringMetrics.protocolSocketTotals")}
          </Typography>
          <ValueLineChart
            capacity={capacity}
            formatValue={formatMetricValue}
            stamps={protocols.protocolTotalsSeries.stamps}
            valuesByKey={protocols.protocolTotalsSeries.valuesByKey}
          />
        </Stack>
      </Paper>
    </Stack>
  );
};

/**
 * Renders the scrollable detail tables shown under the main protocols workspace.
 */
const renderDetailTables = (
  protocols: NetworkProtocolsPanelData,
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack
      data-testid="network-protocols-detail-columns"
      direction="row"
      flexWrap="wrap"
      gap={1.5}
      useFlexGap
    >
      {renderPortsTable(protocols.portRows, getLabel)}
      {renderRemoteIPsTable(protocols.remoteIPRows, getLabel)}
    </Stack>
  );
};

/**
 * Renders one labeled metadata section inside the left protocols column.
 */
const renderMetadataSection = ({
  title,
  testId,
  testClass,
  labels,
  getLabel,
}: {
  title: string;
  testId: string;
  testClass: string;
  labels: MetricChartLabel[];
  getLabel: (key: string) => string;
}): ReactElement => {
  return (
    <Stack spacing={0.625}>
      <Typography
        data-test-class="network-protocols-row-title"
        data-test-name={title}
        variant="body2"
      >
        {title}
      </Typography>
      <Stack data-testid={testId} spacing={0.5}>
        {labels.map((label) => {
          return renderLabel(label, testClass, getLabel);
        })}
      </Stack>
    </Stack>
  );
};

/**
 * Renders the local ports and common app mapping table.
 */
const renderPortsTable = (
  rows: NetworkProtocolsPanelPortRow[],
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Paper sx={monitoringDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="network-protocols-table-title"
          data-test-name="Local ports and apps"
          variant="body2"
        >
          {getLabel("admin.monitoringMetrics.localPortsAndApps")}
        </Typography>
        <TableContainer
          data-testid="network-protocols-ports-table"
          sx={networkProtocolScrollableTableSx}
        >
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>{getLabel("admin.monitoringMetrics.app")}</TableCell>
                <TableCell>{getLabel("admin.monitoringMetrics.connection")}</TableCell>
                <TableCell align="right">{getLabel("admin.monitoringMetrics.sockets")}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows.length > 0 ? (
                rows.map((row) => {
                  return (
                    <TableRow
                      data-test-class="network-protocols-port-row"
                      data-test-name={row.connection}
                      key={row.connection}
                    >
                      <TableCell>
                        {row.app === "Unknown app"
                          ? getLabel("admin.monitoringMetrics.unknownApp")
                          : row.app}
                      </TableCell>
                      <TableCell>{row.connection}</TableCell>
                      <TableCell align="right">{row.count}</TableCell>
                    </TableRow>
                  );
                })
              ) : (
                <TableRow>
                  <TableCell colSpan={3}>
                    {getLabel("admin.monitoringMetrics.noLocalPorts")}
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
 * Renders the remote IP connection table.
 */
const renderRemoteIPsTable = (
  rows: NetworkProtocolsPanelRemoteIPRow[],
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Paper sx={monitoringDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="network-protocols-table-title"
          data-test-name="Remote IPs"
          variant="body2"
        >
          {getLabel("admin.monitoringMetrics.remoteIps")}
        </Typography>
        <TableContainer
          data-testid="network-protocols-remote-ips-table"
          sx={networkProtocolScrollableTableSx}
        >
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>{getLabel("admin.monitoringMetrics.remoteIp")}</TableCell>
                <TableCell align="right">
                  {getLabel("admin.monitoringMetrics.connections")}
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows.length > 0 ? (
                rows.map((row) => {
                  return (
                    <TableRow
                      data-test-class="network-protocols-remote-ip-row"
                      data-test-name={row.ip}
                      key={row.ip}
                    >
                      <TableCell>{row.ip}</TableCell>
                      <TableCell align="right">{row.count}</TableCell>
                    </TableRow>
                  );
                })
              ) : (
                <TableRow>
                  <TableCell colSpan={2}>
                    {getLabel("admin.monitoringMetrics.noRemoteIps")}
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
 * Renders one compact key/value metadata row inside the protocols-and-sockets panel.
 */
const renderLabel = (
  label: MetricChartLabel,
  testClass: string,
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Typography
      data-test-class={testClass}
      data-test-name={label.key}
      key={label.key}
      variant="body2"
    >
      {localizeMonitoringMetricLabel(label.key, getLabel)}: {label.value}
    </Typography>
  );
};
