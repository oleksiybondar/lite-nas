import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useLanguage } from "@lite-nas/language";
import DeviceThermostatRoundedIcon from "@mui/icons-material/DeviceThermostatRounded";
import ElectricBoltRoundedIcon from "@mui/icons-material/ElectricBoltRounded";
import SpeedRoundedIcon from "@mui/icons-material/SpeedRounded";
import WarningAmberRoundedIcon from "@mui/icons-material/WarningAmberRounded";
import type { ReactElement } from "react";

/**
 * Landing page for Raspberry Pi sensor telemetry.
 */
export const SystemSensorsLandingPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.sensorTemperature"),
          icon: <DeviceThermostatRoundedIcon />,
          path: "/system/sensors/temperature",
          title: getLabel("navigation.Temperature"),
        },
        {
          description: getLabel("admin.landing.sensorVoltage"),
          icon: <ElectricBoltRoundedIcon />,
          path: "/system/sensors/voltage",
          title: getLabel("navigation.Voltage"),
        },
        {
          description: getLabel("admin.landing.sensorClock"),
          icon: <SpeedRoundedIcon />,
          path: "/system/sensors/clock",
          title: getLabel("navigation.Clock"),
        },
        {
          description: getLabel("admin.landing.sensorThrottling"),
          icon: <WarningAmberRoundedIcon />,
          path: "/system/sensors/throttling",
          title: getLabel("navigation.Throttling"),
        },
        {
          description: getLabel("admin.landing.sensorFan"),
          icon: <SpeedRoundedIcon />,
          path: "/system/sensors/fan",
          title: getLabel("navigation.Fan"),
        },
      ]}
      overline={getLabel("navigation.System")}
      summary={getLabel("admin.landing.sensorsSummary")}
      title={getLabel("navigation.Sensors")}
    />
  );
};
