import type { MetricContextValue } from "@dto/monitoring/metric";
import type { SystemMetricSnapshotDTO } from "@dto/monitoring/system-metric";
import { useMetric } from "@hooks/useMetric";

/**
 * Reads typed system CPU and memory metrics from the nearest system metrics provider.
 */
export const useSystemMetric = (): MetricContextValue<SystemMetricSnapshotDTO> => {
  return useMetric<SystemMetricSnapshotDTO>();
};
