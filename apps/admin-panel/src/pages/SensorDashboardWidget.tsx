import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { defaultSensorChart, useSensorMarkers } from "@hooks/useSensorMarkers";
import { useSensorMetric } from "@hooks/useSensorMetric";
import { useLanguage } from "@lite-nas/language";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Shows one saved sensor key as an independent dashboard tile. */
export const SensorDashboardWidget = ({ sensor }: { sensor: string }): ReactElement => {
  const { items } = useSensorMetric();
  const { charts } = useSensorMarkers();
  const { getLabel } = useLanguage();
  const separator = sensor.indexOf(":");
  const section = sensor.slice(0, separator);
  const key = sensor.slice(separator + 1);
  const chart = charts[sensor] ?? defaultSensorChart(sensor);
  if (section === "throttling") {
    return renderThrottlingValue(items, key, getLabel);
  }
  if (!isNumericSection(section))
    return <Typography>{getLabel("admin.dashboard.noData")}</Typography>;
  const samples = items.flatMap((snapshot) => {
    const reading = snapshot[section]?.find((candidate) => candidate.key === key);
    return reading ? [{ stamp: snapshot.timestamp, reading }] : [];
  });
  const reading = samples.at(-1)?.reading;
  return (
    <ValueLineChart
      capacity={Math.max(items.length, 60)}
      colorStops={chart.mode === "gradient" ? chart.stops : []}
      guideStops={chart.stops}
      formatValue={(value) => `${value.toFixed(2)} ${reading?.unit ?? ""}`}
      stamps={samples.map((sample) => sample.stamp)}
      valuesByKey={{
        [reading?.name ?? key ?? sensor]: samples.map((sample) => sample.reading.value),
      }}
    />
  );
};

/** Identifies sensor sections whose snapshots contain numeric readings. */
const isNumericSection = (
  section: string,
): section is "temperatures" | "voltages" | "clocks" | "fans" =>
  section === "temperatures" ||
  section === "voltages" ||
  section === "clocks" ||
  section === "fans";

/** Displays the current state of a saved throttling sensor. */
const renderThrottlingValue = (
  items: ReturnType<typeof useSensorMetric>["items"],
  key: string,
  getLabel: (key: string) => string,
): ReactElement => {
  const state = items.at(-1)?.throttling?.find((candidate) => candidate.key === key);
  return (
    <Typography>
      {state?.name ?? key}: {getLabel(`admin.sensors.${state?.active ? "active" : "clear"}`)}
    </Typography>
  );
};
