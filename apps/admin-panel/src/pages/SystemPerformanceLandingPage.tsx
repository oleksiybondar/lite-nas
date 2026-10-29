import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useLanguage } from "@lite-nas/language";
import DnsRoundedIcon from "@mui/icons-material/DnsRounded";
import MemoryRoundedIcon from "@mui/icons-material/MemoryRounded";
import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import type { ReactElement } from "react";

/**
 * Landing page for system performance telemetry.
 */
export const SystemPerformanceLandingPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.performanceSystem"),
          icon: <MemoryRoundedIcon />,
          path: "/system/performance/system",
          title: getLabel("navigation.System"),
        },
        {
          description: getLabel("admin.landing.performanceNetwork"),
          icon: <DnsRoundedIcon />,
          path: "/system/performance/network",
          title: getLabel("navigation.Network"),
        },
        {
          description: getLabel("admin.landing.performanceDisk"),
          icon: <StorageRoundedIcon />,
          path: "/system/performance/disk",
          title: getLabel("navigation.Disk"),
        },
        {
          description: getLabel("admin.landing.performanceZfs"),
          icon: <StorageRoundedIcon />,
          path: "/system/performance/zfs",
          title: getLabel("navigation.ZFS"),
        },
      ]}
      overline={getLabel("navigation.System")}
      summary={getLabel("admin.landing.performanceSummary")}
      title={getLabel("navigation.Performance")}
    />
  );
};
