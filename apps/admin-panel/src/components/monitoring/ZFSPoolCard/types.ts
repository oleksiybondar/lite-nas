import type { ZFSPoolCardData } from "@helpers/zfs-metric-chart";

/**
 * Input contract accepted by the ZFS pool card component.
 */
export type ZFSPoolCardProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Browser-facing pool card data built from ZFS metric history.
   */
  pool: ZFSPoolCardData;
};
