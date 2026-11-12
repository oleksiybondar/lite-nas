import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import { Stack, Typography } from "@mui/material";
import {
  OperationalSafetyLists,
  OperationalSafetyStatus,
  OperationalSafetyToolbar,
} from "@pages/OperationalSafetyAutomationPage/index";
import { useOperationalSafetyAutomationContext } from "@providers/OperationalSafetyAutomationProvider";
import { type ReactElement, useState } from "react";

/** Reports a failed operator request without letting it escape a click handler. */
const runOperatorAction = async (
  action: () => Promise<void>,
  setActionError: (failed: boolean) => void,
): Promise<void> => {
  setActionError(false);
  try {
    await action();
  } catch {
    setActionError(true);
  }
};

/** Gives administrators visibility and control over operational safety rules and active tasks. */
export const OperationalSafetyAutomationPage = (): ReactElement | null => {
  const { getLabel } = useLanguage();
  const { requireAdmin } = useRbac();
  const {
    rules,
    tasks,
    loading,
    error,
    outcome,
    refresh,
    setEnabled,
    forceSet,
    forceStop,
    search,
    setSearch,
  } = useOperationalSafetyAutomationContext();
  const [actionError, setActionError] = useState(false);
  const allowed = requireAdmin();
  if (!allowed) return null;

  const runAction = (action: () => Promise<void>): void => {
    void runOperatorAction(action, setActionError);
  };

  return (
    <Stack data-testid="operational-safety-automation-page" spacing={3}>
      <Typography variant="h1" data-testid="operational-safety-automation-title">
        {getLabel("admin.operationalSafety.title")}
      </Typography>
      <Typography>{getLabel("admin.operationalSafety.summary")}</Typography>
      <OperationalSafetyToolbar
        search={search}
        onSearch={setSearch}
        onRefresh={() => void refresh()}
      />
      <OperationalSafetyStatus
        loading={loading}
        error={error}
        actionError={actionError}
        outcome={outcome}
      />
      <OperationalSafetyLists
        rules={rules}
        tasks={tasks}
        loading={loading}
        error={error}
        onSetEnabled={(id, enabled) => void runAction(() => setEnabled(id, enabled))}
        onForceSet={(id, level) => void runAction(() => forceSet(id, level))}
        onForceStop={(id) => void runAction(() => forceStop(id))}
      />
    </Stack>
  );
};
