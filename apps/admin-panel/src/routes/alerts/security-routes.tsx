import { AlertsDashboardRoute } from "@pages/AlertsDashboardRoute";
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
    element: <AlertsDashboardRoute category="unacknowledged" domain="security" />,
    path: "/alerts/security/unacknowledged",
  },
  {
    element: <AlertsDashboardRoute category="active" domain="security" />,
    path: "/alerts/security/active",
  },
  {
    element: <AlertsDashboardRoute category="all" domain="security" />,
    path: "/alerts/security/all",
  },
];
