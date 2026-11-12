import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Displays the dashboard title and its edit and download actions. */
export const DashboardHeader = ({
  editing,
  onToggleEditing,
  onDownload,
}: {
  editing: boolean;
  onToggleEditing: () => void;
  onDownload: () => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack
      alignItems={{ sm: "center" }}
      direction={{ sm: "row", xs: "column" }}
      justifyContent="space-between"
      spacing={1}
    >
      <Stack>
        <Typography color="primary" variant="overline">
          {getLabel("admin.landing.dashboardOverline")}
        </Typography>
        <Typography data-testid="dashboard-page-title" variant="h1">
          {getLabel("admin.landing.dashboardTitle")}
        </Typography>
      </Stack>
      <Stack direction="row" spacing={1}>
        <Button
          data-testid="dashboard-edit-button"
          onClick={onToggleEditing}
          variant={editing ? "contained" : "outlined"}
        >
          {getLabel(editing ? "admin.dashboard.done" : "admin.dashboard.edit")}
        </Button>
        <Button data-testid="dashboard-download-button" onClick={onDownload}>
          {getLabel("admin.dashboard.download")}
        </Button>
      </Stack>
    </Stack>
  );
};
