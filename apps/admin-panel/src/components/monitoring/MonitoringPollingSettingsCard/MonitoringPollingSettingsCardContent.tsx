import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import type { MonitoringPollingSettingsCardProps } from "@lite-nas/settings";
import { MonitoringPollingSettingsCard as SharedMonitoringPollingSettingsCard } from "@lite-nas/settings";
import type { ReactElement } from "react";

/** Connects the shared monitoring form to one admin polling settings provider. */
export const MonitoringPollingSettingsCardContent = (
  props: MonitoringPollingSettingsCardProps,
): ReactElement => {
  const settings = useMonitoringPollingSettings();
  return (
    <SharedMonitoringPollingSettingsCard
      {...props}
      settings={settings}
      onSettingsChange={settings.setSettings}
    />
  );
};
