import { SystemMetricsCard } from "@components/monitoring/SystemMetricsCard";
import { buildSystemMetricsCardData } from "@helpers/system-metric-chart";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useSystemMetric } from "@hooks/useSystemMetric";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Route state rendered when gateway-backed system CPU and memory metrics are available.
 */
export const SystemTelemetrySystemMetricState = (): ReactElement => {
  const { items } = useSystemMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const metrics = buildSystemMetricsCardData(items);

  return (
    <Stack spacing={3}>
      <SystemMetricsCard capacity={maxRecords} metrics={metrics} />
    </Stack>
  );
};
