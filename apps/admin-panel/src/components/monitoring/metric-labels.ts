/** Label translations shared by disk and network telemetry summary rows. */
const monitoringMetricLabelKeys: Record<string, string> = {
  "Active TCP": "activeTcp",
  "All sockets": "allSockets",
  "Close wait": "closeWait",
  "Fullest mount": "fullestMount",
  "In flight": "inFlight",
  "Loop devices": "loopDevices",
  "Other block devices": "otherBlockDevices",
  "Snap-backed": "snapBacked",
  "Sockets used": "socketsUsed",
  "Storage mounts": "storageMounts",
  TCP6: "tcp6",
  "Total RX": "totalRx",
  "Total TX": "totalTx",
  UDP6: "udp6",
  Written: "written",
  "mounted image": "mountedImage",
  "removable device": "removableDevice",
  "snap image": "snapImage",
  "virtual loop device": "virtualLoopDevice",
  partition: "partition",
  virtual: "virtual",
  Capacity: "capacity",
  Close: "close",
  Established: "established",
  Filesystems: "filesystems",
  Listen: "listen",
  Listening: "listening",
  Ops: "operations",
  Read: "read",
  TCP: "tcp",
  Time: "timeWait",
  "Time wait": "timeWait",
  UDP: "udp",
  Used: "used",
};

/** Resolves known telemetry-summary labels and preserves data-defined labels. */
export const localizeMonitoringMetricLabel = (
  label: string,
  getLabel: (key: string) => string,
): string => {
  const key = monitoringMetricLabelKeys[label];
  return key ? getLabel(`admin.monitoringMetrics.${key}`) : label;
};

/** Localizes the shared unavailable marker used in telemetry summaries. */
export const localizeMonitoringMetricValue = (
  value: string,
  getLabel: (key: string) => string,
): string => (value === "N/A" ? getLabel("admin.monitoringMetrics.notAvailable") : value);
