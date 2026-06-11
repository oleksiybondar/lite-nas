import type { MonitoringPollingSettings } from "@dto/monitoring/monitoring-polling-settings";

/**
 * Editable string-backed form state for one monitoring polling settings card.
 */
export type MonitoringPollingSettingsDraft = {
  historyIntervalMs: string;
  historyResetGapMs: string;
  maxRecords: string;
  mode: "history" | "snapshot";
  snapshotIntervalMs: string;
};

/**
 * Props shared by monitoring polling settings card components.
 */
export type MonitoringPollingSettingsCardProps = {
  /**
   * Short explanation shown above the resource-specific polling controls.
   */
  description: string;
  /**
   * Source-scoped storage key for one monitoring polling settings slice.
   */
  storageKey: string;
  /**
   * Human-readable resource title shown in the settings card.
   */
  title: string;
};

/**
 * Runtime validation errors exposed to monitoring polling form fields.
 */
export type MonitoringPollingSettingsFieldErrors = Partial<
  Record<keyof MonitoringPollingSettings, string[]>
>;

/**
 * Converts persisted monitoring polling settings into editable string-backed form state.
 */
export const createDraftFromSettings = (
  settings: MonitoringPollingSettings,
): MonitoringPollingSettingsDraft => {
  return {
    historyIntervalMs: String(settings.historyIntervalMs),
    historyResetGapMs: String(settings.historyResetGapMs),
    maxRecords: String(settings.maxRecords),
    mode: settings.mode,
    snapshotIntervalMs: String(settings.snapshotIntervalMs),
  };
};

/**
 * Reports whether the editable card state differs from persisted provider settings.
 */
export const hasDraftChanged = (
  draft: MonitoringPollingSettingsDraft,
  settings: MonitoringPollingSettings,
): boolean => {
  return (
    draft.mode !== settings.mode ||
    draft.historyIntervalMs !== String(settings.historyIntervalMs) ||
    draft.snapshotIntervalMs !== String(settings.snapshotIntervalMs) ||
    draft.maxRecords !== String(settings.maxRecords) ||
    draft.historyResetGapMs !== String(settings.historyResetGapMs)
  );
};
