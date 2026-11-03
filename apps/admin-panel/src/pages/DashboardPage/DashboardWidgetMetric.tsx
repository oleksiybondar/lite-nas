import { DiskDashboardWidget } from "@pages/DiskDashboardWidget";
import { NetworkDashboardWidget } from "@pages/NetworkDashboardWidget";
import { ProcessesDashboardWidget } from "@pages/ProcessesDashboardWidget";
import { ServicesDashboardWidget } from "@pages/ServicesDashboardWidget";
import { SystemDashboardWidget } from "@pages/SystemDashboardWidget";
import { ZFSDashboardWidget } from "@pages/ZFSDashboardWidget";
import { DiskMetricProvider } from "@providers/DiskMetricProvider";
import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import { NetworkMetricProvider } from "@providers/NetworkMetricProvider";
import { ProcessMetricProvider } from "@providers/ProcessMetricProvider";
import { ServiceMetricProvider } from "@providers/ServiceMetricProvider";
import { SystemMetricProvider } from "@providers/SystemMetricProvider";
import { ZFSMetricProvider } from "@providers/ZFSMetricProvider";
import type { dashboardSchema } from "@schemas/profile/admin-settings";
import type { ReactElement } from "react";
import type { z } from "zod";

type Widget = z.infer<typeof dashboardSchema>["widgets"][number];

const systemKinds = new Set(["cpu", "cpu-cores", "ram"]);
const zfsKinds = new Set(["zfs-io", "zfs-errors"]);

/** Mounts the polling provider and chart for one dashboard metric kind. */
export const DashboardWidgetMetric = ({ kind }: { kind: Widget["kind"] }): ReactElement => {
  if (systemKinds.has(kind))
    return (
      <MonitoringPollingSettingsProvider storageKey="system-metrics">
        <SystemMetricProvider>
          <SystemDashboardWidget kind={kind as "cpu" | "cpu-cores" | "ram"} />
        </SystemMetricProvider>
      </MonitoringPollingSettingsProvider>
    );
  if (zfsKinds.has(kind))
    return (
      <MonitoringPollingSettingsProvider storageKey="zfs-metrics">
        <ZFSMetricProvider>
          <ZFSDashboardWidget kind={kind as "zfs-io" | "zfs-errors"} />
        </ZFSMetricProvider>
      </MonitoringPollingSettingsProvider>
    );
  if (kind === "network")
    return (
      <MonitoringPollingSettingsProvider storageKey="network-metrics">
        <NetworkMetricProvider>
          <NetworkDashboardWidget />
        </NetworkMetricProvider>
      </MonitoringPollingSettingsProvider>
    );
  if (kind === "disk")
    return (
      <MonitoringPollingSettingsProvider storageKey="disk-metrics">
        <DiskMetricProvider>
          <DiskDashboardWidget />
        </DiskMetricProvider>
      </MonitoringPollingSettingsProvider>
    );
  if (kind === "processes")
    return (
      <MonitoringPollingSettingsProvider storageKey="process-metrics">
        <ProcessMetricProvider>
          <ProcessesDashboardWidget />
        </ProcessMetricProvider>
      </MonitoringPollingSettingsProvider>
    );
  return (
    <MonitoringPollingSettingsProvider storageKey="service-metrics">
      <ServiceMetricProvider>
        <ServicesDashboardWidget />
      </ServiceMetricProvider>
    </MonitoringPollingSettingsProvider>
  );
};
