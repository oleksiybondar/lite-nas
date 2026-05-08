import { PreferencesApplicationSettingsPage } from "@pages/PreferencesApplicationSettingsPage";
import { PreferencesLandingPage } from "@pages/PreferencesLandingPage";
import { PreferencesProfilePage } from "@pages/PreferencesProfilePage";
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
    element: <PreferencesApplicationSettingsPage />,
    path: "/preferences/application",
  },
];
