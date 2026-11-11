import { buildMonitoringResponseEnvelopeSchema } from "@schemas/monitoring/response-envelope";
import { z } from "zod";

/** Runtime contract for a numeric sensor reading. */
export const sensorNumericSchema = z.object({
  key: z.string(),
  name: z.string(),
  value: z.number(),
  unit: z.string(),
});
/** Runtime contract for a throttling state. */
export const sensorStateSchema = z.object({
  key: z.string(),
  name: z.string(),
  active: z.boolean(),
});
/** Runtime contract for one complete sensor collection cycle. */
export const sensorSnapshotSchema = z.object({
  timestamp: z.string(),
  temperatures: z.array(sensorNumericSchema).nullable(),
  voltages: z.array(sensorNumericSchema).nullable(),
  fans: z.array(sensorNumericSchema).nullable(),
  clocks: z.array(sensorNumericSchema).nullable(),
  throttling: z.array(sensorStateSchema).nullable(),
});
/** A configurable color stop at one measured value. */
export const sensorMarkerSchema = z.object({
  value: z.number().finite(),
  color: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  label: z.string().max(80).default(""),
  showBar: z.boolean().default(false),
  showLabel: z.boolean().default(false),
});
/** Saved color stops indexed by sensor category and key. */
const sensorMarkerListSchema = z.array(sensorMarkerSchema).max(32);
export const sensorMarkersSchema = z.record(z.string(), sensorMarkerListSchema);
/** Chart presentation stored for one sensor in the user profile. */
export const sensorChartSchema = z.object({
  mode: z.enum(["regular", "gradient"]),
  stops: z.array(sensorMarkerSchema).max(32),
});
export const sensorChartsSchema = z.record(z.string(), sensorChartSchema);
export type SensorChart = z.infer<typeof sensorChartSchema>;
export type SensorSnapshot = z.infer<typeof sensorSnapshotSchema>;
export type SensorMarker = z.infer<typeof sensorMarkerSchema>;

/** Accepts only an object with named profile fields. */
const asRecord = (value: unknown): Record<string, unknown> | null =>
  typeof value === "object" && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;

/** Converts each valid legacy stop list without losing other sensors to one bad entry. */
const readLegacySensorCharts = (raw: unknown): Record<string, SensorChart> =>
  Object.fromEntries(
    Object.entries(asRecord(raw) ?? {}).flatMap(([key, value]) => {
      const stops = sensorMarkerListSchema.safeParse(value);
      return stops.success && stops.data.length > 0
        ? [[key, { mode: "gradient", stops: stops.data }]]
        : [];
    }),
  );

/** Keeps only valid current chart settings from a saved profile. */
const readCurrentSensorCharts = (raw: unknown): Record<string, SensorChart> =>
  Object.fromEntries(
    Object.entries(asRecord(raw) ?? {}).flatMap(([key, value]) => {
      const chart = sensorChartSchema.safeParse(value);
      return chart.success ? [[key, chart.data]] : [];
    }),
  );

/** Reads valid chart settings and converts legacy marker lists one sensor at a time. */
export const readSensorCharts = (raw: unknown): Record<string, SensorChart> => {
  const document = asRecord(raw);
  if (document === null) return {};
  return {
    ...readLegacySensorCharts(document.sensorMarkers),
    ...readCurrentSensorCharts(document.sensorCharts),
  };
};

const historySchema = buildMonitoringResponseEnvelopeSchema(
  z.array(sensorSnapshotSchema).nullable(),
);
const snapshotSchema = buildMonitoringResponseEnvelopeSchema(sensorSnapshotSchema);
/** Validates and extracts sensor history. */
export const parseSensorHistory = (value: unknown): SensorSnapshot[] =>
  historySchema.parse(value).data ?? [];
/** Validates and extracts the latest sensor snapshot. */
export const parseSensorSnapshot = (value: unknown): SensorSnapshot =>
  snapshotSchema.parse(value).data;
