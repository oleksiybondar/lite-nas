import { AlertsDashboardRoute } from "@pages/AlertsDashboardRoute";
import { AlertsSystemLandingPage } from "@pages/AlertsSystemLandingPage";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by system alerts dashboards.
 */
export const systemAlertsRoutes: RouteObject[] = [
  {
    element: <AlertsSystemLandingPage />,
    path: "/alerts/system",
  },
  {
    element: <AlertsDashboardRoute category="unacknowledged" domain="system" />,
    path: "/alerts/system/unacknowledged",
  },
  {
    element: <AlertsDashboardRoute category="active" domain="system" />,
    path: "/alerts/system/active",
  },
  {
    element: <AlertsDashboardRoute category="all" domain="system" />,
    path: "/alerts/system/all",
  },
];
