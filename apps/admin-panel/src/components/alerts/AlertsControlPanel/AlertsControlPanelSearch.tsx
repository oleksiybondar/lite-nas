import { useAlertsControlPanel } from "@hooks/useAlertsControlPanel";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";

/**
 * Renders the local search-highlight input for alerts page content.
 */
export const AlertsControlPanelSearch = (): ReactElement => {
  const { search, setSearch } = useAlertsControlPanel();

  return (
    <TextField
      data-testid="alerts-search-control"
      label="Search current page"
      name="alertsSearch"
      onChange={(event) => {
        setSearch(event.target.value);
      }}
      sx={{ width: 220, maxWidth: "100%" }}
      value={search}
    />
  );
};
