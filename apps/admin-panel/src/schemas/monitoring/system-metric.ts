import type {
  SystemMetricHistoryResponseDTO,
  SystemMetricSnapshotDTO,
  SystemMetricSnapshotResponseDTO,
} from "@dto/monitoring/system-metric";
import { buildMonitoringResponseEnvelopeSchema } from "@schemas/monitoring/response-envelope";
import { z } from "zod";

/**
 * Runtime schema for one CPU metrics sample.
 */
export const systemMetricCpuSampleSchema = z.object({
  PerCoreUsage: z.array(z.number()).nullable(),
  TotalUsagePct: z.number(),
});

/**
 * Runtime schema for one memory metrics sample.
 */
export const systemMetricMemSampleSchema = z.object({
  TotalBytes: z.number(),
  UsedBytes: z.number(),
  UsedPct: z.number(),
});

/**
 * Runtime schema for one timestamped system metrics snapshot item.
 */
export const systemMetricSnapshotSchema = z.object({
  CPU: systemMetricCpuSampleSchema,
  Mem: systemMetricMemSampleSchema,
  Timestamp: z.string(),
});

/**
 * Runtime schema for a system metrics history response envelope.
 */
export const systemMetricHistoryResponseSchema = buildMonitoringResponseEnvelopeSchema(
  z.array(systemMetricSnapshotSchema).nullable(),
);

/**
 * Runtime schema for a system metrics snapshot response envelope.
 */
export const systemMetricSnapshotResponseSchema = buildMonitoringResponseEnvelopeSchema(
  systemMetricSnapshotSchema,
);

/**
 * Parses a system metrics history transport response into browser-facing items.
 */
export const parseSystemMetricHistoryResponse = (value: unknown): SystemMetricSnapshotDTO[] => {
  const response = systemMetricHistoryResponseSchema.parse(value) as SystemMetricHistoryResponseDTO;

  return response.data ?? [];
};

/**
 * Parses a system metrics snapshot transport response into one browser-facing item.
 */
export const parseSystemMetricSnapshotResponse = (value: unknown): SystemMetricSnapshotDTO => {
  const response = systemMetricSnapshotResponseSchema.parse(
    value,
  ) as SystemMetricSnapshotResponseDTO;

  return response.data;
};
