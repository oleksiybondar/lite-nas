/** Labels and identifier for one process polling settings card. */
export interface ProcessPollingSettingsCardProps {
  description: string;
  storageKey: string;
  title: string;
}

/** Editable string value for a process snapshot interval. */
export interface ProcessPollingSettingsDraft {
  snapshotIntervalMs: string;
}
