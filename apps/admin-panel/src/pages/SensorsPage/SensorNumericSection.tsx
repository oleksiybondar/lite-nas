import { defaultSensorChart } from "@hooks/useSensorMarkers";
import { useLanguage } from "@lite-nas/language";
import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { SensorCard } from "@pages/SensorCard";
import type { SensorChart, SensorSnapshot } from "@schemas/monitoring/sensor";
import type { ReactElement } from "react";

type Section = "temperatures" | "voltages" | "clocks" | "fans";

/** Displays and searches one numeric sensor category from the shared snapshot. */
export const SensorNumericSection = ({
  section,
  items,
  charts,
  search,
  onSearch,
  onEdit,
}: {
  section: Section;
  items: SensorSnapshot[];
  charts: Record<string, SensorChart>;
  search: string;
  onSearch: (value: string) => void;
  onEdit: (key: string, name: string) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const query = search.trim().toLocaleLowerCase();
  const readings = (items.at(-1)?.[section] ?? []).filter((reading) =>
    `${reading.key} ${reading.name}`.toLocaleLowerCase().includes(query),
  );
  return (
    <Stack data-test-class="sensor-section" data-test-name={section} spacing={2}>
      <Typography variant="h2">{getLabel(`admin.sensors.${section}`)}</Typography>
      <TextField
        data-test-class="sensor-search"
        data-test-name={section}
        label={getLabel("admin.sensors.search")}
        onChange={(event) => onSearch(event.target.value)}
        size="small"
        value={search}
      />
      <Grid container spacing={2}>
        {readings.map((reading) => {
          const markerKey = `${section}:${reading.key}`;
          return (
            <Grid key={reading.key} size={{ xs: 12, md: 6 }}>
              <SensorCard
                capacity={Math.max(items.length, 60)}
                items={items}
                markerKey={markerKey}
                chart={charts[markerKey] ?? defaultSensorChart(markerKey)}
                onEdit={() => onEdit(markerKey, reading.name)}
                reading={reading}
                section={section}
              />
            </Grid>
          );
        })}
      </Grid>
      {readings.length === 0 && (
        <Typography color="text.secondary">{getLabel("admin.sensors.empty")}</Typography>
      )}
    </Stack>
  );
};
