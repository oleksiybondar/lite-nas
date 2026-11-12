import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Common heading for a preferences page or settings group. */
export const PreferencesPageHeader = ({
  title,
  summary,
  testIdPrefix,
  headerTestId,
}: {
  title: string;
  summary: string;
  testIdPrefix?: string;
  headerTestId?: string;
}): ReactElement => (
  <Stack
    data-testid={headerTestId ?? (testIdPrefix ? `${testIdPrefix}-header` : undefined)}
    maxWidth="820px"
    spacing={1}
  >
    <Typography
      color="primary"
      data-testid={testIdPrefix ? `${testIdPrefix}-overline` : undefined}
      variant="overline"
    >
      Preferences
    </Typography>
    <Typography data-testid={testIdPrefix ? `${testIdPrefix}-title` : undefined} variant="h1">
      {title}
    </Typography>
    <Typography
      color="text.secondary"
      data-testid={testIdPrefix ? `${testIdPrefix}-summary` : undefined}
      variant="body1"
    >
      {summary}
    </Typography>
  </Stack>
);
