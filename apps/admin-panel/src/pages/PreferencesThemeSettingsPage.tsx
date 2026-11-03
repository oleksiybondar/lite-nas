import { saveAdminSettings, saveSharedMode } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import { useAuth } from "@hooks/useAuth";
import { useThemeManager } from "@hooks/useThemeManager";
import { ThemePreferencesPanel } from "@lite-nas/theme";
import { saveThemeSettings } from "@theme/manager/storage";
import type { ThemeSettings } from "@theme/types";
import type { ReactElement } from "react";
import { useState } from "react";

/** Connects the shared appearance screen to admin theme state and profile persistence. */
export const PreferencesThemeSettingsPage = (): ReactElement => {
  const theme = useThemeManager();
  const { put } = useApi();
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
      void Promise.all([saveAdminSettings({ put }), saveSharedMode({ put }, theme.resolvedMode)])
        .then(() => setError(null))
        .catch((error: unknown) => {
          const cause = error instanceof Error ? ` ${error.message}` : "";
          setError(`Saved in this browser, but the user file could not be updated.${cause}`);
        });
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
