import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import { SystemMetricProvider } from "@providers/SystemMetricProvider";
import { ZFSMetricProvider } from "@providers/ZFSMetricProvider";
import type { ReactElement } from "react";
import type { SupportedTelemetryRoute } from "./helpers";

type SystemTelemetryPageContentProps = {
  /**
   * Telemetry provider slice selected for the current route.
   */
  routeType: SupportedTelemetryRoute["type"];
  /**
   * Concrete telemetry state rendered under the selected provider.
   */
  children: ReactElement;
};

/**
 * Wraps telemetry route content with the providers required by the current route type.
 */
export const SystemTelemetryPageContent = ({
  children,
  routeType,
}: SystemTelemetryPageContentProps): ReactElement => {
  if (routeType === "system-metric") {
    return (
      <MonitoringPollingSettingsProvider storageKey="system-metrics">
        <SystemMetricProvider>{children}</SystemMetricProvider>
      </MonitoringPollingSettingsProvider>
    );
  }

  if (routeType === "zfs-metric") {
    return (
      <MonitoringPollingSettingsProvider storageKey="zfs-metrics">
        <ZFSMetricProvider>{children}</ZFSMetricProvider>
      </MonitoringPollingSettingsProvider>
    );
  }

  return children;
};
