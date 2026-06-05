import { AlertsDashboardPage } from "@pages/AlertsDashboardPage";
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
    element: <AlertsDashboardPage />,
    path: "/alerts/system/:category",
  },
];
