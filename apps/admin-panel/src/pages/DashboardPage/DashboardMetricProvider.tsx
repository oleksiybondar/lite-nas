import { DiskMetricProvider } from "@providers/DiskMetricProvider";
import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import { NetworkMetricProvider } from "@providers/NetworkMetricProvider";
import { ProcessMetricProvider } from "@providers/ProcessMetricProvider";
import { SensorMetricProvider } from "@providers/SensorMetricProvider";
import { ServiceMetricProvider } from "@providers/ServiceMetricProvider";
import { SystemMetricProvider } from "@providers/SystemMetricProvider";
import { ZFSMetricProvider } from "@providers/ZFSMetricProvider";
import type { dashboardSchema } from "@schemas/profile/admin-settings";
import type { ComponentType, PropsWithChildren, ReactElement } from "react";
import type { z } from "zod";

type WidgetKind = z.infer<typeof dashboardSchema>["widgets"][number]["kind"];
type MetricSource =
  | "system-metrics"
  | "zfs-metrics"
  | "network-metrics"
  | "disk-metrics"
  | "process-metrics"
  | "service-metrics"
  | "sensor-monitor";

const sourceByKind: Record<WidgetKind, MetricSource> = {
  cpu: "system-metrics",
  "cpu-cores": "system-metrics",
  ram: "system-metrics",
  "zfs-io": "zfs-metrics",
  "zfs-errors": "zfs-metrics",
  network: "network-metrics",
  disk: "disk-metrics",
  processes: "process-metrics",
  services: "service-metrics",
  sensor: "sensor-monitor",
};

const providerBySource: Record<MetricSource, ComponentType<PropsWithChildren>> = {
  "system-metrics": SystemMetricProvider,
  "zfs-metrics": ZFSMetricProvider,
  "network-metrics": NetworkMetricProvider,
  "disk-metrics": DiskMetricProvider,
  "process-metrics": ProcessMetricProvider,
  "service-metrics": ServiceMetricProvider,
  "sensor-monitor": SensorMetricProvider,
};

/** Mounts the existing source context and polling settings around any dashboard tile. */
export const DashboardMetricProvider = ({
  children,
  kind,
}: PropsWithChildren<{ kind: WidgetKind }>): ReactElement => {
  const source = sourceByKind[kind];
  const Provider = providerBySource[source];
  return (
    <MonitoringPollingSettingsProvider storageKey={source}>
      <Provider>{children}</Provider>
    </MonitoringPollingSettingsProvider>
  );
};
