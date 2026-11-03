import { PreferencesTabs } from "@lite-nas/settings";
import Tab from "@mui/material/Tab";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

type MonitoringSettingsTabsProps = {
  /**
   * Currently selected monitoring settings tab.
   */
  value: "metrics" | "processes";
};

/**
 * Route-backed tabs that switch between metrics and process polling settings pages.
 */
export const MonitoringSettingsTabs = ({ value }: MonitoringSettingsTabsProps): ReactElement => {
  return (
    <PreferencesTabs
      label="Monitoring settings sections"
      testId="monitoring-settings-tabs"
      value={value}
    >
      <Tab
        component={RouterLink}
        label="Metrics"
        to="/preferences/application/monitoring"
        value="metrics"
      />
      <Tab
        component={RouterLink}
        label="Processes"
        to="/preferences/application/monitoring/processes"
        value="processes"
      />
    </PreferencesTabs>
  );
};
