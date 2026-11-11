import type { ApiContextValue } from "@dto/api/api";
import { collectAdminSettings } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import {
  readSensorCharts,
  type SensorChart,
  sensorChartSchema,
  sensorChartsSchema,
} from "@schemas/monitoring/sensor";
import { adminSettingsSchema } from "@schemas/profile/admin-settings";
import { useEffect, useState } from "react";

const temperatureStops = [
  [30, "#00bfff"],
  [40, "#00ff00"],
  [50, "#adff2f"],
  [55, "#ffff00"],
  [60, "#ff9900"],
  [65, "#ff4500"],
  [70, "#ff0000"],
  [75, "#8b0000"],
] as const;

/** Returns the chart default for a sensor category. */
export const defaultSensorChart = (key: string): SensorChart => ({
  mode: key.startsWith("temperatures:") ? "gradient" : "regular",
  stops: key.startsWith("temperatures:")
    ? temperatureStops.map(([value, color]) => ({
        value,
        color,
        label: "",
        showBar: false,
        showLabel: false,
      }))
    : [],
});

type SensorMarkersState = {
  charts: Record<string, SensorChart>;
  saveError: boolean;
  saveChart: (key: string, chart: SensorChart) => Promise<void>;
};

/** Loads a profile document and returns its chart presentation settings. */
const loadSensorCharts = async (
  get: ApiContextValue["get"],
): Promise<Record<string, SensorChart> | null> => {
  const response = await get("/f/profile/admin-panel").execute();
  if (!response.ok) return null;
  return readSensorCharts(await response.json());
};

/** Saves validated chart settings while retaining other profile fields. */
const persistSensorChart = async (
  api: Pick<ApiContextValue, "get" | "put">,
  key: string,
  chart: SensorChart,
): Promise<Record<string, SensorChart>> => {
  const response = await api.get("/f/profile/admin-panel").execute();
  const document = await mergeSensorChartDocument(response, key, chart);
  const saved = await api.put("/f/profile/admin-panel", { document }).execute();
  if (!saved.ok) throw new Error("Save failed");
  return document.sensorCharts ?? {};
};

/** Validates an existing profile before merging its sensor chart changes. */
const mergeSensorChartDocument = async (response: Response, key: string, chart: SensorChart) => {
  if (!response.ok)
    return adminSettingsSchema.parse({
      ...collectAdminSettings(),
      sensorCharts: sensorChartsSchema.parse({ [key]: chart }),
    });
  const raw: unknown = await response.json();
  if (typeof raw !== "object" || raw === null || Array.isArray(raw))
    throw new Error("Invalid profile");
  const charts = sensorChartsSchema.parse({ ...readSensorCharts(raw), [key]: chart });
  const parsed = adminSettingsSchema.safeParse({
    ...collectAdminSettings(),
    ...raw,
    sensorMarkers: undefined,
    sensorCharts: charts,
  });
  if (!parsed.success) throw new Error("Invalid profile");
  return parsed.data;
};

/** Loads and saves per-sensor chart presentation in the authenticated profile. */
export const useSensorMarkers = (): SensorMarkersState => {
  const { get, put } = useApi();
  const [charts, setCharts] = useState<Record<string, SensorChart>>({});
  const [saveError, setSaveError] = useState(false);

  useEffect(() => {
    let active = true;
    void loadSensorCharts(get)
      .then((loaded) => {
        if (active && loaded) setCharts(loaded);
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, [get]);

  const saveChart = async (key: string, chart: SensorChart): Promise<void> => {
    try {
      const validChart = sensorChartSchema.parse(chart);
      const savedCharts = await persistSensorChart({ get, put }, key, validChart);
      setCharts(savedCharts);
      setSaveError(false);
    } catch {
      setSaveError(true);
      throw new Error("Save failed");
    }
  };

  return { charts, saveError, saveChart };
};
