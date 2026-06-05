import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useMatch, useParams } from "react-router-dom";

/**
 * Draft page for status-specific alerts dashboards.
 */
export const AlertsDashboardPage = (): ReactElement => {
  const { category = "unacknowledged" } = useParams();
  const match = useMatch("/alerts/:group/:category");
  const group = match?.params.group ?? "system";
  const title = buildAlertsDashboardTitle(category);
  const groupTitle = formatRouteLabel(group);

  return (
    <Stack data-testid="alerts-dashboard-page" maxWidth="860px" spacing={1}>
      <Typography color="primary" data-testid="alerts-dashboard-overline" variant="overline">
        {groupTitle}
      </Typography>
      <Typography data-testid="alerts-dashboard-title" variant="h1">
        {title}
      </Typography>
      <Typography color="text.secondary" data-testid="alerts-dashboard-summary" variant="body1">
        Alert dashboard panels for this route will be wired to gateway-backed alert queries.
      </Typography>
    </Stack>
  );
};

/**
 * Formats a route segment for display in placeholder alert dashboards.
 */
const formatRouteLabel = (value: string): string => {
  return value.slice(0, 1).toUpperCase() + value.slice(1).replaceAll("-", " ");
};

/**
 * Formats a status route segment into the visible alerts dashboard title.
 */
const buildAlertsDashboardTitle = (category: string): string => {
  return `${formatRouteLabel(category)} alerts`;
};
