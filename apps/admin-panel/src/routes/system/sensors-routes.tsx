import { SystemSensorsLandingPage } from "@pages/SystemSensorsLandingPage";
import { SystemTelemetryPage } from "@pages/SystemTelemetryPage";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by Raspberry Pi sensor telemetry.
 */
export const sensorsRoutes: RouteObject[] = [
  {
    element: <SystemSensorsLandingPage />,
    path: "/system/sensors",
  },
  {
    element: <SystemTelemetryPage />,
    path: "/system/sensors/:category",
  },
];
