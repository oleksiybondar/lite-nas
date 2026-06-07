import { PreferencesApplicationSettingsPage } from "@pages/PreferencesApplicationSettingsPage";
import { PreferencesMonitoringSettingsPage } from "@pages/PreferencesMonitoringSettingsPage";
import { PreferencesThemeSettingsPage } from "@pages/PreferencesThemeSettingsPage";
import type { RouteObject } from "react-router-dom";

/**
 * Routes owned by application-level preferences.
 */
export const applicationPreferencesRoutes: RouteObject[] = [
  {
    element: <PreferencesApplicationSettingsPage />,
    path: "/preferences/application",
  },
  {
    element: <PreferencesThemeSettingsPage />,
    path: "/preferences/application/theme",
  },
  {
    element: <PreferencesMonitoringSettingsPage />,
    path: "/preferences/application/monitoring",
  },
];
