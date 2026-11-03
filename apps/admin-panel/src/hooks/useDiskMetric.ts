import type { DiskMetricSnapshotDTO } from "@dto/monitoring/disk-metric";
import type { MetricContextValue } from "@dto/monitoring/metric";
import { useMetric } from "@hooks/useMetric";

/**
 * Reads typed disk telemetry from the nearest disk metrics provider.
 */
export const useDiskMetric = (): MetricContextValue<DiskMetricSnapshotDTO> => {
  return useMetric<DiskMetricSnapshotDTO>();
};
