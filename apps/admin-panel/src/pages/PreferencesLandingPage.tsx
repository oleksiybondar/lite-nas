import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useLanguage } from "@lite-nas/language";
import ManageAccountsRoundedIcon from "@mui/icons-material/ManageAccountsRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import type { ReactElement } from "react";

/**
 * Landing page for authenticated user preferences.
 */
export const PreferencesLandingPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.profileOverview"),
          icon: <ManageAccountsRoundedIcon />,
          path: "/preferences/profile",
          title: getLabel("navigation.User profile"),
        },
        {
          description: getLabel("admin.landing.applicationSettingsOverview"),
          icon: <SettingsRoundedIcon />,
          path: "/preferences/application",
          title: getLabel("navigation.Application settings"),
        },
      ]}
      overline={getLabel("navigation.Preferences")}
      summary={getLabel("admin.landing.preferencesSummary")}
      title={getLabel("navigation.Preferences")}
    />
  );
};
