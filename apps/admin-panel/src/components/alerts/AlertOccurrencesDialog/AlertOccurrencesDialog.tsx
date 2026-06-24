import { AlertOccurrencesChart } from "@components/alerts/AlertOccurrencesDialog/AlertOccurrencesChart";
import { AlertOccurrencesDialogLayout } from "@components/alerts/AlertOccurrencesDialog/AlertOccurrencesDialogLayout";
import { AlertOccurrencesTable } from "@components/alerts/AlertOccurrencesDialog/AlertOccurrencesTable";
import { hasNumericAlertOccurrences } from "@components/alerts/AlertOccurrencesDialog/helpers";
import { useAlertOccurrences } from "@domain/alerts/hooks/useAlertOccurrences";
import type { AlertOccurrenceItemDTO } from "@dto/alerts/alerts";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import { type ReactElement, useState } from "react";
import type { AlertOccurrencesDialogProps } from "./types";

/**
 * Trigger button and modal dialog used to inspect one alert's full occurrence history.
 */
export const AlertOccurrencesDialog = ({
  domain,
  eventId,
  triggerIcon,
  triggerLabel,
}: AlertOccurrencesDialogProps): ReactElement => {
  const dialogState = useAlertOccurrencesDialogState();

  return renderAlertOccurrencesDialog({
    onClose: dialogState.closeDialog,
    onOpen: dialogState.openDialog,
    open: dialogState.open,
    triggerIcon,
    triggerLabel,
    body: dialogState.open ? (
      <AlertOccurrencesDialogContent domain={domain} eventId={eventId} />
    ) : null,
    eventId,
  });
};

type AlertOccurrencesDialogState = {
  closeDialog: () => void;
  open: boolean;
  openDialog: () => void;
};

/**
 * Owns the open/close state and query state for the occurrences dialog.
 */
const useAlertOccurrencesDialogState = (): AlertOccurrencesDialogState => {
  const [open, setOpen] = useState(false);

  return {
    closeDialog: () => {
      setOpen(false);
    },
    open,
    openDialog: () => {
      setOpen(true);
    },
  };
};

type RenderAlertOccurrencesDialogInput = {
  body: ReactElement | null;
  eventId: string;
  onClose: () => void;
  onOpen: () => void;
  open: boolean;
  triggerIcon: AlertOccurrencesDialogProps["triggerIcon"];
  triggerLabel: string;
};

/**
 * Renders the trigger button and the shared dialog shell.
 */
const renderAlertOccurrencesDialog = ({
  body,
  eventId,
  onClose,
  onOpen,
  open,
  triggerIcon,
  triggerLabel,
}: RenderAlertOccurrencesDialogInput): ReactElement => {
  return (
    <>
      <Button
        data-testid={`alert-occurrences-trigger-${eventId}`}
        endIcon={triggerIcon}
        onClick={onOpen}
        size="small"
        sx={alertOccurrencesTriggerButtonSx}
        variant="text"
      >
        {triggerLabel}
      </Button>
      <AlertOccurrencesDialogLayout
        onClose={onClose}
        open={open}
        summary={`Occurrence history for event ${eventId}.`}
        title="Alert occurrences"
      >
        {body}
      </AlertOccurrencesDialogLayout>
    </>
  );
};

type AlertOccurrencesDialogContentProps = {
  /**
   * Alert domain owning the occurrences endpoint.
   */
  domain: AlertOccurrencesDialogProps["domain"];
  /**
   * Event identifier used by the occurrences endpoint and dialog copy.
   */
  eventId: string;
};

/**
 * Loads and renders the occurrence content only after the dialog is opened.
 */
const AlertOccurrencesDialogContent = ({
  domain,
  eventId,
}: AlertOccurrencesDialogContentProps): ReactElement => {
  const occurrencesQuery = useAlertOccurrences({
    domain,
    enabled: true,
    eventId,
  });
  const occurrences = occurrencesQuery.data ?? [];

  return renderAlertOccurrencesDialogBody({
    error: occurrencesQuery.error,
    eventId,
    isError: occurrencesQuery.isError,
    isLoading: occurrencesQuery.isLoading,
    occurrences,
  });
};

type RenderAlertOccurrencesDialogBodyInput = {
  /**
   * Query error returned by the occurrences request.
   */
  error: Error | null;
  /**
   * Event identifier shown in the empty/error states.
   */
  eventId: string;
  /**
   * Whether the occurrences query failed.
   */
  isError: boolean;
  /**
   * Whether the occurrences query is still loading.
   */
  isLoading: boolean;
  /**
   * Loaded occurrence rows used to choose table or chart mode.
   */
  occurrences: AlertOccurrenceItemDTO[];
};

/**
 * Resolves the dialog body for loading, error, empty, chart, and table states.
 */
const renderAlertOccurrencesDialogBody = ({
  error,
  eventId,
  isError,
  isLoading,
  occurrences,
}: RenderAlertOccurrencesDialogBodyInput): ReactElement => {
  if (isLoading) {
    return (
      <Box
        alignItems="center"
        data-testid="alert-occurrences-loading-state"
        display="flex"
        justifyContent="center"
        minHeight={240}
      >
        <CircularProgress size={28} />
      </Box>
    );
  }

  if (isError) {
    return (
      <Alert data-testid="alert-occurrences-error-state" severity="error">
        {error?.message ?? `Failed to load occurrences for ${eventId}.`}
      </Alert>
    );
  }

  if (occurrences.length === 0) {
    return (
      <Typography
        color="text.secondary"
        data-testid="alert-occurrences-empty-state"
        variant="body2"
      >
        No occurrences were returned for {eventId}.
      </Typography>
    );
  }

  if (hasNumericAlertOccurrences(occurrences)) {
    return <AlertOccurrencesChart occurrences={occurrences} />;
  }

  return <AlertOccurrencesTable occurrences={occurrences} />;
};

const alertOccurrencesTriggerButtonSx = {
  alignItems: "center",
  color: "primary.main",
  fontWeight: 600,
  minWidth: 0,
  px: 0.5,
  py: 0,
  textAlign: "left",
  textTransform: "none",
} as const;
