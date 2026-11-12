import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { buildDiskDevicePanelItems } from "@helpers/disk-metric-panel";
import { formatMetricBytesPerSecond } from "@helpers/metric-display";
import { useDiskMetric } from "@hooks/useDiskMetric";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Shows read and write throughput for each primary disk in the dashboard. */
export const DiskDashboardWidget = (): ReactElement => {
  const { items } = useDiskMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  return (
    <Stack spacing={1}>
      {buildDiskDevicePanelItems(items).map((device) => (
        <Stack key={device.name} spacing={1}>
          <Typography variant="subtitle1">{device.name}</Typography>
          <ValueLineChart
            capacity={maxRecords}
            formatValue={formatMetricBytesPerSecond}
            stamps={device.readWriteSeries.stamps}
            valuesByKey={device.readWriteSeries.valuesByKey}
          />
        </Stack>
      ))}
    </Stack>
  );
};
