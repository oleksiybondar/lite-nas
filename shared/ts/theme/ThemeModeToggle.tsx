import { useLanguage } from "@lite-nas/language";
import FormControlLabel from "@mui/material/FormControlLabel";
import Stack from "@mui/material/Stack";
import Switch from "@mui/material/Switch";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { ThemeMode, ThemeSource } from "./types";

/** Controlled light/dark preference switch for LiteNAS settings pages. */
export const ThemeModeToggle = ({
  mode,
  onChange,
  source,
}: {
  mode: ThemeMode;
  onChange: (mode: ThemeMode) => void;
  source: ThemeSource;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack data-testid="theme-mode-control" spacing={1}>
      <Typography data-testid="theme-mode-label" variant="subtitle2">
        {getLabel("theme.mode")}
      </Typography>
      <FormControlLabel
        control={
          <Switch
            checked={mode === "dark"}
            data-testid="theme-mode-switch"
            disabled={source !== "user"}
            name="themeMode"
            onChange={(event) => onChange(event.target.checked ? "dark" : "light")}
          />
        }
        data-testid="theme-mode-toggle"
        label={getLabel(mode === "dark" ? "theme.modes.dark" : "theme.modes.light")}
      />
    </Stack>
  );
};
