import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { MonitoringPollingSettingsCardProps } from "./helpers";

/**
 * Header block rendered at the top of one monitoring polling settings card.
 */
export const MonitoringPollingSettingsCardHeader = ({
  description,
  storageKey,
  title,
}: MonitoringPollingSettingsCardProps): ReactElement => {
  return (
    <Stack spacing={1}>
      <Typography data-testid={`monitoring-settings-title-${storageKey}`} variant="h2">
        {title}
      </Typography>
      <Typography
        color="text.secondary"
        data-testid={`monitoring-settings-summary-${storageKey}`}
        variant="body2"
      >
        {description}
      </Typography>
    </Stack>
  );
};
