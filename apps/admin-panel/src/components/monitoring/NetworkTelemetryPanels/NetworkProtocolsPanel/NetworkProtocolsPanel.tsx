import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { formatMetricValue } from "@helpers/metric-display";
import type {
  NetworkProtocolsPanelData,
  NetworkProtocolsPanelPortRow,
  NetworkProtocolsPanelRemoteIPRow,
} from "@helpers/network-protocols-panel";
import type { MetricChartLabel } from "@helpers/system-metric-chart";
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

const networkProtocolDetailColumnSx = {
  flex: "1 1 480px",
  minWidth: 0,
  p: 1,
} as const;

const networkProtocolScrollableTableSx = {
  maxHeight: 260,
  minHeight: 260,
  overflowY: "auto",
} as const;

/**
 * Renders the protocols-and-sockets panel with a wider totals chart and scrollable detail tables.
 */
export const NetworkProtocolsPanel = ({
  capacity,
  protocols,
}: NetworkProtocolsPanelProps): ReactElement => {
  return (
    <Paper sx={{ p: 2, width: "100%" }}>
      <Stack spacing={1.25}>
        <Typography data-testid="network-protocols-panel-title" variant="h4">
          Protocols & sockets
        </Typography>
        {renderSummaryRow(protocols.summaryLabels)}
        <Divider />
        {renderWorkspaceRow(capacity, protocols)}
        {renderDetailTables(protocols)}
      </Stack>
    </Paper>
  );
};

/**
 * Renders the summary metadata row above the main protocols workspace.
 */
const renderSummaryRow = (summaryLabels: MetricChartLabel[]): ReactElement => {
  return (
    <Stack
      data-testid="network-protocols-summary-row"
      direction="row"
      flexWrap="wrap"
      gap={1}
      useFlexGap
    >
      {summaryLabels.map((label) => {
        return renderLabel(label, "network-protocols-summary-label");
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
          {renderMetadataSection(
            "Connection state",
            "network-protocols-state-row",
            "network-protocols-state-label",
            protocols.socketStateLabels,
          )}
          <Divider />
          {renderMetadataSection(
            "Protocol distribution",
            "network-protocols-distribution-row",
            "network-protocols-distribution-label",
            protocols.protocolTotalsLabels,
          )}
        </Stack>
      </Paper>
      <Paper sx={networkProtocolChartColumnSx} variant="outlined">
        <Stack spacing={0.625}>
          <Typography
            data-test-class="network-protocols-chart-title"
            data-test-name="Protocol socket totals"
            variant="body2"
          >
            Protocol socket totals
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
const renderDetailTables = (protocols: NetworkProtocolsPanelData): ReactElement => {
  return (
    <Stack
      data-testid="network-protocols-detail-columns"
      direction="row"
      flexWrap="wrap"
      gap={1.5}
      useFlexGap
    >
      {renderPortsTable(protocols.portRows)}
      {renderRemoteIPsTable(protocols.remoteIPRows)}
    </Stack>
  );
};

/**
 * Renders one labeled metadata section inside the left protocols column.
 */
const renderMetadataSection = (
  title: string,
  testId: string,
  testClass: string,
  labels: MetricChartLabel[],
): ReactElement => {
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
          return renderLabel(label, testClass);
        })}
      </Stack>
    </Stack>
  );
};

/**
 * Renders the local ports and common app mapping table.
 */
const renderPortsTable = (rows: NetworkProtocolsPanelPortRow[]): ReactElement => {
  return (
    <Paper sx={networkProtocolDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="network-protocols-table-title"
          data-test-name="Local ports and apps"
          variant="body2"
        >
          Local ports and apps
        </Typography>
        <TableContainer
          data-testid="network-protocols-ports-table"
          sx={networkProtocolScrollableTableSx}
        >
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>App</TableCell>
                <TableCell>Connection</TableCell>
                <TableCell align="right">Sockets</TableCell>
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
                      <TableCell>{row.app}</TableCell>
                      <TableCell>{row.connection}</TableCell>
                      <TableCell align="right">{row.count}</TableCell>
                    </TableRow>
                  );
                })
              ) : (
                <TableRow>
                  <TableCell colSpan={3}>No observed local ports</TableCell>
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
const renderRemoteIPsTable = (rows: NetworkProtocolsPanelRemoteIPRow[]): ReactElement => {
  return (
    <Paper sx={networkProtocolDetailColumnSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="network-protocols-table-title"
          data-test-name="Remote IPs"
          variant="body2"
        >
          Remote IPs
        </Typography>
        <TableContainer
          data-testid="network-protocols-remote-ips-table"
          sx={networkProtocolScrollableTableSx}
        >
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>Remote IP</TableCell>
                <TableCell align="right">Connections</TableCell>
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
                  <TableCell colSpan={2}>No observed remote IPs</TableCell>
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
const renderLabel = (label: MetricChartLabel, testClass: string): ReactElement => {
  return (
    <Typography
      data-test-class={testClass}
      data-test-name={label.key}
      key={label.key}
      variant="body2"
    >
      {label.key}: {label.value}
    </Typography>
  );
};
