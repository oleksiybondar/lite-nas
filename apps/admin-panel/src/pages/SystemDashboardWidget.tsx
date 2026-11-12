import { PercentGradientChart } from "@components/monitoring/PercentGradientChart";
import { PercentGradientMultiChart } from "@components/monitoring/PercentGradientMultiChart";
import { buildSystemMetricsCardData } from "@helpers/system-metric-chart";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useSystemMetric } from "@hooks/useSystemMetric";
import type { ReactElement } from "react";

/** Renders one system chart under its matching metrics provider. */
export const SystemDashboardWidget = ({
  kind,
}: {
  kind: "cpu" | "cpu-cores" | "ram";
}): ReactElement => {
  const { items } = useSystemMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const metrics = buildSystemMetricsCardData(items.filter((item) => item.CPU && item.Mem));
  if (kind === "cpu-cores")
    return (
      <PercentGradientMultiChart
        capacity={maxRecords}
        stamps={metrics.perCoreCpuSeries.stamps}
        valuesByKey={metrics.perCoreCpuSeries.valuesByKey}
      />
    );
  const series = kind === "cpu" ? metrics.totalCpuSeries : metrics.ramSeries;
  return (
    <PercentGradientChart capacity={maxRecords} stamps={series.stamps} values={series.values} />
  );
};
