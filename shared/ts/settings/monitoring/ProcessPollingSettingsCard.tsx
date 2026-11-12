import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";
import { useEffect, useMemo, useState } from "react";
import { MonitoringPollingSettingsCardActions } from "./MonitoringPollingSettingsCardActions";
import { MonitoringPollingSettingsCardHeader } from "./MonitoringPollingSettingsCardHeader";
import { createProcessPollingDraft, processPollingSettingsFormSchema } from "./processHelpers";
import type { ProcessPollingSettingsCardProps, ProcessPollingSettingsDraft } from "./processTypes";
import type { MonitoringPollingSettings } from "./types";

/**
 * Stateful content rendered inside one process polling settings provider.
 */
export const ProcessPollingSettingsCard = ({
  description,
  storageKey,
  title,
  settings,
  onSettingsChange,
}: ProcessPollingSettingsCardProps & {
  settings: MonitoringPollingSettings;
  onSettingsChange: (settings: MonitoringPollingSettings) => void;
}): ReactElement => {
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
        {renderActions({ isChanged, setDraft, settings, storageKey, validation, onSettingsChange })}
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
  settings: MonitoringPollingSettings;
  onSettingsChange: (settings: MonitoringPollingSettings) => void;
  storageKey: string;
  validation: ProcessPollingSettingsValidation;
};

const renderActions = ({
  isChanged,
  setDraft,
  settings,
  storageKey,
  validation,
  onSettingsChange,
}: RenderActionsOptions): ReactElement | null => {
  if (!isChanged) {
    return null;
  }

  return (
    <MonitoringPollingSettingsCardActions
      isApplyDisabled={!validation.success}
      onApply={() => {
        if (validation.success) {
          onSettingsChange({
            historyIntervalMs: 1,
            historyResetGapMs: 1,
            maxRecords: 1,
            mode: "snapshot",
            snapshotIntervalMs: validation.data.snapshotIntervalMs,
          });
        }
      }}
      onCancel={() => {
        setDraft(createProcessPollingDraft(settings.snapshotIntervalMs));
      }}
      storageKey={storageKey}
    />
  );
};
