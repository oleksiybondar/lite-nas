import {
  formatMetricBytesPerSecond,
  formatZFSHealthLabel,
  resolveMetricPercentColor,
  resolveZFSHealthColor,
} from "@helpers/metric-display";

describe("metric display helpers", () => {
  test("formats byte rates using compact binary throughput units", () => {
    expect(formatMetricBytesPerSecond(999)).toBe("999 B/s");
    expect(formatMetricBytesPerSecond(1126.4)).toBe("1.1 KiB/s");
    expect(formatMetricBytesPerSecond(1_058_576.64)).toBe("1.01 MiB/s");
  });

  test("maps percent thresholds to semantic theme tokens", () => {
    expect(resolveMetricPercentColor(10)).toBe("info.main");
    expect(resolveMetricPercentColor(40)).toBe("success.main");
    expect(resolveMetricPercentColor(60)).toBe("warning.main");
    expect(resolveMetricPercentColor(90)).toBe("error.main");
  });

  test("maps zfs health values to semantic theme tokens", () => {
    expect(resolveZFSHealthColor("ONLINE")).toBe("success.main");
    expect(resolveZFSHealthColor("DEGRADED")).toBe("warning.main");
    expect(resolveZFSHealthColor("FAULTED")).toBe("error.main");
  });

  test("formats zfs health labels for display", () => {
    expect(formatZFSHealthLabel("ONLINE")).toBe("Online");
    expect(formatZFSHealthLabel("DEGRADED")).toBe("Degraded");
  });
});
