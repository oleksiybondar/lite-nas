import type { SystemMetricSnapshotDTO } from "@dto/monitoring/system-metric";
import { formatMetricBytes, resolveMetricPercentColor } from "@helpers/metric-display";

/**
 * Browser-facing chart arrays consumed by simple SVG metric visualizations.
 */
export type MetricChartSeries = {
  stamps: string[];
  values: number[];
};

/**
 * Browser-facing keyed chart arrays consumed by multi-series SVG metric visualizations.
 */
export type MetricMultiChartSeries = {
  stamps: string[];
  valuesByKey: Record<string, number[]>;
};

/**
 * One human-readable key/value label rendered beside one metric chart.
 */
export type MetricChartLabel = {
  key: string;
  value: string;
};

/**
 * Converts system metrics snapshots into the total CPU percent chart series.
 */
export const toSystemTotalCpuChartSeries = (
  items: SystemMetricSnapshotDTO[],
): MetricChartSeries => {
  return {
    stamps: items.map((item) => item.Timestamp),
    values: items.map((item) => item.CPU.TotalUsagePct),
  };
};

/**
 * Converts system metrics snapshots into one keyed percent series per CPU core.
 */
export const toSystemPerCoreCpuChartSeries = (
  items: SystemMetricSnapshotDTO[],
): MetricMultiChartSeries => {
  const stamps = items.map((item) => item.Timestamp);
  const coreCount = items.reduce((maxCoreCount, item) => {
    return Math.max(maxCoreCount, item.CPU.PerCoreUsage.length);
  }, 0);
  const valuesByKey = Object.fromEntries(
    Array.from({ length: coreCount }, (_, index) => {
      return [
        `CPU${index + 1}`,
        items.map((item) => {
          return item.CPU.PerCoreUsage[index] ?? 0;
        }),
      ];
    }),
  );

  return {
    stamps,
    valuesByKey,
  };
};

/**
 * Converts system metrics snapshots into the RAM percent chart series.
 */
export const toSystemRamChartSeries = (items: SystemMetricSnapshotDTO[]): MetricChartSeries => ({
  stamps: items.map((item) => item.Timestamp),
  values: items.map((item) => item.Mem.UsedPct),
});

/**
 * Browser-facing summary and series data consumed by the system metrics card.
 */
export type SystemMetricsCardData = {
  cpuUsageColor: string;
  cpuUsageLabel: string;
  cpuUsagePct: number;
  ramUsageColor: string;
  ramUsageLabel: string;
  ramUsagePct: number;
  totalCores: number;
  totalRam: string;
  usedRam: string;
  availableRam: string;
  totalCpuSeries: MetricChartSeries;
  ramSeries: MetricChartSeries;
  perCoreCpuSeries: MetricMultiChartSeries;
};

/**
 * Calculates the maximum core count seen across all provided metric snapshots.
 */
const calculateMaxCoreCount = (items: SystemMetricSnapshotDTO[]): number => {
  return items.reduce((maxCoreCount, item) => {
    return Math.max(maxCoreCount, item.CPU.PerCoreUsage?.length ?? 0);
  }, 0);
};

/**
 * Converts system metrics snapshots into the unified data structure used by the system metrics card.
 */
export const buildSystemMetricsCardData = (
  items: SystemMetricSnapshotDTO[],
): SystemMetricsCardData => {
  const latestItem = items[items.length - 1] ?? null;

  if (latestItem === null) {
    return buildEmptySystemMetricsCardData(items);
  }

  const cpuUsagePct = latestItem.CPU.TotalUsagePct;
  const ramUsagePct = latestItem.Mem.UsedPct;

  return {
    cpuUsageColor: resolveMetricPercentColor(cpuUsagePct),
    cpuUsageLabel: `${cpuUsagePct.toFixed(1)}%`,
    cpuUsagePct,
    ramUsageColor: resolveMetricPercentColor(ramUsagePct),
    ramUsageLabel: `${ramUsagePct.toFixed(1)}%`,
    ramUsagePct,
    totalCores: calculateMaxCoreCount(items),
    totalRam: formatMetricBytes(latestItem.Mem.TotalBytes),
    usedRam: formatMetricBytes(latestItem.Mem.UsedBytes),
    availableRam: formatMetricBytes(latestItem.Mem.TotalBytes - latestItem.Mem.UsedBytes),
    totalCpuSeries: toSystemTotalCpuChartSeries(items),
    ramSeries: toSystemRamChartSeries(items),
    perCoreCpuSeries: toSystemPerCoreCpuChartSeries(items),
  };
};

/**
 * Provides a default empty state for the system metrics card data.
 */
const buildEmptySystemMetricsCardData = (
  items: SystemMetricSnapshotDTO[],
): SystemMetricsCardData => ({
  cpuUsageColor: resolveMetricPercentColor(0),
  cpuUsageLabel: "0.0%",
  cpuUsagePct: 0,
  ramUsageColor: resolveMetricPercentColor(0),
  ramUsageLabel: "0.0%",
  ramUsagePct: 0,
  totalCores: calculateMaxCoreCount(items),
  totalRam: "N/A",
  usedRam: "N/A",
  availableRam: "N/A",
  totalCpuSeries: toSystemTotalCpuChartSeries(items),
  ramSeries: toSystemRamChartSeries(items),
  perCoreCpuSeries: toSystemPerCoreCpuChartSeries(items),
});
