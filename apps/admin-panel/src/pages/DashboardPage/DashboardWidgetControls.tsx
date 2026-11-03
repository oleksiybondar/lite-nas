import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import FormControl from "@mui/material/FormControl";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import type { dashboardSchema } from "@schemas/profile/admin-settings";
import type { Dispatch, ReactElement, SetStateAction } from "react";
import type { z } from "zod";

type Widget = z.infer<typeof dashboardSchema>["widgets"][number];

/** Moves a dashboard widget one position, retaining the original order at the bounds. */
const moveWidget = (widgets: Widget[], id: string, offset: number): Widget[] => {
  const next = [...widgets];
  const index = next.findIndex((item) => item.id === id);
  const target = index + offset;
  if (index < 0 || target < 0 || target >= next.length) return next;
  const moved = next[index];
  if (!moved) return next;
  next.splice(index, 1);
  next.splice(target, 0, moved);
  return next;
};

/** Edits one dashboard tile's dimensions, order, and presence. */
export const DashboardWidgetControls = ({
  widget,
  widgets,
  setWidgets,
}: {
  widget: Widget;
  widgets: Widget[];
  setWidgets: Dispatch<SetStateAction<Widget[]>>;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack alignItems="center" direction="row" flexWrap="wrap" gap={1}>
      <FormControl size="small">
        <Select
          aria-label={getLabel("admin.dashboard.width")}
          value={widget.width}
          onChange={(event) =>
            setWidgets((current) =>
              current.map((item) =>
                item.id === widget.id ? { ...item, width: Number(event.target.value) } : item,
              ),
            )
          }
        >
          {[1, 2, 3, 4, 5, 6].map((width) => (
            <MenuItem key={width} value={width}>
              {width}/6
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small">
        <Select
          aria-label={getLabel("admin.dashboard.height")}
          value={widget.height}
          onChange={(event) =>
            setWidgets((current) =>
              current.map((item) =>
                item.id === widget.id
                  ? { ...item, height: Number(event.target.value) as Widget["height"] }
                  : item,
              ),
            )
          }
        >
          <MenuItem value={1}>1</MenuItem>
          <MenuItem value={2}>2</MenuItem>
        </Select>
      </FormControl>
      <Button
        disabled={widgets[0]?.id === widget.id}
        onClick={() => setWidgets((current) => moveWidget(current, widget.id, -1))}
      >
        {getLabel("admin.dashboard.up")}
      </Button>
      <Button
        disabled={widgets.at(-1)?.id === widget.id}
        onClick={() => setWidgets((current) => moveWidget(current, widget.id, 1))}
      >
        {getLabel("admin.dashboard.down")}
      </Button>
      <Button
        onClick={() => setWidgets((current) => current.filter((item) => item.id !== widget.id))}
      >
        {getLabel("admin.dashboard.remove")}
      </Button>
    </Stack>
  );
};
