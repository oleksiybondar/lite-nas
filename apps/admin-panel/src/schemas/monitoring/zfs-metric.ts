import type {
  ZFSMetricHistoryResponseDTO,
  ZFSMetricSnapshotDTO,
  ZFSMetricSnapshotResponseDTO,
  ZFSMetricVdevSnapshotDTO,
} from "@dto/monitoring/zfs-metric";
import { buildMonitoringResponseEnvelopeSchema } from "@schemas/monitoring/response-envelope";
import { z } from "zod";

/**
 * Runtime schema for one ZFS read/write sample group.
 */
export const zfsMetricIoStatValuesSchema = z.object({
  Read: z.number(),
  Write: z.number(),
});

/**
 * Runtime schema for ZFS I/O error counts.
 */
export const zfsMetricIoErrorsSchema = z.object({
  Checksum: z.number(),
  Read: z.number(),
  Write: z.number(),
});

/**
 * Runtime schema for one ZFS I/O statistics block.
 */
export const zfsMetricIoStatSchema = z.object({
  Bandwidth: zfsMetricIoStatValuesSchema,
  Operations: zfsMetricIoStatValuesSchema,
});

/**
 * Runtime schema for one ZFS usage block.
 */
export const zfsMetricUsageSchema = z.object({
  AllocatedBytes: z.number(),
  CapacityPct: z.number(),
  FreeBytes: z.number(),
  SizeBytes: z.number(),
});

/**
 * Runtime schema for one recursive ZFS vdev snapshot.
 */
export const zfsMetricVdevSnapshotSchema: z.ZodType<ZFSMetricVdevSnapshotDTO> = z.lazy(() =>
  z.object({
    Children: z.array(zfsMetricVdevSnapshotSchema).nullable(),
    Errors: zfsMetricIoErrorsSchema,
    Name: z.string(),
    Path: z.string(),
    Type: z.string(),
  }),
);

/**
 * Runtime schema for one ZFS pool snapshot.
 */
export const zfsMetricPoolSnapshotSchema = z.object({
  Errors: z.string(),
  Health: z.string(),
  IOStat: zfsMetricIoStatSchema,
  Name: z.string(),
  Root: zfsMetricVdevSnapshotSchema,
  Scan: z.string(),
  Usage: zfsMetricUsageSchema,
});

/**
 * Runtime schema for one timestamped ZFS metrics snapshot item.
 */
export const zfsMetricSnapshotSchema = z.object({
  Pools: z.array(zfsMetricPoolSnapshotSchema).nullable(),
  Timestamp: z.string(),
});

/**
 * Runtime schema for a ZFS metrics history response envelope.
 */
export const zfsMetricHistoryResponseSchema = buildMonitoringResponseEnvelopeSchema(
  z.array(zfsMetricSnapshotSchema).nullable(),
);

/**
 * Runtime schema for a ZFS metrics snapshot response envelope.
 */
export const zfsMetricSnapshotResponseSchema =
  buildMonitoringResponseEnvelopeSchema(zfsMetricSnapshotSchema);

/**
 * Parses a ZFS metrics history transport response into browser-facing items.
 */
export const parseZFSMetricHistoryResponse = (value: unknown): ZFSMetricSnapshotDTO[] => {
  const response = zfsMetricHistoryResponseSchema.parse(value) as ZFSMetricHistoryResponseDTO;

  return response.data ?? [];
};

/**
 * Parses a ZFS metrics snapshot transport response into one browser-facing item.
 */
export const parseZFSMetricSnapshotResponse = (value: unknown): ZFSMetricSnapshotDTO => {
  const response = zfsMetricSnapshotResponseSchema.parse(value) as ZFSMetricSnapshotResponseDTO;

  return response.data;
};
