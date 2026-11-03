import { z } from "zod";

/**
 * Runtime schema for one alert occurrence row returned by the gateway alerts API.
 */
export const alertOccurrenceItemSchema = z.object({
  EventID: z.string(),
  EventRecID: z.number(),
  RecID: z.number(),
  Timestamp: z.string(),
  ValueBool: z.boolean().nullable(),
  ValueNum: z.number().nullable(),
  ValueText: z.string().nullable(),
  ValueType: z.string(),
  ValueUnit: z.string().nullable(),
});

/**
 * Runtime schema for alert occurrences response envelopes returned by the gateway.
 */
export const alertOccurrencesResponseSchema = z.object({
  code: z.string().optional(),
  data: z.array(alertOccurrenceItemSchema),
  message: z.string().optional(),
  request_id: z.string().optional(),
  success: z.boolean(),
  timestamp: z.string(),
  trace_id: z.string().optional(),
});
