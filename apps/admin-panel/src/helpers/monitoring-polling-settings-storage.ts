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
  maxRecords: 180,
  mode: "snapshot",
  snapshotIntervalMs: 1000,
};

const processScopedMonitoringPollingSettings: MonitoringPollingSettings = {
  historyIntervalMs: 1,
  historyResetGapMs: 1,
  maxRecords: 1,
  mode: "snapshot",
  snapshotIntervalMs: 5000,
};

const serviceScopedMonitoringPollingSettings: MonitoringPollingSettings = {
  historyIntervalMs: 1,
  historyResetGapMs: 1,
  maxRecords: 1,
  mode: "snapshot",
  snapshotIntervalMs: 15000,
};

const monitoringPollingSettingsDefaultsByScope: Partial<Record<string, MonitoringPollingSettings>> =
  {
    "process-metrics": processScopedMonitoringPollingSettings,
    "service-metrics": serviceScopedMonitoringPollingSettings,
  };

/**
 * Resolves the default polling settings for one persisted monitoring scope.
 */
export const resolveDefaultMonitoringPollingSettings = (
  scope: string,
): MonitoringPollingSettings => {
  return monitoringPollingSettingsDefaultsByScope[scope] ?? defaultMonitoringPollingSettings;
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
export const normalizeMonitoringPollingSettings = (
  value: unknown,
  defaults: MonitoringPollingSettings = defaultMonitoringPollingSettings,
): MonitoringPollingSettings => {
  if (typeof value !== "object" || value === null) {
    return defaults;
  }

  const record = value as Record<string, unknown>;

  return {
    historyIntervalMs: normalizePositiveInteger(
      record.historyIntervalMs,
      defaults.historyIntervalMs,
    ),
    historyResetGapMs: normalizePositiveInteger(
      record.historyResetGapMs,
      defaults.historyResetGapMs,
    ),
    maxRecords: normalizePositiveInteger(record.maxRecords, defaults.maxRecords),
    mode: normalizeMode(record.mode, defaults.mode),
    snapshotIntervalMs: normalizePositiveInteger(
      record.snapshotIntervalMs,
      defaults.snapshotIntervalMs,
    ),
  };
};

/**
 * Loads source-scoped monitoring polling settings from local storage.
 */
export const loadMonitoringPollingSettings = (scope: string): MonitoringPollingSettings => {
  const defaults = resolveDefaultMonitoringPollingSettings(scope);

  if (typeof window === "undefined") {
    return defaults;
  }

  const rawSettings = window.localStorage.getItem(buildMonitoringPollingSettingsStorageKey(scope));

  if (rawSettings === null) {
    return defaults;
  }

  try {
    const parsedSettings: unknown = JSON.parse(rawSettings);
    return normalizeMonitoringPollingSettings(parsedSettings, defaults);
  } catch {
    return defaults;
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
    JSON.stringify(
      normalizeMonitoringPollingSettings(settings, resolveDefaultMonitoringPollingSettings(scope)),
    ),
  );
};

const normalizeMode = (value: unknown, fallback: MonitoringPollingMode): MonitoringPollingMode => {
  const result = monitoringPollingModeSchema.safeParse(value);

  if (!result.success) {
    return fallback;
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
