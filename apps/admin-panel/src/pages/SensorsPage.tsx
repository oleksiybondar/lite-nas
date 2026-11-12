import { defaultSensorChart, useSensorMarkers } from "@hooks/useSensorMarkers";
import { useSensorMetric } from "@hooks/useSensorMetric";
import { useLanguage } from "@lite-nas/language";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { SensorChartEditor } from "@pages/SensorChartEditor";
import {
  SensorNumericSection,
  SensorPageStatus,
  SensorThrottlingSection,
} from "@pages/SensorsPage/index";
import { type ReactElement, useState } from "react";

const numericSections = ["temperatures", "voltages", "clocks", "fans"] as const;

/** Displays all sensor types from one shared snapshot, with per-section search. */
export const SensorsPage = (): ReactElement => {
  const { items, isLoading, isError } = useSensorMetric();
  const { charts, saveError, saveChart } = useSensorMarkers();
  const [editing, setEditing] = useState<{ key: string; name: string } | null>(null);
  const { getLabel } = useLanguage();
  const [search, setSearch] = useState<Record<string, string>>({});
  const latest = items.at(-1);
  return (
    <Stack data-testid="sensors-page" spacing={3}>
      <Typography variant="h1">{getLabel("navigation.Sensors")}</Typography>
      <SensorPageStatus isLoading={isLoading} isError={isError} saveError={saveError} />
      {numericSections.map((section) => (
        <SensorNumericSection
          key={section}
          section={section}
          items={items}
          charts={charts}
          search={search[section] ?? ""}
          onSearch={(value) => setSearch((current) => ({ ...current, [section]: value }))}
          onEdit={(key, name) => setEditing({ key, name })}
        />
      ))}
      {editing && (
        <SensorChartEditor
          key={editing.key}
          chart={charts[editing.key] ?? defaultSensorChart(editing.key)}
          name={editing.name}
          onClose={() => setEditing(null)}
          onSave={(chart) => saveChart(editing.key, chart)}
        />
      )}
      <SensorThrottlingSection
        latest={latest}
        search={search.throttling ?? ""}
        onSearch={(value) => setSearch((current) => ({ ...current, throttling: value }))}
      />
    </Stack>
  );
};
