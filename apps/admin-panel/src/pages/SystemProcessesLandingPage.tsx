import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import MiscellaneousServicesIcon from "@mui/icons-material/MiscellaneousServices";
import TerminalIcon from "@mui/icons-material/Terminal";
import type { ReactElement } from "react";

/**
 * Landing page for process and service runtime inspection.
 */
export const SystemProcessesLandingPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description: "Inspect runtime processes, resource-heavy workloads, and execution state.",
          icon: <TerminalIcon />,
          path: "/system/processes/processes",
          title: "Processes",
        },
        {
          description: "Review system services, startup state, and managed unit health.",
          icon: <MiscellaneousServicesIcon />,
          path: "/system/processes/services",
          title: "Services",
        },
      ]}
      overline="System"
      summary="Track operating-system workloads from individual processes up to managed services."
      title="Processes"
    />
  );
};
