import { useAlertsControlPanel } from "@hooks/useAlertsControlPanel";
import { useLanguage } from "@lite-nas/language";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";

/**
 * Renders the local search-highlight input for alerts page content.
 */
export const AlertsControlPanelSearch = (): ReactElement => {
  const { search, setSearch } = useAlertsControlPanel();
  const { getLabel } = useLanguage();

  return (
    <TextField
      data-testid="alerts-search-control"
      label={getLabel("admin.alerts.searchCurrentPage")}
      name="alertsSearch"
      onChange={(event) => {
        setSearch(event.target.value);
      }}
      sx={{ width: 220, maxWidth: "100%" }}
      value={search}
    />
  );
};
