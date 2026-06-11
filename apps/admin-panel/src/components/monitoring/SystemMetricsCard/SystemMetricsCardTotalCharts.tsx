import { PercentGradientChart } from "@components/monitoring/PercentGradientChart";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { SystemMetricsCardProps } from "./types";

const systemMetricsCardSectionSx = {
  flex: "1 1 320px",
  minWidth: 0,
  p: 1,
  width: "100%",
} as const;

type SystemMetricsCardTotalChartsProps = {
  /**
   * Capacity of the charts.
   */
  capacity: number;
  /**
   * Metrics data containing the series for the charts.
   */
  metrics: SystemMetricsCardProps["metrics"];
};

/**
 * Renders the total CPU and RAM percent charts for the system metrics card.
 */
export const SystemMetricsCardTotalCharts = ({
  capacity,
  metrics,
}: SystemMetricsCardTotalChartsProps): ReactElement => {
  return (
    <Stack
      data-testid="system-metrics-total-charts"
      direction="row"
      flexWrap="wrap"
      gap={0.625}
      useFlexGap
    >
      <Paper sx={systemMetricsCardSectionSx} variant="outlined">
        <Stack spacing={0.625}>
          <Typography data-testid="system-metrics-total-cpu-chart-title" variant="body2">
            Total CPU %
          </Typography>
          <PercentGradientChart
            capacity={capacity}
            stamps={metrics.totalCpuSeries.stamps}
            values={metrics.totalCpuSeries.values}
          />
        </Stack>
      </Paper>
      <Paper sx={systemMetricsCardSectionSx} variant="outlined">
        <Stack spacing={0.625}>
          <Typography data-testid="system-metrics-ram-chart-title" variant="body2">
            RAM %
          </Typography>
          <PercentGradientChart
            capacity={capacity}
            stamps={metrics.ramSeries.stamps}
            values={metrics.ramSeries.values}
          />
        </Stack>
      </Paper>
    </Stack>
  );
};
