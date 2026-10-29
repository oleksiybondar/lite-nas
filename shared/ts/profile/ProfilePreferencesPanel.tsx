import type { ChangeEvent, ReactElement } from "react";
import { useState } from "react";
import { avatarFileSchema } from "./avatar-validation";
import { ProfileAvatarCropDialog } from "./ProfileAvatarCropDialog";
import { ProfileIdentityCard } from "./ProfileIdentityCard";

/** Data and actions supplied by an application using the shared profile screen. */
export interface ProfilePreferencesPanelProps {
  username: string;
  avatarUrl?: string | null;
  fallbackAvatarUrl?: string | null;
  busy: boolean;
  message: string | null;
  onAvatarSelected: (file: File | undefined) => void;
}

/** Displays profile identity and avatar controls without owning application transport. */
export const ProfilePreferencesPanel = ({
  username,
  avatarUrl,
  fallbackAvatarUrl,
  busy,
  message,
  onAvatarSelected,
}: ProfilePreferencesPanelProps): ReactElement => {
  const [source, setSource] = useState<string | null>(null);

  /** Opens the crop dialog for a valid image or reports an invalid selection. */
  const handleSelection = (event: ChangeEvent<HTMLInputElement>): void => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    if (!avatarFileSchema.safeParse(file).success) {
      onAvatarSelected(file);
      return;
    }
    setSource(URL.createObjectURL(file));
  };

  /** Closes the crop dialog and releases its temporary browser URL. */
  const closeCrop = (): void => {
    if (source) URL.revokeObjectURL(source);
    setSource(null);
  };

  return (
    <>
      <ProfileIdentityCard
        username={username}
        avatarUrl={avatarUrl}
        fallbackAvatarUrl={fallbackAvatarUrl}
        busy={busy}
        message={message}
        onFileSelected={handleSelection}
      />
      <ProfileAvatarCropDialog
        key={source}
        source={source}
        onClose={closeCrop}
        onSave={(file) => {
          onAvatarSelected(file);
          closeCrop();
        }}
      />
    </>
  );
};
