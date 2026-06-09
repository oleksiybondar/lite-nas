import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { formatMetricValue } from "@helpers/metric-display";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { ZFSPoolCardProps } from "./types";

type ZFSPoolCardErrorsProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Pool card error labels rendered in the secondary responsive cell.
   */
  pool: ZFSPoolCardProps["pool"];
};

/**
 * Renders the error-history chart without duplicating summary metadata.
 */
export const ZFSPoolCardErrors = ({ capacity, pool }: ZFSPoolCardErrorsProps): ReactElement => {
  return (
    <Stack spacing={0.625}>
      <Typography
        data-test-class="zfs-pool-card-section-title"
        data-test-name="Errors"
        variant="subtitle1"
      >
        Errors
      </Typography>
      <ValueLineChart
        capacity={capacity}
        formatValue={formatMetricValue}
        stamps={pool.errorSeries.stamps}
        valuesByKey={pool.errorSeries.valuesByKey}
      />
    </Stack>
  );
};
