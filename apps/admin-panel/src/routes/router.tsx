import { DashboardPage } from "@pages/DashboardPage";
import { createBrowserRouter, Navigate } from "react-router-dom";

export const router = createBrowserRouter([
  {
    element: <DashboardPage />,
    path: "/",
  },
  {
    element: <Navigate replace to="/" />,
    path: "*",
  },
]);
