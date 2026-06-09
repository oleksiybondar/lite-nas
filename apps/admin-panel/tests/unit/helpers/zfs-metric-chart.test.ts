import type { ZFSMetricSnapshotDTO } from "@dto/monitoring/zfs-metric";
import { buildZFSPoolCardData } from "@helpers/zfs-metric-chart";

const zfsMetricItems: ZFSMetricSnapshotDTO[] = [
  {
    Pools: [
      {
        Errors: "none",
        Health: "ONLINE",
        IOStat: {
          Bandwidth: { Read: 10, Write: 20 },
          Operations: { Read: 3, Write: 4 },
        },
        Name: "tank",
        Root: {
          Children: null,
          Errors: { Checksum: 1, Read: 2, Write: 3 },
          Name: "root",
          Path: "/dev/root",
          Type: "disk",
        },
        Scan: "scrub repaired 0B",
        Usage: {
          AllocatedBytes: 300,
          CapacityPct: 60,
          FreeBytes: 200,
          SizeBytes: 500,
        },
      },
    ],
    Timestamp: "2026-06-07T13:00:00Z",
  },
  {
    Pools: [
      {
        Errors: "none",
        Health: "ONLINE",
        IOStat: {
          Bandwidth: { Read: 12, Write: 24 },
          Operations: { Read: 5, Write: 6 },
        },
        Name: "tank",
        Root: {
          Children: null,
          Errors: { Checksum: 2, Read: 4, Write: 6 },
          Name: "root",
          Path: "/dev/root",
          Type: "disk",
        },
        Scan: "scrub repaired 0B",
        Usage: {
          AllocatedBytes: 320,
          CapacityPct: 64,
          FreeBytes: 180,
          SizeBytes: 500,
        },
      },
    ],
    Timestamp: "2026-06-07T13:00:01Z",
  },
] as const;

describe("zfs metric chart helpers", () => {
  test("builds one pool card per latest pool with metadata and chart series", () => {
    expect(buildZFSPoolCardData([...zfsMetricItems])).toEqual([
      {
        bandwidthSeries: {
          stamps: ["2026-06-07T13:00:00Z", "2026-06-07T13:00:01Z"],
          valuesByKey: {
            Read: [10, 12],
            Write: [20, 24],
          },
        },
        errorSeries: {
          stamps: ["2026-06-07T13:00:00Z", "2026-06-07T13:00:01Z"],
          valuesByKey: {
            Checksum: [1, 2],
            Read: [2, 4],
            Write: [3, 6],
          },
        },
        health: "ONLINE",
        healthColor: "success.main",
        metadataLabels: [
          { key: "Total", value: "500 B" },
          { key: "Available", value: "180 B" },
          { key: "Allocated", value: "320 B" },
        ],
        name: "tank",
        operationsSeries: {
          stamps: ["2026-06-07T13:00:00Z", "2026-06-07T13:00:01Z"],
          valuesByKey: {
            Read: [3, 5],
            Write: [4, 6],
          },
        },
        poolErrorSummary: "No known data errors",
        scan: "scrub repaired 0B",
        usedPercentColor: "warning.main",
        usedPercentLabel: "64%",
      },
    ]);
  });
});
