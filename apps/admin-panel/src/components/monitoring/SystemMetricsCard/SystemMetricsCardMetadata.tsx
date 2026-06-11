import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { SystemMetricsCardProps } from "./types";

type SystemMetricsCardMetadataProps = {
  /**
   * System metrics metadata rendered above the charts.
   */
  metrics: SystemMetricsCardProps["metrics"];
};

/**
 * Renders the top-level CPU and memory metadata for the system metrics card.
 */
export const SystemMetricsCardMetadata = ({
  metrics,
}: SystemMetricsCardMetadataProps): ReactElement => {
  return (
    <Stack spacing={1}>
      {/* Row 1: CPU X% RAM Y% */}
      <Stack alignItems="center" direction="row" justifyContent="space-between" spacing={1}>
        <Typography
          data-testid="system-metrics-cpu-percent"
          sx={{ color: metrics.cpuUsageColor, fontWeight: 700 }}
          variant="h5"
        >
          CPU {metrics.cpuUsageLabel}
        </Typography>
        <Typography
          data-testid="system-metrics-ram-percent"
          sx={{ color: metrics.ramUsageColor, fontWeight: 700 }}
          variant="h5"
        >
          RAM {metrics.ramUsageLabel}
        </Typography>
      </Stack>

      {/* Row 2: Total cores */}
      <Typography data-testid="system-metrics-total-cores" variant="body2">
        Total cores: {metrics.totalCores}
      </Typography>

      {/* Row 3: Total RAM, Used RAM, Available RAM */}
      <Stack
        data-testid="system-metrics-ram-details"
        direction="row"
        flexWrap="wrap"
        gap={1.5}
        useFlexGap
      >
        <Typography
          data-test-class="system-metrics-ram-detail"
          data-test-name="total"
          variant="body2"
        >
          Total RAM: {metrics.totalRam}
        </Typography>
        <Typography
          data-test-class="system-metrics-ram-detail"
          data-test-name="used"
          variant="body2"
        >
          Used RAM: {metrics.usedRam}
        </Typography>
        <Typography
          data-test-class="system-metrics-ram-detail"
          data-test-name="available"
          variant="body2"
        >
          Available RAM: {metrics.availableRam}
        </Typography>
      </Stack>
    </Stack>
  );
};
