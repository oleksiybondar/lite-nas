import Box from "@mui/material/Box";
import Tab from "@mui/material/Tab";
import Tabs from "@mui/material/Tabs";
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
    <Box data-testid="monitoring-settings-tabs" sx={{ borderBottom: 1, borderColor: "divider" }}>
      <Tabs aria-label="Monitoring settings sections" value={value} variant="fullWidth">
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
      </Tabs>
    </Box>
  );
};
