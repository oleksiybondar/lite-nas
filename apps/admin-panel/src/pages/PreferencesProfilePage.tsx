import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Stub preferences page for the authenticated Unix-backed user identity.
 */
export const PreferencesProfilePage = (): ReactElement => {
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
        <Stack spacing={1}>
          <Typography data-testid="preferences-profile-details-title" variant="h2">
            Profile details
          </Typography>
          <Typography
            color="text.secondary"
            data-testid="preferences-profile-details-summary"
            variant="body2"
          >
            User profile preferences will be added after the identity contract exposes the fields we
            want to manage here.
          </Typography>
        </Stack>
      </Paper>
    </Stack>
  );
};
