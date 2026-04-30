import { DashboardPage } from "@pages/DashboardPage";
import { createBrowserRouter } from "react-router-dom";

export const router = createBrowserRouter([
  {
    element: <DashboardPage />,
    path: "/",
  },
]);
