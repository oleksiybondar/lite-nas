import { useZFSMetric } from "@hooks/useZFSMetric";
import Paper from "@mui/material/Paper";
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
  const { error, isError, isFetching, isLoading, items, latestItem, mode } = useZFSMetric();

  return (
    <Paper data-testid="system-telemetry-metric-card" sx={{ p: 3 }}>
      <Stack spacing={1.5}>
        <Typography data-testid="system-telemetry-metric-title" variant="h2">
          ZFS metrics
        </Typography>
        <Typography
          color="text.secondary"
          data-testid="system-telemetry-metric-summary"
          variant="body2"
        >
          {route.summary}
        </Typography>
        <Typography data-testid="system-telemetry-metric-mode" variant="body2">
          Polling mode: {mode}
        </Typography>
        <Typography data-testid="system-telemetry-metric-points" variant="body2">
          Cached points: {items.length}
        </Typography>
        <Typography data-testid="system-telemetry-metric-latest" variant="body2">
          Latest timestamp: {latestItem?.Timestamp ?? "None yet"}
        </Typography>
        <Typography data-testid="system-telemetry-metric-loading" variant="body2">
          Loading: {String(isLoading)}
        </Typography>
        <Typography data-testid="system-telemetry-metric-fetching" variant="body2">
          Fetching: {String(isFetching)}
        </Typography>
        <Typography data-testid="system-telemetry-metric-error" variant="body2">
          Error: {isError ? (error?.message ?? "Unknown error") : "None"}
        </Typography>
      </Stack>
    </Paper>
  );
};
