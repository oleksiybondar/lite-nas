import { AppDashboardLayout } from "@components/layout/AppDashboardLayout";
import { BackupRestorePage } from "@pages/BackupRestorePage";
import { DashboardPage } from "@pages/DashboardPage";
import { alertsRoutes } from "@routes/alerts/routes";
import { applicationsRoutes } from "@routes/applications/routes";
import { preferencesRoutes } from "@routes/preferences/routes";
import { securityRoutes } from "@routes/security/routes";
import { systemRoutes } from "@routes/system/routes";
import { Navigate, type RouteObject } from "react-router-dom";

/**
 * Protected app routes rendered inside the dashboard layout.
 */
export const rootRoutes: RouteObject[] = [
  {
    children: [
      {
        element: <DashboardPage />,
        path: "/",
      },
      ...alertsRoutes,
      ...applicationsRoutes,
      ...systemRoutes,
      ...securityRoutes,
      { element: <BackupRestorePage />, path: "/backup-restore" },
      ...preferencesRoutes,
      {
        element: <Navigate replace to="/" />,
        path: "*",
      },
    ],
    element: <AppDashboardLayout />,
  },
];
