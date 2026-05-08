import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import ManageAccountsRoundedIcon from "@mui/icons-material/ManageAccountsRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import type { ReactElement } from "react";

/**
 * Landing page for authenticated user preferences.
 */
export const PreferencesLandingPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description: "Review the authenticated Unix account represented by this browser session.",
          icon: <ManageAccountsRoundedIcon />,
          path: "/preferences/profile",
          title: "User profile",
        },
        {
          description: "Adjust admin-panel behavior, appearance, and local interface preferences.",
          icon: <SettingsRoundedIcon />,
          path: "/preferences/application",
          title: "Application settings",
        },
      ]}
      overline="Preferences"
      summary="Manage settings that belong to your current browser session and admin-panel experience."
      title="Preferences"
    />
  );
};
