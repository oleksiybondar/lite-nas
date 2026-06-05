import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import NotificationImportantIcon from "@mui/icons-material/NotificationImportant";
import PrivacyTipIcon from "@mui/icons-material/PrivacyTip";
import TableRowsIcon from "@mui/icons-material/TableRows";
import type { ReactElement } from "react";

/**
 * Landing page for system alert dashboards.
 */
export const AlertsSystemLandingPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description: "System alerts that still need acknowledgement from an operator.",
          icon: <NotificationImportantIcon />,
          path: "/alerts/system/unacknowledged",
          title: "Unacknowledged alerts",
        },
        {
          description: "Currently active system alerts that still affect runtime operations.",
          icon: <PrivacyTipIcon />,
          path: "/alerts/system/active",
          title: "Active alerts",
        },
        {
          description:
            "Full system alert history across acknowledged, cleared, and active entries.",
          icon: <TableRowsIcon />,
          path: "/alerts/system/all",
          title: "All alerts",
        },
      ]}
      overline="Alerts"
      summary="Inspect system alert queues by lifecycle state, from immediate operator action to complete history."
      title="System"
    />
  );
};
