import { z } from "zod";

const ruleId = z.string().regex(/^[a-z][a-z0-9_]{0,63}$/);
const taskId = z.string().regex(/^[A-Za-z0-9]{1,64}$/);

/** Validates one operational safety rule returned by the gateway. */
export const operationalSafetyRuleSchema = z.object({
  id: ruleId,
  name: z.string(),
  enabled: z.boolean(),
  configured_enabled: z.boolean(),
  output_id_prefix: z.string(),
  cooldown: z.string(),
});

/** Validates the gateway's configured rule response. */
export const operationalSafetyRulesResponseSchema = z.object({
  data: z.object({ rules: z.array(operationalSafetyRuleSchema) }),
});

/** Validates one active operational safety task. */
export const operationalSafetyTaskSchema = z.object({
  id: taskId,
  rule_id: ruleId,
  correlation_key: z.string(),
  level: z.enum(["normal", "warning", "error", "critical"]),
  active_alert_ids: z.array(z.string()),
  cooldown_until: z.string().optional(),
  last_outcome: z.string(),
});

/** Validates the gateway's active task response. */
export const operationalSafetyTasksResponseSchema = z.object({
  data: z.object({ items: z.array(operationalSafetyTaskSchema) }),
});

/** Validates the result returned after an operational safety action. */
export const operationalSafetyOperationResponseSchema = z.object({
  data: z.object({
    task_id: z.string(),
    state: z.string(),
    level: z.string(),
    outcome: z.string(),
  }),
});

/** Validates an admin rule state change before it is sent. */
export const operationalSafetyRuleStatePayloadSchema = z.object({
  ruleId,
  enabled: z.boolean(),
});

/** Validates a task level action before it is sent. */
export const operationalSafetyForceSetPayloadSchema = z.object({
  taskId,
  level: z.enum(["normal", "warning", "error", "critical"]),
});

/** Validates a staged stop request before it is sent. */
export const operationalSafetyForceStopPayloadSchema = z.object({ taskId });

export type OperationalSafetyRule = z.infer<typeof operationalSafetyRuleSchema>;
export type OperationalSafetyTask = z.infer<typeof operationalSafetyTaskSchema>;
