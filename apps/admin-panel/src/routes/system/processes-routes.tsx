import { SystemProcessesLandingPage } from "@pages/SystemProcessesLandingPage";
import { SystemTelemetryPage } from "@pages/SystemTelemetryPage/SystemTelemetryPage";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by system process and service inspection.
 */
export const processesRoutes: RouteObject[] = [
  {
    element: <SystemProcessesLandingPage />,
    path: "/system/processes",
  },
  {
    element: <SystemTelemetryPage />,
    path: "/system/processes/:category",
  },
];
