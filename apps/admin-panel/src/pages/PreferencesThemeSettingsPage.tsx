import type { ApiContextValue } from "@dto/api/api";
import { saveAdminSettings, saveSharedMode } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import { useAuth } from "@hooks/useAuth";
import { useThemeManager } from "@hooks/useThemeManager";
import { ThemePreferencesPanel } from "@lite-nas/theme";
import { saveThemeSettings } from "@theme/manager/storage";
import type { ThemeSettings } from "@theme/types";
import type { ReactElement } from "react";
import { useState } from "react";

/** Persists a changed theme to the authenticated admin and shared UI profiles. */
const persistThemePreferences = async (
  api: Pick<ApiContextValue, "get" | "put">,
  mode: ThemeSettings["mode"],
  setError: (error: string | null) => void,
): Promise<void> => {
  try {
    await Promise.all([saveAdminSettings(api), saveSharedMode(api, mode)]);
    setError(null);
  } catch (error) {
    const cause = error instanceof Error ? ` ${error.message}` : "";
    setError(`Saved in this browser, but the user file could not be updated.${cause}`);
  }
};

/** Connects the shared appearance screen to admin theme state and profile persistence. */
export const PreferencesThemeSettingsPage = (): ReactElement => {
  const theme = useThemeManager();
  const { get, put } = useApi();
  const { isAuthenticated } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const current: ThemeSettings = {
    mode: theme.mode,
    source: theme.source,
    templateName: theme.templateName,
  };
  const [saved, setSaved] = useState<ThemeSettings>(current);
  const isChanged =
    current.mode !== saved.mode ||
    current.source !== saved.source ||
    current.templateName !== saved.templateName;

  const apply = (): void => {
    if (!isChanged) return;
    saveThemeSettings(current);
    setSaved(current);
    if (isAuthenticated) {
      void persistThemePreferences({ get, put }, theme.resolvedMode, setError);
    }
  };

  return (
    <ThemePreferencesPanel
      appName="admin-panel"
      availableTemplates={theme.availableTemplates}
      error={error}
      isChanged={isChanged}
      mode={theme.source === "user" ? theme.mode : theme.resolvedMode}
      source={theme.source}
      templateName={theme.source === "user" ? theme.templateName : theme.resolvedTemplateName}
      onApply={apply}
      onCancel={() => theme.setSettings(saved)}
      onModeChange={theme.setMode}
      onSourceChange={theme.setSource}
      onTemplateChange={theme.setTemplateName}
    />
  );
};
