import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import type { ProcessPollingSettingsCardProps } from "@lite-nas/settings";
import { ProcessPollingSettingsCard as SharedProcessPollingSettingsCard } from "@lite-nas/settings";
import type { ReactElement } from "react";

/** Connects the shared process interval form to one admin polling settings provider. */
export const ProcessPollingSettingsCardContent = (
  props: ProcessPollingSettingsCardProps,
): ReactElement => {
  const settings = useMonitoringPollingSettings();
  return (
    <SharedProcessPollingSettingsCard
      {...props}
      settings={settings}
      onSettingsChange={settings.setSettings}
    />
  );
};
