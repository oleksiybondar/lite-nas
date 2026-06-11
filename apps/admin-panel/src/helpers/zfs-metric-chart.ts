import type { ZFSMetricPoolSnapshotDTO, ZFSMetricSnapshotDTO } from "@dto/monitoring/zfs-metric";
import {
  formatMetricBytes,
  formatZFSPoolErrorSummary,
  resolveMetricPercentColor,
  resolveZFSHealthColor,
} from "@helpers/metric-display";
import type { MetricChartLabel, MetricMultiChartSeries } from "@helpers/system-metric-chart";

/**
 * Browser-facing data consumed by one ZFS pool card.
 */
export type ZFSPoolCardData = {
  bandwidthSeries: MetricMultiChartSeries;
  errorSeries: MetricMultiChartSeries;
  health: string;
  healthColor: string;
  metadataLabels: MetricChartLabel[];
  name: string;
  operationsSeries: MetricMultiChartSeries;
  poolErrorSummary: string;
  scan: string;
  usedPercentColor: string;
  usedPercentLabel: string;
};

/**
 * Builds one browser-facing pool card per latest known ZFS pool.
 */
export const buildZFSPoolCardData = (items: ZFSMetricSnapshotDTO[]): ZFSPoolCardData[] => {
  const latestItem = items.length === 0 ? null : items[items.length - 1];
  const latestPools = latestItem?.Pools ?? [];

  return latestPools.map((pool) => {
    return buildSingleZFSPoolCardData(items, pool);
  });
};

/**
 * Builds one pool-card view model from the full ZFS metric history.
 */
const buildSingleZFSPoolCardData = (
  items: ZFSMetricSnapshotDTO[],
  latestPool: ZFSMetricPoolSnapshotDTO,
): ZFSPoolCardData => {
  const poolHistory = items.map((item) => {
    return item.Pools?.find((pool) => pool.Name === latestPool.Name) ?? null;
  });

  return {
    bandwidthSeries: {
      stamps: items.map((item) => item.Timestamp),
      valuesByKey: {
        Read: poolHistory.map((pool) => pool?.IOStat.Bandwidth.Read ?? 0),
        Write: poolHistory.map((pool) => pool?.IOStat.Bandwidth.Write ?? 0),
      },
    },
    errorSeries: {
      stamps: items.map((item) => item.Timestamp),
      valuesByKey: {
        Checksum: poolHistory.map((pool) => pool?.Root.Errors.Checksum ?? 0),
        Read: poolHistory.map((pool) => pool?.Root.Errors.Read ?? 0),
        Write: poolHistory.map((pool) => pool?.Root.Errors.Write ?? 0),
      },
    },
    health: latestPool.Health,
    healthColor: resolveZFSHealthColor(latestPool.Health),
    metadataLabels: [
      { key: "Total", value: formatMetricBytes(latestPool.Usage.SizeBytes) },
      { key: "Available", value: formatMetricBytes(latestPool.Usage.FreeBytes) },
      { key: "Allocated", value: formatMetricBytes(latestPool.Usage.AllocatedBytes) },
    ],
    name: latestPool.Name,
    operationsSeries: {
      stamps: items.map((item) => item.Timestamp),
      valuesByKey: {
        Read: poolHistory.map((pool) => pool?.IOStat.Operations.Read ?? 0),
        Write: poolHistory.map((pool) => pool?.IOStat.Operations.Write ?? 0),
      },
    },
    poolErrorSummary: formatZFSPoolErrorSummary(latestPool.Errors),
    scan: latestPool.Scan,
    usedPercentColor: resolveMetricPercentColor(latestPool.Usage.CapacityPct),
    usedPercentLabel: `${Math.round(latestPool.Usage.CapacityPct)}%`,
  };
};
