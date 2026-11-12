import { LanguageSelector, useLanguage } from "@lite-nas/language";
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
  const { getLabel } = useLanguage();
  return (
    <Paper variant="outlined" sx={{ p: 1 }} data-testid="file-explorer-actions">
      <Stack
        direction="row"
        alignItems="center"
        justifyContent="space-between"
        flexWrap="wrap"
        gap={1}
      >
        <Stack direction="row" gap={0.5} flexWrap="wrap">
          {doublePanel && (
            <>
              <Button
                startIcon={<ContentCopyIcon />}
                disabled={!hasSelection}
                onClick={onCopy}
                data-testid="file-explorer-copy-button"
              >
                {getLabel("common.copy")}
              </Button>
              <Button
                startIcon={<DriveFileMoveIcon />}
                disabled={!hasSelection}
                onClick={onMove}
                data-testid="file-explorer-move-button"
              >
                {getLabel("common.move")}
              </Button>
            </>
          )}
          <Button
            startIcon={<PreviewIcon />}
            disabled={!canPreview}
            onClick={onPreview}
            data-testid="file-explorer-preview-button"
          >
            {getLabel("common.preview")}
          </Button>
          <Button
            startIcon={<DeleteIcon />}
            disabled={!canDelete}
            onClick={onDelete}
            data-testid="file-explorer-delete-button"
          >
            {getLabel("common.delete")}
          </Button>
          <Button
            startIcon={<UploadFileIcon />}
            onClick={onUpload}
            data-testid="file-explorer-upload-button"
          >
            {getLabel("common.upload")}
          </Button>
          <Button
            startIcon={<CreateNewFolderIcon />}
            onClick={onNewFolder}
            data-testid="file-explorer-new-folder-button"
          >
            {getLabel("common.newFolder")}
          </Button>
        </Stack>
        <Stack direction="row" flexWrap="wrap" gap={0.5}>
          <LanguageSelector />
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
  const { getLabel } = useLanguage();
  return (
    <IconButton
      aria-label={getLabel(showHidden ? "explorer.hideHidden" : "explorer.showHidden")}
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
  const { getLabel } = useLanguage();
  return (
    <IconButton
      aria-label={getLabel(doublePanel ? "explorer.onePanel" : "explorer.twoPanels")}
      onClick={onToggle}
      data-testid="file-explorer-view-toggle"
    >
      {doublePanel ? <ViewAgendaIcon /> : <SplitscreenIcon />}
    </IconButton>
  );
}
