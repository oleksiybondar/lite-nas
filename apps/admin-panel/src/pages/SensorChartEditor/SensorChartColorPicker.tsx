import { useLanguage } from "@lite-nas/language";
import Autocomplete from "@mui/material/Autocomplete";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import { type Theme, useTheme } from "@mui/material/styles";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { type HTMLAttributes, type Key, type ReactElement, useEffect, useState } from "react";

const themeColors = ["primary", "secondary", "info", "success", "warning", "error"] as const;
const presetColors = [
  { name: "red", value: "#ff0000" },
  { name: "orange", value: "#ff9800" },
  { name: "yellow", value: "#ffeb3b" },
  { name: "green", value: "#4caf50" },
  { name: "cyan", value: "#00bcd4" },
  { name: "blue", value: "#2196f3" },
  { name: "purple", value: "#9c27b0" },
  { name: "pink", value: "#e91e63" },
  { name: "gray", value: "#808080" },
  { name: "black", value: "#000000" },
  { name: "white", value: "#ffffff" },
] as const;

type ColorOption = { label: string; value: string };
type Props = { color: string; index: number; onChange: (color: string) => void };

/** Builds the color choices from the currently selected theme and fixed presets. */
const colorOptions = (theme: Theme, getLabel: (key: string) => string): ColorOption[] => [
  ...themeColors.map((name) => ({
    label: `${getLabel("admin.sensors.themeColor")}: ${getLabel(`admin.sensors.${name}`)}`,
    value: theme.palette[name].main,
  })),
  ...presetColors.map(({ name, value }) => ({
    label: getLabel(`admin.sensors.colors.${name}`),
    value,
  })),
];

/** Renders a searchable choice with a color preview and its hex value. */
const renderColorOption = (
  props: HTMLAttributes<HTMLLIElement> & { key: Key },
  option: ColorOption,
): ReactElement => (
  <li {...props} key={`${option.label}-${option.value}`}>
    <Box
      aria-hidden="true"
      sx={{
        bgcolor: option.value,
        border: 1,
        borderColor: "divider",
        borderRadius: 0.5,
        height: 18,
        mr: 1,
        width: 18,
      }}
    />
    <Typography variant="body2">
      {option.label} · {option.value}
    </Typography>
  </li>
);

/** Returns only complete hex colors, leaving search terms out of saved settings. */
const validColor = (option: string | ColorOption | null): string | null => {
  const value = typeof option === "string" ? option : option?.value;
  return value && /^#[0-9a-fA-F]{6}$/.test(value) ? value : null;
};

/** Offers searchable theme and common colors alongside a custom hex and native color input. */
export const SensorChartColorPicker = ({ color, index, onChange }: Props): ReactElement => {
  const { getLabel } = useLanguage();
  const [inputValue, setInputValue] = useState(color);

  useEffect(() => setInputValue(color), [color]);

  const options = colorOptions(useTheme(), getLabel);

  return (
    <Stack direction="row" gap={1} sx={{ minWidth: 220 }}>
      <TextField
        data-test-class="sensor-chart-color-picker"
        data-test-name={String(index)}
        onChange={(event) => onChange(event.target.value)}
        size="small"
        slotProps={{ htmlInput: { "aria-label": getLabel("admin.sensors.colorPicker") } }}
        sx={{ width: 48, "& input": { cursor: "pointer", height: 28, p: 0.5 } }}
        type="color"
        value={color}
      />
      <Autocomplete<ColorOption, false, false, true>
        data-test-class="sensor-chart-color-dropdown"
        data-test-name={String(index)}
        freeSolo
        getOptionLabel={(option) =>
          typeof option === "string" ? option : `${option.label} ${option.value}`
        }
        inputValue={inputValue}
        onBlur={() => setInputValue((current) => validColor(current) ?? color)}
        onChange={(_, option) => {
          const selected = validColor(option);
          if (selected === null) return;
          setInputValue(selected);
          onChange(selected);
        }}
        onInputChange={(_, value, reason) => {
          if (reason !== "input") return;
          setInputValue(value);
          if (validColor(value) !== null) onChange(value);
        }}
        options={options}
        renderInput={(params) => (
          <TextField {...params} label={getLabel("admin.sensors.color")} size="small" />
        )}
        renderOption={renderColorOption}
        sx={{ minWidth: 170, flex: 1 }}
        value={null}
      />
    </Stack>
  );
};
