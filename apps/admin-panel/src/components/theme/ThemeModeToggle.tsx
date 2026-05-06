import { useThemeManager } from "@hooks/useThemeManager";

import FormControlLabel from "@mui/material/FormControlLabel";
import Stack from "@mui/material/Stack";
import Switch from "@mui/material/Switch";
import Typography from "@mui/material/Typography";
import type { ThemeMode, ThemeSource } from "@theme/index";
import type { ReactElement } from "react";

interface ThemeModeToggleProps {
  mode?: ThemeMode;
  onChange?: (mode: ThemeMode) => void;
  source?: ThemeSource;
}

export const ThemeModeToggle = ({ mode, onChange, source }: ThemeModeToggleProps): ReactElement => {
  const { mode: currentMode, resolvedMode, setMode, source: currentSource } = useThemeManager();
  const selectedSource = source ?? currentSource;
  const selectedMode = mode ?? currentMode;
  const isDisabled = selectedSource !== "user";

  return (
    <Stack data-testid="theme-mode-control" spacing={1}>
      <Typography data-testid="theme-mode-label" variant="subtitle2">
        Theme mode
      </Typography>
      <FormControlLabel
        control={
          <Switch
            checked={selectedMode === "dark"}
            data-testid="theme-mode-switch"
            disabled={isDisabled}
            name="themeMode"
            onChange={(event) => {
              const nextMode = event.target.checked ? "dark" : "light";

              if (onChange !== undefined) {
                onChange(nextMode);
                return;
              }

              setMode(nextMode);
            }}
          />
        }
        data-testid="theme-mode-toggle"
        label={(mode ?? resolvedMode) === "dark" ? "Dark" : "Light"}
      />
    </Stack>
  );
};
