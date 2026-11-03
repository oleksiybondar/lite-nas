import type {
  ServiceMetricSnapshotDTO,
  ServiceMetricSnapshotResponseDTO,
} from "@dto/monitoring/service-metric";
import { buildMonitoringResponseEnvelopeSchema } from "@schemas/monitoring/response-envelope";
import { z } from "zod";

/**
 * Runtime schema for one service resource usage block.
 */
export const serviceMetricResourcesSchema = z.object({
  cpu_usage_usec: z.number().optional().nullable(),
  memory_bytes: z.number().optional().nullable(),
});

/**
 * Runtime schema for one service unit snapshot item.
 */
export const serviceMetricUnitSchema = z.object({
  active_state: z.string().optional().nullable(),
  cgroup: z.string().optional().nullable(),
  control_pid: z.number().optional().nullable(),
  description: z.string().optional().nullable(),
  enabled_state: z.string().optional().nullable(),
  fragment_path: z.string().optional().nullable(),
  load_state: z.string().optional().nullable(),
  main_pid: z.number().optional().nullable(),
  manager: z.string(),
  name: z.string(),
  pids: z.array(z.number()).optional().nullable(),
  resources: serviceMetricResourcesSchema.optional().nullable(),
  result: z.string().optional().nullable(),
  slice: z.string().optional().nullable(),
  started_at: z.string().optional().nullable(),
  sub_state: z.string().optional().nullable(),
  unit_type: z.string(),
});

/**
 * Runtime schema for one timestamped service metrics snapshot item.
 */
export const serviceMetricSnapshotSchema = z.object({
  services: z.array(serviceMetricUnitSchema).optional().nullable(),
  timestamp: z.string(),
});

/**
 * Runtime schema for a service metrics snapshot response envelope.
 */
export const serviceMetricSnapshotResponseSchema = buildMonitoringResponseEnvelopeSchema(
  serviceMetricSnapshotSchema,
);

/**
 * Parses a service metrics snapshot transport response into one browser-facing item.
 */
export const parseServiceMetricSnapshotResponse = (value: unknown): ServiceMetricSnapshotDTO => {
  const response = serviceMetricSnapshotResponseSchema.parse(
    value,
  ) as ServiceMetricSnapshotResponseDTO;

  return {
    ...response.data,
    services: response.data.services ?? [],
  };
};
