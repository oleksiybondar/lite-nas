import { collectAdminSettings } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";
import { DashboardEditor } from "@pages/DashboardPage/DashboardEditor";
import { DashboardHeader } from "@pages/DashboardPage/DashboardHeader";
import { DashboardWidgetTile } from "@pages/DashboardPage/DashboardWidgetTile";
import { adminSettingsSchema, dashboardSchema } from "@schemas/profile/admin-settings";
import { type ReactElement, useEffect, useState } from "react";
import { z } from "zod";

const widgetKindSchema = z.enum([
  "cpu",
  "cpu-cores",
  "ram",
  "zfs-io",
  "zfs-errors",
  "network",
  "disk",
  "processes",
  "services",
]);
type Widget = z.infer<typeof dashboardSchema>["widgets"][number];
const defaults: z.infer<typeof dashboardSchema> = {
  version: 2,
  widgets: [
    { id: "cpu", kind: "cpu", width: 3, height: 1 },
    { id: "ram", kind: "ram", width: 3, height: 1 },
    { id: "zfs-io", kind: "zfs-io", width: 3, height: 1 },
    { id: "zfs-errors", kind: "zfs-errors", width: 3, height: 1 },
  ],
};

/** Keeps one dashboard tile per metric kind when loading older saved layouts. */
const uniqueWidgets = (widgets: Widget[]): Widget[] =>
  widgets.filter(
    (widget, index) => widgets.findIndex((candidate) => candidate.kind === widget.kind) === index,
  );

/** Applies a validated add request from telemetry navigation to the loaded layout. */
const addRequestedWidget = (
  current: Widget[],
  setEditing: (editing: boolean) => void,
): Widget[] => {
  const requested = widgetKindSchema.safeParse(
    new URLSearchParams(window.location.search).get("add"),
  );
  if (!requested.success) return current;
  window.history.replaceState(window.history.state, "", window.location.pathname);
  if (current.length >= 16) return current;
  if (current.some((widget) => widget.kind === requested.data)) return current;
  setEditing(true);
  return [...current, { id: crypto.randomUUID(), kind: requested.data, width: 3, height: 1 }];
};

/** Reads the saved dashboard, falling back to defaults for missing or invalid profiles. */
const readDashboard = async (get: ReturnType<typeof useApi>["get"]): Promise<Widget[]> => {
  const response = await get("/f/profile/admin-panel").execute();
  if (!response.ok) return defaults.widgets;
  const parsed = adminSettingsSchema.safeParse(await response.json());
  if (!parsed.success || !parsed.data.dashboard) return defaults.widgets;
  window.localStorage.setItem(
    "lite-nas.admin-panel.dashboard",
    JSON.stringify(parsed.data.dashboard),
  );
  return uniqueWidgets(parsed.data.dashboard.widgets);
};

/** Saves the dashboard while preserving the rest of the current profile document. */
const persistDashboard = async (
  get: ReturnType<typeof useApi>["get"],
  put: ReturnType<typeof useApi>["put"],
  widgets: Widget[],
): Promise<void> => {
  const dashboard = dashboardSchema.parse({ version: 2, widgets });
  const response = await get("/f/profile/admin-panel").execute();
  const current = response.ok ? adminSettingsSchema.safeParse(await response.json()) : null;
  if (current && !current.success) throw new Error("Invalid profile document");
  const document = adminSettingsSchema.parse({
    ...(current?.success ? current.data : collectAdminSettings()),
    dashboard,
  });
  const saved = await put("/f/profile/admin-panel", { document }).execute();
  if (!saved.ok) throw new Error("Profile save failed");
  window.localStorage.setItem("lite-nas.admin-panel.dashboard", JSON.stringify(dashboard));
};

/** Downloads the validated dashboard layout as JSON. */
const downloadDashboard = (widgets: Widget[]): void => {
  const blob = new Blob([JSON.stringify(dashboardSchema.parse({ version: 2, widgets }), null, 2)], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = "dashboard.json";
  anchor.click();
  URL.revokeObjectURL(url);
};

/** Loads the saved layout and applies any tile request from telemetry navigation. */
const useLoadedDashboard = (
  get: ReturnType<typeof useApi>["get"],
  setWidgets: (widgets: Widget[]) => void,
  setEditing: (editing: boolean) => void,
): void => {
  useEffect(() => {
    let active = true;
    void readDashboard(get)
      .then((loaded) => {
        if (active) setWidgets(addRequestedWidget(loaded, setEditing));
      })
      .catch(() => {
        if (active) setWidgets(addRequestedWidget(defaults.widgets, setEditing));
      });
    return () => {
      active = false;
    };
  }, [get, setEditing, setWidgets]);
};

/** Renders a configurable six-column telemetry dashboard and persists its JSON profile. */
export const DashboardPage = (): ReactElement => {
  const { get, put } = useApi();
  const [widgets, setWidgets] = useState<Widget[]>(defaults.widgets);
  const [editing, setEditing] = useState(false);
  const [saved, setSaved] = useState(false);
  const [saveFailed, setSaveFailed] = useState(false);
  useLoadedDashboard(get, setWidgets, setEditing);
  const save = async (): Promise<void> => {
    try {
      await persistDashboard(get, put, widgets);
      setSaved(true);
      setSaveFailed(false);
    } catch {
      setSaved(false);
      setSaveFailed(true);
    }
  };
  return (
    <Stack data-testid="dashboard-page" spacing={3}>
      <DashboardHeader
        editing={editing}
        onToggleEditing={() => {
          setEditing(!editing);
          setSaved(false);
        }}
        onDownload={() => downloadDashboard(widgets)}
      />
      {editing && (
        <DashboardEditor
          widgets={widgets}
          setWidgets={setWidgets}
          save={save}
          saved={saved}
          saveFailed={saveFailed}
        />
      )}
      <Grid columns={{ xs: 1, md: 6 }} container data-testid="dashboard-widget-grid" spacing={2}>
        {widgets.map((widget) => (
          <DashboardWidgetTile
            editing={editing}
            key={widget.id}
            setWidgets={setWidgets}
            widget={widget}
            widgets={widgets}
          />
        ))}
      </Grid>
    </Stack>
  );
};
