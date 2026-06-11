import type { MetricContextValue } from "@dto/monitoring/metric";
import type { ZFSMetricSnapshotDTO } from "@dto/monitoring/zfs-metric";
import { useMetric } from "@hooks/useMetric";

/**
 * Reads typed ZFS pool metrics from the nearest ZFS metrics provider.
 */
export const useZFSMetric = (): MetricContextValue<ZFSMetricSnapshotDTO> => {
  return useMetric<ZFSMetricSnapshotDTO>();
};
