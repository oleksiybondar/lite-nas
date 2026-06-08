import type { SystemMetricSnapshotDTO } from "@dto/monitoring/system-metric";
import { MetricProvider } from "@providers/MetricProvider";
import {
  parseSystemMetricHistoryResponse,
  parseSystemMetricSnapshotResponse,
} from "@schemas/monitoring/system-metric";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Metrics provider configured for gateway-backed system CPU and memory telemetry.
 */
export const SystemMetricProvider = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <MetricProvider<SystemMetricSnapshotDTO>
      getTimestamp={(item) => item.Timestamp}
      historyPath="/api/system-metrics/history"
      parseHistoryResponse={parseSystemMetricHistoryResponse}
      parseSnapshotResponse={parseSystemMetricSnapshotResponse}
      snapshotPath="/api/system-metrics/snapshot"
      storageKey="system-metrics"
    >
      {children}
    </MetricProvider>
  );
};
