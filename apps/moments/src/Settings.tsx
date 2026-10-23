import { ProfilePreferencesPanel, writeProfileAvatar } from "@lite-nas/profile";
import { type ThemeMode, ThemePreferencesPanel } from "@lite-nas/theme";
import type { ReactElement } from "react";
import { useState } from "react";
import { request } from "./api";

/** Shows the shared profile or appearance screen without admin-only settings. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: Both shared panels use the same local draft and feedback state.
export const Settings = ({
  view,
  username,
  avatarUrl,
  fallbackAvatarUrl,
  mode,
  message,
  onModeChange,
  onSaveMode,
}: {
  view: "profile" | "appearance";
  username: string;
  avatarUrl: string | null;
  fallbackAvatarUrl?: string | null;
  mode: ThemeMode;
  message: string | null;
  onModeChange: (value: ThemeMode) => void;
  onSaveMode: (value: ThemeMode) => void;
}): ReactElement => {
  const [savedMode, setSavedMode] = useState(mode);
  const [avatarMessage, setAvatarMessage] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const upload = async (file: File | undefined): Promise<void> => {
    if (!file) return;
    setBusy(true);
    try {
      await writeProfileAvatar(
        (path, image) => request(path, { method: "PUT", body: image }),
        file,
      );
      window.dispatchEvent(new Event("lite-nas-avatar-updated"));
      setAvatarMessage("Avatar saved.");
    } catch {
      setAvatarMessage("Could not save avatar.");
    } finally {
      setBusy(false);
    }
  };

  if (view === "profile") {
    return (
      <ProfilePreferencesPanel
        username={username}
        avatarUrl={avatarUrl}
        fallbackAvatarUrl={fallbackAvatarUrl}
        busy={busy}
        message={avatarMessage}
        onAvatarSelected={(file) => void upload(file)}
      />
    );
  }

  return (
    <ThemePreferencesPanel
      appName="Moments"
      availableTemplates={["default"]}
      error={message}
      isChanged={mode !== savedMode}
      mode={mode}
      source="user"
      templateName="default"
      showSourceAndTemplate={false}
      onApply={() => {
        setSavedMode(mode);
        onSaveMode(mode);
      }}
      onCancel={() => onModeChange(savedMode)}
      onModeChange={onModeChange}
      onSourceChange={() => undefined}
      onTemplateChange={() => undefined}
    />
  );
};
