import { useLanguage } from "@lite-nas/language";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { DashboardWidgetControls } from "@pages/DashboardPage/DashboardWidgetControls";
import { DashboardWidgetMetric } from "@pages/DashboardPage/DashboardWidgetMetric";
import type { dashboardSchema } from "@schemas/profile/admin-settings";
import type { Dispatch, ReactElement, SetStateAction } from "react";
import type { z } from "zod";

type Widget = z.infer<typeof dashboardSchema>["widgets"][number];

/** Renders the controls and metric content for one dashboard tile. */
export const DashboardWidgetTile = ({
  widget,
  widgets,
  editing,
  setWidgets,
}: {
  widget: Widget;
  widgets: Widget[];
  editing: boolean;
  setWidgets: Dispatch<SetStateAction<Widget[]>>;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Grid
      data-test-class="dashboard-widget"
      data-test-name={widget.kind}
      key={widget.id}
      size={{ xs: 1, md: widget.width }}
    >
      <Stack spacing={1}>
        {editing && (
          <DashboardWidgetControls setWidgets={setWidgets} widget={widget} widgets={widgets} />
        )}
        <Paper sx={{ minHeight: widget.height === 2 ? 480 : 240, p: 2 }}>
          <Stack spacing={1}>
            <Typography variant="h2">{getLabel(`admin.dashboard.${widget.kind}`)}</Typography>
            <DashboardWidgetMetric kind={widget.kind} />
          </Stack>
        </Paper>
      </Stack>
    </Grid>
  );
};
