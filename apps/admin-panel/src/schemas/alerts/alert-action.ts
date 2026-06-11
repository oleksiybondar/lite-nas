import { z } from "zod";

/**
 * Runtime schema for alert action response envelopes returned by the gateway.
 */
export const alertActionResponseSchema = z.object({
  code: z.string().optional(),
  data: z.record(z.string(), z.never()).optional(),
  message: z.string().optional(),
  request_id: z.string().optional(),
  success: z.boolean(),
  timestamp: z.string(),
  trace_id: z.string().optional(),
});
