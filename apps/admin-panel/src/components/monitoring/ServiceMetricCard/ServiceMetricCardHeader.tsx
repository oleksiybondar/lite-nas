import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { ServiceMetricCardProps } from "./types";

type ServiceMetricCardHeaderProps = {
  /**
   * Status LED color derived from the service state.
   */
  ledColor: string;
  /**
   * Service currently being displayed.
   */
  service: ServiceMetricCardProps["service"];
};

/**
 * Header row containing the service state LED, name, and optional description.
 */
export const ServiceMetricCardHeader = ({
  ledColor,
  service,
}: ServiceMetricCardHeaderProps): ReactElement => {
  return (
    <Stack alignItems="center" direction="row" spacing={1.25}>
      <Paper
        aria-hidden="true"
        component="span"
        elevation={0}
        sx={{
          backgroundColor: ledColor,
          borderRadius: "50%",
          flexShrink: 0,
          height: 12,
          width: 12,
        }}
      />
      <Stack minWidth={0} spacing={0.25}>
        <Typography data-testid={`service-metric-card-title-${service.name}`} variant="h6">
          {service.name}
        </Typography>
        {service.description ? (
          <Typography color="text.secondary" variant="body2">
            {service.description}
          </Typography>
        ) : null}
      </Stack>
    </Stack>
  );
};
