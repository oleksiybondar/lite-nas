import type { DiskMetricSnapshotDTO } from "@dto/monitoring/disk-metric";
import { MetricProvider } from "@providers/MetricProvider";
import {
  parseDiskMetricHistoryResponse,
  parseDiskMetricSnapshotResponse,
} from "@schemas/monitoring/disk-metric";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Metrics provider configured for gateway-backed disk telemetry.
 */
export const DiskMetricProvider = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <MetricProvider<DiskMetricSnapshotDTO>
      getTimestamp={(item) => item.timestamp}
      historyPath="/api/disk-metrics/history"
      parseHistoryResponse={parseDiskMetricHistoryResponse}
      parseSnapshotResponse={parseDiskMetricSnapshotResponse}
      snapshotPath="/api/disk-metrics/snapshot"
      storageKey="disk-metrics"
    >
      {children}
    </MetricProvider>
  );
};
