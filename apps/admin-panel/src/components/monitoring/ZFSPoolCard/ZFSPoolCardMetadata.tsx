import { formatZFSHealthLabel } from "@helpers/metric-display";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { ZFSPoolCardProps } from "./types";

type ZFSPoolCardMetadataProps = {
  /**
   * Pool card metadata rendered above the charts.
   */
  pool: ZFSPoolCardProps["pool"];
};

/**
 * Renders the pool-level status and capacity metadata for one ZFS pool card.
 */
export const ZFSPoolCardMetadata = ({ pool }: ZFSPoolCardMetadataProps): ReactElement => {
  return (
    <Stack spacing={1}>
      <Stack alignItems="center" direction="row" justifyContent="space-between" spacing={1}>
        <Stack alignItems="center" direction="row" flexWrap="wrap" gap={1} useFlexGap>
          <Typography data-test-class="zfs-pool-card-name" data-test-name={pool.name} variant="h4">
            {pool.name}
          </Typography>
          <Typography
            data-test-class="zfs-pool-card-health"
            data-test-name={pool.health}
            sx={{ color: pool.healthColor, fontWeight: 600 }}
            variant="h5"
          >
            {formatZFSHealthLabel(pool.health)}
          </Typography>
        </Stack>
        <Typography
          data-test-class="zfs-pool-card-used-percent"
          data-test-name={pool.name}
          sx={{ color: pool.usedPercentColor, fontWeight: 700, whiteSpace: "nowrap" }}
          variant="h5"
        >
          Used {pool.usedPercentLabel}
        </Typography>
      </Stack>
      <Stack
        data-test-class="zfs-pool-card-metadata"
        direction="row"
        flexWrap="wrap"
        gap={1}
        useFlexGap
      >
        {pool.metadataLabels.map((label) => {
          return (
            <Typography
              data-test-class="zfs-pool-card-metadata-label"
              data-test-name={label.key}
              key={label.key}
              variant="body2"
            >
              {label.key}: {label.value}
            </Typography>
          );
        })}
      </Stack>
    </Stack>
  );
};
