import { SystemLandingPage } from "@pages/SystemLandingPage";
import { performanceRoutes } from "@routes/system/performance-routes";
import { sensorsRoutes } from "@routes/system/sensors-routes";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by the system administration area.
 */
export const systemRoutes: RouteObject[] = [
  {
    element: <SystemLandingPage />,
    path: "/system",
  },
  ...performanceRoutes,
  ...sensorsRoutes,
];
