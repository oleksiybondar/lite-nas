import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import CreateNewFolderIcon from "@mui/icons-material/CreateNewFolder";
import DeleteIcon from "@mui/icons-material/Delete";
import DriveFileMoveIcon from "@mui/icons-material/DriveFileMove";
import PreviewIcon from "@mui/icons-material/Preview";
import SplitscreenIcon from "@mui/icons-material/Splitscreen";
import UploadFileIcon from "@mui/icons-material/UploadFile";
import ViewAgendaIcon from "@mui/icons-material/ViewAgenda";
import VisibilityIcon from "@mui/icons-material/Visibility";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import { Button, IconButton, Paper, Stack } from "@mui/material";
import type { ReactElement } from "react";

/** Commands for the active panel, with transfer commands in two-panel view. */
export function ExplorerActionBar({
  doublePanel,
  hasSelection,
  canPreview,
  canDelete,
  onCopy,
  onMove,
  onPreview,
  onDelete,
  onUpload,
  onNewFolder,
  onToggleView,
  showHidden,
  onToggleHidden,
}: {
  doublePanel: boolean;
  hasSelection: boolean;
  canPreview: boolean;
  canDelete: boolean;
  onCopy: () => void;
  onMove: () => void;
  onPreview: () => void;
  onDelete: () => void;
  onUpload: () => void;
  onNewFolder: () => void;
  onToggleView?: () => void;
  showHidden: boolean;
  onToggleHidden: () => void;
}): ReactElement {
  return (
    <Paper variant="outlined" sx={{ p: 1 }} data-testid="file-explorer-actions">
      <Stack direction="row" alignItems="center" justifyContent="space-between" gap={1}>
        <Stack direction="row" gap={0.5} flexWrap="wrap">
          {doublePanel && (
            <>
              <Button
                startIcon={<ContentCopyIcon />}
                disabled={!hasSelection}
                onClick={onCopy}
                data-testid="file-explorer-copy-button"
              >
                Copy
              </Button>
              <Button
                startIcon={<DriveFileMoveIcon />}
                disabled={!hasSelection}
                onClick={onMove}
                data-testid="file-explorer-move-button"
              >
                Move
              </Button>
            </>
          )}
          <Button
            startIcon={<PreviewIcon />}
            disabled={!canPreview}
            onClick={onPreview}
            data-testid="file-explorer-preview-button"
          >
            Preview
          </Button>
          <Button
            startIcon={<DeleteIcon />}
            disabled={!canDelete}
            onClick={onDelete}
            data-testid="file-explorer-delete-button"
          >
            Delete
          </Button>
          <Button
            startIcon={<UploadFileIcon />}
            onClick={onUpload}
            data-testid="file-explorer-upload-button"
          >
            Upload
          </Button>
          <Button
            startIcon={<CreateNewFolderIcon />}
            onClick={onNewFolder}
            data-testid="file-explorer-new-folder-button"
          >
            New folder
          </Button>
        </Stack>
        <Stack direction="row" gap={0.5}>
          <HiddenFilesToggle showHidden={showHidden} onToggle={onToggleHidden} />
          <PanelViewToggle doublePanel={doublePanel} onToggle={onToggleView} />
        </Stack>
      </Stack>
    </Paper>
  );
}

/** Toggles visibility of dot-prefixed files and folders with an eye-state icon. */
function HiddenFilesToggle({
  showHidden,
  onToggle,
}: {
  showHidden: boolean;
  onToggle: () => void;
}): ReactElement {
  return (
    <IconButton
      aria-label={showHidden ? "Hide hidden files" : "Show hidden files"}
      aria-pressed={showHidden}
      onClick={onToggle}
      data-testid="file-explorer-hidden-toggle"
      sx={{
        bgcolor: showHidden ? "success.light" : "action.selected",
        color: showHidden ? "success.dark" : "text.secondary",
        "&:hover": { bgcolor: showHidden ? "success.main" : "action.hover" },
      }}
    >
      {showHidden ? <VisibilityIcon /> : <VisibilityOffIcon />}
    </IconButton>
  );
}

/** Switches between one-panel and two-panel workspace layouts. */
function PanelViewToggle({
  doublePanel,
  onToggle,
}: {
  doublePanel: boolean;
  onToggle?: () => void;
}): ReactElement {
  return (
    <IconButton
      aria-label={doublePanel ? "Show one panel" : "Show two panels"}
      onClick={onToggle}
      data-testid="file-explorer-view-toggle"
    >
      {doublePanel ? <ViewAgendaIcon /> : <SplitscreenIcon />}
    </IconButton>
  );
}
