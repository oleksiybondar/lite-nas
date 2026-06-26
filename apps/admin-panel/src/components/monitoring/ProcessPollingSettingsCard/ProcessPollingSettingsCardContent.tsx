import {
  MonitoringPollingSettingsCardActions,
  MonitoringPollingSettingsCardHeader,
} from "@components/monitoring/MonitoringPollingSettingsCard";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";
import { useEffect, useMemo, useState } from "react";
import { createProcessPollingDraft, processPollingSettingsFormSchema } from "./helpers";
import type { ProcessPollingSettingsCardProps, ProcessPollingSettingsDraft } from "./types";

/**
 * Stateful content rendered inside one process polling settings provider.
 */
export const ProcessPollingSettingsCardContent = ({
  description,
  storageKey,
  title,
}: ProcessPollingSettingsCardProps): ReactElement => {
  const settings = useMonitoringPollingSettings();
  const syncedDraft = useMemo(() => {
    return createProcessPollingDraft(settings.snapshotIntervalMs);
  }, [settings.snapshotIntervalMs]);
  const [draft, setDraft] = useState(syncedDraft);
  const validation = useMemo(() => processPollingSettingsFormSchema.safeParse(draft), [draft]);
  const fieldErrors = validation.success ? {} : validation.error.flatten().fieldErrors;
  const isChanged = draft.snapshotIntervalMs !== String(settings.snapshotIntervalMs);

  useEffect(() => {
    setDraft(syncedDraft);
  }, [syncedDraft]);

  return (
    <Paper
      data-test-class="monitoring-settings-card"
      data-test-name={storageKey}
      data-testid={`monitoring-settings-card-${storageKey}`}
      sx={{ p: 3 }}
    >
      <Stack spacing={2.5}>
        <MonitoringPollingSettingsCardHeader
          description={description}
          storageKey={storageKey}
          title={title}
        />
        <TextField
          data-test-class="monitoring-settings-input"
          data-test-name="snapshotIntervalMs"
          error={fieldErrors.snapshotIntervalMs?.[0] !== undefined}
          fullWidth
          helperText={
            fieldErrors.snapshotIntervalMs?.[0] ??
            "Interval in milliseconds between snapshot refresh requests."
          }
          label="Snapshot interval (ms)"
          name={`${storageKey}-snapshotIntervalMs`}
          onChange={(event) => {
            setDraft({ snapshotIntervalMs: event.target.value });
          }}
          slotProps={{ htmlInput: { min: 1 } }}
          type="number"
          value={draft.snapshotIntervalMs}
        />
        {renderActions({ isChanged, setDraft, settings, storageKey, validation })}
      </Stack>
    </Paper>
  );
};

type ProcessPollingSettingsValidation = ReturnType<
  (typeof processPollingSettingsFormSchema)["safeParse"]
>;

type RenderActionsOptions = {
  isChanged: boolean;
  setDraft: (draft: ProcessPollingSettingsDraft) => void;
  settings: ReturnType<typeof useMonitoringPollingSettings>;
  storageKey: string;
  validation: ProcessPollingSettingsValidation;
};

const renderActions = ({
  isChanged,
  setDraft,
  settings,
  storageKey,
  validation,
}: RenderActionsOptions): ReactElement | null => {
  if (!isChanged) {
    return null;
  }

  return (
    <MonitoringPollingSettingsCardActions
      isApplyDisabled={!validation.success}
      onApply={() => {
        if (validation.success) {
          settings.setSettings({
            historyIntervalMs: 1,
            historyResetGapMs: 1,
            maxRecords: 1,
            mode: "snapshot",
            snapshotIntervalMs: validation.data.snapshotIntervalMs,
          });
        }
      }}
      onCancel={() => {
        setDraft(draftFromSettings(settings.snapshotIntervalMs));
      }}
      storageKey={storageKey}
    />
  );
};

const draftFromSettings = (snapshotIntervalMs: number): ProcessPollingSettingsDraft => {
  return createProcessPollingDraft(snapshotIntervalMs);
};
