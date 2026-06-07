import type { RbacContextValue } from "@dto/rbac/rbac";
import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import DeviceThermostatRoundedIcon from "@mui/icons-material/DeviceThermostatRounded";
import DnsRoundedIcon from "@mui/icons-material/DnsRounded";
import ElectricBoltRoundedIcon from "@mui/icons-material/ElectricBoltRounded";
import ManageAccountsRoundedIcon from "@mui/icons-material/ManageAccountsRounded";
import MemoryIcon from "@mui/icons-material/Memory";
import MemoryRoundedIcon from "@mui/icons-material/MemoryRounded";
import MonitorHeartRoundedIcon from "@mui/icons-material/MonitorHeartRounded";
import NotificationImportantIcon from "@mui/icons-material/NotificationImportant";
import NotificationsRoundedIcon from "@mui/icons-material/NotificationsRounded";
import PaletteRoundedIcon from "@mui/icons-material/PaletteRounded";
import PrivacyTipIcon from "@mui/icons-material/PrivacyTip";
import SecurityIcon from "@mui/icons-material/Security";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import SpeedRoundedIcon from "@mui/icons-material/SpeedRounded";
import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import TableRowsIcon from "@mui/icons-material/TableRows";
import TuneRoundedIcon from "@mui/icons-material/TuneRounded";
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
 * RBAC guards used when building role-aware dashboard navigation.
 */
export type AppNavigationRbac = Pick<RbacContextValue, "requireOperator" | "requireSecurity">;

/**
 * RBAC fallback used by static navigation tests and anonymous states.
 */
const denyAllNavigationRbac: AppNavigationRbac = {
  requireOperator: () => false,
  requireSecurity: () => false,
};

/**
 * Sidebar navigation for authenticated admin-panel routes.
 *
 * The shape intentionally mirrors Toolpad Core navigation concepts so the
 * Material-only draft can be replaced by `DashboardLayout` later without
 * changing route ownership.
 */
export const buildAppNavigationItems = ({
  requireOperator,
  requireSecurity,
}: AppNavigationRbac): AppNavigationItem[] => {
  return [
    buildDashboardNavigationItem(),
    ...buildAlertsNavigationItems({ requireOperator, requireSecurity }),
    buildSystemNavigationItem(),
  ];
};

/**
 * Builds the dashboard root navigation entry.
 */
const buildDashboardNavigationItem = (): AppNavigationItem => {
  return {
    icon: <DashboardRoundedIcon />,
    path: "/",
    title: "Dashboard",
  };
};

/**
 * Builds the alerts root navigation entry when at least one alerts domain is visible.
 */
const buildAlertsNavigationItems = ({
  requireOperator,
  requireSecurity,
}: AppNavigationRbac): AppNavigationItem[] => {
  const domainItems = [
    buildSystemAlertsNavigationItem(requireOperator),
    buildSecurityAlertsNavigationItem(requireSecurity),
  ].filter((item): item is AppNavigationItem => item !== null);

  if (domainItems.length === 0) {
    return [];
  }

  return [
    {
      children: domainItems,
      icon: <NotificationsRoundedIcon />,
      path: "/alerts",
      title: "Alerts",
    },
  ];
};

/**
 * Builds the system alerts navigation branch when operator access is allowed.
 */
const buildSystemAlertsNavigationItem = (
  requireOperator: AppNavigationRbac["requireOperator"],
): AppNavigationItem | null => {
  if (!requireOperator()) {
    return null;
  }

  return {
    children: buildAlertStatusNavigationItems("/alerts/system"),
    icon: <MemoryIcon />,
    path: "/alerts/system",
    title: "System",
  };
};

/**
 * Builds the security alerts navigation branch when security access is allowed.
 */
const buildSecurityAlertsNavigationItem = (
  requireSecurity: AppNavigationRbac["requireSecurity"],
): AppNavigationItem | null => {
  if (!requireSecurity()) {
    return null;
  }

  return {
    children: buildAlertStatusNavigationItems("/alerts/security"),
    icon: <SecurityIcon />,
    path: "/alerts/security",
    title: "Security",
  };
};

/**
 * Builds the shared status pages nested under one alerts domain.
 */
const buildAlertStatusNavigationItems = (domainPath: string): AppNavigationItem[] => {
  return [
    {
      icon: <NotificationImportantIcon />,
      path: `${domainPath}/unacknowledged`,
      title: "Unacknowledged alerts",
    },
    {
      icon: <PrivacyTipIcon />,
      path: `${domainPath}/active`,
      title: "Active alerts",
    },
    {
      icon: <TableRowsIcon />,
      path: `${domainPath}/all`,
      title: "All alerts",
    },
  ];
};

/**
 * Builds the system metrics and sensors navigation branch.
 */
const buildSystemNavigationItem = (): AppNavigationItem => {
  return {
    children: [buildPerformanceNavigationItem(), buildSensorsNavigationItem()],
    icon: <MemoryRoundedIcon />,
    path: "/system",
    title: "System",
  };
};

/**
 * Builds the system performance navigation branch.
 */
const buildPerformanceNavigationItem = (): AppNavigationItem => {
  return {
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
  };
};

/**
 * Builds the Raspberry Pi sensors navigation branch.
 */
const buildSensorsNavigationItem = (): AppNavigationItem => {
  return {
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
  };
};

/**
 * Default admin navigation used in static selection logic and tests.
 */
export const appNavigationItems: AppNavigationItem[] =
  buildAppNavigationItems(denyAllNavigationRbac);

/**
 * Sidebar navigation for the authenticated preferences area.
 */
export const preferencesNavigationItems: AppNavigationItem[] = [
  {
    icon: <TuneRoundedIcon />,
    path: "/preferences",
    title: "Preferences",
  },
  {
    icon: <ManageAccountsRoundedIcon />,
    path: "/preferences/profile",
    title: "User profile",
  },
  {
    children: [
      {
        icon: <PaletteRoundedIcon />,
        path: "/preferences/application/theme",
        title: "Theme",
      },
      {
        icon: <MonitorHeartRoundedIcon />,
        path: "/preferences/application/monitoring",
        title: "Monitoring",
      },
    ],
    icon: <SettingsRoundedIcon />,
    path: "/preferences/application",
    title: "Application settings",
  },
];

/**
 * Returns the sidebar tree that owns the current URL.
 */
export const resolveNavigationItems = (
  pathname: string,
  rbac: AppNavigationRbac = denyAllNavigationRbac,
): AppNavigationItem[] => {
  if (pathname === "/preferences" || pathname.startsWith("/preferences/")) {
    return preferencesNavigationItems;
  }

  if (!rbac.requireOperator() && !rbac.requireSecurity()) {
    return appNavigationItems;
  }

  return buildAppNavigationItems(rbac);
};

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
