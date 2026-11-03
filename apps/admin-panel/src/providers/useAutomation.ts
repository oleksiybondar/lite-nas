import { useApi } from "@hooks/useApi";
import {
  type ActiveAutomation,
  type AutomationRule,
  activeAutomationsResponseSchema,
  automationRulesResponseSchema,
  detectionResetPayloadSchema,
  forceMitigatePayloadSchema,
  incidentCommandSchema,
  ruleStatePayloadSchema,
} from "@schemas/security-automation";
import { useCallback, useEffect, useState } from "react";
import { z } from "zod";

const base = "/api/security/automations";
const operationResponseSchema = z.object({
  data: z.object({
    completed: z.boolean(),
    pending_effects: z.number().int().nonnegative(),
    outcome: z.string(),
  }),
});

/** Loads automation state and validates every operator payload at the API boundary. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The hook keeps related automation requests and response validation together.
export const useAutomation = () => {
  const { get, post, put } = useApi();
  const [rules, setRules] = useState<AutomationRule[]>([]);
  const [active, setActive] = useState<ActiveAutomation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [outcome, setOutcome] = useState("");

  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Both snapshots must be validated before state updates.
  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [ruleResponse, activeResponse] = await Promise.all([
        get(`${base}/rules`).execute(),
        get(`${base}/active`).execute(),
      ]);
      if (!ruleResponse.ok || !activeResponse.ok) throw new Error("automation request failed");
      setRules(automationRulesResponseSchema.parse(await ruleResponse.json()).data.rules);
      setActive(activeAutomationsResponseSchema.parse(await activeResponse.json()).data.items);
      setError(false);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [get]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const finish = async (response: Response) => {
    if (!response.ok) throw new Error("automation command failed");
    const result = operationResponseSchema.parse(await response.json()).data;
    setOutcome(result.outcome);
    await refresh();
  };

  const setEnabled = async (ruleId: string, enabled: boolean) => {
    const payload = ruleStatePayloadSchema.parse({ ruleId, enabled });
    await finish(
      await put(`${base}/rules/${encodeURIComponent(payload.ruleId)}/enabled`, {
        enabled: payload.enabled,
      }).execute(),
    );
  };

  const rollback = async (incidentId: string, forceNormalize = false) => {
    const payload = incidentCommandSchema.parse({ incidentId });
    const verb = forceNormalize ? "force-normalize" : "rollback";
    await finish(
      await post(`${base}/active/${encodeURIComponent(payload.incidentId)}/${verb}`, {}).execute(),
    );
  };

  const forceMitigate = async (ruleId: string, operands: Record<string, string>) => {
    const payload = forceMitigatePayloadSchema.parse({ ruleId, operands });
    const { ruleId: validatedRuleID, ...body } = payload;
    await finish(
      await post(`${base}/rules/${encodeURIComponent(validatedRuleID)}/mitigate`, body).execute(),
    );
  };

  const resetCooldown = async (ruleId: string, correlationKey: string) => {
    const payload = detectionResetPayloadSchema.parse({ ruleId, correlationKey });
    await finish(
      await post(`${base}/rules/${encodeURIComponent(payload.ruleId)}/cooldown/reset`, {
        correlation_key: payload.correlationKey,
      }).execute(),
    );
  };

  return {
    rules,
    active,
    loading,
    error,
    outcome,
    refresh,
    setEnabled,
    rollback,
    forceMitigate,
    resetCooldown,
  };
};
