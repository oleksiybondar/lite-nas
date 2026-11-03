import type {
  DiskMetricHistoryResponseDTO,
  DiskMetricSnapshotDTO,
  DiskMetricSnapshotResponseDTO,
} from "@dto/monitoring/disk-metric";
import { buildMonitoringResponseEnvelopeSchema } from "@schemas/monitoring/response-envelope";
import { z } from "zod";

/**
 * Runtime schema for one optional raw device I/O counters block.
 */
export const diskMetricDeviceIoCountersSchema = z.object({
  discards_completed: z.number().nullable().optional(),
  discards_merged: z.number().nullable().optional(),
  discard_time_ms: z.number().nullable().optional(),
  flushes_completed: z.number().nullable().optional(),
  flush_time_ms: z.number().nullable().optional(),
  io_in_progress: z.number(),
  io_time_ms: z.number(),
  reads_completed: z.number(),
  reads_merged: z.number(),
  read_time_ms: z.number(),
  sectors_discarded: z.number().nullable().optional(),
  sectors_read: z.number(),
  sectors_written: z.number(),
  weighted_io_time_ms: z.number(),
  writes_completed: z.number(),
  writes_merged: z.number(),
  write_time_ms: z.number(),
});

/**
 * Runtime schema for one disk device snapshot item.
 */
export const diskMetricDeviceSnapshotSchema = z.object({
  connection_type: z.string(),
  description: z.string().nullable().optional(),
  io: diskMetricDeviceIoCountersSchema.nullable().optional(),
  kind: z.string(),
  major: z.number(),
  minor: z.number(),
  model: z.string().nullable().optional(),
  mounts: z.array(z.string()).nullable().optional(),
  name: z.string().nullable().optional(),
  node: z.string(),
  parent: z.string().nullable().optional(),
  partitions: z.array(z.string()).nullable().optional(),
  removable: z.boolean().nullable().optional(),
  rotational: z.boolean().nullable().optional(),
  serial: z.string().nullable().optional(),
  size_bytes: z.number(),
  vendor: z.string().nullable().optional(),
  wwn: z.string().nullable().optional(),
  zpool_memberships: z.array(z.string()).nullable().optional(),
});

/**
 * Runtime schema for one filesystem summary entry.
 */
export const diskMetricFilesystemSummarySchema = z.object({
  mount_count: z.number(),
  type: z.string(),
});

/**
 * Runtime schema for one active mountpoint snapshot.
 */
export const diskMetricMountSnapshotSchema = z.object({
  available_bytes: z.number().nullable().optional(),
  device: z.string().nullable().optional(),
  filesystem: z.string(),
  free_bytes: z.number().nullable().optional(),
  memory_backed: z.boolean(),
  mountpoint: z.string(),
  readonly: z.boolean(),
  remote: z.boolean(),
  source: z.string(),
  total_bytes: z.number().nullable().optional(),
  used_bytes: z.number().nullable().optional(),
  used_percent: z.number().nullable().optional(),
});

/**
 * Runtime schema for one timestamped disk metrics snapshot item.
 */
export const diskMetricSnapshotSchema = z.object({
  devices: z.array(diskMetricDeviceSnapshotSchema).nullable(),
  filesystems: z.record(z.string(), diskMetricFilesystemSummarySchema).nullable(),
  mounts: z.record(z.string(), diskMetricMountSnapshotSchema).nullable(),
  timestamp: z.string(),
});

/**
 * Runtime schema for a disk metrics history response envelope.
 */
export const diskMetricHistoryResponseSchema = buildMonitoringResponseEnvelopeSchema(
  z.array(diskMetricSnapshotSchema).nullable(),
);

/**
 * Runtime schema for a disk metrics snapshot response envelope.
 */
export const diskMetricSnapshotResponseSchema =
  buildMonitoringResponseEnvelopeSchema(diskMetricSnapshotSchema);

/**
 * Parses a disk metrics history transport response into browser-facing items.
 */
export const parseDiskMetricHistoryResponse = (value: unknown): DiskMetricSnapshotDTO[] => {
  const response = diskMetricHistoryResponseSchema.parse(value) as DiskMetricHistoryResponseDTO;

  return response.data ?? [];
};

/**
 * Parses a disk metrics snapshot transport response into one browser-facing item.
 */
export const parseDiskMetricSnapshotResponse = (value: unknown): DiskMetricSnapshotDTO => {
  const response = diskMetricSnapshotResponseSchema.parse(value) as DiskMetricSnapshotResponseDTO;

  return response.data;
};
