import {
  parseSystemMetricHistoryResponse,
  parseSystemMetricSnapshotResponse,
} from "@schemas/monitoring/system-metric";

const systemMetricSnapshotBody = {
  data: {
    CPU: {
      PerCoreUsage: null,
      TotalUsagePct: 15,
    },
    Mem: {
      TotalBytes: 1000,
      UsedBytes: 500,
      UsedPct: 50,
    },
    Timestamp: "2026-06-07T12:00:00Z",
  },
  success: true,
  timestamp: "2026-06-07T12:00:01Z",
};

describe("system metrics history schemas", () => {
  test("parses history envelopes and falls back null history data to an empty list", () => {
    expect(
      parseSystemMetricHistoryResponse({
        data: null,
        success: true,
        timestamp: "2026-06-07T12:00:00Z",
      }),
    ).toEqual([]);

    expect(
      parseSystemMetricHistoryResponse({
        data: [
          { ...systemMetricSnapshotBody.data, CPU: { PerCoreUsage: [10, 20], TotalUsagePct: 15 } },
        ],
        success: true,
        timestamp: "2026-06-07T12:00:01Z",
      }),
    ).toHaveLength(1);
  });
});

describe("system metrics snapshot schemas", () => {
  test("parses snapshot envelopes", () => {
    expect(parseSystemMetricSnapshotResponse(systemMetricSnapshotBody)).toEqual({
      CPU: {
        PerCoreUsage: null,
        TotalUsagePct: 15,
      },
      Mem: {
        TotalBytes: 1000,
        UsedBytes: 500,
        UsedPct: 50,
      },
      Timestamp: "2026-06-07T12:00:00Z",
    });
  });
});
