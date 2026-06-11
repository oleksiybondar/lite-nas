import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { formatRouteLabel, resolveTelemetryRoute } from "@pages/SystemTelemetryPage/helpers";
import { SystemTelemetryPageContent } from "@pages/SystemTelemetryPage/SystemTelemetryPageContent";
import { SystemTelemetrySystemMetricState } from "@pages/SystemTelemetryPage/SystemTelemetrySystemMetricState";
import { SystemTelemetryUnsupportedState } from "@pages/SystemTelemetryPage/SystemTelemetryUnsupportedState";
import { SystemTelemetryZFSMetricState } from "@pages/SystemTelemetryPage/SystemTelemetryZFSMetricState";
import type { ReactElement } from "react";
import { useLocation, useParams } from "react-router-dom";

/**
 * Telemetry page host for system performance and Raspberry Pi sensor routes.
 */
export const SystemTelemetryPage = (): ReactElement => {
  const { pathname } = useLocation();
  const { category = "system" } = useParams();
  const route = resolveTelemetryRoute(pathname, category);

  return (
    <Stack data-testid="system-telemetry-page" spacing={3}>
      <Stack spacing={1}>
        <Typography color="primary" data-testid="system-telemetry-overline" variant="overline">
          {formatRouteLabel(route.group)}
        </Typography>
        <Typography data-testid="system-telemetry-title" variant="h1">
          {route.title}
        </Typography>
      </Stack>
      <SystemTelemetryPageContent routeType={route.type}>
        {renderTelemetryState(route)}
      </SystemTelemetryPageContent>
    </Stack>
  );
};

/**
 * Resolves the telemetry state component that matches the current route contract.
 */
const renderTelemetryState = (route: ReturnType<typeof resolveTelemetryRoute>): ReactElement => {
  if (route.type === "system-metric") {
    return <SystemTelemetrySystemMetricState />;
  }

  if (route.type === "zfs-metric") {
    return <SystemTelemetryZFSMetricState route={route} />;
  }

  return <SystemTelemetryUnsupportedState route={route} />;
};
