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
 * Preferences page for admin-panel application behavior and appearance.
 */
export const PreferencesApplicationSettingsPage = (): ReactElement => {
  const themePreferences = useThemePreferencesForm();

  return (
    <Stack data-testid="preferences-application-settings-page" maxWidth="720px" spacing={3}>
      <Stack data-testid="preferences-application-settings-header" spacing={1}>
        <Typography
          color="primary"
          data-testid="preferences-application-settings-overline"
          variant="overline"
        >
          Preferences
        </Typography>
        <Typography data-testid="preferences-application-settings-title" variant="h1">
          Application settings
        </Typography>
      </Stack>
      <Paper data-testid="preferences-theme-card" sx={{ p: 3 }}>
        <Stack spacing={3}>
          <Stack spacing={1}>
            <Typography data-testid="preferences-theme-title" variant="h2">
              Theme
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

/**
 * Props for the application theme preferences action row.
 */
type ThemePreferencesActionsProps = {
  /**
   * Editable theme preference state and handlers.
   */
  preferences: ThemePreferencesForm;
};

/**
 * Action row shown when application theme preferences have unsaved changes.
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

/**
 * Editable theme preference state and commands consumed by the page view.
 */
type ThemePreferencesForm = {
  /**
   * Mode displayed by the mode control.
   */
  displayedMode: ThemeMode;
  /**
   * Template displayed by the template control.
   */
  displayedTemplateName: ThemeTemplateName;
  /**
   * Persists the current theme settings.
   */
  handleApply: () => void;
  /**
   * Reverts controls to the last saved theme settings.
   */
  handleCancel: () => void;
  /**
   * Updates the editable theme mode.
   */
  handleModeChange: (nextMode: ThemeMode) => void;
  /**
   * Updates the editable theme source.
   */
  handleSourceChange: (nextSource: ThemeSource) => void;
  /**
   * Updates the editable theme template.
   */
  handleTemplateChange: (nextTemplateName: ThemeTemplateName) => void;
  /**
   * Whether editable settings differ from persisted settings.
   */
  isChanged: boolean;
  /**
   * Current editable theme source.
   */
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

/**
 * Extracts persistable settings from the theme manager context.
 */
const pickThemeSettings = ({
  mode,
  source,
  templateName,
}: ThemeManagerContextValue): ThemeSettings => {
  return { mode, source, templateName };
};

/**
 * Reports whether editable settings differ from the last saved settings.
 */
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

/**
 * Resolves the mode shown by the mode control for the current source.
 */
const resolveDisplayedMode = ({
  mode,
  resolvedMode,
  source,
}: ThemeManagerContextValue): ThemeMode => {
  return source === "user" ? mode : resolvedMode;
};

/**
 * Resolves the template shown by the template control for the current source.
 */
const resolveDisplayedTemplateName = ({
  resolvedTemplateName,
  source,
  templateName,
}: ThemeManagerContextValue): ThemeTemplateName => {
  return source === "user" ? templateName : resolvedTemplateName;
};

/**
 * Saves current theme settings when unsaved changes exist.
 */
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
