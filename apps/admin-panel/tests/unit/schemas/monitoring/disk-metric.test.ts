import {
  parseDiskMetricHistoryResponse,
  parseDiskMetricSnapshotResponse,
} from "@schemas/monitoring/disk-metric";

const diskMetricSnapshotBody = {
  data: {
    devices: [
      {
        connection_type: "SATA",
        io: {
          io_in_progress: 0,
          io_time_ms: 10,
          reads_completed: 11,
          reads_merged: 12,
          read_time_ms: 13,
          sectors_read: 14,
          sectors_written: 15,
          weighted_io_time_ms: 16,
          writes_completed: 17,
          writes_merged: 18,
          write_time_ms: 19,
        },
        kind: "disk",
        major: 8,
        minor: 0,
        mounts: ["/"],
        node: "sda",
        size_bytes: 1000,
      },
    ],
    filesystems: {
      ext4: {
        mount_count: 1,
        type: "local_block",
      },
    },
    mounts: {
      "/": {
        filesystem: "ext4",
        memory_backed: false,
        mountpoint: "/",
        readonly: false,
        remote: false,
        source: "/dev/sda1",
      },
    },
    timestamp: "2026-06-15T12:00:00Z",
  },
  success: true,
  timestamp: "2026-06-15T12:00:01Z",
};

describe("disk metrics history schemas", () => {
  test("parses history envelopes and falls back null history data to an empty list", () => {
    expect(
      parseDiskMetricHistoryResponse({
        data: null,
        success: true,
        timestamp: "2026-06-15T12:00:00Z",
      }),
    ).toEqual([]);

    expect(
      parseDiskMetricHistoryResponse({
        ...diskMetricSnapshotBody,
        data: [diskMetricSnapshotBody.data],
      }),
    ).toHaveLength(1);
  });
});

describe("disk metrics snapshot schemas", () => {
  test("parses snapshot envelopes with nested maps and optional I/O counters", () => {
    expect(parseDiskMetricSnapshotResponse(diskMetricSnapshotBody)).toHaveProperty(
      "mounts./.source",
      "/dev/sda1",
    );
  });
});
