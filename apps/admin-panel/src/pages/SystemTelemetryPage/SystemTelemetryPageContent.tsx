import { DiskMetricProvider } from "@providers/DiskMetricProvider";
import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import { NetworkMetricProvider } from "@providers/NetworkMetricProvider";
import { SystemMetricProvider } from "@providers/SystemMetricProvider";
import { ZFSMetricProvider } from "@providers/ZFSMetricProvider";
import type { ReactElement } from "react";
import type { SupportedTelemetryRoute } from "./helpers";

type SystemTelemetryPageContentProps = {
  /**
   * Concrete telemetry state rendered under the selected provider.
   */
  children: ReactElement;
  /**
   * Current telemetry route metadata resolved from the browser URL.
   */
  route: SupportedTelemetryRoute;
};

type PerformanceRouteWrapper = (children: ReactElement) => ReactElement;

const performanceRouteWrappers: Record<string, PerformanceRouteWrapper> = {
  disk: (children) => (
    <MonitoringPollingSettingsProvider storageKey="disk-metrics">
      <DiskMetricProvider>{children}</DiskMetricProvider>
    </MonitoringPollingSettingsProvider>
  ),
  network: (children) => (
    <MonitoringPollingSettingsProvider storageKey="network-metrics">
      <NetworkMetricProvider>{children}</NetworkMetricProvider>
    </MonitoringPollingSettingsProvider>
  ),
  system: (children) => (
    <MonitoringPollingSettingsProvider storageKey="system-metrics">
      <SystemMetricProvider>{children}</SystemMetricProvider>
    </MonitoringPollingSettingsProvider>
  ),
  zfs: (children) => (
    <MonitoringPollingSettingsProvider storageKey="zfs-metrics">
      <ZFSMetricProvider>{children}</ZFSMetricProvider>
    </MonitoringPollingSettingsProvider>
  ),
};

/**
 * Wraps telemetry route content with the providers required by the current route type.
 */
export const SystemTelemetryPageContent = ({
  children,
  route,
}: SystemTelemetryPageContentProps): ReactElement => {
  if (route.group !== "performance") {
    return children;
  }

  return performanceRouteWrappers[route.category]?.(children) ?? children;
};
