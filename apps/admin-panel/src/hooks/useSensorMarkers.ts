import type { ApiContextValue } from "@dto/api/api";
import { collectAdminSettings } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import {
  type SensorChart,
  sensorChartSchema,
  sensorChartsSchema,
  sensorMarkersSchema,
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

/** Converts validated saved charts and legacy stops into chart settings. */
const readSensorCharts = (raw: unknown): Record<string, SensorChart> => {
  if (typeof raw !== "object" || raw === null) return {};
  const document = raw as Record<string, unknown>;
  const valid = readLegacyCharts(document.sensorMarkers);
  const saved = document.sensorCharts;
  if (typeof saved !== "object" || saved === null || Array.isArray(saved)) return valid;
  for (const [key, value] of Object.entries(saved)) {
    valid[key] = readSavedChart(key, value);
  }
  return valid;
};

/** Keeps valid legacy marker sets for profiles saved before chart settings existed. */
const readLegacyCharts = (raw: unknown): Record<string, SensorChart> => {
  const legacy = sensorMarkersSchema.safeParse(raw);
  if (!legacy.success) return {};
  const charts: Record<string, SensorChart> = {};
  for (const [key, stops] of Object.entries(legacy.data)) {
    if (stops.length > 0) charts[key] = { mode: "gradient", stops };
  }
  return charts;
};

/** Falls back to a sensor default when its stored chart is invalid. */
const readSavedChart = (key: string, value: unknown): SensorChart => {
  const parsed = sensorChartSchema.safeParse(value);
  return parsed.success ? parsed.data : defaultSensorChart(key);
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
  next: Record<string, SensorChart>,
): Promise<Record<string, SensorChart>> => {
  const response = await api.get("/f/profile/admin-panel").execute();
  const document = await mergeSensorChartDocument(response, next);
  const saved = await api.put("/f/profile/admin-panel", { document }).execute();
  if (!saved.ok) throw new Error("Save failed");
  return next;
};

/** Validates an existing profile before merging its sensor chart changes. */
const mergeSensorChartDocument = async (response: Response, next: Record<string, SensorChart>) => {
  if (!response.ok)
    return adminSettingsSchema.parse({ ...collectAdminSettings(), sensorCharts: next });
  const raw: unknown = await response.json();
  if (typeof raw !== "object" || raw === null || Array.isArray(raw))
    throw new Error("Invalid profile");
  const parsed = adminSettingsSchema.safeParse({ ...raw, sensorCharts: next });
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
      const next = sensorChartsSchema.parse({ ...charts, [key]: chart });
      await persistSensorChart({ get, put }, next);
      setCharts(next);
      setSaveError(false);
    } catch {
      setSaveError(true);
      throw new Error("Save failed");
    }
  };

  return { charts, saveError, saveChart };
};
