import { MonitoringSettingsTabs } from "@components/monitoring/MonitoringSettingsTabs";
import { ProcessPollingSettingsCard } from "@components/monitoring/ProcessPollingSettingsCard";
import { PreferencesPageHeader } from "@lite-nas/settings";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Preferences page for process- and service-oriented snapshot polling intervals.
 */
export const PreferencesProcessesPollingSettingsPage = (): ReactElement => {
  return (
    <Stack data-testid="preferences-processes-polling-settings-page" maxWidth="720px" spacing={3}>
      <PreferencesPageHeader
        headerTestId="preferences-processes-polling-settings-header"
        testIdPrefix="preferences-processes-polling"
        title="Processes"
        summary="Configure the refresh interval used by process and service snapshot views. History and cache retention stay pinned to minimal internal defaults because these views do not keep time-series history yet."
      />
      <MonitoringSettingsTabs value="processes" />
      <Stack data-testid="preferences-processes-polling-section" spacing={2}>
        <ProcessPollingSettingsCard
          description="Polling interval used by process list and runtime workload snapshot views."
          storageKey="process-metrics"
          title="Process metrics"
        />
        <ProcessPollingSettingsCard
          description="Polling interval used by service state and systemd-managed workload snapshot views."
          storageKey="service-metrics"
          title="Service metrics"
        />
      </Stack>
    </Stack>
  );
};
