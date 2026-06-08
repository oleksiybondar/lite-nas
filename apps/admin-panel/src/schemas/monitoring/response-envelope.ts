import { z } from "zod";

/**
 * Builds the common monitoring response envelope schema around one data payload schema.
 */
export const buildMonitoringResponseEnvelopeSchema = <TDataSchema extends z.ZodTypeAny>(
  dataSchema: TDataSchema,
) => {
  return z.object({
    code: z.string().optional(),
    data: dataSchema,
    message: z.string().optional(),
    request_id: z.string().optional(),
    success: z.boolean(),
    timestamp: z.string(),
    trace_id: z.string().optional(),
  });
};
