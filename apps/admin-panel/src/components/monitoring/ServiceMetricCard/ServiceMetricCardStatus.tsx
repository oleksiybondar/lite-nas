import Chip from "@mui/material/Chip";
import Stack from "@mui/material/Stack";
import type { Theme } from "@mui/material/styles";
import type { ReactElement } from "react";
import { buildServiceToneSx } from "./helpers";
import type { ServiceMetricCardProps, ServiceMetricTone } from "./types";

type ServiceMetricCardStatusProps = {
  /**
   * Service name used in stable selectors.
   */
  serviceName: ServiceMetricCardProps["service"]["name"];
  /**
   * Tone displayed for the active state.
   */
  statusTone: ServiceMetricTone;
  /**
   * Tone displayed for the substate.
   */
  substatusTone: ServiceMetricTone;
  /**
   * Current MUI theme used to derive chip colors.
   */
  theme: Theme;
};

/**
 * Status row containing the colored service state and substate chips.
 */
export const ServiceMetricCardStatus = ({
  serviceName,
  statusTone,
  substatusTone,
  theme,
}: ServiceMetricCardStatusProps): ReactElement => {
  return (
    <Stack direction="row" flexWrap="wrap" gap={1} useFlexGap>
      <Chip
        data-testid={`service-metric-status-${serviceName}`}
        label={`Status: ${statusTone.label}`}
        size="small"
        sx={buildServiceToneSx(theme, statusTone.color)}
      />
      <Chip
        data-testid={`service-metric-substatus-${serviceName}`}
        label={`Substatus: ${substatusTone.label}`}
        size="small"
        sx={buildServiceToneSx(theme, substatusTone.color)}
      />
    </Stack>
  );
};
