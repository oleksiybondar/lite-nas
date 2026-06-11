import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { SupportedTelemetryRoute } from "@pages/SystemTelemetryPage/helpers";
import type { ReactElement } from "react";

type SystemTelemetryUnsupportedStateProps = {
  /**
   * Current route metadata resolved from the browser URL.
   */
  route: SupportedTelemetryRoute;
};

/**
 * Placeholder state rendered for telemetry routes that do not yet have a backend contract.
 */
export const SystemTelemetryUnsupportedState = ({
  route,
}: SystemTelemetryUnsupportedStateProps): ReactElement => {
  return (
    <Paper data-testid="system-telemetry-placeholder-card" sx={{ p: 3 }}>
      <Stack spacing={1.5}>
        <Typography data-testid="system-telemetry-placeholder-title" variant="h2">
          Route pending backend support
        </Typography>
        <Typography
          color="text.secondary"
          data-testid="system-telemetry-placeholder-summary"
          variant="body2"
        >
          {route.summary}
        </Typography>
      </Stack>
    </Paper>
  );
};
