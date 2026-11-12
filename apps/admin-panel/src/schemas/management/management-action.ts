import { z } from "zod";

/**
 * Runtime schema for a completed service management action response.
 */
export const managementActionResponseSchema = z.object({
  completed: z.literal(true),
  success: z.literal(true),
  timestamp: z.string(),
});
