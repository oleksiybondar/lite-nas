import { SystemProcessesLandingPage } from "@pages/SystemProcessesLandingPage";
import { SystemTelemetryPage } from "@pages/SystemTelemetryPage/SystemTelemetryPage";
import { Navigate, type RouteObject } from "react-router-dom";

/**
 * Routes owned by system process and service inspection.
 */
export const processesRoutes: RouteObject[] = [
  {
    element: <SystemProcessesLandingPage />,
    path: "/system/processes",
  },
  {
    element: <Navigate replace to="/security/sessions" />,
    path: "/system/processes/sessions",
  },
  {
    element: <SystemTelemetryPage />,
    path: "/system/processes/:category",
  },
];
