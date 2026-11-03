import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import FactCheckIcon from "@mui/icons-material/FactCheck";
import TableRowsIcon from "@mui/icons-material/TableRows";
import type { ReactElement } from "react";

/** Presents the journal and automated detection catalog to security users. */
export const LogsLandingPage = (): ReactElement | null => {
  const { requireSecurity } = useRbac();
  const { getLabel } = useLanguage();
  if (!requireSecurity()) return null;

  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.securityJournal"),
          icon: <TableRowsIcon />,
          path: "/security/logs/journal",
          title: getLabel("navigation.Journal"),
        },
        {
          description: getLabel("admin.logAudit.summary"),
          icon: <FactCheckIcon />,
          path: "/security/logs/automated-audit",
          title: getLabel("navigation.Automated Audit"),
        },
      ]}
      overline={getLabel("navigation.Security")}
      summary={getLabel("admin.landing.securityLogs")}
      title={getLabel("navigation.Logs")}
    />
  );
};
