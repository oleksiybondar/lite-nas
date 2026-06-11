import Divider from "@mui/material/Divider";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";
import { SystemMetricsCardMetadata } from "./SystemMetricsCardMetadata";
import { SystemMetricsCardPerCoreChart } from "./SystemMetricsCardPerCoreChart";
import { SystemMetricsCardTotalCharts } from "./SystemMetricsCardTotalCharts";
import type { SystemMetricsCardProps } from "./types";

/**
 * Large responsive card that groups system CPU and memory visualizations.
 */
export const SystemMetricsCard = ({ capacity, metrics }: SystemMetricsCardProps): ReactElement => {
  return (
    <Paper data-testid="system-metrics-card" sx={{ p: 2 }}>
      <Stack spacing={1}>
        <SystemMetricsCardMetadata metrics={metrics} />
        <Divider />
        <SystemMetricsCardTotalCharts capacity={capacity} metrics={metrics} />
        <Divider />
        <SystemMetricsCardPerCoreChart capacity={capacity} metrics={metrics} />
      </Stack>
    </Paper>
  );
};
