import { MonitoringPollingSettingsCard } from "@components/monitoring/MonitoringPollingSettingsCard";
import { MonitoringSettingsTabs } from "@components/monitoring/MonitoringSettingsTabs";
import { useLanguage } from "@lite-nas/language";
import { PreferencesPageHeader } from "@lite-nas/settings";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Preferences page for metrics-oriented monitoring polling behavior.
 */
export const PreferencesMonitoringSettingsPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack data-testid="preferences-monitoring-settings-page" maxWidth="720px" spacing={3}>
      <PreferencesPageHeader
        testIdPrefix="preferences-monitoring-settings"
        title={getLabel("navigation.Monitoring")}
        summary={getLabel("admin.landing.monitoringPageSummary")}
      />
      <MonitoringSettingsTabs value="metrics" />
      <Stack data-testid="preferences-monitoring-settings-performance-section" spacing={2}>
        <MonitoringPollingSettingsCard
          description={getLabel("admin.landing.pollSystem")}
          storageKey="system-metrics"
          title={getLabel("admin.landing.systemMetrics")}
        />
        <MonitoringPollingSettingsCard
          description={getLabel("admin.landing.pollNetwork")}
          storageKey="network-metrics"
          title={getLabel("admin.landing.networkMetrics")}
        />
        <MonitoringPollingSettingsCard
          description={getLabel("admin.landing.pollDisk")}
          storageKey="disk-metrics"
          title={getLabel("admin.landing.diskMetrics")}
        />
        <MonitoringPollingSettingsCard
          description={getLabel("admin.landing.pollZfs")}
          storageKey="zfs-metrics"
          title={getLabel("admin.landing.zfsMetrics")}
        />
      </Stack>
    </Stack>
  );
};
