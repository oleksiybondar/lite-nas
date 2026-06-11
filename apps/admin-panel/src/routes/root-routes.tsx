import { AppDashboardLayout } from "@components/layout/AppDashboardLayout";
import { DashboardPage } from "@pages/DashboardPage";
import { alertsRoutes } from "@routes/alerts/routes";
import { preferencesRoutes } from "@routes/preferences/routes";
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
      ...systemRoutes,
      ...preferencesRoutes,
      {
        element: <Navigate replace to="/" />,
        path: "*",
      },
    ],
    element: <AppDashboardLayout />,
  },
];
