/**
 * One telemetry route supported by the current admin-panel metrics integrations.
 */
export type SupportedTelemetryRoute = {
  category: string;
  group: "performance" | "sensors";
  summary: string;
  title: string;
  type: "network-metric" | "system-metric" | "zfs-metric" | "unsupported";
};

type SupportedPerformanceTelemetryType = Exclude<SupportedTelemetryRoute["type"], "unsupported">;

type SupportedPerformanceRouteConfig = {
  summary: string;
  type: SupportedPerformanceTelemetryType;
};

const supportedPerformanceRoutesByCategory: Record<string, SupportedPerformanceRouteConfig> = {
  network: {
    summary: "Gateway-backed network telemetry is available for this route.",
    type: "network-metric",
  },
  system: {
    summary: "Gateway-backed CPU and memory telemetry is available for this route.",
    type: "system-metric",
  },
  zfs: {
    summary: "Gateway-backed ZFS pool telemetry is available for this route.",
    type: "zfs-metric",
  },
};

/**
 * Resolves the current telemetry route metadata from the browser pathname.
 */
export const resolveTelemetryRoute = (
  pathname: string,
  category: string,
): SupportedTelemetryRoute => {
  const group = pathname.startsWith("/system/sensors/") ? "sensors" : "performance";
  const title = resolveTelemetryRouteTitle(group, category);
  const performanceRouteConfig =
    group === "performance" ? supportedPerformanceRoutesByCategory[category] : undefined;

  if (performanceRouteConfig !== undefined) {
    return {
      category,
      group,
      summary: performanceRouteConfig.summary,
      title,
      type: performanceRouteConfig.type,
    };
  }

  return {
    category,
    group,
    summary: "This telemetry route does not have a backend metrics contract yet.",
    title,
    type: "unsupported",
  };
};

/**
 * Resolves route-specific heading overrides used by supported telemetry pages.
 */
const resolveTelemetryRouteTitle = (
  group: SupportedTelemetryRoute["group"],
  category: string,
): string => {
  if (group === "performance" && category === "system") {
    return "System (CPU & RAM)";
  }

  return formatRouteLabel(category);
};

/**
 * Formats one telemetry route segment for user-facing page headings.
 */
export const formatRouteLabel = (value: string): string => {
  return value.slice(0, 1).toUpperCase() + value.slice(1).replaceAll("-", " ");
};
