import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useLanguage } from "@lite-nas/language";
import NotificationImportantIcon from "@mui/icons-material/NotificationImportant";
import PrivacyTipIcon from "@mui/icons-material/PrivacyTip";
import TableRowsIcon from "@mui/icons-material/TableRows";
import type { ReactElement } from "react";

/**
 * Landing page for system alert dashboards.
 */
export const AlertsSystemLandingPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.systemAlertsUnacknowledged"),
          icon: <NotificationImportantIcon />,
          path: "/alerts/system/unacknowledged",
          title: getLabel("navigation.Unacknowledged alerts"),
        },
        {
          description: getLabel("admin.landing.systemAlertsActive"),
          icon: <PrivacyTipIcon />,
          path: "/alerts/system/active",
          title: getLabel("navigation.Active alerts"),
        },
        {
          description: getLabel("admin.landing.systemAlertsHistory"),
          icon: <TableRowsIcon />,
          path: "/alerts/system/all",
          title: getLabel("navigation.All alerts"),
        },
      ]}
      overline={getLabel("navigation.Alerts")}
      summary={getLabel("admin.landing.systemAlertsSummary")}
      title={getLabel("navigation.System")}
    />
  );
};
