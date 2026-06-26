import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import MonitorHeartRoundedIcon from "@mui/icons-material/MonitorHeartRounded";
import PaletteRoundedIcon from "@mui/icons-material/PaletteRounded";
import type { ReactElement } from "react";

/**
 * Landing page for admin-panel application settings categories.
 */
export const PreferencesApplicationSettingsPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description:
            "Adjust the admin-panel theme source, mode, and template for this browser session.",
          icon: <PaletteRoundedIcon />,
          path: "/preferences/application/theme",
          title: "Theme",
        },
        {
          description:
            "Configure metrics history polling and separate process or service snapshot refresh intervals.",
          icon: <MonitorHeartRoundedIcon />,
          path: "/preferences/application/monitoring",
          title: "Monitoring",
        },
      ]}
      overline="Preferences"
      summary="Manage application-level behavior and presentation without mixing independent settings into one page."
      title="Application settings"
    />
  );
};
