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
export const sensorMarkersSchema = z.record(z.string(), z.array(sensorMarkerSchema).max(32));
/** Chart presentation stored for one sensor in the user profile. */
export const sensorChartSchema = z.object({
  mode: z.enum(["regular", "gradient"]),
  stops: z.array(sensorMarkerSchema).max(32),
});
export const sensorChartsSchema = z.record(z.string(), sensorChartSchema);
export type SensorChart = z.infer<typeof sensorChartSchema>;
export type SensorSnapshot = z.infer<typeof sensorSnapshotSchema>;
export type SensorMarker = z.infer<typeof sensorMarkerSchema>;
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
