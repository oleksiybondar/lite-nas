import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import DeviceThermostatRoundedIcon from "@mui/icons-material/DeviceThermostatRounded";
import DnsRoundedIcon from "@mui/icons-material/DnsRounded";
import ElectricBoltRoundedIcon from "@mui/icons-material/ElectricBoltRounded";
import MemoryRoundedIcon from "@mui/icons-material/MemoryRounded";
import SpeedRoundedIcon from "@mui/icons-material/SpeedRounded";
import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import WarningAmberRoundedIcon from "@mui/icons-material/WarningAmberRounded";
import type { ReactNode } from "react";

/**
 * Page item rendered in the protected dashboard sidebar.
 */
export type AppNavigationPageItem = {
  /**
   * Optional child page items shown under this item.
   */
  children?: AppNavigationPageItem[];
  /**
   * Icon rendered before the item label.
   */
  icon?: ReactNode;
  /**
   * Stable label shown in the sidebar.
   */
  title: string;
  /**
   * Route path represented by this navigation item.
   */
  path: string;
};

/**
 * Item supported by the protected dashboard navigation model.
 */
export type AppNavigationItem = AppNavigationPageItem;

/**
 * Sidebar navigation for authenticated admin-panel routes.
 *
 * The shape intentionally mirrors Toolpad Core navigation concepts so the
 * Material-only draft can be replaced by `DashboardLayout` later without
 * changing route ownership.
 */
export const appNavigationItems: AppNavigationItem[] = [
  {
    icon: <DashboardRoundedIcon />,
    path: "/",
    title: "Dashboard",
  },
  {
    children: [
      {
        children: [
          {
            icon: <MemoryRoundedIcon />,
            path: "/system/performance/system",
            title: "System",
          },
          {
            icon: <DnsRoundedIcon />,
            path: "/system/performance/network",
            title: "Network",
          },
          {
            icon: <StorageRoundedIcon />,
            path: "/system/performance/disk",
            title: "Disk",
          },
          {
            icon: <StorageRoundedIcon />,
            path: "/system/performance/zfs",
            title: "ZFS",
          },
        ],
        icon: <SpeedRoundedIcon />,
        path: "/system/performance",
        title: "Performance",
      },
      {
        children: [
          {
            icon: <DeviceThermostatRoundedIcon />,
            path: "/system/sensors/temperature",
            title: "Temperature",
          },
          {
            icon: <ElectricBoltRoundedIcon />,
            path: "/system/sensors/voltage",
            title: "Voltage",
          },
          {
            icon: <SpeedRoundedIcon />,
            path: "/system/sensors/clock",
            title: "Clock",
          },
          {
            icon: <WarningAmberRoundedIcon />,
            path: "/system/sensors/throttling",
            title: "Throttling",
          },
          {
            icon: <SpeedRoundedIcon />,
            path: "/system/sensors/fan",
            title: "Fan",
          },
        ],
        icon: <DeviceThermostatRoundedIcon />,
        path: "/system/sensors",
        title: "Sensors",
      },
    ],
    icon: <MemoryRoundedIcon />,
    path: "/system",
    title: "System",
  },
];

/**
 * Resolves the deepest navigation path that should appear selected for a URL.
 */
export const resolveSelectedNavigationPath = (
  pathname: string,
  items: AppNavigationItem[] = appNavigationItems,
): string | null => {
  const pages = flattenNavigationPages(items);
  const matchingPages = pages.filter((item) => {
    return pathname === item.path || pathname.startsWith(`${item.path}/`);
  });

  if (matchingPages.length === 0) {
    return null;
  }

  const [selectedPage] = matchingPages.sort(
    (first, second) => second.path.length - first.path.length,
  );

  return selectedPage?.path ?? null;
};

/**
 * Flattens nested sidebar navigation into clickable page items.
 */
const flattenNavigationPages = (items: AppNavigationItem[]): AppNavigationPageItem[] => {
  return items.flatMap((item) => {
    return [item, ...flattenNavigationPages(item.children ?? [])];
  });
};
