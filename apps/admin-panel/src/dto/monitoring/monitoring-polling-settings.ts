/**
 * Polling modes supported by generic monitoring time-series providers.
 */
export type MonitoringPollingMode = "history" | "snapshot";

/**
 * Shared persisted polling settings consumed by monitoring pages and providers.
 */
export type MonitoringPollingSettings = {
  /**
   * Active polling strategy after the initial history bootstrap request.
   */
  mode: MonitoringPollingMode;
  /**
   * Poll interval used when the active mode is history polling.
   */
  historyIntervalMs: number;
  /**
   * Poll interval used when the active mode is snapshot polling.
   */
  snapshotIntervalMs: number;
  /**
   * Maximum number of time-series points retained in the browser cache.
   */
  maxRecords: number;
  /**
   * Gap threshold after which a history poll replaces the local cache.
   */
  historyResetGapMs: number;
};

/**
 * Shared context contract for monitoring polling settings.
 */
export type MonitoringPollingSettingsContextValue = MonitoringPollingSettings & {
  /**
   * Replaces the full polling settings state after normalization.
   */
  setSettings: (settings: MonitoringPollingSettings) => void;
  /**
   * Switches the active monitoring polling mode.
   */
  setMode: (mode: MonitoringPollingMode) => void;
  /**
   * Updates the history polling interval.
   */
  setHistoryIntervalMs: (value: number) => void;
  /**
   * Updates the snapshot polling interval.
   */
  setSnapshotIntervalMs: (value: number) => void;
  /**
   * Updates the maximum rolling cache size.
   */
  setMaxRecords: (value: number) => void;
  /**
   * Updates the gap threshold that triggers a full history reset.
   */
  setHistoryResetGapMs: (value: number) => void;
  /**
   * Restores the default monitoring polling settings.
   */
  resetSettings: () => void;
};
