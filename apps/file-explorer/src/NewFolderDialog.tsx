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

import { newFolderNameSchema } from "./schemas";

/** Creates one validated folder name and returns any displayable service error. */
async function createFolder(
  name: string,
  onCreate: (name: string) => Promise<void>,
): Promise<string | null> {
  const parsed = newFolderNameSchema.safeParse(name);
  if (!parsed.success) {
    return "Enter a folder name without path separators.";
  }
  try {
    await onCreate(parsed.data);
    return null;
  } catch (caught) {
    return caught instanceof Error ? caught.message : "Could not create the folder.";
  }
}

/** Requests a validated name for one directory in the active panel. */
export function NewFolderDialog({
  open,
  directory,
  onClose,
  onCreate,
}: {
  open: boolean;
  directory: string;
  onClose: () => void;
  onCreate: (name: string) => Promise<void>;
}): ReactElement {
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  /** Validates the input before invoking the File Service action. */
  const submit = async (): Promise<void> => {
    setBusy(true);
    setError("");
    const creationError = await createFolder(name, onCreate);
    setBusy(false);
    if (creationError !== null) {
      setError(creationError);
      return;
    }
    setName("");
    onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={busy ? undefined : onClose}
      fullWidth
      maxWidth="xs"
      data-testid="file-explorer-new-folder-dialog"
    >
      <DialogTitle>New folder</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <Typography variant="body2" color="text.secondary">
            Create in {directory}
          </Typography>
          <TextField
            autoFocus
            label="Folder name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                void submit();
              }
            }}
            data-testid="file-explorer-new-folder-name"
          />
          {error && <Alert severity="error">{error}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={busy}>
          Cancel
        </Button>
        <Button
          onClick={() => void submit()}
          disabled={busy || name.length === 0}
          variant="contained"
          data-testid="file-explorer-new-folder-submit"
        >
          Create
        </Button>
      </DialogActions>
    </Dialog>
  );
}
