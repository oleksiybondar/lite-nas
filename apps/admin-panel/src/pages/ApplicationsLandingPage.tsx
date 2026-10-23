import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import { Button, Paper, Stack, Typography } from "@mui/material";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

/** Introduces the installed File Explorer from the admin panel. */
export const ApplicationsLandingPage = (): ReactElement => {
  return (
    <Stack spacing={3} data-testid="applications-landing-page">
      <Typography variant="h1">Applications</Typography>
      <Paper sx={{ maxWidth: 640, p: 3 }}>
        <Stack spacing={2} alignItems="flex-start">
          <FolderOpenIcon color="primary" fontSize="large" />
          <Typography variant="h2">File Explorer</Typography>
          <Typography color="text.secondary">
            Browse authorized NAS directories, open files, and upload files to directories where you
            have permission.
          </Typography>
          <Button
            component={RouterLink}
            to="/applications/file-explorer"
            variant="contained"
            data-testid="open-file-explorer-button"
          >
            Open File Explorer
          </Button>
        </Stack>
      </Paper>
    </Stack>
  );
};
