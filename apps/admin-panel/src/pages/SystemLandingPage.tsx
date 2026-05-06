import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import DeviceThermostatRoundedIcon from "@mui/icons-material/DeviceThermostatRounded";
import SpeedRoundedIcon from "@mui/icons-material/SpeedRounded";
import type { ReactElement } from "react";

/**
 * Landing page for system-level administration categories.
 */
export const SystemLandingPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description: "CPU, memory, disk, network, and ZFS runtime telemetry.",
          icon: <SpeedRoundedIcon />,
          path: "/system/performance",
          title: "Performance",
        },
        {
          description: "Raspberry Pi temperature, voltage, clocks, throttling, and fan state.",
          icon: <DeviceThermostatRoundedIcon />,
          path: "/system/sensors",
          title: "Sensors",
        },
      ]}
      overline="System"
      summary="Inspect runtime health, host telemetry, and board-level signals from one place."
      title="System"
    />
  );
};
