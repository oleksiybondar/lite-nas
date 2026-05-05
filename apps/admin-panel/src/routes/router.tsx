import { DashboardPage } from "@pages/DashboardPage";
import { AuthRouteGuard } from "@routes/AuthRouteGuard";
import { createBrowserRouter, Navigate } from "react-router-dom";

export const router = createBrowserRouter([
  {
    children: [
      {
        element: <DashboardPage />,
        path: "/",
      },
      {
        element: <Navigate replace to="/" />,
        path: "*",
      },
    ],
    element: <AuthRouteGuard />,
  },
]);
