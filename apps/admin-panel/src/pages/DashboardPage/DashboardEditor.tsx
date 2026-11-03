import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { dashboardSchema } from "@schemas/profile/admin-settings";
import type { Dispatch, ReactElement, SetStateAction } from "react";
import type { z } from "zod";

type Widget = z.infer<typeof dashboardSchema>["widgets"][number];
const widgetKinds = [
  "cpu",
  "cpu-cores",
  "ram",
  "zfs-io",
  "zfs-errors",
  "network",
  "disk",
  "processes",
  "services",
] as const;

/** Displays the dashboard's tile picker and save feedback while editing. */
export const DashboardEditor = ({
  widgets,
  setWidgets,
  save,
  saved,
  saveFailed,
}: {
  widgets: Widget[];
  setWidgets: Dispatch<SetStateAction<Widget[]>>;
  save: () => Promise<void>;
  saved: boolean;
  saveFailed: boolean;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Paper sx={{ p: 2 }}>
      <Stack spacing={1}>
        <Typography variant="h2">{getLabel("admin.dashboard.widgets")}</Typography>
        <Stack direction="row" flexWrap="wrap" gap={1}>
          {widgetKinds.map((kind) => (
            <Button
              data-test-class="dashboard-add-widget"
              data-test-name={kind}
              disabled={widgets.length >= 16 || widgets.some((widget) => widget.kind === kind)}
              key={kind}
              onClick={() =>
                setWidgets((current) =>
                  current.some((widget) => widget.kind === kind)
                    ? current
                    : [...current, { id: crypto.randomUUID(), kind, width: 3, height: 1 }],
                )
              }
            >
              {getLabel(`admin.dashboard.${kind}`)}
            </Button>
          ))}
          <Button data-testid="dashboard-save-button" onClick={() => void save()}>
            {getLabel("admin.dashboard.save")}
          </Button>
          {saved && <Typography role="status">{getLabel("admin.dashboard.saved")}</Typography>}
          {saveFailed && (
            <Typography color="error" role="alert">
              {getLabel("admin.dashboard.saveFailed")}
            </Typography>
          )}
        </Stack>
      </Stack>
    </Paper>
  );
};
