import { useLanguage } from "@lite-nas/language";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";

import { newFolderNameSchema } from "./schemas";

/** Creates one validated folder name and returns any displayable service error. */
async function createFolder(
  name: string,
  privateDirectory: boolean,
  onCreate: (name: string, privateDirectory: boolean) => Promise<void>,
): Promise<string | null> {
  const parsed = newFolderNameSchema.safeParse(name);
  if (!parsed.success) {
    return "Enter a folder name without path separators.";
  }
  try {
    await onCreate(parsed.data, privateDirectory);
    return null;
  } catch (caught) {
    return caught instanceof Error ? caught.message : "Could not create the folder.";
  }
}

/** Lets a user decide whether others can access a newly created folder. */
function FolderVisibilityControl({
  privateDirectory,
  busy,
  onChange,
}: {
  privateDirectory: boolean;
  busy: boolean;
  onChange: (value: boolean) => void;
}): ReactElement {
  const { getLabel } = useLanguage();
  return (
    <FormControlLabel
      control={
        <Checkbox
          checked={!privateDirectory}
          disabled={busy}
          onChange={(_, checked) => onChange(!checked)}
        />
      }
      label={getLabel("explorer.folderVisibility")}
    />
  );
}

/** Edits the single child name and submits it when Enter is pressed. */
function FolderNameInput({
  name,
  onChange,
  onSubmit,
}: {
  name: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
}): ReactElement {
  const { getLabel } = useLanguage();
  return (
    <TextField
      autoFocus
      label={getLabel("explorer.folderName")}
      value={name}
      onChange={(event) => onChange(event.target.value)}
      onKeyDown={(event) => {
        if (event.key === "Enter") onSubmit();
      }}
      data-testid="file-explorer-new-folder-name"
    />
  );
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
  onCreate: (name: string, privateDirectory: boolean) => Promise<void>;
}): ReactElement {
  const { getLabel } = useLanguage();
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [privateDirectory, setPrivateDirectory] = useState(false);
  useEffect(() => {
    if (!open) setPrivateDirectory(false);
  }, [open]);

  /** Validates the input before invoking the File Service action. */
  const submit = async (): Promise<void> => {
    setBusy(true);
    setError("");
    const creationError = await createFolder(name, privateDirectory, onCreate);
    setBusy(false);
    if (creationError !== null) {
      setError(creationError);
      return;
    }
    setName("");
    setPrivateDirectory(false);
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
      <DialogTitle>{getLabel("common.newFolder")}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <Typography variant="body2" color="text.secondary">
            {getLabel("explorer.createIn")} {directory}
          </Typography>
          <FolderNameInput name={name} onChange={setName} onSubmit={() => void submit()} />
          <FolderVisibilityControl
            privateDirectory={privateDirectory}
            busy={busy}
            onChange={setPrivateDirectory}
          />
          {error && <Alert severity="error">{error}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={busy}>
          {getLabel("common.cancel")}
        </Button>
        <Button
          onClick={() => void submit()}
          disabled={busy || name.length === 0}
          variant="contained"
          data-testid="file-explorer-new-folder-submit"
        >
          {getLabel("explorer.create")}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
