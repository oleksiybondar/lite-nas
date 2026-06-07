import { MonitoringPollingSettingsContext } from "@contexts/monitoring-polling-settings-context";
import type { MonitoringPollingSettingsContextValue } from "@dto/monitoring/monitoring-polling-settings";
import { useContext } from "react";

/**
 * Reads source-scoped monitoring polling settings from the nearest provider.
 */
export const useMonitoringPollingSettings = (): MonitoringPollingSettingsContextValue => {
  const context = useContext(MonitoringPollingSettingsContext);

  if (context === undefined) {
    throw new Error(
      "useMonitoringPollingSettings must be used inside MonitoringPollingSettingsProvider",
    );
  }

  return context;
};
