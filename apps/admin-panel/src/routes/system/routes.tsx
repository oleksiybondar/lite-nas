import { OperationalSafetyAutomationPage } from "@pages/OperationalSafetyAutomationPage";
import { SystemLandingPage } from "@pages/SystemLandingPage";
import { OperationalSafetyAutomationProvider } from "@providers/OperationalSafetyAutomationProvider";
import { performanceRoutes } from "@routes/system/performance-routes";
import { processesRoutes } from "@routes/system/processes-routes";
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
  ...processesRoutes,
  ...sensorsRoutes,
  {
    element: (
      <OperationalSafetyAutomationProvider>
        <OperationalSafetyAutomationPage />
      </OperationalSafetyAutomationProvider>
    ),
    path: "/system/automation",
  },
];
