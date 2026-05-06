import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import DeviceThermostatRoundedIcon from "@mui/icons-material/DeviceThermostatRounded";
import ElectricBoltRoundedIcon from "@mui/icons-material/ElectricBoltRounded";
import SpeedRoundedIcon from "@mui/icons-material/SpeedRounded";
import WarningAmberRoundedIcon from "@mui/icons-material/WarningAmberRounded";
import type { ReactElement } from "react";

/**
 * Landing page for Raspberry Pi sensor telemetry.
 */
export const SystemSensorsLandingPage = (): ReactElement => {
  return (
    <CategoryLandingPage
      cards={[
        {
          description: "Board and SoC thermal readings for cooling and load diagnostics.",
          icon: <DeviceThermostatRoundedIcon />,
          path: "/system/sensors/temperature",
          title: "Temperature",
        },
        {
          description: "Voltage signals that can reveal power-supply instability.",
          icon: <ElectricBoltRoundedIcon />,
          path: "/system/sensors/voltage",
          title: "Voltage",
        },
        {
          description: "Clock frequencies for CPU, core, and related Raspberry Pi domains.",
          icon: <SpeedRoundedIcon />,
          path: "/system/sensors/clock",
          title: "Clock",
        },
        {
          description: "Thermal and undervoltage throttling flags reported by the board.",
          icon: <WarningAmberRoundedIcon />,
          path: "/system/sensors/throttling",
          title: "Throttling",
        },
        {
          description: "Fan state and speed signals when active cooling is available.",
          icon: <SpeedRoundedIcon />,
          path: "/system/sensors/fan",
          title: "Fan",
        },
      ]}
      overline="System"
      summary="Track board-level Raspberry Pi signals that affect reliability, cooling, and power behavior."
      title="Sensors"
    />
  );
};
