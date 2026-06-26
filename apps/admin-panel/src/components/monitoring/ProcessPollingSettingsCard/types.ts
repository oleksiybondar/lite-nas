/**
 * Props shared by process polling settings components.
 */
export type ProcessPollingSettingsCardProps = {
  /**
   * Short explanation shown above the process-specific polling controls.
   */
  description: string;
  /**
   * Source-scoped storage key for one process polling settings slice.
   */
  storageKey: string;
  /**
   * Human-readable resource title shown in the settings card.
   */
  title: string;
};

export type ProcessPollingSettingsDraft = {
  snapshotIntervalMs: string;
};
