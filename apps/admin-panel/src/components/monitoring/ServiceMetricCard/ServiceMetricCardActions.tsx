import { useLanguage } from "@lite-nas/language";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import RestartAltIcon from "@mui/icons-material/RestartAlt";
import StopIcon from "@mui/icons-material/Stop";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Switch from "@mui/material/Switch";
import type { ReactElement } from "react";
import type { ServiceMetricPrimaryAction, ServiceMetricStartupAction } from "./types";

type ServiceMetricCardActionsProps = {
  /**
   * Current primary runtime action for the service.
   */
  primaryAction: ServiceMetricPrimaryAction;
  /**
   * Restart handler for the service.
   */
  onRestart: () => void;
  /**
   * Service name used by accessible labels.
   */
  serviceName: string;
  /**
   * Startup toggle contract for the service.
   */
  startupAction: ServiceMetricStartupAction;
  busy: boolean;
};

/**
 * Action row containing runtime commands and startup toggle controls.
 */
export const ServiceMetricCardActions = ({
  onRestart,
  primaryAction,
  serviceName,
  startupAction,
  busy,
}: ServiceMetricCardActionsProps): ReactElement => {
  const { getLabel } = useLanguage();
  const ActionIcon = primaryAction.label === "Stop" ? StopIcon : PlayArrowIcon;

  return (
    <Stack direction="row" flexWrap="wrap" gap={1} useFlexGap>
      <Button
        color={primaryAction.color}
        disabled={busy}
        data-testid={`service-metric-toggle-state-${serviceName}`}
        onClick={primaryAction.onClick}
        size="small"
        startIcon={<ActionIcon fontSize="small" />}
        variant="contained"
      >
        {getLabel(`admin.telemetry.service.${primaryAction.label.toLowerCase()}`)}
      </Button>
      <Button
        color="warning"
        disabled={busy}
        data-testid={`service-metric-restart-${serviceName}`}
        onClick={onRestart}
        size="small"
        startIcon={<RestartAltIcon fontSize="small" />}
        variant="contained"
      >
        {getLabel("admin.telemetry.service.restart")}
      </Button>
      <ServiceMetricStartupToggle
        busy={busy}
        serviceName={serviceName}
        startupAction={startupAction}
      />
    </Stack>
  );
};

type ServiceMetricStartupToggleProps = {
  busy: boolean;
  serviceName: string;
  startupAction: ServiceMetricStartupAction;
};

/**
 * Startup toggle with a neutral switch and color-coded action label.
 */
const ServiceMetricStartupToggle = ({
  busy,
  serviceName,
  startupAction,
}: ServiceMetricStartupToggleProps): ReactElement => {
  return (
    <Stack
      alignItems="center"
      data-testid={`service-metric-toggle-startup-${serviceName}`}
      direction="row"
      spacing={1}
    >
      <Switch
        checked={startupAction.isEnabled}
        disabled={busy}
        inputProps={{
          "aria-label": `${startupAction.label} startup for ${serviceName}`,
        }}
        onChange={startupAction.onToggle}
        size="medium"
        sx={{
          ml: -0.5,
          transform: "scale(1.2)",
        }}
      />
      <Button
        color={startupAction.isEnabled ? "error" : "info"}
        disableRipple
        sx={{
          color: startupAction.labelColor,
          minWidth: 0,
          px: 0,
          "&:hover": {
            backgroundColor: "transparent",
          },
        }}
        variant="text"
      >
        {startupAction.label}
      </Button>
    </Stack>
  );
};
