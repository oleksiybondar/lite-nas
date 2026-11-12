import { MonitoringPollingSettingsContext } from "@contexts/monitoring-polling-settings-context";
import type { ApiContextValue } from "@dto/api/api";
import type {
  MonitoringPollingSettings,
  MonitoringPollingSettingsContextValue,
} from "@dto/monitoring/monitoring-polling-settings";
import {
  loadMonitoringPollingSettings,
  normalizeMonitoringPollingSettings,
  resolveDefaultMonitoringPollingSettings,
  saveMonitoringPollingSettings,
} from "@helpers/monitoring-polling-settings-storage";
import { saveAdminSettings } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import { useAuth } from "@hooks/useAuth";
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
  const { get, put } = useApi();
  const { isAuthenticated } = useAuth();
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
      get,
      put,
      isAuthenticated,
    });
  }, [settings, storageKey, get, put, isAuthenticated]);

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
  get: ApiContextValue["get"];
  put: ApiContextValue["put"];
  isAuthenticated: boolean;
};

const buildMonitoringPollingSettingsContextValue = ({
  setSettingsState,
  settings,
  storageKey,
  get,
  put,
  isAuthenticated,
}: BuildMonitoringPollingSettingsContextValueOptions): MonitoringPollingSettingsContextValue => {
  const updateSettings = createSettingsUpdater(setSettingsState, storageKey, {
    get,
    put,
    isAuthenticated,
  });

  return {
    ...settings,
    resetSettings: () => {
      updateSettings(resolveDefaultMonitoringPollingSettings(storageKey));
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
  api: Pick<ApiContextValue, "get" | "put"> & { isAuthenticated: boolean },
): ((settings: MonitoringPollingSettings) => void) => {
  return (nextSettings: MonitoringPollingSettings): void => {
    const normalizedSettings = normalizeMonitoringPollingSettings(
      nextSettings,
      resolveDefaultMonitoringPollingSettings(storageKey),
    );

    setSettingsState(normalizedSettings);
    saveMonitoringPollingSettings(storageKey, normalizedSettings);
    if (api.isAuthenticated) void saveAdminSettings(api).catch(() => undefined);
  };
};
