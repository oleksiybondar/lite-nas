import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Stub preferences page for the authenticated Unix-backed user identity.
 */
export const PreferencesProfilePage = (): ReactElement => {
  return (
    <Stack maxWidth="720px" spacing={3}>
      <Stack spacing={1}>
        <Typography color="primary" variant="overline">
          Preferences
        </Typography>
        <Typography variant="h1">User profile</Typography>
      </Stack>
      <Paper sx={{ p: 3 }}>
        <Stack spacing={1}>
          <Typography variant="h2">Profile details</Typography>
          <Typography color="text.secondary" variant="body2">
            User profile preferences will be added after the identity contract exposes the fields we
            want to manage here.
          </Typography>
        </Stack>
      </Paper>
    </Stack>
  );
};
