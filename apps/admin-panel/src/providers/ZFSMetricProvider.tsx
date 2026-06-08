import type { ZFSMetricSnapshotDTO } from "@dto/monitoring/zfs-metric";
import { MetricProvider } from "@providers/MetricProvider";
import {
  parseZFSMetricHistoryResponse,
  parseZFSMetricSnapshotResponse,
} from "@schemas/monitoring/zfs-metric";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Metrics provider configured for gateway-backed ZFS pool telemetry.
 */
export const ZFSMetricProvider = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <MetricProvider<ZFSMetricSnapshotDTO>
      getTimestamp={(item) => item.Timestamp}
      historyPath="/api/zfs-metrics/history"
      parseHistoryResponse={parseZFSMetricHistoryResponse}
      parseSnapshotResponse={parseZFSMetricSnapshotResponse}
      snapshotPath="/api/zfs-metrics/snapshot"
      storageKey="zfs-metrics"
    >
      {children}
    </MetricProvider>
  );
};
