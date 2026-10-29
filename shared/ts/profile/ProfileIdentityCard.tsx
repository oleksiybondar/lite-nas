import { useLanguage } from "@lite-nas/language";
import Alert from "@mui/material/Alert";
import Avatar from "@mui/material/Avatar";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ChangeEvent, ReactElement } from "react";

type ProfileIdentityCardProps = {
  username: string;
  avatarUrl?: string | null | undefined;
  fallbackAvatarUrl?: string | null | undefined;
  busy: boolean;
  message: string | null;
  onFileSelected: (event: ChangeEvent<HTMLInputElement>) => void;
};

/** Displays the current account identity and avatar upload control. */
export const ProfileIdentityCard = ({
  username,
  avatarUrl,
  fallbackAvatarUrl,
  busy,
  message,
  onFileSelected,
}: ProfileIdentityCardProps): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack data-testid="preferences-profile-page" maxWidth="720px" spacing={3}>
      <Stack data-testid="preferences-profile-header" spacing={1}>
        <Typography color="primary" data-testid="preferences-profile-overline" variant="overline">
          {getLabel("profile.preferences")}
        </Typography>
        <Typography data-testid="preferences-profile-title" variant="h1">
          {getLabel("profile.userProfile")}
        </Typography>
      </Stack>
      <Paper data-testid="preferences-profile-card" sx={{ p: 3 }}>
        <Stack spacing={2} alignItems="flex-start">
          <Typography data-testid="preferences-profile-details-title" variant="h2">
            {getLabel("profile.profileDetails")}
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
            {busy ? getLabel("profile.uploading") : getLabel("profile.uploadAvatar")}
            <input
              hidden
              type="file"
              accept="image/png,image/jpeg"
              onChange={onFileSelected}
              data-testid="profile-avatar-file-input"
            />
          </Button>
          <Typography color="text.secondary" variant="body2">
            {getLabel("profile.avatarHint")}
          </Typography>
          {message ? (
            <Alert severity={message === getLabel("profile.avatarSaved") ? "success" : "warning"}>
              {message}
            </Alert>
          ) : null}
        </Stack>
      </Paper>
    </Stack>
  );
};
