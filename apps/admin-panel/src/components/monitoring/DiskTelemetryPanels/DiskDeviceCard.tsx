import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import type { DiskDevicePanelItem } from "@helpers/disk-metric-panel";
import { formatDiskReadWriteValue, formatDiskUtilizationValue } from "@helpers/disk-metric-panel";
import type { MetricChartLabel } from "@helpers/system-metric-chart";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

type DiskDeviceCardProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Browser-facing device card data rendered by this card.
   */
  device: DiskDevicePanelItem;
};

const diskDeviceChartSectionSx = {
  flex: "1 1 320px",
  minWidth: 0,
  p: 1,
} as const;

/**
 * Renders one disk device card with metadata, cumulative totals, and time-series charts.
 */
export const DiskDeviceCard = ({ capacity, device }: DiskDeviceCardProps): ReactElement => {
  return (
    <Paper
      data-test-class="disk-device-card"
      data-test-name={device.name}
      sx={{ p: 1.5, width: "100%" }}
      variant="outlined"
    >
      <Stack spacing={1}>
        {renderDeviceHeader(device.name, device.sizeLabel)}
        {renderBodyText("disk-device-subtitle", device.name, device.subtitle)}
        {renderBodyText("disk-device-connection", device.name, device.connectionLabel)}
        {renderBodyText("disk-device-mounts", device.name, device.mountsLabel)}
        {renderLabels(device.ioSummaryLabels)}
        <Divider />
        {renderChartsRow(capacity, device)}
      </Stack>
    </Paper>
  );
};

/**
 * Renders the device name and size chip row.
 */
const renderDeviceHeader = (name: string, sizeLabel: string): ReactElement => {
  return (
    <Stack alignItems="center" direction="row" justifyContent="space-between" spacing={1}>
      <Typography data-test-class="disk-device-name" data-test-name={name} variant="h4">
        {name}
      </Typography>
      <Chip
        data-test-class="disk-device-size"
        data-test-name={name}
        label={sizeLabel}
        size="small"
        variant="outlined"
      />
    </Stack>
  );
};

/**
 * Renders one compact device metadata line.
 */
const renderBodyText = (testClass: string, testName: string, value: string): ReactElement => {
  return (
    <Typography data-test-class={testClass} data-test-name={testName} variant="body2">
      {value}
    </Typography>
  );
};

/**
 * Renders the current read/write total labels row.
 */
const renderLabels = (labels: MetricChartLabel[]): ReactElement => {
  return (
    <Stack data-test-class="disk-device-totals" direction="row" flexWrap="wrap" gap={1} useFlexGap>
      {labels.map((label) => {
        return (
          <Typography
            data-test-class="disk-device-total"
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
 * Renders the responsive chart row inside one disk device card.
 */
const renderChartsRow = (capacity: number, device: DiskDevicePanelItem): ReactElement => {
  return (
    <Stack
      data-test-class="disk-device-chart-row"
      direction="row"
      flexWrap="wrap"
      gap={1.5}
      useFlexGap
    >
      {renderChartSection(
        capacity,
        "Read/write throughput",
        device.readWriteSeries,
        formatDiskReadWriteValue,
      )}
      {renderChartSection(
        capacity,
        "Busy time",
        device.utilizationSeries,
        formatDiskUtilizationValue,
      )}
    </Stack>
  );
};

/**
 * Renders one chart section inside the disk device chart row.
 */
const renderChartSection = (
  capacity: number,
  title: string,
  series: DiskDevicePanelItem["readWriteSeries"],
  formatValue: (value: number) => string,
): ReactElement => {
  return (
    <Paper sx={diskDeviceChartSectionSx} variant="outlined">
      <Stack spacing={0.625}>
        <Typography
          data-test-class="disk-device-chart-title"
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
