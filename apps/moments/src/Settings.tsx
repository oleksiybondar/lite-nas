// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: tab content and its feedback stay together in the settings screen.

import { useLanguage } from "@lite-nas/language";
import { ProfilePreferencesPanel, writeProfileAvatar } from "@lite-nas/profile";
import { PreferencesPageHeader, PreferencesTabs } from "@lite-nas/settings";
import { type ThemeMode, ThemePreferencesPanel } from "@lite-nas/theme";
import CollectionsIcon from "@mui/icons-material/Collections";
import GroupIcon from "@mui/icons-material/Group";
import ManageAccountsIcon from "@mui/icons-material/ManageAccounts";
import PaletteIcon from "@mui/icons-material/Palette";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Tab from "@mui/material/Tab";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useState } from "react";
import { z } from "zod";
import { request } from "./api";
import { FriendsSettings } from "./FriendsSettings";
import { imageConcurrencySchema } from "./schemas";

/** Settings sections exposed in the Moments preferences sidebar. */
const settingsViewSchema = z.enum(["profile", "appearance", "gallery", "friends"]);
export type SettingsView = z.infer<typeof settingsViewSchema>;

/** Shows profile, appearance, and gallery preferences in one tabbed settings area. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: sidebar tabs keep the three related preference sections in one screen.
export const Settings = ({
  view,
  onViewChange,
  username,
  avatarUrl,
  fallbackAvatarUrl,
  mode,
  message,
  onModeChange,
  onSaveMode,
  maxConcurrent,
  galleryMessage,
  onMaxConcurrentChange,
}: {
  view: SettingsView;
  onViewChange: (view: SettingsView) => void;
  username: string;
  avatarUrl: string | null;
  fallbackAvatarUrl?: string | null;
  mode: ThemeMode;
  message: string | null;
  onModeChange: (value: ThemeMode) => void;
  onSaveMode: (value: ThemeMode) => void;
  maxConcurrent: number;
  galleryMessage: string | null;
  onMaxConcurrentChange: (value: number) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
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
      setAvatarMessage(getLabel("profile.avatarSaved"));
    } catch {
      setAvatarMessage(getLabel("profile.avatarSaveFailed"));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Stack spacing={3} data-testid="moments-settings" sx={{ width: "100%" }}>
      <PreferencesPageHeader
        title={getLabel("moments.settings.title")}
        summary={getLabel("moments.settings.summary")}
      />
      <Stack direction={{ xs: "column", sm: "row" }} spacing={3} alignItems="stretch">
        <PreferencesTabs
          label={getLabel("moments.settings.tabs")}
          layout="sidebar"
          value={view}
          onChange={(next) => {
            const parsed = settingsViewSchema.safeParse(next);
            if (parsed.success) onViewChange(parsed.data);
          }}
        >
          <Tab
            icon={<ManageAccountsIcon />}
            iconPosition="start"
            label={getLabel("moments.settings.profile")}
            value="profile"
          />
          <Tab
            icon={<PaletteIcon />}
            iconPosition="start"
            label={getLabel("moments.settings.appearance")}
            value="appearance"
          />
          <Tab
            icon={<CollectionsIcon />}
            iconPosition="start"
            label={getLabel("moments.settings.gallery")}
            value="gallery"
          />
          <Tab
            icon={<GroupIcon />}
            iconPosition="start"
            label={getLabel("moments.settings.friends")}
            value="friends"
          />
        </PreferencesTabs>
        <Box sx={{ flex: 1, minWidth: 0, maxWidth: 900 }}>
          {view === "profile" && (
            <ProfilePreferencesPanel
              username={username}
              avatarUrl={avatarUrl}
              fallbackAvatarUrl={fallbackAvatarUrl}
              busy={busy}
              message={avatarMessage}
              onAvatarSelected={(file) => void upload(file)}
            />
          )}
          {view === "appearance" && (
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
          )}
          {view === "gallery" && (
            <Stack spacing={2}>
              <Typography variant="h5">{getLabel("moments.settings.gallery")}</Typography>
              <Typography color="text.secondary" variant="body2">
                {getLabel("moments.settings.galleryHint")}
              </Typography>
              <TextField
                label={getLabel("moments.settings.maxConcurrent")}
                type="number"
                value={maxConcurrent}
                inputProps={{ min: 1, max: 20 }}
                onChange={(event) => {
                  const result = imageConcurrencySchema.safeParse(Number(event.target.value));
                  if (result.success) onMaxConcurrentChange(result.data);
                }}
                sx={{ maxWidth: 320 }}
              />
              {galleryMessage && <Typography color="error">{galleryMessage}</Typography>}
            </Stack>
          )}
          {view === "friends" && <FriendsSettings />}
        </Box>
      </Stack>
    </Stack>
  );
};
