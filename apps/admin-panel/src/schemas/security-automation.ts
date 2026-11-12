import { z } from "zod";

const identifier = z.string().regex(/^[A-Za-z0-9_]{1,64}$/);

/** Validates one rule returned by the automation gateway. */
export const automationRuleSchema = z.object({
  id: identifier,
  name: z.string().min(1),
  description: z.string(),
  explanation: z.string().optional(),
  action_kind: z.enum(["alert_only", "alert", "builtin", "command_profile"]),
  action_name: z.string().optional(),
  enabled: z.boolean(),
  configured_enabled: z.boolean(),
  required_operands: z.array(z.string()).optional(),
  state: z.enum(["normal", "pending", "mitigating", "restoring", "failed"]),
  active_count: z.number().int().nonnegative(),
  last_outcome: z.string(),
});

/** Validates the complete rule-list response at the HTTP boundary. */
export const automationRulesResponseSchema = z.object({
  data: z.object({ rules: z.array(automationRuleSchema) }),
});

/** Validates one active incident returned by the automation gateway. */
export const activeAutomationSchema = z.object({
  incident_id: z.string().min(1),
  rule: automationRuleSchema,
  target: z.string(),
  mitigation: z.string(),
  started_at: z.string(),
  expires_at: z.string().optional(),
  state: automationRuleSchema.shape.state,
  last_outcome: z.string(),
});

/** Validates active incident responses before displaying operator controls. */
export const activeAutomationsResponseSchema = z.object({
  data: z.object({ items: z.array(activeAutomationSchema) }),
});

export type AutomationRule = z.infer<typeof automationRuleSchema>;
export type ActiveAutomation = z.infer<typeof activeAutomationSchema>;

/** Validates a temporary rule override before the provider sends it. */
export const ruleStatePayloadSchema = z.object({
  ruleId: identifier,
  enabled: z.boolean(),
  reason: z.string().max(256).optional(),
});

/** Validates one incident command before the provider sends it. */
export const incidentCommandSchema = z.object({
  incidentId: z.string().regex(/^[A-Za-z0-9]{1,64}$/),
  reason: z.string().max(256).optional(),
});

/** Validates one keyed detector reset before it leaves the browser. */
export const detectionResetPayloadSchema = z.object({
  ruleId: identifier,
  correlationKey: z.string().min(1).max(256),
});

/** Validates a force mitigation request; the service checks rule-specific operands. */
export const forceMitigatePayloadSchema = z.object({
  ruleId: identifier,
  operands: z
    .record(z.string().regex(/^[a-z][a-z0-9_]{0,63}$/), z.string().min(1).max(256))
    .refine((values) => Object.keys(values).length <= 16),
  reason: z.string().max(256).optional(),
});
