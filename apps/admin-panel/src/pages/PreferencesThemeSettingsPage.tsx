import { ThemeModeToggle, ThemeSourceSelector, ThemeTemplateSelector } from "@components/theme";
import { useThemeManager } from "@hooks/useThemeManager";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { saveThemeSettings } from "@theme/manager/storage";
import type {
  ThemeManagerContextValue,
  ThemeMode,
  ThemeSettings,
  ThemeSource,
  ThemeTemplateName,
} from "@theme/types";
import type { ReactElement } from "react";
import { useState } from "react";

/**
 * Preferences page for admin-panel appearance settings.
 */
export const PreferencesThemeSettingsPage = (): ReactElement => {
  const themePreferences = useThemePreferencesForm();

  return (
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
              Adjust the admin-panel appearance and persist the selected theme behavior for this
              browser.
            </Typography>
          </Stack>
          <ThemeSourceSelector
            onChange={themePreferences.handleSourceChange}
            value={themePreferences.source}
          />
          <ThemeModeToggle
            mode={themePreferences.displayedMode}
            onChange={themePreferences.handleModeChange}
            source={themePreferences.source}
          />
          <ThemeTemplateSelector
            onChange={themePreferences.handleTemplateChange}
            source={themePreferences.source}
            value={themePreferences.displayedTemplateName}
          />
          {themePreferences.isChanged ? (
            <ThemePreferencesActions preferences={themePreferences} />
          ) : null}
        </Stack>
      </Paper>
    </Stack>
  );
};

type ThemePreferencesActionsProps = {
  /**
   * Editable theme preference state and handlers.
   */
  preferences: ThemePreferencesForm;
};

/**
 * Action row shown when appearance settings have unsaved changes.
 */
const ThemePreferencesActions = ({ preferences }: ThemePreferencesActionsProps): ReactElement => {
  return (
    <Stack
      data-testid="theme-preferences-actions"
      direction={{ sm: "row", xs: "column-reverse" }}
      justifyContent="flex-end"
      spacing={1.5}
    >
      <Button
        data-testid="theme-preferences-cancel-button"
        onClick={preferences.handleCancel}
        variant="outlined"
      >
        Cancel
      </Button>
      <Button
        data-testid="theme-preferences-apply-button"
        onClick={preferences.handleApply}
        variant="contained"
      >
        Apply
      </Button>
    </Stack>
  );
};

type ThemePreferencesForm = {
  displayedMode: ThemeMode;
  displayedTemplateName: ThemeTemplateName;
  handleApply: () => void;
  handleCancel: () => void;
  handleModeChange: (nextMode: ThemeMode) => void;
  handleSourceChange: (nextSource: ThemeSource) => void;
  handleTemplateChange: (nextTemplateName: ThemeTemplateName) => void;
  isChanged: boolean;
  source: ThemeSource;
};

/**
 * State and actions for editable application theme preferences.
 */
const useThemePreferencesForm = (): ThemePreferencesForm => {
  const themeManager = useThemeManager();
  const currentSettings = pickThemeSettings(themeManager);
  const [savedSettings, setSavedSettings] = useState<ThemeSettings>(currentSettings);
  const isChanged = hasThemeSettingsChanged(currentSettings, savedSettings);

  return {
    displayedMode: resolveDisplayedMode(themeManager),
    displayedTemplateName: resolveDisplayedTemplateName(themeManager),
    handleApply: () => persistThemeSettings(currentSettings, isChanged, setSavedSettings),
    handleCancel: () => themeManager.setSettings(savedSettings),
    handleModeChange: themeManager.setMode,
    handleSourceChange: themeManager.setSource,
    handleTemplateChange: themeManager.setTemplateName,
    isChanged,
    source: themeManager.source,
  };
};

const pickThemeSettings = ({
  mode,
  source,
  templateName,
}: ThemeManagerContextValue): ThemeSettings => {
  return { mode, source, templateName };
};

const hasThemeSettingsChanged = (
  currentSettings: ThemeSettings,
  savedSettings: ThemeSettings,
): boolean => {
  return (
    currentSettings.source !== savedSettings.source ||
    currentSettings.mode !== savedSettings.mode ||
    currentSettings.templateName !== savedSettings.templateName
  );
};

const resolveDisplayedMode = ({
  mode,
  resolvedMode,
  source,
}: ThemeManagerContextValue): ThemeMode => {
  return source === "user" ? mode : resolvedMode;
};

const resolveDisplayedTemplateName = ({
  resolvedTemplateName,
  source,
  templateName,
}: ThemeManagerContextValue): ThemeTemplateName => {
  return source === "user" ? templateName : resolvedTemplateName;
};

const persistThemeSettings = (
  currentSettings: ThemeSettings,
  isChanged: boolean,
  setSavedSettings: (nextSettings: ThemeSettings) => void,
): void => {
  if (!isChanged) {
    return;
  }

  saveThemeSettings(currentSettings);
  setSavedSettings(currentSettings);
};
