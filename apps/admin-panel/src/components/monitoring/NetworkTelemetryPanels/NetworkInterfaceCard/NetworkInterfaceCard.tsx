import {
  localizeMonitoringMetricLabel,
  localizeMonitoringMetricValue,
} from "@components/monitoring/metric-labels";
import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { formatMetricBytesPerSecond, formatMetricValue } from "@helpers/metric-display";
import type { MetricChartLabel, MetricMultiChartSeries } from "@helpers/system-metric-chart";
import { useLanguage } from "@lite-nas/language";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { NetworkInterfaceCardProps } from "./types";

const networkInterfaceChartSectionSx = {
  flex: "1 1 320px",
  minWidth: 0,
  p: 1,
} as const;

/**
 * Renders one interface card with status, adapter details, totals, and chart columns.
 */
export const NetworkInterfaceCard = ({
  capacity,
  networkInterface,
}: NetworkInterfaceCardProps): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Paper
      data-test-class="network-interface-card"
      data-test-name={networkInterface.name}
      sx={{ p: 1.5, width: "100%" }}
      variant="outlined"
    >
      <Stack spacing={1}>
        {renderInterfaceHeader(networkInterface.name, networkInterface.status, getLabel)}
        {renderBodyText(
          "network-interface-adapter",
          networkInterface.name,
          networkInterface.adapterLabel,
          getLabel,
        )}
        {renderBodyText(
          "network-interface-link-details",
          networkInterface.name,
          networkInterface.linkDetailsLabel,
          getLabel,
        )}
        {renderTotals(networkInterface.totalsLabels, getLabel)}
        <Divider />
        {renderChartsRow(capacity, networkInterface, getLabel)}
      </Stack>
    </Paper>
  );
};

/**
 * Renders the interface name and status chip row.
 */
const renderInterfaceHeader = (
  name: string,
  status: string,
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack alignItems="center" direction="row" justifyContent="space-between" spacing={1}>
      <Typography data-test-class="network-interface-name" data-test-name={name} variant="h4">
        {name}
      </Typography>
      <Chip
        color={status === "up" ? "success" : "default"}
        data-test-class="network-interface-status"
        data-test-name={name}
        label={
          status === "up" || status === "down"
            ? getLabel(`admin.monitoringMetrics.${status}`)
            : status
        }
        size="small"
        variant="outlined"
      />
    </Stack>
  );
};

/**
 * Renders one compact interface metadata line.
 */
const renderBodyText = (
  testClass: string,
  testName: string,
  value: string,
  getLabel: (key: string) => string,
): ReactElement => {
  const unavailableLabel: Record<string, string> = {
    "Adapter details unavailable": "admin.monitoringMetrics.adapterDetailsUnavailable",
    "Link details unavailable": "admin.monitoringMetrics.linkDetailsUnavailable",
  };
  const labelKey = unavailableLabel[value];
  const localizedValue = labelKey ? getLabel(labelKey) : value;
  return (
    <Typography data-test-class={testClass} data-test-name={testName} variant="body2">
      {localizedValue}
    </Typography>
  );
};

/**
 * Renders the total RX/TX labels row.
 */
const renderTotals = (
  labels: MetricChartLabel[],
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack
      data-test-class="network-interface-totals"
      direction="row"
      flexWrap="wrap"
      gap={1}
      useFlexGap
    >
      {labels.map((label) => {
        return (
          <Typography
            data-test-class="network-interface-total"
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
 * Renders the responsive chart row inside one interface card.
 */
const renderChartsRow = (
  capacity: number,
  networkInterface: NetworkInterfaceCardProps["networkInterface"],
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack
      data-test-class="network-interface-chart-row"
      direction="row"
      flexWrap="wrap"
      gap={1.5}
      useFlexGap
    >
      {renderChartSection(
        capacity,
        getLabel("admin.monitoringMetrics.rxTxThroughput"),
        networkInterface.throughputSeries,
        formatMetricBytesPerSecond,
      )}
      {renderChartSection(
        capacity,
        getLabel("admin.monitoringMetrics.rxTxPackets"),
        networkInterface.packetRateSeries,
        formatMetricValue,
      )}
      {renderChartSection(
        capacity,
        getLabel("admin.monitoringMetrics.rxTxErrorsDrops"),
        networkInterface.errorAndDropSeries,
        formatMetricValue,
      )}
    </Stack>
  );
};

/**
 * Renders one chart section inside the interface chart row.
 */
const renderChartSection = (
  capacity: number,
  title: string,
  series: MetricMultiChartSeries,
  formatValue: (value: number) => string,
): ReactElement => {
  return (
    <Paper sx={networkInterfaceChartSectionSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="network-interface-chart-title"
          data-test-name={title}
          variant="body2"
        >
          {title}
        </Typography>
        <ValueLineChart
          capacity={capacity}
          formatValue={formatValue}
          stamps={series.stamps}
          valuesByKey={series.valuesByKey}
        />
      </Stack>
    </Paper>
  );
};
