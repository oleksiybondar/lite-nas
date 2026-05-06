import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useParams } from "react-router-dom";

/**
 * Draft page for system performance and Raspberry Pi sensor routes.
 */
export const SystemTelemetryPage = (): ReactElement => {
  const { category = "system", group = "performance" } = useParams();
  const title = formatRouteLabel(category);
  const groupTitle = formatRouteLabel(group);

  return (
    <Stack maxWidth="860px" spacing={1}>
      <Typography color="primary" variant="overline">
        {groupTitle}
      </Typography>
      <Typography variant="h1">{title}</Typography>
      <Typography color="text.secondary" variant="body1">
        Telemetry panels for this route will be wired to gateway-backed metrics slices.
      </Typography>
    </Stack>
  );
};

/**
 * Formats a route segment for display in placeholder telemetry pages.
 */
const formatRouteLabel = (value: string): string => {
  return value.slice(0, 1).toUpperCase() + value.slice(1).replaceAll("-", " ");
};
