import { z } from "zod";

/** Browser-side validation for a journal query before it reaches the gateway. */
export const journalQuerySchema = z.object({
  since: z.string(),
  until: z.string(),
  message: z.string().max(256),
  unit: z
    .string()
    .max(256)
    .refine((value) => value === "" || /^[A-Za-z0-9][A-Za-z0-9_.@:-]*\.service$/.test(value)),
  priority: z.string().regex(/^[0-7]?$/),
  cursor: z
    .string()
    .max(1024)
    .regex(/^[A-Za-z0-9;=_.:-]*$/),
  limit: z.number().int().min(1).max(100),
});

/** Browser-side validation for user-selected journal retention. */
export const journalVacuumSchema = z.object({ days: z.number().int().min(1).max(365) });

/** Runtime contract for selected structured journal fields. */
export const journalEntrySchema = z.object({
  cursor: z.string(),
  timestamp: z.string(),
  message: z.string(),
  unit: z.string().optional(),
  priority: z.string().optional(),
  identifier: z.string().optional(),
  pid: z.string().optional(),
});

/** Runtime contract for a journal chunk returned by the gateway. */
export const journalResponseSchema = z.object({
  success: z.boolean(),
  data: z.object({
    entries: z.array(journalEntrySchema),
    next_cursor: z.string().optional(),
    has_more: z.boolean(),
  }),
});

/** Runtime contract for service units with entries matching journal filters. */
export const journalUnitsResponseSchema = z.object({
  success: z.boolean(),
  data: z.object({ units: z.array(z.string()) }),
});

/** Runtime contract for the completed vacuum response. */
export const journalVacuumResponseSchema = z.object({
  success: z.boolean(),
  completed: z.boolean(),
});

export type JournalEntry = z.infer<typeof journalEntrySchema>;
