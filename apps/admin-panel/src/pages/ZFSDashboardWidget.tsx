import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { formatMetricBytesPerSecond, formatMetricValue } from "@helpers/metric-display";
import { buildZFSPoolCardData } from "@helpers/zfs-metric-chart";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useZFSMetric } from "@hooks/useZFSMetric";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Renders pool I/O or errors under the ZFS metrics provider. */
export const ZFSDashboardWidget = ({ kind }: { kind: "zfs-io" | "zfs-errors" }): ReactElement => {
  const { items } = useZFSMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const pools = buildZFSPoolCardData(items);
  const formatValue = kind === "zfs-io" ? formatMetricBytesPerSecond : formatMetricValue;
  return (
    <Stack spacing={1}>
      {pools.map((pool) => (
        <Stack key={pool.name} spacing={1}>
          <Typography variant="subtitle1">{pool.name}</Typography>
          <ValueLineChart
            capacity={maxRecords}
            formatValue={formatValue}
            stamps={kind === "zfs-io" ? pool.bandwidthSeries.stamps : pool.errorSeries.stamps}
            valuesByKey={
              kind === "zfs-io" ? pool.bandwidthSeries.valuesByKey : pool.errorSeries.valuesByKey
            }
          />
        </Stack>
      ))}
    </Stack>
  );
};
