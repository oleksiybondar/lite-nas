import type { SystemMetricSnapshotDTO } from "@dto/monitoring/system-metric";
import {
  buildSystemMetricsCardData,
  toSystemPerCoreCpuChartSeries,
  toSystemTotalCpuChartSeries,
} from "@helpers/system-metric-chart";

const systemMetricItems: SystemMetricSnapshotDTO[] = [
  {
    CPU: { PerCoreUsage: [10, 20], TotalUsagePct: 15 },
    Mem: { TotalBytes: 1000, UsedBytes: 400, UsedPct: 40 },
    Timestamp: "2026-06-07T12:00:00Z",
  },
  {
    CPU: { PerCoreUsage: [30, 40], TotalUsagePct: 35 },
    Mem: { TotalBytes: 1000, UsedBytes: 500, UsedPct: 50 },
    Timestamp: "2026-06-07T12:00:01Z",
  },
] as const;

describe("system metric chart helpers", () => {
  test("builds the total cpu series", () => {
    expect(toSystemTotalCpuChartSeries([...systemMetricItems])).toEqual({
      stamps: ["2026-06-07T12:00:00Z", "2026-06-07T12:00:01Z"],
      values: [15, 35],
    });
  });

  test("builds keyed per-core cpu series", () => {
    expect(toSystemPerCoreCpuChartSeries([...systemMetricItems])).toEqual({
      stamps: ["2026-06-07T12:00:00Z", "2026-06-07T12:00:01Z"],
      valuesByKey: {
        CPU1: [10, 30],
        CPU2: [20, 40],
      },
    });
  });

  test("builds unified system metrics card data", () => {
    expect(buildSystemMetricsCardData([...systemMetricItems])).toEqual({
      availableRam: "500 B",
      cpuUsageColor: "success.main",
      cpuUsageLabel: "35.0%",
      cpuUsagePct: 35,
      perCoreCpuSeries: {
        stamps: ["2026-06-07T12:00:00Z", "2026-06-07T12:00:01Z"],
        valuesByKey: {
          CPU1: [10, 30],
          CPU2: [20, 40],
        },
      },
      ramSeries: {
        stamps: ["2026-06-07T12:00:00Z", "2026-06-07T12:00:01Z"],
        values: [40, 50],
      },
      ramUsageColor: "warning.main",
      ramUsageLabel: "50.0%",
      ramUsagePct: 50,
      totalCores: 2,
      totalCpuSeries: {
        stamps: ["2026-06-07T12:00:00Z", "2026-06-07T12:00:01Z"],
        values: [15, 35],
      },
      totalRam: "1000 B",
      usedRam: "500 B",
    });
  });
});
