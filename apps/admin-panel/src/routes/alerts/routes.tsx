import { AlertsLandingPage } from "@pages/AlertsLandingPage";
import { securityAlertsRoutes } from "@routes/alerts/security-routes";
import { systemAlertsRoutes } from "@routes/alerts/system-routes";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by the alerts dashboard area.
 */
export const alertsRoutes: RouteObject[] = [
  {
    element: <AlertsLandingPage />,
    path: "/alerts",
  },
  ...systemAlertsRoutes,
  ...securityAlertsRoutes,
];
