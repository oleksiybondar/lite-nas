import { PreferencesLandingPage } from "@pages/PreferencesLandingPage";
import { PreferencesProfilePage } from "@pages/PreferencesProfilePage";
import { PreferencesSecurityPage } from "@pages/PreferencesSecurityPage";
import { applicationPreferencesRoutes } from "@routes/preferences/application-routes";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by authenticated user preferences.
 */
export const preferencesRoutes: RouteObject[] = [
  {
    element: <PreferencesLandingPage />,
    path: "/preferences",
  },
  {
    element: <PreferencesProfilePage />,
    path: "/preferences/profile",
  },
  {
    element: <PreferencesSecurityPage />,
    path: "/preferences/security",
  },
  ...applicationPreferencesRoutes,
];
