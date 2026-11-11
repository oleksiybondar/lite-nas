import { useLanguage } from "@lite-nas/language";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";
import { SensorChartThemeSwatches } from "./SensorChartThemeSwatches";

type Props = { color: string; index: number; onChange: (color: string) => void };

/** Offers a native color chooser and theme colors for one sensor chart stop. */
export const SensorChartColorPicker = ({ color, index, onChange }: Props): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack direction="row" flexWrap="wrap" gap={0.5} sx={{ minWidth: 140 }} useFlexGap>
      <TextField
        data-test-class="sensor-chart-color-picker"
        data-test-name={String(index)}
        onChange={(event) => onChange(event.target.value)}
        size="small"
        slotProps={{ htmlInput: { "aria-label": getLabel("admin.sensors.color") } }}
        sx={{ width: 48, "& input": { cursor: "pointer", height: 28, p: 0.5 } }}
        type="color"
        value={color}
      />
      <SensorChartThemeSwatches color={color} index={index} onChange={onChange} />
    </Stack>
  );
};
