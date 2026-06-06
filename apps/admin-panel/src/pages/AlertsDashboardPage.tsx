import { buildAlertsPageSummary, buildAlertsPageTitle, formatAlertsLabel } from "@helpers/alerts";
import { useAlerts } from "@hooks/useAlerts";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Draft page for status-specific alerts dashboards.
 */
export const AlertsDashboardPage = (): ReactElement => {
  const { category, domain } = useAlerts();
  const title = buildAlertsPageTitle(category);
  const groupTitle = formatAlertsLabel(domain);

  return (
    <Stack data-testid="alerts-dashboard-page" maxWidth="860px" spacing={1}>
      <Typography color="primary" data-testid="alerts-dashboard-overline" variant="overline">
        {groupTitle}
      </Typography>
      <Typography data-testid="alerts-dashboard-title" variant="h1">
        {title}
      </Typography>
      <Typography color="text.secondary" data-testid="alerts-dashboard-summary" variant="body1">
        {buildAlertsPageSummary(domain, category)}
      </Typography>
    </Stack>
  );
};
