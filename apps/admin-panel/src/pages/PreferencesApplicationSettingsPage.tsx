import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useLanguage } from "@lite-nas/language";
import MonitorHeartRoundedIcon from "@mui/icons-material/MonitorHeartRounded";
import PaletteRoundedIcon from "@mui/icons-material/PaletteRounded";
import type { ReactElement } from "react";

/**
 * Landing page for admin-panel application settings categories.
 */
export const PreferencesApplicationSettingsPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.themeSettingsOverview"),
          icon: <PaletteRoundedIcon />,
          path: "/preferences/application/theme",
          title: getLabel("navigation.Theme"),
        },
        {
          description: getLabel("admin.landing.monitoringSettingsOverview"),
          icon: <MonitorHeartRoundedIcon />,
          path: "/preferences/application/monitoring",
          title: getLabel("navigation.Monitoring"),
        },
      ]}
      overline={getLabel("navigation.Preferences")}
      summary={getLabel("admin.landing.applicationSettingsSummary")}
      title={getLabel("navigation.Application settings")}
    />
  );
};
