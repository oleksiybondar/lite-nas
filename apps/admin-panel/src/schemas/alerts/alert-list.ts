import { z } from "zod";

/**
 * Runtime schema for one alert row returned by the gateway alerts API.
 */
export const alertListItemSchema = z.object({
  Acknowledged: z.boolean(),
  AcknowledgedAt: z.string(),
  AcknowledgedBy: z.string(),
  Category: z.string(),
  CreatedAt: z.string(),
  EventID: z.string(),
  EventRecID: z.number(),
  LastEventID: z.string().nullable(),
  LastEventRecID: z.number().nullable(),
  LastRecID: z.number().nullable(),
  LastTimestamp: z.string().nullable(),
  LastValueBool: z.boolean().nullable(),
  LastValueNum: z.number().nullable(),
  LastValueText: z.string().nullable(),
  LastValueType: z.string().nullable(),
  LastValueUnit: z.string().nullable(),
  Message: z.string(),
  Meta: z.record(z.string(), z.string()).optional(),
  Muted: z.boolean(),
  MutedAt: z.string(),
  MutedBy: z.string(),
  Priority: z.number(),
  RecID: z.number(),
  Severity: z.enum(["critical", "error", "info", "warning"]),
  Source: z.string(),
  Status: z.string(),
});

/**
 * Runtime schema for browser-facing alert list pagination metadata.
 */
export const alertListMetadataSchema = z.object({
  page: z.number(),
  size: z.number(),
  total_count: z.number(),
  total_pages: z.number(),
});

/**
 * Runtime schema for browser-facing alert list payloads.
 */
export const alertListSchema = z.object({
  items: z.array(alertListItemSchema),
  metadata: alertListMetadataSchema,
});

/**
 * Runtime schema for alert list response envelopes returned by the gateway.
 */
export const alertListResponseSchema = z.object({
  code: z.string().optional(),
  data: alertListSchema,
  message: z.string().optional(),
  request_id: z.string().optional(),
  success: z.boolean(),
  timestamp: z.string(),
  trace_id: z.string().optional(),
});
