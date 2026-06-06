import { AlertsControlPanelFilters } from "@components/alerts/AlertsControlPanel/AlertsControlPanelFilters";
import { AlertsControlPanelSearch } from "@components/alerts/AlertsControlPanel/AlertsControlPanelSearch";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Composes filters and search for one alerts route slice.
 */
export const AlertsControlPanel = (): ReactElement => {
  return (
    <Paper data-testid="alerts-control-panel" sx={{ p: 2 }}>
      <Stack spacing={2}>
        <AlertsControlPanelFilters />
        <AlertsControlPanelSearch />
      </Stack>
    </Paper>
  );
};
