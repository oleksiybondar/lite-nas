import type { ApiContextValue } from "@dto/api/api";
import { useApi } from "@hooks/useApi";
import {
  operationalSafetyForceSetPayloadSchema,
  operationalSafetyForceStopPayloadSchema,
  operationalSafetyOperationResponseSchema,
  operationalSafetyRuleStatePayloadSchema,
  operationalSafetyRulesResponseSchema,
  operationalSafetyTasksResponseSchema,
} from "@schemas/operational-safety-automation";
import { useCallback, useEffect, useState } from "react";

const base = "/api/system/automations";

/** Loads and validates the automation rules and tasks together. */
const loadAutomationData = async (get: ApiContextValue["get"]) => {
  const [rulesResponse, tasksResponse] = await Promise.all([
    get(`${base}/rules`).execute(),
    get(`${base}/tasks`).execute(),
  ]);
  if (!rulesResponse.ok || !tasksResponse.ok) throw new Error("request failed");
  return {
    rules: operationalSafetyRulesResponseSchema.parse(await rulesResponse.json()).data.rules,
    tasks: operationalSafetyTasksResponseSchema.parse(await tasksResponse.json()).data.items,
  };
};

/** Records a successful operator outcome and refreshes the visible task list. */
const finishOperation = async (
  response: Response,
  setOutcome: (outcome: string) => void,
  refresh: () => Promise<void>,
): Promise<void> => {
  if (!response.ok) throw new Error("operation failed");
  setOutcome(operationalSafetyOperationResponseSchema.parse(await response.json()).data.outcome);
  await refresh();
};

/** Submits one validated rule enablement change. */
const updateRuleEnabled = async (
  put: ApiContextValue["put"],
  ruleId: string,
  enabled: boolean,
): Promise<Response> => {
  const payload = operationalSafetyRuleStatePayloadSchema.parse({ ruleId, enabled });
  return put(`${base}/rules/${encodeURIComponent(payload.ruleId)}/enabled`, {
    enabled: payload.enabled,
  }).execute();
};

/** Submits one validated task level change. */
const requestForceSet = async (
  post: ApiContextValue["post"],
  taskId: string,
  level: "normal" | "warning" | "error" | "critical",
): Promise<Response> => {
  const payload = operationalSafetyForceSetPayloadSchema.parse({ taskId, level });
  return post(`${base}/tasks/${encodeURIComponent(payload.taskId)}/force-set`, {
    level: payload.level,
  }).execute();
};

/** Submits one validated task stop request. */
const requestForceStop = async (
  post: ApiContextValue["post"],
  taskId: string,
): Promise<Response> => {
  const payload = operationalSafetyForceStopPayloadSchema.parse({ taskId });
  return post(`${base}/tasks/${encodeURIComponent(payload.taskId)}/force-stop`, {}).execute();
};

/** Loads and operates the gateway's existing system automation endpoints. */
export const useOperationalSafetyAutomation = (enabled: boolean) => {
  const { get, post, put } = useApi();
  const [rules, setRules] = useState<
    ReturnType<typeof operationalSafetyRulesResponseSchema.parse>["data"]["rules"]
  >([]);
  const [tasks, setTasks] = useState<
    ReturnType<typeof operationalSafetyTasksResponseSchema.parse>["data"]["items"]
  >([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [outcome, setOutcome] = useState("");

  const refresh = useCallback(async () => {
    if (!enabled) return;
    setLoading(true);
    try {
      const data = await loadAutomationData(get);
      setRules(data.rules);
      setTasks(data.tasks);
      setError(false);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [enabled, get]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const setEnabled = async (ruleId: string, ruleEnabled: boolean) => {
    await finishOperation(await updateRuleEnabled(put, ruleId, ruleEnabled), setOutcome, refresh);
  };

  const forceSet = async (taskId: string, level: "normal" | "warning" | "error" | "critical") => {
    await finishOperation(await requestForceSet(post, taskId, level), setOutcome, refresh);
  };

  const forceStop = async (taskId: string) => {
    await finishOperation(await requestForceStop(post, taskId), setOutcome, refresh);
  };

  return { rules, tasks, loading, error, outcome, refresh, setEnabled, forceSet, forceStop };
};
