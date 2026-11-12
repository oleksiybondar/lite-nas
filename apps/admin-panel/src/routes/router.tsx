import { LoginPage } from "@pages/LoginPage";
import { AuthRouteGuard } from "@routes/AuthRouteGuard";
import { rootRoutes } from "@routes/root-routes";
import { createBrowserRouter } from "react-router-dom";

export const router = createBrowserRouter([
  {
    children: [{ path: "/login", element: <LoginPage /> }, ...rootRoutes],
    element: <AuthRouteGuard />,
  },
]);
