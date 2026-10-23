import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";
import { useState } from "react";

/** Runs a confirmed action and returns its displayable error when it fails. */
async function runConfirmedAction(onConfirm: () => Promise<void>): Promise<string | null> {
  try {
    await onConfirm();
    return null;
  } catch (caught) {
    return caught instanceof Error ? caught.message : "The action failed.";
  }
}

/** Generic typed-confirmation dialog for consequential actions. */
export function ConfirmActionDialog({
  open,
  title,
  description,
  confirmLabel,
  onClose,
  onConfirm,
}: {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}): ReactElement {
  const [typed, setTyped] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  /** Requires the exact confirmation word before starting the action. */
  const confirm = async (): Promise<void> => {
    if (typed !== "yes" || busy) {
      return;
    }
    setBusy(true);
    setError("");
    const actionError = await runConfirmedAction(onConfirm);
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
    <Dialog open={open} onClose={close} fullWidth maxWidth="sm" data-testid="confirm-action-dialog">
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <Typography>{description}</Typography>
          <TextField
            label="Type yes to confirm"
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
          {error && <Alert severity="error">{error}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={close} disabled={busy}>
          Cancel
        </Button>
        <Button
          onClick={() => void confirm()}
          disabled={typed !== "yes" || busy}
          color="error"
          variant="contained"
          data-testid="confirm-action-submit"
        >
          {confirmLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
