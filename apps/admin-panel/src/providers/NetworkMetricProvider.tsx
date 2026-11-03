import type { NetworkMetricSnapshotDTO } from "@dto/monitoring/network-metric";
import { MetricProvider } from "@providers/MetricProvider";
import {
  parseNetworkMetricHistoryResponse,
  parseNetworkMetricSnapshotResponse,
} from "@schemas/monitoring/network-metric";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Metrics provider configured for gateway-backed network telemetry.
 */
export const NetworkMetricProvider = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <MetricProvider<NetworkMetricSnapshotDTO>
      getTimestamp={(item) => item.timestamp}
      historyPath="/api/network-metrics/history"
      parseHistoryResponse={parseNetworkMetricHistoryResponse}
      parseSnapshotResponse={parseNetworkMetricSnapshotResponse}
      snapshotPath="/api/network-metrics/snapshot"
      storageKey="network-metrics"
    >
      {children}
    </MetricProvider>
  );
};
