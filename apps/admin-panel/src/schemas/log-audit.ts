import { z } from "zod";

/** Validates one configured Log Analysis detection returned by the gateway. */
export const logAuditRuleSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  explanation: z.string(),
  possible_reactions: z
    .array(z.string())
    .nullable()
    .transform((reactions) => reactions ?? []),
  enabled: z.boolean(),
  source: z.string(),
  category: z.string(),
  severity: z.string(),
  instance: z.string(),
  threshold: z.number().int(),
  window: z.string(),
  email_on_critical: z.boolean(),
  lockdown: z.boolean(),
});

/** Validates the read-only rule catalog at the browser API boundary. */
export const logAuditRulesResponseSchema = z.object({
  data: z.object({ rules: z.array(logAuditRuleSchema) }),
});

export type LogAuditRule = z.infer<typeof logAuditRuleSchema>;
