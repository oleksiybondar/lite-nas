import { MonitoringSettingsTabs } from "@components/monitoring/MonitoringSettingsTabs";
import { ProcessPollingSettingsCard } from "@components/monitoring/ProcessPollingSettingsCard";
import { useLanguage } from "@lite-nas/language";
import { PreferencesPageHeader } from "@lite-nas/settings";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Preferences page for process- and service-oriented snapshot polling intervals.
 */
export const PreferencesProcessesPollingSettingsPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack data-testid="preferences-processes-polling-settings-page" maxWidth="720px" spacing={3}>
      <PreferencesPageHeader
        headerTestId="preferences-processes-polling-settings-header"
        testIdPrefix="preferences-processes-polling"
        title={getLabel("navigation.Processes")}
        summary={getLabel("admin.landing.processPollingSummary")}
      />
      <MonitoringSettingsTabs value="processes" />
      <Stack data-testid="preferences-processes-polling-section" spacing={2}>
        <ProcessPollingSettingsCard
          description={getLabel("admin.landing.processPollProcesses")}
          storageKey="process-metrics"
          title={getLabel("admin.landing.processMetrics")}
        />
        <ProcessPollingSettingsCard
          description={getLabel("admin.landing.processPollServices")}
          storageKey="service-metrics"
          title={getLabel("admin.landing.serviceMetrics")}
        />
      </Stack>
    </Stack>
  );
};
