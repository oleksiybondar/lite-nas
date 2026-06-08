import { PercentGradientChart } from "@components/monitoring/PercentGradientChart";
import { PercentGradientMultiChart } from "@components/monitoring/PercentGradientMultiChart";
import {
  toSystemMemoryUsageChartCardData,
  toSystemPerCoreCpuChartSeries,
  toSystemTotalCpuChartSeries,
} from "@helpers/system-metric-chart";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useSystemMetric } from "@hooks/useSystemMetric";
import Divider from "@mui/material/Divider";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Route state rendered when gateway-backed system CPU and memory metrics are available.
 */
export const SystemTelemetrySystemMetricState = (): ReactElement => {
  const { items } = useSystemMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const memoryCardData = toSystemMemoryUsageChartCardData(items);
  const totalCpuSeries = toSystemTotalCpuChartSeries(items);
  const perCoreCpuSeries = toSystemPerCoreCpuChartSeries(items);

  return (
    <Stack spacing={3}>
      <Paper data-testid="system-telemetry-total-cpu-card" sx={{ p: 3 }}>
        <Stack spacing={2}>
          <Typography data-testid="system-telemetry-total-cpu-title" variant="h2">
            Total CPU
          </Typography>
          <PercentGradientChart
            capacity={maxRecords}
            stamps={totalCpuSeries.stamps}
            values={totalCpuSeries.values}
          />
        </Stack>
      </Paper>
      <Divider />
      <Paper data-testid="system-telemetry-per-core-cpu-card" sx={{ p: 3 }}>
        <Stack spacing={2}>
          <Typography data-testid="system-telemetry-per-core-cpu-title" variant="h2">
            Per-core CPU
          </Typography>
          <PercentGradientMultiChart
            capacity={maxRecords}
            stamps={perCoreCpuSeries.stamps}
            valuesByKey={perCoreCpuSeries.valuesByKey}
          />
        </Stack>
      </Paper>
      <Divider />
      <Paper data-testid="system-telemetry-ram-card" sx={{ p: 3 }}>
        <Stack spacing={2}>
          <Typography data-testid="system-telemetry-ram-title" variant="h2">
            RAM
          </Typography>
          <Stack
            data-test-class="system-telemetry-ram-labels"
            direction="row"
            flexWrap="wrap"
            gap={1.5}
            useFlexGap
          >
            {memoryCardData.labels.map((label) => {
              return (
                <Typography
                  data-test-class="system-telemetry-ram-label"
                  data-test-name={label.key}
                  key={label.key}
                  variant="body2"
                >
                  {label.key}: {label.value}
                </Typography>
              );
            })}
          </Stack>
          <PercentGradientChart
            capacity={maxRecords}
            stamps={memoryCardData.series.stamps}
            values={memoryCardData.series.values}
          />
        </Stack>
      </Paper>
    </Stack>
  );
};
