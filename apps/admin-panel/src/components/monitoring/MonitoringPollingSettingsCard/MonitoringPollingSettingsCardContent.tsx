import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import Paper from "@mui/material/Paper";
import { monitoringPollingSettingsFormSchema } from "@schemas/monitoring/monitoring-polling-settings";
import type { ReactElement } from "react";
import { useEffect, useMemo, useState } from "react";
import {
  createDraftFromSettings,
  hasDraftChanged,
  type MonitoringPollingSettingsCardProps,
  type MonitoringPollingSettingsDraft,
} from "./helpers";
import { MonitoringPollingSettingsCardActions } from "./MonitoringPollingSettingsCardActions";
import { MonitoringPollingSettingsCardFields } from "./MonitoringPollingSettingsCardFields";

/**
 * Stateful resource form rendered inside one monitoring polling settings provider.
 */
export const MonitoringPollingSettingsCardContent = ({
  description,
  storageKey,
  title,
}: MonitoringPollingSettingsCardProps): ReactElement => {
  const settings = useMonitoringPollingSettings();
  const [draft, setDraft] = useState<MonitoringPollingSettingsDraft>(() =>
    createDraftFromSettings(settings),
  );
  const validation = useMemo(() => monitoringPollingSettingsFormSchema.safeParse(draft), [draft]);
  const fieldErrors = validation.success ? {} : validation.error.flatten().fieldErrors;
  const isChanged = hasDraftChanged(draft, settings);

  useEffect(() => {
    setDraft(createDraftFromSettings(settings));
  }, [settings]);

  return (
    <Paper
      data-test-class="monitoring-settings-card"
      data-test-name={storageKey}
      data-testid={`monitoring-settings-card-${storageKey}`}
      sx={{ p: 3 }}
    >
      <MonitoringPollingSettingsCardFields
        description={description}
        draft={draft}
        fieldErrors={fieldErrors}
        onDraftChange={(name, value) => {
          setDraft((currentDraft) => ({
            ...currentDraft,
            [name]: value,
          }));
        }}
        storageKey={storageKey}
        title={title}
      />
      {isChanged ? (
        <MonitoringPollingSettingsCardActions
          isApplyDisabled={!validation.success}
          onApply={() => {
            if (validation.success) {
              settings.setSettings(validation.data);
            }
          }}
          onCancel={() => {
            setDraft(createDraftFromSettings(settings));
          }}
          storageKey={storageKey}
        />
      ) : null}
    </Paper>
  );
};
