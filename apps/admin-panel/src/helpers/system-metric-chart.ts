import type { SystemMetricSnapshotDTO } from "@dto/monitoring/system-metric";

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
 * Browser-facing chart data and summary labels consumed by the RAM card.
 */
export type MetricChartCardData = {
  labels: MetricChartLabel[];
  series: MetricChartSeries;
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
 * Converts system metrics snapshots into the RAM usage percent chart and summary labels.
 */
export const toSystemMemoryUsageChartCardData = (
  items: SystemMetricSnapshotDTO[],
): MetricChartCardData => {
  const latestItem = items.length === 0 ? null : items[items.length - 1];

  return {
    labels: [
      {
        key: "Total",
        value: latestItem === null ? "N/A" : formatMetricBytes(latestItem.Mem.TotalBytes),
      },
      {
        key: "Used",
        value: latestItem === null ? "N/A" : formatMetricBytes(latestItem.Mem.UsedBytes),
      },
    ],
    series: {
      stamps: items.map((item) => item.Timestamp),
      values: items.map((item) => item.Mem.UsedPct),
    },
  };
};

/**
 * Formats one byte count into the compact binary unit labels used on telemetry cards.
 */
const formatMetricBytes = (value: number): string => {
  const units = ["B", "KiB", "MiB", "GiB", "TiB"] as const;
  let currentValue = value;
  let unitIndex = 0;

  while (currentValue >= 1024 && unitIndex < units.length - 1) {
    currentValue /= 1024;
    unitIndex += 1;
  }

  const roundedValue =
    currentValue >= 10 || unitIndex === 0 ? currentValue.toFixed(0) : currentValue.toFixed(1);

  return `${roundedValue} ${units[unitIndex]}`;
};
