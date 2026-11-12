import type {
  ProcessMetricSnapshotDTO,
  ProcessMetricSnapshotResponseDTO,
} from "@dto/monitoring/process-metric";
import { buildMonitoringResponseEnvelopeSchema } from "@schemas/monitoring/response-envelope";
import { z } from "zod";

/**
 * Runtime schema for one process CPU usage block.
 */
export const processMetricCpuSchema = z.object({
  cpu_pct: z.number(),
  system_ticks: z.number(),
  total_ticks: z.number(),
  user_ticks: z.number(),
});

/**
 * Runtime schema for one process memory usage block.
 */
export const processMetricMemorySchema = z.object({
  rss_bytes: z.number(),
  vms_bytes: z.number(),
});

/**
 * Runtime schema for one process snapshot row.
 */
export const processMetricProcessSchema = z.object({
  cmdline: z.string().optional().nullable(),
  cpu: processMetricCpuSchema,
  cwd: z.string().optional().nullable(),
  exe: z.string().optional().nullable(),
  gid: z.number(),
  memory: processMetricMemorySchema,
  name: z.string(),
  open_fds: z.number(),
  pid: z.number(),
  ppid: z.number(),
  start_time: z.string(),
  state: z.string(),
  threads: z.number(),
  uid: z.number(),
  username: z.string().optional().nullable(),
});

/**
 * Runtime schema for one timestamped process metrics snapshot item.
 */
export const processMetricSnapshotSchema = z.object({
  processes: z.array(processMetricProcessSchema).optional().nullable(),
  timestamp: z.string(),
});

/**
 * Runtime schema for a process metrics snapshot response envelope.
 */
export const processMetricSnapshotResponseSchema = buildMonitoringResponseEnvelopeSchema(
  processMetricSnapshotSchema,
);

/**
 * Parses a process metrics snapshot transport response into one browser-facing item.
 */
export const parseProcessMetricSnapshotResponse = (value: unknown): ProcessMetricSnapshotDTO => {
  const response = processMetricSnapshotResponseSchema.parse(
    value,
  ) as ProcessMetricSnapshotResponseDTO;

  return {
    ...response.data,
    processes: response.data.processes ?? [],
  };
};
