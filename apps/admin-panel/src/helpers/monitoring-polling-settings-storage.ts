import type {
  MonitoringPollingMode,
  MonitoringPollingSettings,
} from "@dto/monitoring/monitoring-polling-settings";
import {
  monitoringPollingModeSchema,
  monitoringPollingPositiveIntegerSchema,
} from "@schemas/monitoring/monitoring-polling-settings";

const monitoringPollingSettingsStorageKeyPrefix =
  "lite-nas.admin-panel.monitoring-polling-settings";

/**
 * Default monitoring polling settings shared across providers and settings UI.
 */
export const defaultMonitoringPollingSettings: MonitoringPollingSettings = {
  historyIntervalMs: 15000,
  historyResetGapMs: 10000,
  maxRecords: 300,
  mode: "history",
  snapshotIntervalMs: 1000,
};

/**
 * Builds the source-scoped local-storage key for monitoring polling settings.
 */
export const buildMonitoringPollingSettingsStorageKey = (scope: string): string => {
  return `${monitoringPollingSettingsStorageKeyPrefix}.${scope}`;
};

/**
 * Normalizes one unknown settings object against monitoring polling defaults.
 *
 * Invalid or missing fields fall back individually so future partial backend
 * settings can reuse this function without discarding valid values.
 */
export const normalizeMonitoringPollingSettings = (value: unknown): MonitoringPollingSettings => {
  if (typeof value !== "object" || value === null) {
    return defaultMonitoringPollingSettings;
  }

  const record = value as Record<string, unknown>;

  return {
    historyIntervalMs: normalizePositiveInteger(
      record.historyIntervalMs,
      defaultMonitoringPollingSettings.historyIntervalMs,
    ),
    historyResetGapMs: normalizePositiveInteger(
      record.historyResetGapMs,
      defaultMonitoringPollingSettings.historyResetGapMs,
    ),
    maxRecords: normalizePositiveInteger(
      record.maxRecords,
      defaultMonitoringPollingSettings.maxRecords,
    ),
    mode: normalizeMode(record.mode),
    snapshotIntervalMs: normalizePositiveInteger(
      record.snapshotIntervalMs,
      defaultMonitoringPollingSettings.snapshotIntervalMs,
    ),
  };
};

/**
 * Loads source-scoped monitoring polling settings from local storage.
 */
export const loadMonitoringPollingSettings = (scope: string): MonitoringPollingSettings => {
  if (typeof window === "undefined") {
    return defaultMonitoringPollingSettings;
  }

  const rawSettings = window.localStorage.getItem(buildMonitoringPollingSettingsStorageKey(scope));

  if (rawSettings === null) {
    return defaultMonitoringPollingSettings;
  }

  try {
    const parsedSettings: unknown = JSON.parse(rawSettings);
    return normalizeMonitoringPollingSettings(parsedSettings);
  } catch {
    return defaultMonitoringPollingSettings;
  }
};

/**
 * Saves source-scoped monitoring polling settings to local storage.
 */
export const saveMonitoringPollingSettings = (
  scope: string,
  settings: MonitoringPollingSettings,
): void => {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(
    buildMonitoringPollingSettingsStorageKey(scope),
    JSON.stringify(normalizeMonitoringPollingSettings(settings)),
  );
};

const normalizeMode = (value: unknown): MonitoringPollingMode => {
  const result = monitoringPollingModeSchema.safeParse(value);

  if (!result.success) {
    return defaultMonitoringPollingSettings.mode;
  }

  return result.data;
};

const normalizePositiveInteger = (value: unknown, fallback: number): number => {
  const result = monitoringPollingPositiveIntegerSchema.safeParse(value);

  if (!result.success) {
    return fallback;
  }

  return result.data;
};
