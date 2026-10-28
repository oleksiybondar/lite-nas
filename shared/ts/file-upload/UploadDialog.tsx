import {
  Alert,
  Box,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import type {
  ChangeEvent,
  Dispatch,
  DragEvent,
  MutableRefObject,
  ReactElement,
  RefObject,
  SetStateAction,
} from "react";
import { useEffect, useRef, useState } from "react";

import { hostPathSchema, uploadFilenameSchema } from "./schemas";
import type { FileUploadProgress } from "./upload";
import { uploadFile } from "./upload";

/** Reusable upload modal contract for a File Service destination directory. */
export interface UploadDialogProps {
  open: boolean;
  directory: string;
  onClose: () => void;
  onUploaded?: (file: File) => void | Promise<void>;
}

/** Upload progress state displayed beside the active filename. */
interface UploadProgressViewProps {
  busy: boolean;
  activeName: string;
  progress: FileUploadProgress | null;
  percent?: number;
}

/** Inputs required to upload the selected queue in order. */
interface UploadQueueOptions {
  files: File[];
  directory: string;
  privateUpload: boolean;
  signal: AbortSignal;
  setActiveName: (name: string) => void;
  setProgress: (progress: FileUploadProgress | null) => void;
  removeFile: (file: File) => void;
  onUploaded?: (file: File) => void | Promise<void>;
}

/** State and callbacks required to submit or cancel one upload queue. */
interface UploadSubmitOptions extends Omit<UploadQueueOptions, "signal" | "removeFile"> {
  busy: boolean;
  controllerRef: MutableRefObject<AbortController | null>;
  setBusy: (busy: boolean) => void;
  setError: (message: string) => void;
  setFiles: Dispatch<SetStateAction<File[]>>;
  onClose: () => void;
}

/** Props for the stateless upload dialog view. */
interface UploadDialogViewProps {
  open: boolean;
  directory: string;
  inputRef: RefObject<HTMLInputElement>;
  files: File[];
  busy: boolean;
  activeName: string;
  progress: FileUploadProgress | null;
  error: string;
  privateUpload: boolean;
  onPrivateChange: (value: boolean) => void;
  controllerRef: MutableRefObject<AbortController | null>;
  onClose: () => void;
  onUpload: () => void;
  onFilesSelected: (files: FileList | null) => void;
}

/** State setters and callback needed to dismiss an idle upload dialog. */
interface CloseUploadOptions {
  busy: boolean;
  setFiles: Dispatch<SetStateAction<File[]>>;
  setProgress: (progress: FileUploadProgress | null) => void;
  setError: (message: string) => void;
  onClose: () => void;
}

/** Adds browser files after validating their service-facing names. */
function addFiles(
  selected: FileList | null,
  setFiles: Dispatch<SetStateAction<File[]>>,
  setError: (message: string) => void,
): void {
  if (selected === null) {
    return;
  }
  const next = Array.from(selected);
  const invalid = next.find((file) => !uploadFilenameSchema.safeParse(file.name).success);
  if (invalid) {
    setError(`Invalid file name: ${invalid.name}`);
    return;
  }
  setFiles((current) => [...current, ...next]);
  setError("");
}

/** Uploads the selected files sequentially while publishing progress updates. */
async function uploadQueuedFiles(options: UploadQueueOptions): Promise<void> {
  for (const file of options.files) {
    options.setActiveName(file.name);
    options.setProgress(null);
    await uploadFile({
      file,
      directory: options.directory,
      signal: options.signal,
      onProgress: options.setProgress,
      privateUpload: options.privateUpload,
    });
    options.removeFile(file);
    await options.onUploaded?.(file);
  }
}

/** Submits queued files while preserving cancellation and error state. */
async function submitUpload(options: UploadSubmitOptions): Promise<void> {
  if (options.busy || options.files.length === 0) {
    return;
  }
  if (!hostPathSchema.safeParse(options.directory).success) {
    options.setError("The destination directory is invalid.");
    return;
  }
  const controller = new AbortController();
  options.controllerRef.current = controller;
  options.setBusy(true);
  options.setError("");
  try {
    await uploadQueuedFiles({
      ...options,
      signal: controller.signal,
      removeFile: (file) =>
        options.setFiles((current) => current.filter((queued) => queued !== file)),
    });
    options.setActiveName("");
    options.setProgress(null);
    options.onClose();
  } catch (caught) {
    options.setError(uploadErrorMessage(caught, controller.signal));
  } finally {
    options.controllerRef.current = null;
    options.setBusy(false);
  }
}

/** Returns a user-facing upload error, including explicit cancellation. */
function uploadErrorMessage(caught: unknown, signal: AbortSignal): string {
  if (signal.aborted) {
    return "Upload cancelled.";
  }
  return caught instanceof Error ? caught.message : "Upload failed.";
}

/** Clears queued upload state and closes the dialog when no transfer is active. */
function closeUpload(options: CloseUploadOptions): void {
  if (options.busy) {
    return;
  }
  options.setFiles([]);
  options.setProgress(null);
  options.setError("");
  options.onClose();
}

/** Resets the visibility choice when a dialog closes so a later upload starts shared. */
function useUploadPrivacy(open: boolean): [boolean, Dispatch<SetStateAction<boolean>>] {
  const [privateUpload, setPrivateUpload] = useState(false);
  useEffect(() => {
    if (!open) setPrivateUpload(false);
  }, [open]);
  return [privateUpload, setPrivateUpload];
}

/** Renders the hidden chooser and drag-and-drop target. */
function UploadDropZone({
  busy,
  inputRef,
  onFilesSelected,
}: {
  busy: boolean;
  inputRef: RefObject<HTMLInputElement>;
  onFilesSelected: (files: FileList | null) => void;
}): ReactElement {
  const handleDrop = (event: DragEvent<HTMLElement>): void => {
    event.preventDefault();
    if (!busy) {
      onFilesSelected(event.dataTransfer.files);
    }
  };
  const handleSelection = (event: ChangeEvent<HTMLInputElement>): void => {
    onFilesSelected(event.target.files);
    event.target.value = "";
  };
  return (
    <Box
      data-testid="file-upload-drop-zone"
      onDragOver={(event) => event.preventDefault()}
      onDrop={handleDrop}
      sx={{ border: 1, borderColor: "divider", borderRadius: 1, p: 3, textAlign: "center" }}
    >
      <Stack spacing={1} alignItems="center">
        <Typography>Drag files here or select them from your device.</Typography>
        <Button
          onClick={() => inputRef.current?.click()}
          disabled={busy}
          variant="outlined"
          data-testid="file-upload-select-button"
        >
          Select files
        </Button>
        <input ref={inputRef} type="file" multiple hidden onChange={handleSelection} />
      </Stack>
    </Box>
  );
}

/** Displays queued filenames, current upload progress, and service errors. */
function UploadDetails({
  files,
  busy,
  activeName,
  progress,
  percent,
  error,
}: {
  files: File[];
  busy: boolean;
  activeName: string;
  progress: FileUploadProgress | null;
  percent?: number;
  error: string;
}): ReactElement {
  return (
    <>
      {files.length > 0 && <UploadSelection files={files} />}
      {busy && <UploadProgressView activeName={activeName} progress={progress} percent={percent} />}
      {error && (
        <Alert severity="error" data-testid="file-upload-error">
          {error}
        </Alert>
      )}
      <Typography variant="caption" color="text.secondary">
        Existing files with the same name may be replaced when your permissions allow it.
      </Typography>
    </>
  );
}

/** Lists the names currently waiting in the upload queue. */
function UploadSelection({ files }: { files: File[] }): ReactElement {
  const countLabel = files.length === 1 ? "file" : "files";
  return (
    <Typography variant="body2" data-testid="file-upload-selection">
      {files.length} {countLabel} selected: {files.map((file) => file.name).join(", ")}
    </Typography>
  );
}

/** Shows the current file, phase, and byte progress. */
function UploadProgressView({
  activeName,
  progress,
  percent,
}: Omit<UploadProgressViewProps, "busy">): ReactElement {
  return (
    <Stack spacing={1} data-testid="file-upload-progress">
      <Typography variant="body2">
        {activeName}: {progress?.phase ?? "starting"}
        {percent === undefined ? "" : ` ${percent}%`}
      </Typography>
      <LinearProgress
        variant={percent === undefined ? "indeterminate" : "determinate"}
        value={percent}
      />
    </Stack>
  );
}

/** Renders close, cancel, and submit controls for the upload dialog. */
function UploadDialogActions({
  busy,
  canUpload,
  onClose,
  onCancel,
  onUpload,
}: {
  busy: boolean;
  canUpload: boolean;
  onClose: () => void;
  onCancel: () => void;
  onUpload: () => void;
}): ReactElement {
  return (
    <DialogActions>
      {busy ? (
        <Button onClick={onCancel} data-testid="file-upload-cancel-button">
          Cancel upload
        </Button>
      ) : (
        <Button onClick={onClose}>Close</Button>
      )}
      <Button
        onClick={onUpload}
        disabled={!canUpload || busy}
        variant="contained"
        data-testid="file-upload-submit-button"
      >
        Upload
      </Button>
    </DialogActions>
  );
}

/** Renders the complete upload modal using state supplied by its controller. */
function UploadDialogView(props: UploadDialogViewProps): ReactElement {
  const percent =
    props.progress && props.progress.total > 0
      ? Math.round((props.progress.completed / props.progress.total) * 100)
      : undefined;
  return (
    <Dialog
      open={props.open}
      onClose={props.onClose}
      fullWidth
      maxWidth="sm"
      data-testid="file-upload-dialog"
    >
      <DialogTitle>Upload files</DialogTitle>
      <UploadDialogContent {...props} percent={percent} />
      <UploadDialogActions
        busy={props.busy}
        canUpload={props.files.length > 0}
        onClose={props.onClose}
        onCancel={() => props.controllerRef.current?.abort()}
        onUpload={props.onUpload}
      />
    </Dialog>
  );
}

/** Selects or drops files and sends them through the proprietary upload flow. */
export function UploadDialog({
  open,
  directory,
  onClose,
  onUploaded,
}: UploadDialogProps): ReactElement {
  const inputRef = useRef<HTMLInputElement>(null);
  const controllerRef = useRef<AbortController | null>(null);
  const [files, setFiles] = useState<File[]>([]);
  const [activeName, setActiveName] = useState("");
  const [progress, setProgress] = useState<FileUploadProgress | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [privateUpload, setPrivateUpload] = useUploadPrivacy(open);

  /** Validates the destination and uploads queued files in order. */
  const handleUpload = (): void => {
    void submitUpload({
      busy,
      files,
      directory,
      privateUpload,
      controllerRef,
      setBusy,
      setError,
      setActiveName,
      setProgress,
      setFiles,
      onUploaded,
      onClose,
    });
  };

  return (
    <UploadDialogView
      open={open}
      directory={directory}
      inputRef={inputRef}
      files={files}
      busy={busy}
      activeName={activeName}
      progress={progress}
      error={error}
      privateUpload={privateUpload}
      onPrivateChange={setPrivateUpload}
      controllerRef={controllerRef}
      onClose={() => closeUpload({ busy, setFiles, setProgress, setError, onClose })}
      onUpload={handleUpload}
      onFilesSelected={(selected) => addFiles(selected, setFiles, setError)}
    />
  );
}

/** Renders the upload destination, picker, and current queue state. */
function UploadDialogContent({
  directory,
  inputRef,
  files,
  busy,
  activeName,
  progress,
  percent,
  error,
  privateUpload,
  onPrivateChange,
  onFilesSelected,
}: {
  directory: string;
  inputRef: RefObject<HTMLInputElement>;
  files: File[];
  busy: boolean;
  activeName: string;
  progress: FileUploadProgress | null;
  percent?: number;
  error: string;
  privateUpload: boolean;
  onPrivateChange: (value: boolean) => void;
  onFilesSelected: (files: FileList | null) => void;
}): ReactElement {
  return (
    <DialogContent>
      <Stack spacing={2}>
        <Typography variant="body2" color="text.secondary">
          Destination: {directory}
        </Typography>
        <FormControlLabel
          control={
            <Checkbox
              checked={!privateUpload}
              disabled={busy}
              onChange={(_, checked) => onPrivateChange(!checked)}
            />
          }
          label="Readable by other users"
        />
        <UploadDropZone busy={busy} inputRef={inputRef} onFilesSelected={onFilesSelected} />
        <UploadDetails
          files={files}
          busy={busy}
          activeName={activeName}
          progress={progress}
          percent={percent}
          error={error}
        />
      </Stack>
    </DialogContent>
  );
}
