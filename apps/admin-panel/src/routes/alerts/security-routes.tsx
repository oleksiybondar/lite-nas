import { AlertsDashboardPage } from "@pages/AlertsDashboardPage";
import { AlertsSecurityLandingPage } from "@pages/AlertsSecurityLandingPage";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by security alerts dashboards.
 */
export const securityAlertsRoutes: RouteObject[] = [
  {
    element: <AlertsSecurityLandingPage />,
    path: "/alerts/security",
  },
  {
    element: <AlertsDashboardPage />,
    path: "/alerts/security/:category",
  },
];
