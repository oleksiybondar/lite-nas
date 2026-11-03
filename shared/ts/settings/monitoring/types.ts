/** Values edited by a monitoring polling settings card. */
export interface MonitoringPollingSettings {
  mode: "history" | "snapshot";
  historyIntervalMs: number;
  snapshotIntervalMs: number;
  maxRecords: number;
  historyResetGapMs: number;
}
