import { MonitoringSettingsTabs } from "@components/monitoring/MonitoringSettingsTabs";
import { ProcessPollingSettingsCard } from "@components/monitoring/ProcessPollingSettingsCard";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Preferences page for process- and service-oriented snapshot polling intervals.
 */
export const PreferencesProcessesPollingSettingsPage = (): ReactElement => {
  return (
    <Stack data-testid="preferences-processes-polling-settings-page" maxWidth="720px" spacing={3}>
      <Stack data-testid="preferences-processes-polling-settings-header" spacing={1}>
        <Typography
          color="primary"
          data-testid="preferences-processes-polling-overline"
          variant="overline"
        >
          Preferences
        </Typography>
        <Typography data-testid="preferences-processes-polling-title" variant="h1">
          Processes
        </Typography>
        <Typography
          color="text.secondary"
          data-testid="preferences-processes-polling-summary"
          variant="body1"
        >
          Configure the refresh interval used by process and service snapshot views. History and
          cache retention stay pinned to minimal internal defaults because these views do not keep
          time-series history yet.
        </Typography>
      </Stack>
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
