import {
  toSystemMemoryUsageChartCardData,
  toSystemPerCoreCpuChartSeries,
  toSystemTotalCpuChartSeries,
} from "@helpers/system-metric-chart";

const systemMetricItems = [
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

  test("builds RAM chart data and latest summary labels", () => {
    expect(toSystemMemoryUsageChartCardData([...systemMetricItems])).toEqual({
      labels: [
        { key: "Total", value: "1000 B" },
        { key: "Used", value: "500 B" },
      ],
      series: {
        stamps: ["2026-06-07T12:00:00Z", "2026-06-07T12:00:01Z"],
        values: [40, 50],
      },
    });
  });
});
