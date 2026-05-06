import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import DnsRoundedIcon from "@mui/icons-material/DnsRounded";
import MemoryRoundedIcon from "@mui/icons-material/MemoryRounded";
import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import type { ReactElement } from "react";

/**
 * Landing page for system performance telemetry.
 */
export const SystemPerformanceLandingPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description: "CPU, memory, load, and process-level runtime indicators.",
          icon: <MemoryRoundedIcon />,
          path: "/system/performance/system",
          title: "System",
        },
        {
          description: "Interface throughput, link state, and network error counters.",
          icon: <DnsRoundedIcon />,
          path: "/system/performance/network",
          title: "Network",
        },
        {
          description: "Disk utilization, latency, and storage device throughput.",
          icon: <StorageRoundedIcon />,
          path: "/system/performance/disk",
          title: "Disk",
        },
        {
          description: "Pool health, vdev state, ARC behavior, and ZFS-specific trends.",
          icon: <StorageRoundedIcon />,
          path: "/system/performance/zfs",
          title: "ZFS",
        },
      ]}
      overline="System"
      summary="Review host performance from high-level runtime indicators down to storage-specific telemetry."
      title="Performance"
    />
  );
};
