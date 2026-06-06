import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Temporary alerts table placeholder until the real list component is implemented.
 */
export const AlertsTablePlaceholder = (): ReactElement => {
  return (
    <Paper data-testid="alerts-table-placeholder" sx={{ p: 3 }}>
      <Stack spacing={1}>
        <Typography variant="h2">Alerts table</Typography>
        <Typography color="text.secondary" variant="body2">
          The gateway-backed alerts table will be rendered here.
        </Typography>
      </Stack>
    </Paper>
  );
};
