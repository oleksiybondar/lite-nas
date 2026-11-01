import { useLanguage } from "@lite-nas/language";
import { Alert, Button, Stack, TextField, Typography } from "@mui/material";
import type { ReactElement } from "react";
import { useState } from "react";
import { StandardDialog } from "./StandardDialog";

/** Runs a confirmed action and returns its displayable error when it fails. */
async function runConfirmedAction(
  onConfirm: () => Promise<void>,
  getLabel: (key: string) => string,
): Promise<string | null> {
  try {
    await onConfirm();
    return null;
  } catch (caught) {
    return caught instanceof Error ? caught.message : getLabel("dialogs.actionFailed");
  }
}

/** Generic typed-confirmation dialog for consequential actions. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The dialog keeps confirmation state and action feedback together.
export function ConfirmActionDialog({
  open,
  title,
  description,
  confirmLabel,
  onClose,
  onConfirm,
  requireYes = true,
  confirmColor = "error",
}: {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  onClose: () => void;
  onConfirm: () => Promise<void>;
  /** Requires typing `yes` for destructive actions; otherwise a button confirms. */
  requireYes?: boolean;
  /** Semantic color for the confirmation button. */
  confirmColor?: "error" | "primary" | "warning";
}): ReactElement {
  const { getLabel } = useLanguage();
  const [typed, setTyped] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  /** Requires the exact confirmation word before starting the action. */
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The dialog keeps confirmation state and action feedback together.
  const confirm = async (): Promise<void> => {
    if ((requireYes && typed !== "yes") || busy) {
      return;
    }
    setBusy(true);
    setError("");
    const actionError = await runConfirmedAction(onConfirm, getLabel);
    setBusy(false);
    if (actionError !== null) {
      setError(actionError);
      return;
    }
    setTyped("");
    onClose();
  };

  /** Clears confirmation state when the dialog is dismissed. */
  const close = (): void => {
    if (busy) {
      return;
    }
    setTyped("");
    setError("");
    onClose();
  };

  return (
    <StandardDialog
      open={open}
      onClose={close}
      closeLabel={getLabel("dialogs.close")}
      closeDisabled={busy}
      title={title}
      fullWidth
      maxWidth="sm"
      data-testid="confirm-action-dialog"
      actions={
        <>
          <Button onClick={close} disabled={busy}>
            {getLabel("common.cancel")}
          </Button>
          <Button
            onClick={() => void confirm()}
            disabled={(requireYes && typed !== "yes") || busy}
            color={confirmColor}
            variant="contained"
            data-testid="confirm-action-submit"
          >
            {confirmLabel}
          </Button>
        </>
      }
    >
      <Stack spacing={2} sx={{ pt: 1 }}>
        <Typography>{description}</Typography>
        {requireYes && (
          <TextField
            label={getLabel("dialogs.confirmYes")}
            value={typed}
            onChange={(event) => setTyped(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                void confirm();
              }
            }}
            autoComplete="off"
            data-testid="confirm-action-input"
          />
        )}
        {error && <Alert severity="error">{error}</Alert>}
      </Stack>
    </StandardDialog>
  );
}
