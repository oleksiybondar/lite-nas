import { z } from "zod";

/**
 * Runtime schema for alert-count payloads returned by the gateway.
 */
export const alertCountSchema = z.object({
  count: z.number().int().nonnegative(),
});

/**
 * Runtime schema for alert-count response envelopes returned by the gateway.
 */
export const alertCountResponseSchema = z.object({
  code: z.string().optional(),
  data: alertCountSchema,
  message: z.string().optional(),
  request_id: z.string().optional(),
  success: z.boolean(),
  timestamp: z.string(),
  trace_id: z.string().optional(),
});
