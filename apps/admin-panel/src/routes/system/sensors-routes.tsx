import { SensorsPage } from "@pages/SensorsPage";
import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import { SensorMetricProvider } from "@providers/SensorMetricProvider";
import type { RouteObject } from "react-router-dom";

/** The complete sensor snapshot is rendered on one page. */
export const sensorsRoutes: RouteObject[] = [
  {
    path: "/system/sensors",
    element: (
      <MonitoringPollingSettingsProvider storageKey="sensor-monitor">
        <SensorMetricProvider>
          <SensorsPage />
        </SensorMetricProvider>
      </MonitoringPollingSettingsProvider>
    ),
  },
];
