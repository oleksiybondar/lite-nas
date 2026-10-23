import { UploadDialog } from "@lite-nas/file-upload";
import SplitscreenIcon from "@mui/icons-material/Splitscreen";
import UploadFileIcon from "@mui/icons-material/UploadFile";
import ViewAgendaIcon from "@mui/icons-material/ViewAgenda";
import {
  Box,
  Button,
  Container,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Typography,
} from "@mui/material";
import type { Dispatch, ReactElement, SetStateAction } from "react";
import { useEffect, useRef, useState } from "react";
import { ConfirmActionDialog } from "./ConfirmActionDialog";
import { ExplorerActionBar } from "./ExplorerActionBar";
import { ExplorerActionQueue } from "./ExplorerActionQueue";
import { ExplorerPanel } from "./ExplorerPanel";
import type {
  ExplorerActionTask,
  ExplorerQueuedAction,
  FileExplorerContextValue,
} from "./FileExplorerContext";
import { useFileExplorer } from "./FileExplorerContext";
import type { FileEntry, OperationJob } from "./file-service";
import { canPreviewFile, createDirectory, deleteEntry, transferEntry } from "./file-service";
import { NewFolderDialog } from "./NewFolderDialog";
import { displayName, fileServicePath } from "./paths";

/** Snapshot of a consequential action while its confirmation dialog is open. */
interface PendingAction {
  kind: "copy" | "move" | "delete";
  entries: FileEntry[];
  destination?: string;
  panel: 0 | 1;
}

/** Inputs used to snapshot a pending operation from the active selection. */
interface PreparePendingActionOptions {
  kind: PendingAction["kind"];
  selected: FileEntry[];
  panel: 0 | 1;
  doublePanel: boolean;
  panels: [{ directory: string }, { directory: string }];
}

/** Props for all workspace dialogs, collected from the current workspace state. */
interface WorkspaceOverlaysProps {
  directory: string;
  uploadOpen: boolean;
  folderOpen: boolean;
  previewPath: string | null;
  pendingAction: PendingAction | null;
  onCloseUpload: () => void;
  onCloseFolder: () => void;
  onUploaded: () => void;
  onCreateFolder: (name: string) => Promise<void>;
  onCloseAction: () => void;
  onConfirmAction: () => Promise<void>;
  onClosePreview: () => void;
}

/** Builds a stable action snapshot when the current selection can be acted on. */
function preparePendingAction({
  kind,
  selected,
  panel,
  doublePanel,
  panels,
}: PreparePendingActionOptions): PendingAction | null {
  if (!canPrepareAction(kind, selected.length, doublePanel)) {
    return null;
  }
  const destination = kind === "delete" ? undefined : panels[panel === 0 ? 1 : 0].directory;
  return { kind, entries: [...selected], panel, destination };
}

/** Checks whether the current selection supports the requested action. */
function canPrepareAction(
  kind: PendingAction["kind"],
  selectionCount: number,
  doublePanel: boolean,
): boolean {
  return selectionCount > 0 && (kind === "delete" || doublePanel);
}

/** Runs one confirmed copy, move, or delete request. */
function runPendingEntry(action: PendingAction, entry: FileEntry): Promise<OperationJob | null> {
  if (action.kind === "delete") {
    return deleteEntry(entry);
  }
  const targetDirectory = action.destination === "/" ? "" : action.destination;
  return transferEntry(action.kind, entry.path, `${targetDirectory}/${displayName(entry.path)}`);
}

/** Converts the selected entries into one service task per entry. */
function pendingActionTasks(action: PendingAction): ExplorerActionTask[] {
  return action.entries.map((entry) => ({
    label: `${action.kind} ${entry.path}`,
    run: () => runPendingEntry(action, entry),
  }));
}

/** Refreshes panel listings after each server action reaches its done state. */
function useCompletedActionRefresh(
  queue: ExplorerQueuedAction[],
  setRefresh: Dispatch<SetStateAction<number>>,
): void {
  const completed = useRef(new Set<string>());
  useEffect(() => {
    const newlyFinished = queue.filter(
      (item) => item.status === "done" && !completed.current.has(item.id),
    );
    newlyFinished.forEach((item) => {
      completed.current.add(item.id);
    });
    if (newlyFinished.length > 0) {
      setRefresh((value) => value + 1);
    }
  }, [queue, setRefresh]);
}

/** Displays the app title, active panel, upload, and view controls. */
function WorkspaceHeader({
  doublePanel,
  activePanel,
  onUpload,
  onToggleView,
}: {
  doublePanel: boolean;
  activePanel: 0 | 1;
  onUpload: () => void;
  onToggleView: () => void;
}): ReactElement {
  return (
    <Stack direction="row" alignItems="center" justifyContent="space-between" gap={2}>
      <Box>
        <Typography variant="h1" data-testid="file-explorer-title">
          File Explorer
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {doublePanel ? "Two panels" : "One panel"} · {activePanel === 0 ? "Left" : "Right"} panel
          selected
        </Typography>
      </Box>
      <Stack direction="row" alignItems="center" gap={1}>
        <Button startIcon={<UploadFileIcon />} onClick={onUpload}>
          Upload
        </Button>
        <IconButton
          aria-label={doublePanel ? "Show one panel" : "Show two panels"}
          onClick={onToggleView}
          data-testid="file-explorer-view-toggle"
        >
          {doublePanel ? <ViewAgendaIcon /> : <SplitscreenIcon />}
        </IconButton>
      </Stack>
    </Stack>
  );
}

/** Displays the selected action's exact confirmation and service result. */
function ActionConfirmation({
  action,
  onClose,
  onConfirm,
}: {
  action: PendingAction | null;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}): ReactElement {
  return (
    <ConfirmActionDialog
      open={action !== null}
      title={
        action
          ? `${action.kind[0].toUpperCase()}${action.kind.slice(1)} ${action.entries.length} item(s)`
          : "Confirm action"
      }
      description={confirmationDescription(action)}
      confirmLabel={action?.kind ?? "Confirm"}
      onClose={onClose}
      onConfirm={onConfirm}
    />
  );
}

/** Creates the confirmation text from the action snapshot. */
function confirmationDescription(action: PendingAction | null): string {
  if (action === null) {
    return "";
  }
  const destination = action.destination ? ` to ${action.destination}` : "";
  return `Type yes to ${action.kind} ${action.entries.length} selected item(s)${destination}. Existing files at the destination will be replaced.`;
}

/** Displays the upload and new-folder dialogs for the active directory. */
function DirectoryDialogs({
  directory,
  uploadOpen,
  folderOpen,
  onCloseUpload,
  onCloseFolder,
  onUploaded,
  onCreateFolder,
}: Pick<
  WorkspaceOverlaysProps,
  | "directory"
  | "uploadOpen"
  | "folderOpen"
  | "onCloseUpload"
  | "onCloseFolder"
  | "onUploaded"
  | "onCreateFolder"
>): ReactElement {
  return (
    <>
      <UploadDialog
        open={uploadOpen}
        directory={directory}
        onClose={onCloseUpload}
        onUploaded={onUploaded}
      />
      <NewFolderDialog
        open={folderOpen}
        directory={directory}
        onClose={onCloseFolder}
        onCreate={onCreateFolder}
      />
    </>
  );
}

/** Displays an inline preview for the selected File Service resource. */
function PreviewDialog({
  path,
  onClose,
}: {
  path: string | null;
  onClose: () => void;
}): ReactElement {
  return (
    <Dialog open={path !== null} onClose={onClose} fullWidth maxWidth="lg">
      <DialogTitle>Preview</DialogTitle>
      <DialogContent>
        {path !== null && (
          <iframe
            title="File preview"
            src={fileServicePath("/f", path)}
            style={{ width: "100%", height: "70vh", border: 0 }}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

/** Renders upload, folder, action confirmation, and preview dialogs together. */
function WorkspaceOverlays(props: WorkspaceOverlaysProps): ReactElement {
  return (
    <>
      <DirectoryDialogs {...props} />
      <ActionConfirmation
        action={props.pendingAction}
        onClose={props.onCloseAction}
        onConfirm={props.onConfirmAction}
      />
      <PreviewDialog path={props.previewPath} onClose={props.onClosePreview} />
    </>
  );
}

/** Owns dialog state and actions that operate on the active panel selection. */
function useWorkspaceControls(
  explorer: FileExplorerContextValue,
  selected: FileEntry[],
  setRefresh: Dispatch<SetStateAction<number>>,
) {
  const [uploadOpen, setUploadOpen] = useState(false);
  const [folderOpen, setFolderOpen] = useState(false);
  const [previewPath, setPreviewPath] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);

  /** Captures the selected entries and destination before opening confirmation. */
  const prepare = (kind: PendingAction["kind"]): void => {
    setPendingAction(
      preparePendingAction({
        kind,
        selected,
        panel: explorer.activePanel,
        doublePanel: explorer.doublePanel,
        panels: explorer.panels,
      }),
    );
  };

  /** Submits one queued task for every confirmed entry. */
  const confirmAction = async (): Promise<void> => {
    if (pendingAction === null) {
      return;
    }
    await explorer.enqueueActions(pendingActionTasks(pendingAction));
    explorer.clearSelection(pendingAction.panel);
  };

  /** Creates a child folder and refreshes the active directory listing. */
  const makeFolder = async (name: string): Promise<void> => {
    const directory = explorer.panels[explorer.activePanel].directory;
    const path = `${directory === "/" ? "" : directory}/${name}`;
    await createDirectory(path);
    setRefresh((value) => value + 1);
  };

  return {
    uploadOpen,
    setUploadOpen,
    folderOpen,
    setFolderOpen,
    previewPath,
    setPreviewPath,
    pendingAction,
    setPendingAction,
    prepare,
    confirmAction,
    makeFolder,
  };
}

/** Two persistent browser panels and actions targeting the highlighted panel. */
export function FileExplorerWorkspace(): ReactElement {
  const explorer = useFileExplorer();
  const [refresh, setRefresh] = useState(0);
  useCompletedActionRefresh(explorer.queue, setRefresh);
  const selected = explorer.panels[explorer.activePanel].selected;
  const previewable = selected.length === 1 && canPreviewFile(selected[0]);
  const controls = useWorkspaceControls(explorer, selected, setRefresh);

  return (
    <Container component="main" maxWidth={false} sx={{ py: 3 }} data-testid="file-explorer-page">
      <Stack spacing={2}>
        <WorkspaceHeader
          doublePanel={explorer.doublePanel}
          activePanel={explorer.activePanel}
          onUpload={() => controls.setUploadOpen(true)}
          onToggleView={explorer.toggleView}
        />
        <PanelPair doublePanel={explorer.doublePanel} refresh={refresh} />
        <WorkspaceActions
          doublePanel={explorer.doublePanel}
          selected={selected}
          previewable={previewable}
          onPrepare={controls.prepare}
          onPreview={() => controls.setPreviewPath(selected[0]?.path ?? null)}
          onUpload={() => controls.setUploadOpen(true)}
          onNewFolder={() => controls.setFolderOpen(true)}
        />
        <ExplorerActionQueue />
      </Stack>
      <WorkspaceOverlays
        directory={explorer.panels[explorer.activePanel].directory}
        uploadOpen={controls.uploadOpen}
        folderOpen={controls.folderOpen}
        previewPath={controls.previewPath}
        pendingAction={controls.pendingAction}
        onCloseUpload={() => controls.setUploadOpen(false)}
        onCloseFolder={() => controls.setFolderOpen(false)}
        onUploaded={() => setRefresh((value) => value + 1)}
        onCreateFolder={controls.makeFolder}
        onCloseAction={() => controls.setPendingAction(null)}
        onConfirmAction={controls.confirmAction}
        onClosePreview={() => controls.setPreviewPath(null)}
      />
    </Container>
  );
}

/** Displays both panels and the active panel's shared action bar. */
function PanelPair({
  doublePanel,
  refresh,
}: {
  doublePanel: boolean;
  refresh: number;
}): ReactElement {
  return (
    <Stack direction={{ xs: "column", md: "row" }} gap={2} alignItems="stretch">
      <ExplorerPanel panel={0} refresh={refresh} />
      {doublePanel && <ExplorerPanel panel={1} refresh={refresh} />}
    </Stack>
  );
}

/** Connects action-bar controls to the active selection and dialogs. */
function WorkspaceActions({
  doublePanel,
  selected,
  previewable,
  onPrepare,
  onPreview,
  onUpload,
  onNewFolder,
}: {
  doublePanel: boolean;
  selected: FileEntry[];
  previewable: boolean;
  onPrepare: (kind: PendingAction["kind"]) => void;
  onPreview: () => void;
  onUpload: () => void;
  onNewFolder: () => void;
}): ReactElement {
  return (
    <ExplorerActionBar
      doublePanel={doublePanel}
      hasSelection={selected.length > 0}
      canPreview={previewable}
      canDelete={selected.length > 0}
      onCopy={() => onPrepare("copy")}
      onMove={() => onPrepare("move")}
      onPreview={onPreview}
      onDelete={() => onPrepare("delete")}
      onUpload={onUpload}
      onNewFolder={onNewFolder}
    />
  );
}
