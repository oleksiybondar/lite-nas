import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { formatMetricBytesPerSecond } from "@helpers/metric-display";
import { buildNetworkInterfacePanelItems } from "@helpers/network-metric-panel";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useNetworkMetric } from "@hooks/useNetworkMetric";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Renders interface throughput under the network metrics provider. */
export const NetworkDashboardWidget = (): ReactElement => {
  const { items } = useNetworkMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const interfaces = buildNetworkInterfacePanelItems(items);
  return (
    <Stack spacing={1}>
      {interfaces.map((item) => (
        <Stack key={item.name} spacing={1}>
          <Typography variant="subtitle1">{item.name}</Typography>
          <ValueLineChart
            capacity={maxRecords}
            formatValue={formatMetricBytesPerSecond}
            stamps={item.throughputSeries.stamps}
            valuesByKey={item.throughputSeries.valuesByKey}
          />
        </Stack>
      ))}
    </Stack>
  );
};
