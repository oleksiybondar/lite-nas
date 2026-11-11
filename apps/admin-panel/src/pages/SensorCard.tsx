import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { SensorChart, SensorSnapshot } from "@schemas/monitoring/sensor";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

type Section = "temperatures" | "voltages" | "clocks" | "fans";
type Reading = NonNullable<SensorSnapshot[Section]>[number];
type Props = {
  capacity: number;
  chart: SensorChart;
  items: SensorSnapshot[];
  markerKey: string;
  onEdit: () => void;
  reading: Reading;
  section: Section;
};
/** Displays one sensor's current reading and history with a separate chart editor trigger. */
export const SensorCard = ({
  capacity,
  chart,
  items,
  markerKey,
  onEdit,
  reading,
  section,
}: Props): ReactElement => {
  const { getLabel } = useLanguage();
  const samples = items.flatMap((snapshot) => {
    const value = snapshot[section]?.find((candidate) => candidate.key === reading.key)?.value;
    return value === undefined ? [] : [{ stamp: snapshot.timestamp, value }];
  });
  return (
    <Paper data-test-class="sensor-card" data-test-name={markerKey} sx={{ p: 2 }}>
      <Stack spacing={2}>
        <Typography sx={{ overflowWrap: "anywhere" }} variant="h3">
          {reading.name}
        </Typography>
        <Typography>
          {reading.value} {reading.unit}
        </Typography>
        <Button component={RouterLink} to={`/?add=sensor&sensor=${encodeURIComponent(markerKey)}`}>
          {getLabel("admin.dashboard.addToDashboard")}
        </Button>
        <Button data-test-class="sensor-edit-chart" data-test-name={markerKey} onClick={onEdit}>
          {getLabel("admin.sensors.editChart")}
        </Button>
        <details>
          <summary>{getLabel("admin.sensors.historyAndColors")}</summary>
          <Stack spacing={2} sx={{ pt: 2 }}>
            <Typography color="text.secondary" sx={{ overflowWrap: "anywhere" }} variant="body2">
              {reading.key}
            </Typography>
            <ValueLineChart
              capacity={capacity}
              colorStops={chart.mode === "gradient" ? chart.stops : []}
              formatValue={(value) => `${value.toFixed(2)} ${reading.unit}`}
              guideStops={chart.stops}
              stamps={samples.map((sample) => sample.stamp)}
              valuesByKey={{ [reading.name]: samples.map((sample) => sample.value) }}
            />
          </Stack>
        </details>
      </Stack>
    </Paper>
  );
};
