import Paper from "@mui/material/Paper";
import type { ReactElement, SetStateAction } from "react";
import { useEffect, useMemo, useState } from "react";
import {
  createDraftFromSettings,
  hasDraftChanged,
  type MonitoringPollingSettingsCardProps,
  type MonitoringPollingSettingsDraft,
} from "./helpers";
import { MonitoringPollingSettingsCardActions } from "./MonitoringPollingSettingsCardActions";
import { MonitoringPollingSettingsCardFields } from "./MonitoringPollingSettingsCardFields";
import { monitoringPollingSettingsFormSchema } from "./schema";
import type { MonitoringPollingSettings } from "./types";

/**
 * Stateful resource form rendered inside one monitoring polling settings provider.
 */
export const MonitoringPollingSettingsCard = ({
  description,
  storageKey,
  title,
  settings,
  onSettingsChange,
}: MonitoringPollingSettingsCardProps & {
  settings: MonitoringPollingSettings;
  onSettingsChange: (settings: MonitoringPollingSettings) => void;
}): ReactElement => {
  const syncedDraft = useMemo(
    (): MonitoringPollingSettingsDraft => ({
      historyIntervalMs: String(settings.historyIntervalMs),
      historyResetGapMs: String(settings.historyResetGapMs),
      maxRecords: String(settings.maxRecords),
      mode: settings.mode,
      snapshotIntervalMs: String(settings.snapshotIntervalMs),
    }),
    [
      settings.historyIntervalMs,
      settings.historyResetGapMs,
      settings.maxRecords,
      settings.mode,
      settings.snapshotIntervalMs,
    ],
  );
  const [draft, setDraft] = useState<MonitoringPollingSettingsDraft>(() => syncedDraft);
  const validation = useMemo(() => monitoringPollingSettingsFormSchema.safeParse(draft), [draft]);
  const fieldErrors = validation.success ? {} : validation.error.flatten().fieldErrors;
  const isChanged = hasDraftChanged(draft, settings);

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
      <MonitoringPollingSettingsCardFields
        description={description}
        draft={draft}
        fieldErrors={fieldErrors}
        onDraftChange={setDraftFieldValue(setDraft)}
        storageKey={storageKey}
        title={title}
      />
      {renderActions({ isChanged, setDraft, settings, storageKey, validation, onSettingsChange })}
    </Paper>
  );
};

/**
 * Creates one draft field update callback for the polling settings form.
 */
const setDraftFieldValue = (
  setDraft: (value: SetStateAction<MonitoringPollingSettingsDraft>) => void,
) => {
  return (name: keyof MonitoringPollingSettingsDraft, value: string): void => {
    setDraft((currentDraft) => ({
      ...currentDraft,
      [name]: value,
    }));
  };
};

type MonitoringPollingSettingsValidation = ReturnType<
  (typeof monitoringPollingSettingsFormSchema)["safeParse"]
>;

type RenderActionsOptions = {
  isChanged: boolean;
  setDraft: (value: SetStateAction<MonitoringPollingSettingsDraft>) => void;
  settings: MonitoringPollingSettings;
  onSettingsChange: (settings: MonitoringPollingSettings) => void;
  storageKey: string;
  validation: MonitoringPollingSettingsValidation;
};

/**
 * Renders the apply and cancel actions only when the current draft differs from persisted settings.
 */
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
          onSettingsChange(validation.data);
        }
      }}
      onCancel={() => {
        setDraft(createDraftFromSettings(settings));
      }}
      storageKey={storageKey}
    />
  );
};
