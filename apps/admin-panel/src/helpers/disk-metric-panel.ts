import type {
  DiskMetricDeviceIOCountersDTO,
  DiskMetricDeviceSnapshotDTO,
  DiskMetricFilesystemSummaryDTO,
  DiskMetricMountSnapshotDTO,
  DiskMetricSnapshotDTO,
} from "@dto/monitoring/disk-metric";
import {
  formatMetricBytes,
  formatMetricBytesPerSecond,
  formatMetricValue,
} from "@helpers/metric-display";
import type { MetricChartLabel, MetricMultiChartSeries } from "@helpers/system-metric-chart";

const sectorSizeBytes = 512;

/**
 * Browser-facing device card data rendered by the disk telemetry devices panel.
 */
export type DiskDevicePanelItem = {
  connectionLabel: string;
  ioSummaryLabels: MetricChartLabel[];
  mountsLabel: string;
  name: string;
  readWriteSeries: MetricMultiChartSeries;
  sizeLabel: string;
  subtitle: string;
  utilizationSeries: MetricMultiChartSeries;
};

/**
 * Browser-facing auxiliary block-device row rendered by the secondary devices panel.
 */
export type DiskAuxiliaryDeviceItem = {
  connectionLabel: string;
  mountsLabel: string;
  name: string;
  roleLabel: string;
  sizeLabel: string;
  subtitle: string;
};

/**
 * Browser-facing auxiliary block-device data rendered by the secondary devices panel.
 */
export type DiskAuxiliaryDevicesPanelData = {
  loopDevices: DiskAuxiliaryDeviceItem[];
  otherDevices: DiskAuxiliaryDeviceItem[];
  summaryLabels: MetricChartLabel[];
};

/**
 * Browser-facing storage mount row rendered by the disk telemetry mounts table.
 */
export type DiskMountPanelRow = {
  availableBytes: number;
  availableLabel: string;
  deviceLabel: string;
  filesystem: string;
  mountpoint: string;
  source: string;
  totalBytes: number;
  totalLabel: string;
  usedBytes: number;
  usedLabel: string;
  usedPercent: number;
  usageLabel: string;
};

/**
 * Browser-facing filesystem summary row rendered by the disk telemetry filesystems table.
 */
export type DiskFilesystemPanelRow = {
  filesystem: string;
  mountCount: number;
  type: string;
};

/**
 * Browser-facing mount and filesystem data rendered by the disk telemetry storage panel.
 */
export type DiskMountsPanelData = {
  filesystemRows: DiskFilesystemPanelRow[];
  mountRows: DiskMountPanelRow[];
  summaryLabels: MetricChartLabel[];
};

type DiskSeriesExtractor = (
  io: DiskMetricDeviceIOCountersDTO,
  elapsedMs: number,
  previousIo: DiskMetricDeviceIOCountersDTO | null,
) => number;

type DiskSeriesContext = {
  currentIo: DiskMetricDeviceIOCountersDTO | null;
  elapsedMs: number;
  previousIo: DiskMetricDeviceIOCountersDTO | null;
  previousItem: DiskMetricSnapshotDTO | null;
};

/**
 * Converts disk snapshot history into one card per top-level disk device.
 */
export const buildDiskDevicePanelItems = (
  items: DiskMetricSnapshotDTO[],
): DiskDevicePanelItem[] => {
  const latestDevices = (items[items.length - 1]?.devices ?? []).filter(isChartableDiskDevice);

  return latestDevices
    .map((device) => {
      return buildDiskDevicePanelItem(items, device);
    })
    .sort((left, right) => {
      return left.name.localeCompare(right.name);
    });
};

/**
 * Converts the latest disk snapshot into categorized loop and other block-device data.
 */
export const buildDiskAuxiliaryDevicesPanelData = (
  items: DiskMetricSnapshotDTO[],
): DiskAuxiliaryDevicesPanelData => {
  const latestDevices = items[items.length - 1]?.devices ?? [];
  const loopDevices = latestDevices
    .filter(isAuxiliaryLoopDevice)
    .map((device) => {
      return buildDiskAuxiliaryDeviceItem(device, resolveLoopDeviceRole(device));
    })
    .sort(sortAuxiliaryDevices);
  const otherDevices = latestDevices
    .filter(isOtherBlockDevice)
    .map((device) => {
      return buildDiskAuxiliaryDeviceItem(device, resolveOtherBlockDeviceRole(device));
    })
    .sort(sortAuxiliaryDevices);

  return {
    loopDevices,
    otherDevices,
    summaryLabels: buildDiskAuxiliarySummaryLabels(loopDevices, otherDevices),
  };
};

/**
 * Converts the latest disk snapshot into the storage mounts/filesystems workspace data.
 */
export const buildDiskMountsPanelData = (items: DiskMetricSnapshotDTO[]): DiskMountsPanelData => {
  const latestItem = items[items.length - 1];
  const latestMounts = latestItem?.mounts ?? {};
  const latestFilesystems = latestItem?.filesystems ?? {};
  const latestDevices = latestItem?.devices ?? [];
  const mountRows = buildFilteredStorageMountRows(latestMounts, latestDevices, latestFilesystems);
  const filesystemRows = buildFilesystemRowsFromMountRows(mountRows, latestFilesystems);

  return {
    filesystemRows,
    mountRows,
    summaryLabels: buildDiskMountSummaryLabels(mountRows, filesystemRows),
  };
};

/**
 * Builds one top-level disk device card from the latest snapshot and history deltas.
 */
const buildDiskDevicePanelItem = (
  items: DiskMetricSnapshotDTO[],
  device: DiskMetricDeviceSnapshotDTO,
): DiskDevicePanelItem => {
  const latestIo = device.io;

  return {
    connectionLabel: `${device.connection_type} | ${formatDeviceMedium(device)}`,
    ioSummaryLabels: buildDiskDeviceIoSummaryLabels(latestIo),
    mountsLabel: formatDiskMountsLabel(device),
    name: device.name ?? device.node,
    readWriteSeries: buildReadWriteSeries(items, device.node),
    sizeLabel: formatMetricBytes(device.size_bytes),
    subtitle: buildDiskDeviceSubtitle(device),
    utilizationSeries: buildUtilizationSeries(items, device.node),
  };
};

/**
 * Builds one lightweight auxiliary device row from the latest snapshot.
 */
const buildDiskAuxiliaryDeviceItem = (
  device: DiskMetricDeviceSnapshotDTO,
  roleLabel: string,
): DiskAuxiliaryDeviceItem => {
  return {
    connectionLabel: `${device.connection_type} | ${formatDeviceMedium(device)}`,
    mountsLabel: formatDiskMountsLabel(device),
    name: device.name ?? device.node,
    roleLabel,
    sizeLabel: formatMetricBytes(device.size_bytes),
    subtitle: buildDiskAuxiliaryDeviceSubtitle(device),
  };
};

/**
 * Builds the compact subtitle rendered under one disk device heading.
 */
const buildDiskDeviceSubtitle = (device: DiskMetricDeviceSnapshotDTO): string => {
  const parts = [device.node, device.vendor, device.model, device.description].filter(
    (value): value is string => Boolean(value && value.trim().length > 0),
  );

  return parts.join(" | ");
};

/**
 * Builds the auxiliary subtitle without repeating the device node already shown as the row title.
 */
const buildDiskAuxiliaryDeviceSubtitle = (device: DiskMetricDeviceSnapshotDTO): string => {
  const parts = [device.vendor, device.model, device.description].filter((value): value is string =>
    Boolean(value && value.trim().length > 0),
  );

  return parts.length > 0 ? parts.join(" | ") : "No additional metadata reported";
};

/**
 * Builds human-readable mount ownership text for one disk device.
 */
const formatDiskMountsLabel = (device: DiskMetricDeviceSnapshotDTO): string => {
  if ((device.partitions?.length ?? 0) > 0) {
    return `Partitions: ${device.partitions?.join(", ")}`;
  }

  if ((device.mounts?.length ?? 0) > 0) {
    return `Mounts: ${device.mounts?.join(", ")}`;
  }

  return "No direct partitions or mounts reported";
};

/**
 * Formats compact device-medium metadata derived from physical disk flags.
 */
const formatDeviceMedium = (device: DiskMetricDeviceSnapshotDTO): string => {
  const parts = [device.kind];

  if (device.rotational !== null && device.rotational !== undefined) {
    parts.push(device.rotational ? "rotational" : "solid-state");
  }

  if (device.removable) {
    parts.push("removable");
  }

  return parts.join(" | ");
};

/**
 * Builds the latest read/write/queue labels shown in one disk device card.
 */
const buildDiskDeviceIoSummaryLabels = (
  io: DiskMetricDeviceIOCountersDTO | null | undefined,
): MetricChartLabel[] => {
  if (io === null || io === undefined) {
    return [
      { key: "Read", value: "N/A" },
      { key: "Written", value: "N/A" },
      { key: "Ops", value: "N/A" },
      { key: "In flight", value: "N/A" },
    ];
  }

  return [
    { key: "Read", value: formatMetricBytes(io.sectors_read * sectorSizeBytes) },
    { key: "Written", value: formatMetricBytes(io.sectors_written * sectorSizeBytes) },
    {
      key: "Ops",
      value: formatMetricValue(io.reads_completed + io.writes_completed),
    },
    { key: "In flight", value: formatMetricValue(io.io_in_progress) },
  ];
};

/**
 * Builds the dual-series read/write throughput chart data for one disk node.
 */
const buildReadWriteSeries = (
  items: DiskMetricSnapshotDTO[],
  node: string,
): MetricMultiChartSeries => {
  return buildDiskDeltaSeries(items, node, {
    Read: (io) => io.sectors_read * sectorSizeBytes,
    Write: (io) => io.sectors_written * sectorSizeBytes,
  });
};

/**
 * Builds the single-series utilization chart data for one disk node.
 */
const buildUtilizationSeries = (
  items: DiskMetricSnapshotDTO[],
  node: string,
): MetricMultiChartSeries => {
  return buildDiskDeltaSeries(items, node, {
    Busy: (io, elapsedMs, previousIo) => {
      if (elapsedMs <= 0 || previousIo === null) {
        return 0;
      }

      const busyTimeMs = Math.max(io.io_time_ms - previousIo.io_time_ms, 0);

      return Math.min((busyTimeMs / elapsedMs) * 100, 100);
    },
  });
};

/**
 * Converts cumulative device counters into one or more per-interval chart series.
 */
const buildDiskDeltaSeries = (
  items: DiskMetricSnapshotDTO[],
  node: string,
  extractors: Record<string, DiskSeriesExtractor>,
): MetricMultiChartSeries => {
  const stamps = items.map((item) => item.timestamp);
  const valuesByKey = initializeValuesByKey(extractors);

  items.forEach((item, index) => {
    appendDiskSeriesValues(
      valuesByKey,
      extractors,
      buildDiskSeriesContext(items, item, index, node),
    );
  });

  return {
    stamps,
    valuesByKey,
  };
};

/**
 * Prepares one empty numeric series array for each extractor key.
 */
const initializeValuesByKey = (
  extractors: Record<string, DiskSeriesExtractor>,
): Record<string, number[]> => {
  return Object.fromEntries(
    Object.keys(extractors).map((key) => {
      return [key, [] as number[]];
    }),
  );
};

/**
 * Builds the per-snapshot context required to derive interval-based disk values.
 */
const buildDiskSeriesContext = (
  items: DiskMetricSnapshotDTO[],
  item: DiskMetricSnapshotDTO,
  index: number,
  node: string,
): DiskSeriesContext => {
  const previousItem = index > 0 ? (items[index - 1] ?? null) : null;
  const previousIo = previousItem ? findDeviceIo(previousItem, node) : null;

  return {
    currentIo: findDeviceIo(item, node),
    elapsedMs: previousItem
      ? Math.max(Date.parse(item.timestamp) - Date.parse(previousItem.timestamp), 0)
      : 0,
    previousIo,
    previousItem,
  };
};

/**
 * Appends one derived point into each disk chart series for the current snapshot.
 */
const appendDiskSeriesValues = (
  valuesByKey: Record<string, number[]>,
  extractors: Record<string, DiskSeriesExtractor>,
  context: DiskSeriesContext,
): void => {
  Object.entries(extractors).forEach(([key, extractor]) => {
    const seriesValues = valuesByKey[key];

    if (!seriesValues) {
      return;
    }

    seriesValues.push(calculateDiskSeriesValue(extractor, context));
  });
};

/**
 * Calculates one chart point from the current and previous cumulative device counters.
 */
const calculateDiskSeriesValue = (
  extractor: DiskSeriesExtractor,
  context: DiskSeriesContext,
): number => {
  if (context.currentIo === null || context.previousItem === null) {
    return 0;
  }

  if (extractor.length >= 3) {
    return Math.max(extractor(context.currentIo, context.elapsedMs, context.previousIo), 0);
  }

  return calculateDiskRateValue(extractor, context);
};

/**
 * Converts a cumulative byte-like counter into a per-second rate.
 */
const calculateDiskRateValue = (
  extractor: DiskSeriesExtractor,
  context: DiskSeriesContext,
): number => {
  if (context.elapsedMs <= 0) {
    return 0;
  }

  const currentCounter = extractor(
    context.currentIo as DiskMetricDeviceIOCountersDTO,
    context.elapsedMs,
    context.previousIo,
  );
  const previousCounter = context.previousIo
    ? extractor(context.previousIo, context.elapsedMs, null)
    : 0;

  return Math.max(((currentCounter - previousCounter) / context.elapsedMs) * 1000, 0);
};

/**
 * Reads one device I/O counter block from a snapshot by stable node name.
 */
const findDeviceIo = (
  item: DiskMetricSnapshotDTO,
  node: string,
): DiskMetricDeviceIOCountersDTO | null => {
  const device = item.devices?.find((candidate) => candidate.node === node) ?? null;

  return device?.io ?? null;
};

/**
 * Returns whether one snapshot device should be rendered as a top-level disk card.
 */
const isChartableDiskDevice = (device: DiskMetricDeviceSnapshotDTO): boolean => {
  return device.kind === "disk" || isPartitionContainerLoopDevice(device);
};

/**
 * Returns whether one snapshot device represents a loop-backed virtual block device.
 */
const isLoopDevice = (device: DiskMetricDeviceSnapshotDTO): boolean => {
  return (
    device.kind === "loop" ||
    device.node.startsWith("loop") ||
    device.connection_type.toUpperCase() === "LOOP"
  );
};

/**
 * Returns whether one loop-backed device behaves like a partition container and should get full charts.
 */
const isPartitionContainerLoopDevice = (device: DiskMetricDeviceSnapshotDTO): boolean => {
  return isLoopDevice(device) && (device.partitions?.length ?? 0) > 0;
};

/**
 * Returns whether one snapshot device belongs in the auxiliary non-loop block-device list.
 */
const isAuxiliaryLoopDevice = (device: DiskMetricDeviceSnapshotDTO): boolean => {
  return isLoopDevice(device) && !isChartableDiskDevice(device);
};

/**
 * Returns whether one snapshot device belongs in the auxiliary non-loop block-device list.
 */
const isOtherBlockDevice = (device: DiskMetricDeviceSnapshotDTO): boolean => {
  return !isChartableDiskDevice(device) && !isLoopDevice(device) && device.kind !== "partition";
};

/**
 * Returns whether one mount belongs to a storage-oriented filesystem worth surfacing here.
 */
const isStorageMount = (
  mount: DiskMetricMountSnapshotDTO,
  devices: DiskMetricDeviceSnapshotDTO[],
  filesystems: Record<string, DiskMetricFilesystemSummaryDTO>,
): boolean => {
  const filesystemSummary = filesystems[mount.filesystem];

  if (filesystemSummary?.type === "local_block") {
    return true;
  }

  if (!mount.device) {
    return false;
  }

  const device = devices.find((candidate) => candidate.node === mount.device) ?? null;

  return device !== null && !isLoopDevice(device);
};

/**
 * Builds one filtered and sorted storage-mount row list from the latest snapshot maps.
 */
const buildFilteredStorageMountRows = (
  latestMounts: Record<string, DiskMetricMountSnapshotDTO>,
  latestDevices: DiskMetricDeviceSnapshotDTO[],
  latestFilesystems: Record<string, DiskMetricFilesystemSummaryDTO>,
): DiskMountPanelRow[] => {
  return Object.values(latestMounts)
    .filter((mount) => {
      return isStorageMount(mount, latestDevices, latestFilesystems);
    })
    .map((mount) => {
      return buildDiskMountPanelRow(mount, latestDevices);
    })
    .sort((left, right) => {
      return (
        right.usedPercent - left.usedPercent || left.mountpoint.localeCompare(right.mountpoint)
      );
    });
};

/**
 * Builds filesystem rows from the already filtered storage mounts so loop-only filesystems stay hidden.
 */
const buildFilesystemRowsFromMountRows = (
  mountRows: DiskMountPanelRow[],
  latestFilesystems: Record<string, DiskMetricFilesystemSummaryDTO>,
): DiskFilesystemPanelRow[] => {
  const mountCountByFilesystem = mountRows.reduce<Record<string, number>>((accumulator, row) => {
    accumulator[row.filesystem] = (accumulator[row.filesystem] ?? 0) + 1;
    return accumulator;
  }, {});

  return Object.entries(mountCountByFilesystem)
    .map(([filesystem, mountCount]) => {
      return buildDiskFilesystemPanelRow(filesystem, latestFilesystems[filesystem], mountCount);
    })
    .sort((left, right) => {
      return right.mountCount - left.mountCount || left.filesystem.localeCompare(right.filesystem);
    });
};

/**
 * Builds one storage mount table row from the latest snapshot mount item.
 */
const buildDiskMountPanelRow = (
  mount: DiskMetricMountSnapshotDTO,
  devices: DiskMetricDeviceSnapshotDTO[],
): DiskMountPanelRow => {
  const byteTotals = normalizeDiskMountByteTotals(mount);
  const deviceLabel = resolveDiskMountDeviceLabel(mount, devices);
  const usageLabel = `${byteTotals.usedPercent.toFixed(1)}%`;

  return {
    availableBytes: byteTotals.availableBytes,
    availableLabel: formatDiskMountByteLabel(byteTotals.availableBytes, byteTotals.totalBytes),
    deviceLabel,
    filesystem: mount.filesystem,
    mountpoint: mount.mountpoint,
    source: mount.source,
    totalBytes: byteTotals.totalBytes,
    totalLabel: formatDiskMountByteLabel(byteTotals.totalBytes, byteTotals.totalBytes),
    usedBytes: byteTotals.usedBytes,
    usedLabel: formatDiskMountByteLabel(byteTotals.usedBytes, byteTotals.totalBytes),
    usedPercent: byteTotals.usedPercent,
    usageLabel,
  };
};

/**
 * Normalizes the nullable byte counters carried by one mount snapshot.
 */
const normalizeDiskMountByteTotals = (mount: DiskMetricMountSnapshotDTO) => {
  return {
    availableBytes: mount.available_bytes ?? 0,
    totalBytes: mount.total_bytes ?? 0,
    usedBytes: mount.used_bytes ?? 0,
    usedPercent: mount.used_percent ?? 0,
  };
};

/**
 * Formats one mount byte total or returns N/A when the mount has no real capacity.
 */
const formatDiskMountByteLabel = (value: number, totalBytes: number): string => {
  return totalBytes > 0 ? formatMetricBytes(value) : "N/A";
};

/**
 * Resolves the mount table device label, folding partitions back to their parent disk when possible.
 */
const resolveDiskMountDeviceLabel = (
  mount: DiskMetricMountSnapshotDTO,
  devices: DiskMetricDeviceSnapshotDTO[],
): string => {
  if (!mount.device) {
    return "virtual";
  }

  const device = devices.find((candidate) => candidate.node === mount.device) ?? null;

  if (device?.kind === "partition" && device.parent) {
    return device.parent;
  }

  return mount.device;
};

/**
 * Builds one filesystem summary table row from one filtered filesystem count.
 */
const buildDiskFilesystemPanelRow = (
  filesystem: string,
  summary: DiskMetricFilesystemSummaryDTO | undefined,
  mountCount: number,
): DiskFilesystemPanelRow => {
  return {
    filesystem,
    mountCount,
    type: summary?.type.replaceAll("_", " ") ?? "unknown",
  };
};

/**
 * Resolves one human-readable role label for one loop-backed device.
 */
const resolveLoopDeviceRole = (device: DiskMetricDeviceSnapshotDTO): string => {
  if (device.mounts?.some((mount) => mount.startsWith("/snap/"))) {
    return "snap image";
  }

  if ((device.mounts?.length ?? 0) > 0) {
    return "mounted image";
  }

  return "virtual loop device";
};

/**
 * Resolves one human-readable role label for a non-primary non-loop block device.
 */
const resolveOtherBlockDeviceRole = (device: DiskMetricDeviceSnapshotDTO): string => {
  if (device.kind === "partition") {
    return "partition";
  }

  if (device.removable) {
    return "removable device";
  }

  return device.kind;
};

/**
 * Sorts auxiliary devices by role and then by stable display name.
 */
const sortAuxiliaryDevices = (
  left: DiskAuxiliaryDeviceItem,
  right: DiskAuxiliaryDeviceItem,
): number => {
  return left.name.localeCompare(right.name);
};

/**
 * Builds the summary row shown above the auxiliary block-device lists.
 */
const buildDiskAuxiliarySummaryLabels = (
  loopDevices: DiskAuxiliaryDeviceItem[],
  otherDevices: DiskAuxiliaryDeviceItem[],
): MetricChartLabel[] => {
  const snapDeviceCount = loopDevices.filter((device) => device.roleLabel === "snap image").length;

  return [
    { key: "Loop devices", value: formatMetricValue(loopDevices.length) },
    { key: "Snap-backed", value: formatMetricValue(snapDeviceCount) },
    { key: "Other block devices", value: formatMetricValue(otherDevices.length) },
  ];
};

/**
 * Builds the summary row shown above the mounts/filesystems tables.
 */
const buildDiskMountSummaryLabels = (
  mountRows: DiskMountPanelRow[],
  filesystemRows: DiskFilesystemPanelRow[],
): MetricChartLabel[] => {
  const totalBytes = mountRows.reduce((sum, row) => {
    return sum + row.totalBytes;
  }, 0);
  const usedBytes = mountRows.reduce((sum, row) => {
    return sum + row.usedBytes;
  }, 0);
  const fullestMount = mountRows[0] ?? null;

  return [
    { key: "Storage mounts", value: formatMetricValue(mountRows.length) },
    { key: "Filesystems", value: formatMetricValue(filesystemRows.length) },
    { key: "Capacity", value: totalBytes > 0 ? formatMetricBytes(totalBytes) : "N/A" },
    { key: "Used", value: usedBytes > 0 ? formatMetricBytes(usedBytes) : "N/A" },
    {
      key: "Fullest mount",
      value: fullestMount ? `${fullestMount.mountpoint} (${fullestMount.usageLabel})` : "N/A",
    },
  ];
};

/**
 * Formats one chart value for the read/write throughput series.
 */
export const formatDiskReadWriteValue = (value: number): string => {
  return formatMetricBytesPerSecond(value);
};

/**
 * Formats one chart value for the utilization series.
 */
export const formatDiskUtilizationValue = (value: number): string => {
  return `${value.toFixed(value >= 10 ? 0 : 1)}%`;
};
