import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import { useTheme } from "@mui/material/styles";
import type { ReactElement } from "react";
import {
  buildServiceMetadata,
  resolvePrimaryAction,
  resolveServiceLedColor,
  resolveServiceStatusTone,
  resolveServiceSubstatusTone,
  resolveStartupAction,
} from "./helpers";
import { ServiceMetricCardActions } from "./ServiceMetricCardActions";
import { ServiceMetricCardHeader } from "./ServiceMetricCardHeader";
import { ServiceMetricCardMetadata } from "./ServiceMetricCardMetadata";
import { ServiceMetricCardStatus } from "./ServiceMetricCardStatus";
import type { ServiceMetricCardProps } from "./types";

/**
 * Card-style snapshot view for one systemd-managed service unit.
 */
export const ServiceMetricCard = ({
  service,
  serviceActions,
}: ServiceMetricCardProps): ReactElement => {
  const theme = useTheme();
  const ledColor = resolveServiceLedColor(theme, service.active_state, service.result);
  const metadata = buildServiceMetadata(service);
  const primaryAction = resolvePrimaryAction(service, serviceActions);
  const startupAction = resolveStartupAction(service, serviceActions);

  return (
    <Paper
      data-test-class="service-metric-card"
      data-test-name={service.name}
      sx={{ p: 2 }}
      variant="outlined"
    >
      <Stack spacing={1.5}>
        <Stack
          alignItems={{ md: "center", xs: "flex-start" }}
          direction={{ md: "row", xs: "column" }}
          justifyContent="space-between"
          spacing={1.5}
        >
          <ServiceMetricCardHeader ledColor={ledColor} service={service} />
          <ServiceMetricCardActions
            onRestart={() => {
              void serviceActions.restart(service.name);
            }}
            primaryAction={primaryAction}
            serviceName={service.name}
            startupAction={startupAction}
          />
        </Stack>
        <ServiceMetricCardStatus
          serviceName={service.name}
          statusTone={resolveServiceStatusTone(service.active_state)}
          substatusTone={resolveServiceSubstatusTone(service.sub_state, service.result)}
          theme={theme}
        />
        <ServiceMetricCardMetadata metadata={metadata} serviceName={service.name} />
      </Stack>
    </Paper>
  );
};
