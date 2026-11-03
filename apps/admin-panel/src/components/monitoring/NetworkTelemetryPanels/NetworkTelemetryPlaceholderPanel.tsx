import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

type NetworkTelemetryPlaceholderPanelProps = {
  /**
   * Stable section title shown at the top of one placeholder panel.
   */
  title: string;
};

/**
 * Renders one full-width placeholder network telemetry panel until section-specific fields are defined.
 */
export const NetworkTelemetryPlaceholderPanel = ({
  title,
}: NetworkTelemetryPlaceholderPanelProps): ReactElement => {
  return (
    <Paper
      data-test-class="network-telemetry-panel"
      data-test-name={title}
      sx={{ p: 2, width: "100%" }}
    >
      <Stack spacing={1.5}>
        <Typography
          data-test-class="network-telemetry-panel-title"
          data-test-name={title}
          variant="h2"
        >
          {title}
        </Typography>
        <Typography color="text.secondary" variant="body2">
          Section fields will be added after the panel contract is defined.
        </Typography>
      </Stack>
    </Paper>
  );
};
