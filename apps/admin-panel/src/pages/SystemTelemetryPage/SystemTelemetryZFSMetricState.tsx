import { ZFSPoolCard } from "@components/monitoring/ZFSPoolCard";
import { buildZFSPoolCardData } from "@helpers/zfs-metric-chart";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useZFSMetric } from "@hooks/useZFSMetric";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { SupportedTelemetryRoute } from "@pages/SystemTelemetryPage/helpers";
import type { ReactElement } from "react";

type SystemTelemetryZFSMetricStateProps = {
  /**
   * Current route metadata resolved from the browser URL.
   */
  route: SupportedTelemetryRoute;
};

/**
 * Route state rendered when gateway-backed ZFS pool metrics are available.
 */
export const SystemTelemetryZFSMetricState = ({
  route,
}: SystemTelemetryZFSMetricStateProps): ReactElement => {
  const { items } = useZFSMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const poolCards = buildZFSPoolCardData(items);

  if (poolCards.length === 0) {
    return (
      <Typography data-testid="system-telemetry-zfs-empty" variant="body2">
        {route.summary}
      </Typography>
    );
  }

  return (
    <Stack spacing={3}>
      {poolCards.map((pool) => {
        return <ZFSPoolCard capacity={maxRecords} key={pool.name} pool={pool} />;
      })}
    </Stack>
  );
};
