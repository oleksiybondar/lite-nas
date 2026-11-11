import { useLanguage } from "@lite-nas/language";
import IconButton from "@mui/material/IconButton";
import { useTheme } from "@mui/material/styles";
import type { ReactElement } from "react";

const themeColors = ["primary", "secondary", "success", "warning", "error"] as const;
type Props = { color: string; index: number; onChange: (color: string) => void };

/** Shows color choices taken from the active application theme. */
export const SensorChartThemeSwatches = ({ color, index, onChange }: Props): ReactElement => {
  const { getLabel } = useLanguage();
  const theme = useTheme();
  return (
    <>
      {themeColors.map((name) => {
        const choice = theme.palette[name].main;
        const label = `${getLabel("admin.sensors.themeColor")}: ${getLabel(`admin.sensors.${name}`)}`;
        return (
          <IconButton
            aria-label={label}
            aria-pressed={color.toLowerCase() === choice.toLowerCase()}
            data-test-class="sensor-chart-theme-color"
            data-test-name={`${index}-${name}`}
            key={name}
            onClick={() => onChange(choice)}
            size="small"
            sx={{
              bgcolor: choice,
              border: 2,
              borderColor:
                color.toLowerCase() === choice.toLowerCase() ? "text.primary" : "transparent",
              height: 26,
              width: 26,
              "&:hover": { bgcolor: choice },
            }}
            title={label}
          />
        );
      })}
    </>
  );
};
