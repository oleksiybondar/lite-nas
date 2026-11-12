import { DashboardMetricProvider } from "@pages/DashboardPage/DashboardMetricProvider";
import { DiskDashboardWidget } from "@pages/DiskDashboardWidget";
import { NetworkDashboardWidget } from "@pages/NetworkDashboardWidget";
import { ProcessesDashboardWidget } from "@pages/ProcessesDashboardWidget";
import { SensorDashboardWidget } from "@pages/SensorDashboardWidget";
import { ServicesDashboardWidget } from "@pages/ServicesDashboardWidget";
import { SystemDashboardWidget } from "@pages/SystemDashboardWidget";
import { ZFSDashboardWidget } from "@pages/ZFSDashboardWidget";
import type { dashboardSchema } from "@schemas/profile/admin-settings";
import type { ReactElement } from "react";
import type { z } from "zod";

type Widget = z.infer<typeof dashboardSchema>["widgets"][number];

/** Renders the chart for one dashboard metric through its existing source context. */
export const DashboardWidgetMetric = ({
  kind,
  sensor,
}: {
  kind: Widget["kind"];
  sensor?: string | undefined;
}): ReactElement => {
  let content: ReactElement;
  switch (kind) {
    case "cpu":
    case "cpu-cores":
    case "ram":
      content = <SystemDashboardWidget kind={kind} />;
      break;
    case "zfs-io":
    case "zfs-errors":
      content = <ZFSDashboardWidget kind={kind} />;
      break;
    case "network":
      content = <NetworkDashboardWidget />;
      break;
    case "disk":
      content = <DiskDashboardWidget />;
      break;
    case "processes":
      content = <ProcessesDashboardWidget />;
      break;
    case "sensor":
      content = <SensorDashboardWidget sensor={sensor ?? ""} />;
      break;
    case "services":
      content = <ServicesDashboardWidget />;
      break;
  }
  return <DashboardMetricProvider kind={kind}>{content}</DashboardMetricProvider>;
};
