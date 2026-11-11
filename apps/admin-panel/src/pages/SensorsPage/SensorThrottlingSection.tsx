import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { SensorSnapshot } from "@schemas/monitoring/sensor";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

/** Displays searchable throttling state with dashboard entry points. */
export const SensorThrottlingSection = ({
  latest,
  search,
  onSearch,
}: {
  latest: SensorSnapshot | undefined;
  search: string;
  onSearch: (value: string) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const states = latest?.throttling ?? [];
  const query = search.trim().toLocaleLowerCase();
  const visible = states.filter((state) =>
    `${state.key} ${state.name} ${getLabel(`admin.sensors.throttle.${state.key}`)}`
      .toLocaleLowerCase()
      .includes(query),
  );
  return (
    <Stack spacing={1}>
      <Typography variant="h2">{getLabel("admin.sensors.throttling")}</Typography>
      <TextField
        data-test-class="sensor-search"
        data-test-name="throttling"
        label={getLabel("admin.sensors.search")}
        onChange={(event) => onSearch(event.target.value)}
        size="small"
        value={search}
      />
      {visible.map((state) => (
        <Paper
          data-test-class="sensor-throttling"
          data-test-name={state.key}
          key={state.key}
          sx={{ p: 2 }}
        >
          <Typography>
            {state.name}: {getLabel(`admin.sensors.${state.active ? "active" : "clear"}`)}
          </Typography>
          <Typography color="text.secondary">
            {getLabel(`admin.sensors.throttle.${state.key}`)}
          </Typography>
          <Button
            component={RouterLink}
            to={`/?add=sensor&sensor=${encodeURIComponent(`throttling:${state.key}`)}`}
          >
            {getLabel("admin.dashboard.addToDashboard")}
          </Button>
        </Paper>
      ))}
      {states.length === 0 && <Typography>{getLabel("admin.sensors.empty")}</Typography>}
    </Stack>
  );
};
