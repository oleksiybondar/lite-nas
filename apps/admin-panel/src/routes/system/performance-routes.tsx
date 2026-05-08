import { SystemPerformanceLandingPage } from "@pages/SystemPerformanceLandingPage";
import { SystemTelemetryPage } from "@pages/SystemTelemetryPage";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by system performance telemetry.
 */
export const performanceRoutes: RouteObject[] = [
  {
    element: <SystemPerformanceLandingPage />,
    path: "/system/performance",
  },
  {
    element: <SystemTelemetryPage />,
    path: "/system/performance/:category",
  },
];
