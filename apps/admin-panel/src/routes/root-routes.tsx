import { AppDashboardLayout } from "@components/layout/AppDashboardLayout";
import { DashboardPage } from "@pages/DashboardPage";
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
      ...systemRoutes,
      {
        element: <Navigate replace to="/" />,
        path: "*",
      },
    ],
    element: <AppDashboardLayout />,
  },
];
