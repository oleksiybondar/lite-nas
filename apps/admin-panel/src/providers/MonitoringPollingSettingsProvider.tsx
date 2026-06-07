import { MonitoringPollingSettingsContext } from "@contexts/monitoring-polling-settings-context";
import type {
  MonitoringPollingSettings,
  MonitoringPollingSettingsContextValue,
} from "@dto/monitoring/monitoring-polling-settings";
import {
  defaultMonitoringPollingSettings,
  loadMonitoringPollingSettings,
  normalizeMonitoringPollingSettings,
  saveMonitoringPollingSettings,
} from "@helpers/monitoring-polling-settings-storage";
import type { Dispatch, PropsWithChildren, ReactElement, SetStateAction } from "react";
import { useEffect, useMemo, useState } from "react";

type MonitoringPollingSettingsProviderProps = PropsWithChildren<{
  /**
   * Stable source identifier used to scope persisted polling settings.
   */
  storageKey: string;
}>;

/**
 * Provides source-scoped monitoring polling settings for UI and polling providers.
 */
export const MonitoringPollingSettingsProvider = ({
  children,
  storageKey,
}: MonitoringPollingSettingsProviderProps): ReactElement => {
  const [settings, setSettingsState] = useState<MonitoringPollingSettings>(() => {
    return loadMonitoringPollingSettings(storageKey);
  });

  useEffect(() => {
    setSettingsState(loadMonitoringPollingSettings(storageKey));
  }, [storageKey]);

  const value = useMemo<MonitoringPollingSettingsContextValue>(() => {
    return buildMonitoringPollingSettingsContextValue({
      setSettingsState,
      settings,
      storageKey,
    });
  }, [settings, storageKey]);

  return (
    <MonitoringPollingSettingsContext.Provider value={value}>
      {children}
    </MonitoringPollingSettingsContext.Provider>
  );
};

type BuildMonitoringPollingSettingsContextValueOptions = {
  setSettingsState: Dispatch<SetStateAction<MonitoringPollingSettings>>;
  settings: MonitoringPollingSettings;
  storageKey: string;
};

const buildMonitoringPollingSettingsContextValue = ({
  setSettingsState,
  settings,
  storageKey,
}: BuildMonitoringPollingSettingsContextValueOptions): MonitoringPollingSettingsContextValue => {
  const updateSettings = createSettingsUpdater(setSettingsState, storageKey);

  return {
    ...settings,
    resetSettings: () => {
      updateSettings(defaultMonitoringPollingSettings);
    },
    setHistoryIntervalMs: (historyIntervalMs: number) => {
      updateSettings({
        ...settings,
        historyIntervalMs,
      });
    },
    setHistoryResetGapMs: (historyResetGapMs: number) => {
      updateSettings({
        ...settings,
        historyResetGapMs,
      });
    },
    setMaxRecords: (maxRecords: number) => {
      updateSettings({
        ...settings,
        maxRecords,
      });
    },
    setMode: (mode) => {
      updateSettings({
        ...settings,
        mode,
      });
    },
    setSettings: updateSettings,
    setSnapshotIntervalMs: (snapshotIntervalMs: number) => {
      updateSettings({
        ...settings,
        snapshotIntervalMs,
      });
    },
  };
};

/**
 * Creates the shared settings update primitive used by all provider setters.
 */
const createSettingsUpdater = (
  setSettingsState: Dispatch<SetStateAction<MonitoringPollingSettings>>,
  storageKey: string,
): ((settings: MonitoringPollingSettings) => void) => {
  return (nextSettings: MonitoringPollingSettings): void => {
    const normalizedSettings = normalizeMonitoringPollingSettings(nextSettings);

    setSettingsState(normalizedSettings);
    saveMonitoringPollingSettings(storageKey, normalizedSettings);
  };
};
