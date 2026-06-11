import { AlertsControlPanelFilters } from "@components/alerts/AlertsControlPanel/AlertsControlPanelFilters";
import { AlertsControlPanelSearch } from "@components/alerts/AlertsControlPanel/AlertsControlPanelSearch";
import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Composes separate filters and search panels for one alerts route slice.
 */
export const AlertsControlPanel = (): ReactElement => {
  return (
    <Box
      data-testid="alerts-control-panel"
      display="flex"
      flexDirection="column"
      gap={2}
      width="100%"
    >
      <Paper data-testid="alerts-filters-panel" sx={{ p: 2 }}>
        <AlertsControlPanelFilters />
      </Paper>
      <Paper data-testid="alerts-search-panel" sx={{ p: 2 }}>
        <Stack alignItems="flex-start" spacing={2}>
          <AlertsControlPanelSearch />
        </Stack>
      </Paper>
    </Box>
  );
};
