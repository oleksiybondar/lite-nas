import type {
  DiskMetricDeviceIOCountersDTO,
  DiskMetricDeviceSnapshotDTO,
  DiskMetricMountSnapshotDTO,
  DiskMetricSnapshotDTO,
} from "@dto/monitoring/disk-metric";
import {
  buildDiskAuxiliaryDevicesPanelData,
  buildDiskDevicePanelItems,
  buildDiskMountsPanelData,
} from "@helpers/disk-metric-panel";

const diskMetricItems: DiskMetricSnapshotDTO[] = [
  buildDiskSnapshot({
    devices: [
      buildTopLevelDiskDevice({
        io: buildDiskIoCounters({
          io_in_progress: 0,
          io_time_ms: 100,
          read_time_ms: 50,
          reads_completed: 100,
          sectors_read: 2000,
          sectors_written: 4000,
          weighted_io_time_ms: 100,
          write_time_ms: 40,
          writes_completed: 200,
        }),
      }),
      buildLoopDevice({ mounts: ["/snap/core"] }),
      buildPartitionedLoopDevice({
        io: buildDiskIoCounters({
          io_in_progress: 1,
          io_time_ms: 60,
          read_time_ms: 10,
          reads_completed: 20,
          sectors_read: 800,
          sectors_written: 1600,
          weighted_io_time_ms: 60,
          write_time_ms: 12,
          writes_completed: 30,
        }),
      }),
      buildPartitionDevice({
        io: buildDiskIoCounters({
          io_in_progress: 0,
          io_time_ms: 80,
          read_time_ms: 50,
          reads_completed: 100,
          sectors_read: 2000,
          sectors_written: 4000,
          weighted_io_time_ms: 80,
          write_time_ms: 40,
          writes_completed: 200,
        }),
      }),
    ],
    mounts: {
      "/": buildDiskMount({
        available_bytes: 500 * 1024 ** 2,
        device: "sda1",
        mountpoint: "/",
        source: "/dev/sda1",
        total_bytes: 900 * 1024 ** 2,
        used_bytes: 400 * 1024 ** 2,
        used_percent: 44.4,
      }),
      "/snap/core": buildDiskMount({
        device: "loop0",
        filesystem: "squashfs",
        mountpoint: "/snap/core",
        readonly: true,
        source: "/dev/loop0",
        total_bytes: 1000,
        used_bytes: 1000,
        used_percent: 100,
      }),
      "/mnt/archive": buildDiskMount({
        available_bytes: 160 * 1024 ** 2,
        device: "loop7p1",
        mountpoint: "/mnt/archive",
        source: "/dev/loop7p1",
        total_bytes: 256 * 1024 ** 2,
        used_bytes: 96 * 1024 ** 2,
        used_percent: 37.5,
      }),
      "/tank": buildDiskMount({
        available_bytes: 1500 * 1024 ** 2,
        device: null,
        filesystem: "zfs",
        mountpoint: "/tank",
        source: "tank",
        total_bytes: 2 * 1024 ** 3,
        used_bytes: 512 * 1024 ** 2,
        used_percent: 25,
      }),
    },
    timestamp: "2026-06-16T10:57:17.000Z",
  }),
  buildDiskSnapshot({
    devices: [
      buildTopLevelDiskDevice({
        io: buildDiskIoCounters({
          io_in_progress: 2,
          io_time_ms: 350,
          read_time_ms: 55,
          reads_completed: 120,
          sectors_read: 2600,
          sectors_written: 5200,
          weighted_io_time_ms: 350,
          write_time_ms: 50,
          writes_completed: 260,
        }),
      }),
      buildLoopDevice({ mounts: ["/snap/core"] }),
      buildPartitionedLoopDevice({
        io: buildDiskIoCounters({
          io_in_progress: 0,
          io_time_ms: 140,
          read_time_ms: 12,
          reads_completed: 44,
          sectors_read: 1200,
          sectors_written: 2200,
          weighted_io_time_ms: 140,
          write_time_ms: 16,
          writes_completed: 60,
        }),
      }),
      buildPartitionDevice({
        io: buildDiskIoCounters({
          io_in_progress: 0,
          io_time_ms: 120,
          read_time_ms: 55,
          reads_completed: 120,
          sectors_read: 2600,
          sectors_written: 5200,
          weighted_io_time_ms: 120,
          write_time_ms: 50,
          writes_completed: 260,
        }),
      }),
    ],
    mounts: {
      "/": buildDiskMount({
        available_bytes: 450 * 1024 ** 2,
        device: "sda1",
        mountpoint: "/",
        source: "/dev/sda1",
        total_bytes: 900 * 1024 ** 2,
        used_bytes: 450 * 1024 ** 2,
        used_percent: 50,
      }),
      "/snap/core": buildDiskMount({
        device: "loop0",
        filesystem: "squashfs",
        mountpoint: "/snap/core",
        readonly: true,
        source: "/dev/loop0",
        total_bytes: 1000,
        used_bytes: 1000,
        used_percent: 100,
      }),
      "/mnt/archive": buildDiskMount({
        available_bytes: 128 * 1024 ** 2,
        device: "loop7p1",
        mountpoint: "/mnt/archive",
        source: "/dev/loop7p1",
        total_bytes: 256 * 1024 ** 2,
        used_bytes: 128 * 1024 ** 2,
        used_percent: 50,
      }),
      "/tank": buildDiskMount({
        available_bytes: 1400 * 1024 ** 2,
        device: null,
        filesystem: "zfs",
        mountpoint: "/tank",
        source: "tank",
        total_bytes: 2 * 1024 ** 3,
        used_bytes: 648 * 1024 ** 2,
        used_percent: 31.6,
      }),
    },
    timestamp: "2026-06-16T10:57:18.000Z",
  }),
];

test("builds disk device cards from top-level disks and derives interval charts", () => {
  const items = buildDiskDevicePanelItems(diskMetricItems);

  expect(items).toHaveLength(2);

  const firstItem = items.find((item) => item.name === "System Disk");
  const partitionedLoopItem = items.find((item) => item.name === "loop7");

  expect(firstItem).toBeDefined();
  expect(partitionedLoopItem).toBeDefined();

  if (!firstItem || !partitionedLoopItem) {
    throw new Error("Expected both primary and partition-container loop devices");
  }

  expect(firstItem).toMatchObject({
    connectionLabel: "SATA | disk | solid-state",
    mountsLabel: "Partitions: sda1",
    name: "System Disk",
    sizeLabel: "1.0 GiB",
    subtitle: "sda | ATA | LiteNAS SSD",
  });
  expect(firstItem.ioSummaryLabels).toEqual([
    { key: "Read", value: "1.3 MiB" },
    { key: "Written", value: "2.5 MiB" },
    { key: "Ops", value: "380" },
    { key: "In flight", value: "2" },
  ]);
  expect(firstItem.readWriteSeries.valuesByKey).toEqual({
    Read: [0, 307200],
    Write: [0, 614400],
  });
  expect(firstItem.utilizationSeries.valuesByKey).toEqual({
    Busy: [0, 25],
  });
  expect(partitionedLoopItem).toMatchObject({
    connectionLabel: "LOOP | loop | solid-state",
    mountsLabel: "Partitions: loop7p1",
    name: "loop7",
    sizeLabel: "256 MiB",
    subtitle: "loop7",
  });
  expect(partitionedLoopItem.readWriteSeries.valuesByKey).toEqual({
    Read: [0, 204800],
    Write: [0, 307200],
  });
  expect(partitionedLoopItem.utilizationSeries.valuesByKey).toEqual({
    Busy: [0, 8],
  });
});

test("keeps loop-backed snap devices out of primary storage mounts but exposes them separately", () => {
  const panelData = buildDiskMountsPanelData(diskMetricItems);
  const auxiliaryDevices = buildDiskAuxiliaryDevicesPanelData(diskMetricItems);

  expect(panelData.mountRows).toHaveLength(3);
  expect(panelData.mountRows.map((row) => row.mountpoint)).toEqual(["/", "/mnt/archive", "/tank"]);
  expect(panelData.filesystemRows).toEqual([
    { filesystem: "ext4", mountCount: 2, type: "local block" },
    { filesystem: "zfs", mountCount: 1, type: "local block" },
  ]);
  expect(panelData.summaryLabels).toEqual([
    { key: "Storage mounts", value: "3" },
    { key: "Filesystems", value: "2" },
    { key: "Capacity", value: "3.1 GiB" },
    { key: "Used", value: "1.2 GiB" },
    { key: "Fullest mount", value: "/ (50.0%)" },
  ]);
  expect(auxiliaryDevices.loopDevices).toEqual([
    {
      connectionLabel: "LOOP | loop | solid-state",
      mountsLabel: "Mounts: /snap/core",
      name: "loop0",
      roleLabel: "snap image",
      sizeLabel: "1.0 KiB",
      subtitle: "No additional metadata reported",
    },
  ]);
  expect(auxiliaryDevices.otherDevices).toEqual([]);
  expect(auxiliaryDevices.summaryLabels).toEqual([
    { key: "Loop devices", value: "1" },
    { key: "Snap-backed", value: "1" },
    { key: "Other block devices", value: "0" },
  ]);
});

/**
 * Builds one complete disk snapshot test item around the shared defaults.
 */
function buildDiskSnapshot(overrides: Partial<DiskMetricSnapshotDTO>): DiskMetricSnapshotDTO {
  return {
    devices: [],
    filesystems: buildFilesystemSummaryMap(),
    mounts: {},
    timestamp: "2026-06-16T10:57:17.000Z",
    ...overrides,
  };
}

/**
 * Builds one reusable top-level disk device fixture for disk helper tests.
 */
function buildTopLevelDiskDevice(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "SATA",
    io: buildDiskIoCounters({}),
    kind: "disk",
    major: 8,
    minor: 0,
    model: "LiteNAS SSD",
    name: "System Disk",
    node: "sda",
    partitions: ["sda1"],
    removable: false,
    rotational: false,
    size_bytes: 1024 ** 3,
    vendor: "ATA",
    ...overrides,
  };
}

/**
 * Builds one reusable loop device fixture for mount filtering coverage.
 */
function buildLoopDevice(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "LOOP",
    io: buildDiskIoCounters({
      io_time_ms: 20,
      reads_completed: 10,
      read_time_ms: 5,
      sectors_read: 100,
      sectors_written: 0,
      weighted_io_time_ms: 20,
      writes_completed: 0,
      write_time_ms: 0,
    }),
    kind: "loop",
    major: 7,
    minor: 0,
    node: "loop0",
    rotational: false,
    size_bytes: 1024,
    ...overrides,
  };
}

/**
 * Builds one reusable partition-container loop device fixture for chartable image coverage.
 */
function buildPartitionedLoopDevice(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "LOOP",
    io: buildDiskIoCounters({
      io_time_ms: 60,
      reads_completed: 20,
      read_time_ms: 10,
      sectors_read: 800,
      sectors_written: 1600,
      weighted_io_time_ms: 60,
      writes_completed: 30,
      write_time_ms: 12,
    }),
    kind: "loop",
    major: 7,
    minor: 7,
    node: "loop7",
    partitions: ["loop7p1"],
    rotational: false,
    size_bytes: 256 * 1024 ** 2,
    ...overrides,
  };
}

/**
 * Builds one reusable partition device fixture for disk helper tests.
 */
function buildPartitionDevice(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "SATA",
    io: buildDiskIoCounters({}),
    kind: "partition",
    major: 8,
    minor: 1,
    mounts: ["/"],
    node: "sda1",
    parent: "sda",
    size_bytes: 900 * 1024 ** 2,
    ...overrides,
  };
}

/**
 * Builds one disk I/O counter block with explicit defaults for concise fixtures.
 */
function buildDiskIoCounters(
  overrides: Partial<DiskMetricDeviceIOCountersDTO>,
): DiskMetricDeviceIOCountersDTO {
  return {
    io_in_progress: 0,
    io_time_ms: 100,
    reads_completed: 100,
    reads_merged: 0,
    read_time_ms: 50,
    sectors_read: 2000,
    sectors_written: 4000,
    weighted_io_time_ms: 100,
    writes_completed: 200,
    writes_merged: 0,
    write_time_ms: 40,
    ...overrides,
  };
}

/**
 * Builds one storage mount fixture with concise defaults for test snapshots.
 */
function buildDiskMount(
  overrides: Partial<DiskMetricMountSnapshotDTO> &
    Pick<DiskMetricMountSnapshotDTO, "mountpoint" | "source">,
): DiskMetricMountSnapshotDTO {
  return {
    available_bytes: 500 * 1024 ** 2,
    device: "sda1",
    filesystem: "ext4",
    free_bytes: 500 * 1024 ** 2,
    memory_backed: false,
    mountpoint: overrides.mountpoint,
    readonly: false,
    remote: false,
    source: overrides.source,
    total_bytes: 900 * 1024 ** 2,
    used_bytes: 400 * 1024 ** 2,
    used_percent: 44.4,
    ...overrides,
  };
}

/**
 * Builds the shared filesystem-summary fixture map used by disk helper tests.
 */
function buildFilesystemSummaryMap() {
  return {
    ext4: { mount_count: 1, type: "local_block" },
    proc: { mount_count: 4, type: "pseudo" },
    squashfs: { mount_count: 10, type: "unknown" },
    zfs: { mount_count: 1, type: "local_block" },
  };
}
