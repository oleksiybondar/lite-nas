import { AuthRouteGuard } from "@routes/AuthRouteGuard";
import { rootRoutes } from "@routes/root-routes";
import { createBrowserRouter } from "react-router-dom";

export const router = createBrowserRouter([
  {
    children: rootRoutes,
    element: <AuthRouteGuard />,
  },
]);
