import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import CreateNewFolderIcon from "@mui/icons-material/CreateNewFolder";
import DeleteIcon from "@mui/icons-material/Delete";
import DriveFileMoveIcon from "@mui/icons-material/DriveFileMove";
import PreviewIcon from "@mui/icons-material/Preview";
import UploadFileIcon from "@mui/icons-material/UploadFile";
import { Button, Paper, Stack } from "@mui/material";
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
}): ReactElement {
  return (
    <Paper variant="outlined" sx={{ p: 1.5 }} data-testid="file-explorer-actions">
      <Stack direction="row" gap={1} flexWrap="wrap">
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
    </Paper>
  );
}
