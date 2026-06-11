import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { ZFSPoolCardProps } from "./types";

type ZFSPoolCardSummaryProps = {
  /**
   * Pool metadata rendered under the chart row as compact summary rows.
   */
  pool: ZFSPoolCardProps["pool"];
};

/**
 * Renders the scan and normalized error summary rows below the ZFS pool charts.
 */
export const ZFSPoolCardSummary = ({ pool }: ZFSPoolCardSummaryProps): ReactElement => {
  return (
    <Stack data-test-class="zfs-pool-card-summary" spacing={0.5}>
      <Typography data-test-class="zfs-pool-card-summary-row" data-test-name="Scan" variant="body2">
        Scan: {pool.scan}
      </Typography>
      <Typography
        data-test-class="zfs-pool-card-summary-row"
        data-test-name="Errors"
        variant="body2"
      >
        Errors: {pool.poolErrorSummary}
      </Typography>
    </Stack>
  );
};
