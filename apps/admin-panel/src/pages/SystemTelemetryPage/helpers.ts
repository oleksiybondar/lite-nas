/**
 * One telemetry route supported by the current admin-panel metrics integrations.
 */
export type SupportedTelemetryRoute = {
  category: string;
  group: "performance" | "processes" | "sensors";
  summary: string;
  title: string;
  type:
    | "disk-metric"
    | "network-metric"
    | "process-metric"
    | "service-metric"
    | "system-metric"
    | "zfs-metric"
    | "unsupported";
};

type SupportedTelemetryType = Exclude<SupportedTelemetryRoute["type"], "unsupported">;

type SupportedTelemetryRouteConfig = {
  summary: string;
  type: SupportedTelemetryType;
};

const supportedPerformanceRoutesByCategory: Record<string, SupportedTelemetryRouteConfig> = {
  disk: {
    summary: "Gateway-backed disk and filesystem telemetry is available for this route.",
    type: "disk-metric",
  },
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

const supportedProcessesRoutesByCategory: Record<string, SupportedTelemetryRouteConfig> = {
  processes: {
    summary: "Gateway-backed process snapshot polling is available for this route.",
    type: "process-metric",
  },
  services: {
    summary: "Gateway-backed service snapshot polling is available for this route.",
    type: "service-metric",
  },
};

/**
 * Resolves the current telemetry route metadata from the browser pathname.
 */
export const resolveTelemetryRoute = (
  pathname: string,
  category: string,
): SupportedTelemetryRoute => {
  const group = resolveTelemetryRouteGroup(pathname);
  const title = resolveTelemetryRouteTitle(group, category);
  const routeConfig = resolveTelemetryRouteConfig(group, category);

  if (routeConfig !== undefined) {
    return {
      category,
      group,
      summary: routeConfig.summary,
      title,
      type: routeConfig.type,
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
 * Resolves the telemetry subsection from the current browser pathname.
 */
const resolveTelemetryRouteGroup = (pathname: string): SupportedTelemetryRoute["group"] => {
  if (pathname.startsWith("/system/sensors/")) {
    return "sensors";
  }

  if (pathname.startsWith("/system/processes/")) {
    return "processes";
  }

  return "performance";
};

/**
 * Resolves route support metadata from one telemetry subsection and category.
 */
const resolveTelemetryRouteConfig = (
  group: SupportedTelemetryRoute["group"],
  category: string,
): SupportedTelemetryRouteConfig | undefined => {
  if (group === "performance") {
    return supportedPerformanceRoutesByCategory[category];
  }

  if (group === "processes") {
    return supportedProcessesRoutesByCategory[category];
  }

  return undefined;
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
