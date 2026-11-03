import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { formatRouteLabel, resolveTelemetryRoute } from "@pages/SystemTelemetryPage/helpers";
import { SystemTelemetryDiskMetricState } from "@pages/SystemTelemetryPage/SystemTelemetryDiskMetricState";
import { SystemTelemetryNetworkMetricState } from "@pages/SystemTelemetryPage/SystemTelemetryNetworkMetricState";
import { SystemTelemetryPageContent } from "@pages/SystemTelemetryPage/SystemTelemetryPageContent";
import { SystemTelemetryProcessMetricState } from "@pages/SystemTelemetryPage/SystemTelemetryProcessMetricState";
import { SystemTelemetryServiceMetricState } from "@pages/SystemTelemetryPage/SystemTelemetryServiceMetricState";
import { SystemTelemetrySystemMetricState } from "@pages/SystemTelemetryPage/SystemTelemetrySystemMetricState";
import { SystemTelemetryUnsupportedState } from "@pages/SystemTelemetryPage/SystemTelemetryUnsupportedState";
import { SystemTelemetryZFSMetricState } from "@pages/SystemTelemetryPage/SystemTelemetryZFSMetricState";
import type { ReactElement } from "react";
import { Link as RouterLink, useLocation, useParams } from "react-router-dom";

const telemetryStateByType = {
  "disk-metric": <SystemTelemetryDiskMetricState />,
  "network-metric": <SystemTelemetryNetworkMetricState />,
  "process-metric": <SystemTelemetryProcessMetricState />,
  "service-metric": <SystemTelemetryServiceMetricState />,
  "system-metric": <SystemTelemetrySystemMetricState />,
} as const;

const dashboardKindsByRoute: Record<string, readonly string[]> = {
  "system-metric": ["cpu", "cpu-cores", "ram"],
  "zfs-metric": ["zfs-io", "zfs-errors"],
  "network-metric": ["network"],
  "disk-metric": ["disk"],
  "process-metric": ["processes"],
  "service-metric": ["services"],
};

const isDirectTelemetryStateType = (
  type: ReturnType<typeof resolveTelemetryRoute>["type"],
): type is keyof typeof telemetryStateByType => {
  return type in telemetryStateByType;
};

/**
 * Telemetry page host for system performance and Raspberry Pi sensor routes.
 */
export const SystemTelemetryPage = (): ReactElement => {
  const { getLabel } = useLanguage();
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
        <Stack direction="row" flexWrap="wrap" gap={1}>
          {(dashboardKindsByRoute[route.type] ?? []).map((kind) => (
            <Button
              component={RouterLink}
              data-test-class="telemetry-add-dashboard"
              data-test-name={kind}
              key={kind}
              to={`/?add=${kind}`}
            >
              {getLabel("admin.dashboard.addToDashboard")}: {getLabel(`admin.dashboard.${kind}`)}
            </Button>
          ))}
        </Stack>
      </Stack>
      <SystemTelemetryPageContent route={route}>
        {renderTelemetryState(route)}
      </SystemTelemetryPageContent>
    </Stack>
  );
};

/**
 * Resolves the telemetry state component that matches the current route contract.
 */
const renderTelemetryState = (route: ReturnType<typeof resolveTelemetryRoute>): ReactElement => {
  if (route.type === "zfs-metric") {
    return <SystemTelemetryZFSMetricState route={route} />;
  }

  if (isDirectTelemetryStateType(route.type)) {
    return telemetryStateByType[route.type];
  }

  return <SystemTelemetryUnsupportedState route={route} />;
};
