import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useLanguage } from "@lite-nas/language";
import NotificationImportantIcon from "@mui/icons-material/NotificationImportant";
import PrivacyTipIcon from "@mui/icons-material/PrivacyTip";
import TableRowsIcon from "@mui/icons-material/TableRows";
import type { ReactElement } from "react";

/**
 * Landing page for security alert dashboards.
 */
export const AlertsSecurityLandingPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.securityAlertsUnacknowledged"),
          icon: <NotificationImportantIcon />,
          path: "/alerts/security/unacknowledged",
          title: getLabel("navigation.Unacknowledged alerts"),
        },
        {
          description: getLabel("admin.landing.securityAlertsActive"),
          icon: <PrivacyTipIcon />,
          path: "/alerts/security/active",
          title: getLabel("navigation.Active alerts"),
        },
        {
          description: getLabel("admin.landing.securityAlertsHistory"),
          icon: <TableRowsIcon />,
          path: "/alerts/security/all",
          title: getLabel("navigation.All alerts"),
        },
      ]}
      overline={getLabel("navigation.Alerts")}
      summary={getLabel("admin.landing.securityAlertsSummary")}
      title={getLabel("navigation.Security")}
    />
  );
};
