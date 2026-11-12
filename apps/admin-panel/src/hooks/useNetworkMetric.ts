import type { MetricContextValue } from "@dto/monitoring/metric";
import type { NetworkMetricSnapshotDTO } from "@dto/monitoring/network-metric";
import { useMetric } from "@hooks/useMetric";

/**
 * Reads typed network telemetry from the nearest network metrics provider.
 */
export const useNetworkMetric = (): MetricContextValue<NetworkMetricSnapshotDTO> => {
  return useMetric<NetworkMetricSnapshotDTO>();
};
