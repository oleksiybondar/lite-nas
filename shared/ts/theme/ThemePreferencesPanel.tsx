import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { ThemeModeToggle } from "./ThemeModeToggle";
import { ThemeSourceSelector } from "./ThemeSourceSelector";
import { ThemeTemplateSelector } from "./ThemeTemplateSelector";
import type { ThemeMode, ThemeSource, ThemeTemplateName } from "./types";

/** Editable appearance state and actions supplied by the hosting application. */
export interface ThemePreferencesPanelProps {
  appName: string;
  availableTemplates: ThemeTemplateName[];
  error: string | null;
  isChanged: boolean;
  mode: ThemeMode;
  source: ThemeSource;
  templateName: ThemeTemplateName;
  onApply: () => void;
  onCancel: () => void;
  onModeChange: (mode: ThemeMode) => void;
  onSourceChange: (source: ThemeSource) => void;
  onTemplateChange: (name: ThemeTemplateName) => void;
}

/** Reusable appearance settings screen; persistence remains with the host app. */
export const ThemePreferencesPanel = ({
  appName,
  availableTemplates,
  error,
  isChanged,
  mode,
  source,
  templateName,
  onApply,
  onCancel,
  onModeChange,
  onSourceChange,
  onTemplateChange,
}: ThemePreferencesPanelProps): ReactElement => (
  <Stack data-testid="preferences-theme-settings-page" maxWidth="720px" spacing={3}>
    <Stack data-testid="preferences-theme-settings-header" spacing={1}>
      <Typography
        color="primary"
        data-testid="preferences-theme-settings-overline"
        variant="overline"
      >
        Preferences
      </Typography>
      <Typography data-testid="preferences-theme-settings-title" variant="h1">
        Theme
      </Typography>
    </Stack>
    <Paper data-testid="preferences-theme-card" sx={{ p: 3 }}>
      <Stack spacing={3}>
        <Stack spacing={1}>
          <Typography data-testid="preferences-theme-title" variant="h2">
            Theme preferences
          </Typography>
          <Typography
            color="text.secondary"
            data-testid="preferences-theme-summary"
            variant="body2"
          >
            Adjust the {appName} appearance. Signed-in changes are saved to your profile and kept in
            this browser as a fallback.
          </Typography>
        </Stack>
        <ThemeSourceSelector onChange={onSourceChange} value={source} />
        <ThemeModeToggle mode={mode} onChange={onModeChange} source={source} />
        <ThemeTemplateSelector
          availableTemplates={availableTemplates}
          onChange={onTemplateChange}
          source={source}
          value={templateName}
        />
        {isChanged ? (
          <Stack
            data-testid="theme-preferences-actions"
            direction={{ sm: "row", xs: "column-reverse" }}
            justifyContent="flex-end"
            spacing={1.5}
          >
            <Button
              data-testid="theme-preferences-cancel-button"
              onClick={onCancel}
              variant="outlined"
            >
              Cancel
            </Button>
            <Button
              data-testid="theme-preferences-apply-button"
              onClick={onApply}
              variant="contained"
            >
              Apply
            </Button>
          </Stack>
        ) : null}
        {error ? <Alert severity="warning">{error}</Alert> : null}
      </Stack>
    </Paper>
  </Stack>
);
