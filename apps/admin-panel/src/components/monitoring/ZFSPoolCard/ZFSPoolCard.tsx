import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { formatMetricBytesPerSecond, formatMetricValue } from "@helpers/metric-display";
import Divider from "@mui/material/Divider";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { ZFSPoolCardProps } from "./types";
import { ZFSPoolCardErrors } from "./ZFSPoolCardErrors";
import { ZFSPoolCardMetadata } from "./ZFSPoolCardMetadata";
import { ZFSPoolCardSummary } from "./ZFSPoolCardSummary";

const zfsPoolCardSectionSx = {
  flex: "1 1 320px",
  minWidth: 0,
  p: 1,
  width: "100%",
} as const;

/**
 * Large responsive card that groups one ZFS pool's metadata and I/O visualizations.
 */
export const ZFSPoolCard = ({ capacity, pool }: ZFSPoolCardProps): ReactElement => {
  return (
    <Paper data-test-class="zfs-pool-card" data-test-name={pool.name} sx={{ p: 2 }}>
      <Stack spacing={1}>
        <ZFSPoolCardMetadata pool={pool} />
        <Divider />
        <Stack
          data-test-class="zfs-pool-card-sections"
          direction="row"
          flexWrap="wrap"
          gap={0.625}
          useFlexGap
        >
          <Paper sx={zfsPoolCardSectionSx} variant="outlined">
            <Stack spacing={0.625}>
              <Typography
                data-test-class="zfs-pool-card-section-title"
                data-test-name="IO bandwidth"
                variant="body2"
              >
                IO bandwidth
              </Typography>
              <ValueLineChart
                capacity={capacity}
                formatValue={formatMetricBytesPerSecond}
                stamps={pool.bandwidthSeries.stamps}
                valuesByKey={pool.bandwidthSeries.valuesByKey}
              />
            </Stack>
          </Paper>
          <Paper sx={zfsPoolCardSectionSx} variant="outlined">
            <Stack spacing={0.625}>
              <Typography
                data-test-class="zfs-pool-card-section-title"
                data-test-name="IO operations"
                variant="body2"
              >
                IO operations
              </Typography>
              <ValueLineChart
                capacity={capacity}
                formatValue={formatMetricValue}
                stamps={pool.operationsSeries.stamps}
                valuesByKey={pool.operationsSeries.valuesByKey}
              />
            </Stack>
          </Paper>
          <Paper sx={zfsPoolCardSectionSx} variant="outlined">
            <ZFSPoolCardErrors capacity={capacity} pool={pool} />
          </Paper>
        </Stack>
        <Divider />
        <ZFSPoolCardSummary pool={pool} />
      </Stack>
    </Paper>
  );
};
