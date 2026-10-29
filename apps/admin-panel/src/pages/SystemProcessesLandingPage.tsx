import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useLanguage } from "@lite-nas/language";
import MiscellaneousServicesIcon from "@mui/icons-material/MiscellaneousServices";
import TerminalIcon from "@mui/icons-material/Terminal";
import type { ReactElement } from "react";

/**
 * Landing page for process and service runtime inspection.
 */
export const SystemProcessesLandingPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.processesOverview"),
          icon: <TerminalIcon />,
          path: "/system/processes/processes",
          title: getLabel("navigation.Processes"),
        },
        {
          description: getLabel("admin.landing.servicesOverview"),
          icon: <MiscellaneousServicesIcon />,
          path: "/system/processes/services",
          title: getLabel("navigation.Services"),
        },
      ]}
      overline={getLabel("navigation.System")}
      summary={getLabel("admin.landing.processesSummary")}
      title={getLabel("navigation.Processes")}
    />
  );
};
