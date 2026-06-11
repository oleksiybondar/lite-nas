import {
  parseZFSMetricHistoryResponse,
  parseZFSMetricSnapshotResponse,
} from "@schemas/monitoring/zfs-metric";

const zfsMetricSnapshotBody = {
  data: {
    Pools: [
      {
        Errors: "none",
        Health: "ONLINE",
        IOStat: {
          Bandwidth: { Read: 1, Write: 2 },
          Operations: { Read: 3, Write: 4 },
        },
        Name: "tank",
        Root: {
          Children: [
            {
              Children: null,
              Errors: { Checksum: 0, Read: 0, Write: 0 },
              Name: "sda",
              Path: "/dev/sda",
              Type: "disk",
            },
          ],
          Errors: { Checksum: 0, Read: 0, Write: 0 },
          Name: "mirror-0",
          Path: "",
          Type: "mirror",
        },
        Scan: "none requested",
        Usage: {
          AllocatedBytes: 100,
          CapacityPct: 10,
          FreeBytes: 900,
          SizeBytes: 1000,
        },
      },
    ],
    Timestamp: "2026-06-07T12:00:00Z",
  },
  success: true,
  timestamp: "2026-06-07T12:00:01Z",
};

describe("zfs metrics history schemas", () => {
  test("parses history envelopes and falls back null history data to an empty list", () => {
    expect(
      parseZFSMetricHistoryResponse({
        data: null,
        success: true,
        timestamp: "2026-06-07T12:00:00Z",
      }),
    ).toEqual([]);
  });
});

describe("zfs metrics snapshot schemas", () => {
  test("parses snapshot envelopes with recursive vdev children", () => {
    expect(parseZFSMetricSnapshotResponse(zfsMetricSnapshotBody)).toHaveProperty(
      "Pools.0.Root.Children.0.Name",
      "sda",
    );
  });
});
