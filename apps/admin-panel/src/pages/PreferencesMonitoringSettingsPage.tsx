import { MonitoringPollingSettingsCard } from "@components/monitoring/MonitoringPollingSettingsCard";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Preferences page for monitoring polling behavior used by telemetry pages.
 */
export const PreferencesMonitoringSettingsPage = (): ReactElement => {
  return (
    <Stack data-testid="preferences-monitoring-settings-page" maxWidth="720px" spacing={3}>
      <Stack data-testid="preferences-monitoring-settings-header" spacing={1}>
        <Typography
          color="primary"
          data-testid="preferences-monitoring-settings-overline"
          variant="overline"
        >
          Preferences
        </Typography>
        <Typography data-testid="preferences-monitoring-settings-title" variant="h1">
          Monitoring
        </Typography>
        <Typography
          color="text.secondary"
          data-testid="preferences-monitoring-settings-summary"
          variant="body1"
        >
          Configure how telemetry charts bootstrap history, poll live updates, and retain cached
          points in this browser.
        </Typography>
      </Stack>
      <MonitoringPollingSettingsCard
        description="Polling settings used by host CPU and memory telemetry views."
        storageKey="system-metrics"
        title="System metrics"
      />
      <MonitoringPollingSettingsCard
        description="Polling settings used by ZFS pool and storage telemetry views."
        storageKey="zfs-metrics"
        title="ZFS metrics"
      />
    </Stack>
  );
};
