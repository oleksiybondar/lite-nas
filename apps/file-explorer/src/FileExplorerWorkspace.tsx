import { UploadDialog } from "@lite-nas/file-upload";
import { useLanguage } from "@lite-nas/language";
import CloseIcon from "@mui/icons-material/Close";
import {
  Alert,
  Box,
  CircularProgress,
  Container,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Typography,
} from "@mui/material";
import MarkdownIt from "markdown-it";
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
import {
  canPreviewFile,
  createDirectory,
  deleteEntry,
  filePreviewKind,
  transferEntry,
  visibleFileEntries,
} from "./file-service";
import { highlightSource } from "./highlight-source";
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
  onCreateFolder: (name: string, privateDirectory: boolean) => Promise<void>;
  onCloseAction: () => void;
  onConfirmAction: () => Promise<void>;
  onClosePreview: () => void;
}

/** Keeps the workspace panels scroll-bounded while the action bar remains visible. */
const workspaceContainerSx = {
  height: "100dvh",
  boxSizing: "border-box",
  p: 2,
  display: "flex",
  flexDirection: "column",
  gap: 1.5,
  overflow: "hidden",
} as const;

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

/** Displays the app title and any active action queue in the fixed header. */
function WorkspaceHeader(): ReactElement {
  const { getLabel } = useLanguage();
  return (
    <Stack direction="row" alignItems="center" justifyContent="space-between" gap={2}>
      <Typography variant="h1" data-testid="file-explorer-title">
        {getLabel("explorer.title")}
      </Typography>
      <ExplorerActionQueue />
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
  const { getLabel } = useLanguage();
  return (
    <ConfirmActionDialog
      open={action !== null}
      title={getLabel("explorer.confirmAction")}
      description={confirmationDescription(action, getLabel)}
      confirmLabel={action ? getLabel(`common.${action.kind}`) : getLabel("explorer.confirm")}
      onClose={onClose}
      onConfirm={onConfirm}
    />
  );
}

/** Creates the confirmation text from the action snapshot. */
function confirmationDescription(
  action: PendingAction | null,
  getLabel: (key: string) => string,
): string {
  if (action === null) {
    return "";
  }
  return getLabel("explorer.confirmSummary")
    .replace("{action}", getLabel(`common.${action.kind}`))
    .replace("{count}", String(action.entries.length))
    .replace(
      "{destination}",
      action.destination
        ? ` ${getLabel("explorer.confirmDestination").replace("{path}", action.destination)}`
        : "",
    );
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
  const { getLabel } = useLanguage();
  const kind = path === null ? null : filePreviewKind({ kind: "file", path });
  const extension = path?.slice(path.lastIndexOf(".") + 1).toLowerCase() ?? "";
  const PreviewContent = kind === null ? null : previewContentByKind[kind];
  return (
    <Dialog open={path !== null} onClose={onClose} fullWidth maxWidth="lg">
      <DialogTitle sx={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
        {path === null ? getLabel("explorer.preview") : displayName(path)}
        <IconButton
          aria-label={getLabel("explorer.closePreview")}
          onClick={onClose}
          edge="end"
          size="small"
        >
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent sx={{ minHeight: 160 }}>
        {path !== null && PreviewContent !== null && (
          <PreviewContent path={path} extension={extension} />
        )}
      </DialogContent>
    </Dialog>
  );
}

interface PreviewContentProps {
  path: string;
  extension: string;
}

/** Selects the preview renderer assigned to each supported extension group. */
const previewContentByKind: Record<
  NonNullable<ReturnType<typeof filePreviewKind>>,
  (props: PreviewContentProps) => ReactElement
> = {
  image: ImagePreview,
  pdf: PdfPreview,
  media: MediaPreview,
  text: TextFilePreview,
};

/** Shows an image within the available preview area without cropping it. */
function ImagePreview({ path }: PreviewContentProps): ReactElement {
  return (
    <Box
      component="img"
      alt={displayName(path)}
      src={fileServicePath("/f", path)}
      sx={{ display: "block", maxWidth: "100%", maxHeight: "70vh", mx: "auto" }}
    />
  );
}

/** Shows a PDF through the browser's built-in document viewer. */
function PdfPreview({ path }: PreviewContentProps): ReactElement {
  const { getLabel } = useLanguage();
  return (
    <iframe
      title={getLabel("explorer.pdfPreview")}
      src={fileServicePath("/f", path)}
      style={{ width: "100%", height: "70vh", border: 0 }}
    />
  );
}

/** Shows browser controls for a supported video or audio resource. */
function MediaPreview({ path, extension }: PreviewContentProps): ReactElement {
  const { getLabel } = useLanguage();
  const source = fileServicePath("/f", path);
  if (["mp4", "m4v", "mov", "avi", "mkv", "webm", "ogv"].includes(extension)) {
    return (
      <Box
        component="video"
        controls
        preload="metadata"
        src={source}
        sx={{ width: "100%", maxHeight: "70vh" }}
      >
        {getLabel("explorer.videoUnsupported")}
      </Box>
    );
  }
  return (
    <Box component="audio" controls preload="metadata" src={source} sx={{ width: "100%", mt: 2 }}>
      {getLabel("explorer.audioUnsupported")}
    </Box>
  );
}

/** Loads authorized text and displays it as escaped source with its file type. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: Loading, error, and source rendering form one preview lifecycle.
function TextFilePreview({ path, extension }: PreviewContentProps): ReactElement {
  const { getLabel } = useLanguage();
  const [content, setContent] = useState<string | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    setContent(null);
    setError("");
    void loadTextPreview(path, controller.signal, getLabel)
      .then(setContent)
      .catch((caught: unknown) =>
        setError(previewLoadError(caught, controller.signal.aborted, getLabel)),
      );
    return () => controller.abort();
  }, [path, getLabel]);

  const previewStatus = renderTextPreviewStatus(error, content, getLabel);
  if (previewStatus) return previewStatus;
  if (content === null) return null;
  if (extension === "md") return <MarkdownFilePreview content={content} />;
  const highlighted = highlightSource(content, extension);
  return (
    <Box>
      <Typography variant="caption" color="text.secondary">
        {getLabel("explorer.sourceLabel").replace("{extension}", extension.toUpperCase())}
      </Typography>
      <Box
        component="pre"
        sx={{
          maxHeight: "70vh",
          overflow: "auto",
          p: 2,
          bgcolor: "action.hover",
          borderRadius: 1,
          fontFamily: "monospace",
          whiteSpace: "pre",
          "& .token.comment, & .token.prolog, & .token.cdata": {
            color: "text.secondary",
            fontStyle: "italic",
          },
          "& .token.keyword, & .token.boolean, & .token.atrule": {
            color: "primary.main",
            fontWeight: 700,
          },
          "& .token.string, & .token.attr-value, & .token.regex": { color: "success.main" },
          "& .token.number, & .token.constant": { color: "warning.main" },
          "& .token.function, & .token.class-name, & .token.selector": {
            color: "secondary.main",
          },
          "& .token.operator, & .token.punctuation": { color: "text.secondary" },
        }}
      >
        <code>{highlighted ?? content}</code>
      </Box>
    </Box>
  );
}

/** Renders the loading or error state before text content is available. */
function renderTextPreviewStatus(
  error: string,
  content: string | null,
  getLabel: (key: string) => string,
): ReactElement | null {
  if (error) return <Alert severity="error">{error}</Alert>;
  if (content === null) {
    return <CircularProgress aria-label={getLabel("explorer.textPreviewLoading")} />;
  }
  return null;
}

const markdownRenderer = new MarkdownIt({ html: false, linkify: false, typographer: false });

/** Renders Markdown syntax while keeping embedded HTML inert. */
function MarkdownFilePreview({ content }: { content: string }): ReactElement {
  return (
    <Box
      className="file-explorer-markdown-preview"
      sx={{
        maxHeight: "70vh",
        overflow: "auto",
        overflowWrap: "anywhere",
        "& img": { maxWidth: "100%" },
        "& pre": { overflow: "auto", p: 2, bgcolor: "action.hover", borderRadius: 1 },
        "& table": { borderCollapse: "collapse" },
        "& th, & td": { border: 1, borderColor: "divider", p: 1 },
      }}
      // biome-ignore lint/security/noDangerouslySetInnerHtml: MarkdownIt escapes raw HTML and disables linkification for this preview.
      dangerouslySetInnerHTML={{ __html: markdownRenderer.render(content) }}
    />
  );
}

/** Fetches one source file with the current same-origin user session. */
async function loadTextPreview(
  path: string,
  signal: AbortSignal,
  getLabel: (key: string) => string,
): Promise<string> {
  const response = await fetch(fileServicePath("/f", path), {
    credentials: "same-origin",
    signal,
  });
  if (!response.ok) {
    throw new Error(
      getLabel("explorer.fileLoadHttpFailed").replace("{status}", String(response.status)),
    );
  }
  return response.text();
}

/** Converts a preview fetch error to a visible message, ignoring dismissed loads. */
function previewLoadError(
  caught: unknown,
  aborted: boolean,
  getLabel: (key: string) => string,
): string {
  if (aborted) return "";
  return caught instanceof Error ? caught.message : getLabel("explorer.fileLoadFailed");
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
  const makeFolder = async (name: string, privateDirectory: boolean): Promise<void> => {
    const directory = explorer.panels[explorer.activePanel].directory;
    const path = `${directory === "/" ? "" : directory}/${name}`;
    await createDirectory(path, privateDirectory);
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
  const [showHidden, setShowHidden] = useState(true);
  useCompletedActionRefresh(explorer.queue, setRefresh);
  const selected = visibleFileEntries(explorer.panels[explorer.activePanel].selected, showHidden);
  const previewable = selected.length === 1 && canPreviewFile(selected[0]);
  const controls = useWorkspaceControls(explorer, selected, setRefresh);

  return (
    <Container
      component="main"
      maxWidth={false}
      sx={workspaceContainerSx}
      data-testid="file-explorer-page"
    >
      <WorkspaceHeader />
      <WorkspaceBody
        explorer={explorer}
        refresh={refresh}
        showHidden={showHidden}
        setShowHidden={setShowHidden}
        selected={selected}
        previewable={previewable}
        controls={controls}
        setRefresh={setRefresh}
      />
    </Container>
  );
}

/** Composes the panels, action bar, and dialogs for the active workspace state. */
function WorkspaceBody({
  explorer,
  refresh,
  showHidden,
  setShowHidden,
  selected,
  previewable,
  controls,
  setRefresh,
}: {
  explorer: FileExplorerContextValue;
  refresh: number;
  showHidden: boolean;
  setShowHidden: Dispatch<SetStateAction<boolean>>;
  selected: FileEntry[];
  previewable: boolean;
  controls: ReturnType<typeof useWorkspaceControls>;
  setRefresh: Dispatch<SetStateAction<number>>;
}): ReactElement {
  return (
    <>
      <Box sx={{ flex: 1, minHeight: 0 }}>
        <PanelPair doublePanel={explorer.doublePanel} refresh={refresh} showHidden={showHidden} />
      </Box>
      <Box sx={{ flex: "0 0 auto" }}>
        <WorkspaceActions
          doublePanel={explorer.doublePanel}
          showHidden={showHidden}
          selected={selected}
          previewable={previewable}
          onToggleView={explorer.toggleView}
          onToggleHidden={() => setShowHidden((current) => !current)}
          onPrepare={controls.prepare}
          onPreview={() => controls.setPreviewPath(selected[0]?.path ?? null)}
          onUpload={() => controls.setUploadOpen(true)}
          onNewFolder={() => controls.setFolderOpen(true)}
        />
      </Box>
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
    </>
  );
}

/** Displays both panels in the flexible region above the fixed action bar. */
function PanelPair({
  doublePanel,
  refresh,
  showHidden,
}: {
  doublePanel: boolean;
  refresh: number;
  showHidden: boolean;
}): ReactElement {
  return (
    <Stack
      direction={{ xs: "column", md: "row" }}
      gap={1.5}
      alignItems="stretch"
      sx={{ height: "100%", minHeight: 0 }}
    >
      <ExplorerPanel panel={0} refresh={refresh} showHidden={showHidden} />
      {doublePanel && <ExplorerPanel panel={1} refresh={refresh} showHidden={showHidden} />}
    </Stack>
  );
}

/** Connects action-bar controls to the active selection and dialogs. */
function WorkspaceActions({
  doublePanel,
  showHidden,
  selected,
  previewable,
  onToggleView,
  onToggleHidden,
  onPrepare,
  onPreview,
  onUpload,
  onNewFolder,
}: {
  doublePanel: boolean;
  showHidden: boolean;
  selected: FileEntry[];
  previewable: boolean;
  onToggleView: () => void;
  onToggleHidden: () => void;
  onPrepare: (kind: PendingAction["kind"]) => void;
  onPreview: () => void;
  onUpload: () => void;
  onNewFolder: () => void;
}): ReactElement {
  return (
    <ExplorerActionBar
      doublePanel={doublePanel}
      showHidden={showHidden}
      hasSelection={selected.length > 0}
      canPreview={previewable}
      canDelete={selected.length > 0}
      onToggleView={onToggleView}
      onToggleHidden={onToggleHidden}
      onCopy={() => onPrepare("copy")}
      onMove={() => onPrepare("move")}
      onPreview={onPreview}
      onDelete={() => onPrepare("delete")}
      onUpload={onUpload}
      onNewFolder={onNewFolder}
    />
  );
}
