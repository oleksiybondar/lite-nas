import { PercentGradientMultiChart } from "@components/monitoring/PercentGradientMultiChart";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { SystemMetricsCardProps } from "./types";

type SystemMetricsCardPerCoreChartProps = {
  /**
   * Capacity of the chart.
   */
  capacity: number;
  /**
   * Metrics data containing the per-core series.
   */
  metrics: SystemMetricsCardProps["metrics"];
};

/**
 * Renders the per-core CPU percent chart for the system metrics card.
 */
export const SystemMetricsCardPerCoreChart = ({
  capacity,
  metrics,
}: SystemMetricsCardPerCoreChartProps): ReactElement => {
  return (
    <Paper sx={{ p: 1, width: "100%" }} variant="outlined">
      <Stack spacing={0.625}>
        <Typography data-testid="system-metrics-per-core-chart-title" variant="body2">
          Per-core CPU %
        </Typography>
        <PercentGradientMultiChart
          capacity={capacity}
          stamps={metrics.perCoreCpuSeries.stamps}
          valuesByKey={metrics.perCoreCpuSeries.valuesByKey}
        />
      </Stack>
    </Paper>
  );
};
