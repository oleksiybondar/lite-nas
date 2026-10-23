import Alert from "@mui/material/Alert";
import Avatar from "@mui/material/Avatar";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ChangeEvent, ReactElement } from "react";

/** Data and actions supplied by an application using the shared profile screen. */
export interface ProfilePreferencesPanelProps {
  username: string;
  avatarUrl?: string | null;
  fallbackAvatarUrl?: string | null;
  busy: boolean;
  message: string | null;
  onAvatarSelected: (file: File | undefined) => void;
}

/** Displays profile and avatar controls without owning app authentication or transport. */
export const ProfilePreferencesPanel = ({
  username,
  avatarUrl,
  fallbackAvatarUrl,
  busy,
  message,
  onAvatarSelected,
}: ProfilePreferencesPanelProps): ReactElement => {
  const handleSelection = (event: ChangeEvent<HTMLInputElement>): void => {
    const file = event.target.files?.[0];
    event.target.value = "";
    onAvatarSelected(file);
  };

  return (
    <Stack data-testid="preferences-profile-page" maxWidth="720px" spacing={3}>
      <Stack data-testid="preferences-profile-header" spacing={1}>
        <Typography color="primary" data-testid="preferences-profile-overline" variant="overline">
          Preferences
        </Typography>
        <Typography data-testid="preferences-profile-title" variant="h1">
          User profile
        </Typography>
      </Stack>
      <Paper data-testid="preferences-profile-card" sx={{ p: 3 }}>
        <Stack spacing={2} alignItems="flex-start">
          <Typography data-testid="preferences-profile-details-title" variant="h2">
            Profile details
          </Typography>
          <Avatar
            alt={username}
            src={avatarUrl ?? fallbackAvatarUrl ?? undefined}
            sx={{ height: 80, width: 80 }}
            data-testid="profile-avatar"
          >
            {username.slice(0, 2).toUpperCase()}
          </Avatar>
          <Typography>{username}</Typography>
          <Button
            component="label"
            variant="outlined"
            disabled={busy}
            data-testid="profile-avatar-upload-button"
          >
            {busy ? "Uploading…" : "Upload avatar"}
            <input
              hidden
              type="file"
              accept="image/png,image/jpeg"
              onChange={handleSelection}
              data-testid="profile-avatar-file-input"
            />
          </Button>
          <Typography color="text.secondary" variant="body2">
            PNG or JPEG, up to 5 MB.
          </Typography>
          {message ? (
            <Alert severity={message === "Avatar saved." ? "success" : "warning"}>{message}</Alert>
          ) : null}
        </Stack>
      </Paper>
    </Stack>
  );
};
