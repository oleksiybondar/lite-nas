import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import NotificationImportantIcon from "@mui/icons-material/NotificationImportant";
import PrivacyTipIcon from "@mui/icons-material/PrivacyTip";
import TableRowsIcon from "@mui/icons-material/TableRows";
import type { ReactElement } from "react";

/**
 * Landing page for security alert dashboards.
 */
export const AlertsSecurityLandingPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description: "Security alerts that still require acknowledgement or triage.",
          icon: <NotificationImportantIcon />,
          path: "/alerts/security/unacknowledged",
          title: "Unacknowledged alerts",
        },
        {
          description: "Security alerts that remain active and need ongoing attention.",
          icon: <PrivacyTipIcon />,
          path: "/alerts/security/active",
          title: "Active alerts",
        },
        {
          description: "Complete security alert history for reviews, audits, and investigations.",
          icon: <TableRowsIcon />,
          path: "/alerts/security/all",
          title: "All alerts",
        },
      ]}
      overline="Alerts"
      summary="Review security alert queues by status so investigation and acknowledgement workflows stay separated."
      title="Security"
    />
  );
};
