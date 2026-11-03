import type { ApiContextValue } from "@dto/api/api";
import { useApi } from "@hooks/useApi";
import { useAuth } from "@hooks/useAuth";
import { useProfileAvatar } from "@hooks/useProfileAvatar";
import { useLanguage } from "@lite-nas/language";
import { avatarFileSchema, ProfilePreferencesPanel, writeProfileAvatar } from "@lite-nas/profile";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";
import { useState } from "react";

/** Connects the shared profile screen to the authenticated admin API. */
export const PreferencesProfilePage = (): ReactElement => {
  const { getLabel } = useLanguage();
  const { me } = useAuth();
  const { put } = useApi();
  const avatar = useProfileAvatar();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  /** Validates and stores a selected image through the fixed profile upload route. */
  const upload = async (file: File | undefined): Promise<void> => {
    const parsed = avatarFileSchema.safeParse(file);
    if (!parsed.success) {
      setMessage(getLabel("profile.invalidAvatar"));
      return;
    }
    await saveAvatar(parsed.data);
  };

  /** Uploads a validated file and reports the result to the shared profile screen. */
  const saveAvatar = async (file: File): Promise<void> => {
    setBusy(true);
    try {
      await writeProfileAvatar((url, image) => sendAvatar(put, url, image), file);
      window.dispatchEvent(new Event("lite-nas-avatar-updated"));
      setMessage(getLabel("profile.avatarSaved"));
    } catch (error) {
      setMessage(avatarSaveErrorMessage(error, getLabel));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Stack spacing={3}>
      <ProfilePreferencesPanel
        username={me?.user.login ?? ""}
        avatarUrl={avatar}
        fallbackAvatarUrl={me?.user.avatar_url ?? null}
        busy={busy}
        message={message}
        onAvatarSelected={(file) => {
          void upload(file);
        }}
      />
    </Stack>
  );
};

/** Adds an optional transport detail to the localized avatar upload error. */
const avatarSaveErrorMessage = (error: unknown, getLabel: (key: string) => string): string => {
  const detail = error instanceof Error ? ` ${error.message}` : "";
  return `${getLabel("profile.avatarSaveFailed")}${detail}`;
};

/** Sends a validated avatar using the admin API pipeline and its file MIME type. */
const sendAvatar = (put: ApiContextValue["put"], url: string, image: File): Promise<Response> => {
  const action = put(url, image);
  if (image.type) action.header("Content-Type", image.type);
  return action.execute();
};
