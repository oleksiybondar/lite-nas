/**
 * One telemetry route supported by the current admin-panel metrics integrations.
 */
export type SupportedTelemetryRoute = {
  category: string;
  group: "performance" | "sensors";
  summary: string;
  title: string;
  type: "system-metric" | "zfs-metric" | "unsupported";
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

  if (group === "performance" && category === "system") {
    return {
      category,
      group,
      summary: "Gateway-backed CPU and memory telemetry is available for this route.",
      title,
      type: "system-metric",
    };
  }

  if (group === "performance" && category === "zfs") {
    return {
      category,
      group,
      summary: "Gateway-backed ZFS pool telemetry is available for this route.",
      title,
      type: "zfs-metric",
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
