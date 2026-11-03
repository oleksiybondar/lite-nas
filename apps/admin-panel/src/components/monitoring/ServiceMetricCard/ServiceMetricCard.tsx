import { ConfirmActionDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import Alert from "@mui/material/Alert";
import CircularProgress from "@mui/material/CircularProgress";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import { useTheme } from "@mui/material/styles";
import type { ReactElement } from "react";
import { useState } from "react";
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

const confirmationColors = { start: "primary", stop: "error", restart: "warning" } as const;

/**
 * Card-style snapshot view for one systemd-managed service unit.
 */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The card keeps service action state and feedback with its controls.
export const ServiceMetricCard = ({
  service,
  serviceActions,
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The card keeps service action state and feedback with its controls.
}: ServiceMetricCardProps): ReactElement => {
  const { getLabel } = useLanguage();
  const theme = useTheme();
  const [actionPending, setActionPending] = useState(false);
  const [actionResult, setActionResult] = useState<"success" | "error" | null>(null);
  const [confirmation, setConfirmation] = useState<"start" | "stop" | "restart" | null>(null);
  const ledColor = resolveServiceLedColor(theme, service.active_state, service.result);
  const metadata = buildServiceMetadata(service);
  const primaryAction = resolvePrimaryAction(service, serviceActions);
  const startupAction = resolveStartupAction(service, serviceActions);
  const runAction = async (
    action: () => void | Promise<void>,
    propagateError = false,
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The card keeps service action state and feedback with its controls.
  ): Promise<void> => {
    setActionPending(true);
    setActionResult(null);
    try {
      await action();
      setActionResult("success");
    } catch (caught) {
      setActionResult("error");
      if (propagateError) throw caught;
    } finally {
      setActionPending(false);
    }
  };

  return (
    <Paper
      data-test-class="service-metric-card"
      data-test-name={service.name}
      sx={{ p: 2, position: "relative" }}
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
            busy={actionPending}
            onRestart={() => setConfirmation("restart")}
            primaryAction={{
              ...primaryAction,
              onClick: () => setConfirmation(primaryAction.label === "Stop" ? "stop" : "start"),
            }}
            serviceName={service.name}
            startupAction={{
              ...startupAction,
              onToggle: () => void runAction(startupAction.onToggle),
            }}
          />
        </Stack>
        <ServiceMetricCardStatus
          serviceName={service.name}
          statusTone={resolveServiceStatusTone(service.active_state)}
          substatusTone={resolveServiceSubstatusTone(service.sub_state, service.result)}
          theme={theme}
        />
        <ServiceMetricCardMetadata metadata={metadata} serviceName={service.name} />
        {actionResult && (
          <Alert
            severity={actionResult}
            data-testid={`service-metric-action-${actionResult}-${service.name}`}
          >
            {getLabel(
              actionResult === "success"
                ? "admin.telemetry.service.actionSucceeded"
                : "admin.telemetry.service.actionFailed",
            )}
          </Alert>
        )}
      </Stack>
      {actionPending && (
        <Stack
          alignItems="center"
          aria-live="polite"
          data-testid={`service-metric-action-pending-${service.name}`}
          justifyContent="center"
          sx={{
            position: "absolute",
            inset: 0,
            bgcolor: "background.paper",
            opacity: 0.88,
            zIndex: 1,
          }}
        >
          <CircularProgress aria-label={getLabel("admin.telemetry.service.actionInProgress")} />
          <span>{getLabel("admin.telemetry.service.actionInProgress")}</span>
        </Stack>
      )}
      <ConfirmActionDialog
        open={confirmation !== null}
        title={getLabel("dialogs.areYouSure")}
        description={getLabel("admin.telemetry.service.confirmAction")
          .replace(
            "{action}",
            confirmation ? getLabel(`admin.telemetry.service.${confirmation}`) : "",
          )
          .replace("{service}", service.name)}
        confirmLabel={confirmation ? getLabel(`admin.telemetry.service.${confirmation}`) : ""}
        confirmColor={confirmation ? confirmationColors[confirmation] : "primary"}
        requireYes={confirmation === "stop"}
        onClose={() => setConfirmation(null)}
        onConfirm={async () => {
          if (confirmation === "restart")
            await runAction(() => serviceActions.restart(service.name), true);
          else if (confirmation)
            await runAction(() => serviceActions.toggleState(service.name, confirmation), true);
        }}
      />
    </Paper>
  );
};
