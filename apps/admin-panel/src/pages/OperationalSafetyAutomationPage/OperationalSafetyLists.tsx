import { useLanguage } from "@lite-nas/language";
import { Typography } from "@mui/material";
import type {
  OperationalSafetyRule,
  OperationalSafetyTask,
} from "@schemas/operational-safety-automation";
import type { ReactElement } from "react";
import { OperationalSafetyRuleCard } from "./OperationalSafetyRuleCard";
import { OperationalSafetyTaskCard } from "./OperationalSafetyTaskCard";

/** Renders the configured rule inventory and the active automation tasks. */
export const OperationalSafetyLists = ({
  rules,
  tasks,
  loading,
  error,
  onSetEnabled,
  onForceSet,
  onForceStop,
}: {
  rules: OperationalSafetyRule[];
  tasks: OperationalSafetyTask[];
  loading: boolean;
  error: boolean;
  onSetEnabled: (id: string, enabled: boolean) => void;
  onForceSet: (id: string, level: OperationalSafetyTask["level"]) => void;
  onForceStop: (id: string) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <>
      <Typography variant="h2">{getLabel("admin.operationalSafety.rulesTitle")}</Typography>
      {!loading && !error && rules.length === 0 && (
        <Typography>{getLabel("admin.operationalSafety.noRules")}</Typography>
      )}
      {rules.map((rule) => (
        <OperationalSafetyRuleCard key={rule.id} rule={rule} onSetEnabled={onSetEnabled} />
      ))}
      <Typography variant="h2">{getLabel("admin.operationalSafety.tasksTitle")}</Typography>
      {!loading && !error && tasks.length === 0 && (
        <Typography>{getLabel("admin.operationalSafety.noTasks")}</Typography>
      )}
      {tasks.map((task) => (
        <OperationalSafetyTaskCard
          key={task.id}
          task={task}
          onForceSet={onForceSet}
          onForceStop={onForceStop}
        />
      ))}
    </>
  );
};
