import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import DeviceThermostatRoundedIcon from "@mui/icons-material/DeviceThermostatRounded";
import PrecisionManufacturingIcon from "@mui/icons-material/PrecisionManufacturing";
import SpeedRoundedIcon from "@mui/icons-material/SpeedRounded";
import TerminalIcon from "@mui/icons-material/Terminal";
import type { ReactElement } from "react";

/**
 * Landing page for system-level administration categories.
 */
export const SystemLandingPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  const { requireAdmin } = useRbac();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.systemPerformance"),
          icon: <SpeedRoundedIcon />,
          path: "/system/performance",
          title: getLabel("navigation.Performance"),
        },
        {
          description: getLabel("admin.landing.systemProcesses"),
          icon: <TerminalIcon />,
          path: "/system/processes",
          title: getLabel("navigation.Processes"),
        },
        {
          description: getLabel("admin.landing.systemSensors"),
          icon: <DeviceThermostatRoundedIcon />,
          path: "/system/sensors",
          title: getLabel("navigation.Sensors"),
        },
        ...(requireAdmin()
          ? [
              {
                description: getLabel("admin.landing.systemAutomation"),
                icon: <PrecisionManufacturingIcon />,
                path: "/system/automation",
                title: getLabel("navigation.Automation"),
              },
            ]
          : []),
      ]}
      overline={getLabel("navigation.System")}
      summary={getLabel("admin.landing.systemSummary")}
      title={getLabel("navigation.System")}
    />
  );
};
