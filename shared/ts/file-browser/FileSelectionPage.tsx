import CloseIcon from "@mui/icons-material/Close";
import VisibilityIcon from "@mui/icons-material/Visibility";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import {
  Box,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Paper,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";
import { useState } from "react";
import { ExplorerPanel } from "./ExplorerPanel";
import { FileExplorerProvider, useFileExplorer } from "./FileExplorerContext";
import type { FileExtensionFilter, FileServiceRequest } from "./file-service";
import { filterFileEntries, visibleFileEntries } from "./file-service";
import { hostPathSchema } from "./schemas";

/** Configuration for one reusable File Service folder or file picker. */
export interface FileSelectionPageProps {
  kind: "directory" | "file";
  initialDirectory: string;
  extensionFilter?: FileExtensionFilter;
  showHeading?: boolean;
  onSelect: (path: string) => void;
  request?: FileServiceRequest;
}

/** Props shared by the modal folder and file selectors. */
export interface FileSelectionDialogProps extends FileSelectionPageProps {
  open: boolean;
  onClose: () => void;
}

/** Opens a bounded modal browser with its own scrolling directory list. */
function SelectionDialog({
  open,
  onClose,
  onSelect,
  ...pickerProps
}: FileSelectionDialogProps): ReactElement {
  const title = pickerProps.kind === "directory" ? "Select folder" : "Select file";
  const closeLabel =
    pickerProps.kind === "directory" ? "Close folder selector" : "Close file selector";
  return (
    <Dialog
      open={open}
      onClose={onClose}
      fullWidth
      maxWidth="lg"
      PaperProps={{
        sx: { height: "min(82vh, 800px)", display: "flex", overflow: "hidden" },
      }}
    >
      <DialogTitle sx={{ display: "flex", alignItems: "center", gap: 1 }}>
        <Box sx={{ flex: 1 }}>{title}</Box>
        <IconButton aria-label={closeLabel} onClick={onClose} edge="end" size="small">
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent sx={{ display: "flex", flex: 1, minHeight: 0, overflow: "hidden" }}>
        <FileSelectionPage
          {...pickerProps}
          showHeading={false}
          onSelect={(path) => {
            onSelect(path);
            onClose();
          }}
        />
      </DialogContent>
    </Dialog>
  );
}

/** Reusable modal for selecting one folder from File Service. */
export function FolderSelectionDialog(
  props: Omit<FileSelectionDialogProps, "kind" | "extensionFilter">,
): ReactElement {
  return <SelectionDialog {...props} kind="directory" />;
}

/** Reusable modal for selecting one file with optional extension rules. */
export function FileSelectionDialog(props: Omit<FileSelectionDialogProps, "kind">): ReactElement {
  return <SelectionDialog {...props} kind="file" />;
}

/** Owns a picker-scoped browsing context and starts from a validated host path. */
export function FileSelectionPage(props: FileSelectionPageProps): ReactElement {
  return (
    <FileExplorerProvider initialDirectory={hostPathSchema.parse(props.initialDirectory)}>
      <FileSelectionControls {...props} />
    </FileExplorerProvider>
  );
}

/** Shows a single directory panel and confirms one matching path. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: the two picker modes share one small control flow.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: picker layout stays together for this shared component.
function FileSelectionControls({
  kind,
  extensionFilter,
  showHeading = true,
  onSelect,
  request,
}: FileSelectionPageProps): ReactElement {
  const explorer = useFileExplorer();
  const [showHidden, setShowHidden] = useState(false);
  const panel = explorer.panels[0];
  const selected = filterFileEntries(
    visibleFileEntries(panel.selected, showHidden),
    extensionFilter,
  )[0];
  const selectedPath = selected?.kind === kind ? selected.path : null;
  const chosenPath = kind === "directory" ? (selectedPath ?? panel.directory) : selectedPath;
  const label = kind === "directory" ? "Select folder" : "Select file";

  return (
    <Stack
      spacing={2}
      sx={{ flex: 1, height: "100%", width: "100%", minWidth: 0, minHeight: 0, overflow: "hidden" }}
      data-testid={kind === "directory" ? "file-selection-folder" : "file-selection-file"}
    >
      {showHeading && <Typography variant="h5">{label}</Typography>}
      <Box
        sx={{
          display: "flex",
          flex: 1,
          width: "100%",
          minWidth: 0,
          minHeight: 0,
          overflow: "hidden",
        }}
      >
        <ExplorerPanel
          panel={0}
          refresh={0}
          showHidden={showHidden}
          extensionFilter={extensionFilter}
          selectableKind={kind}
          singleSelection
          showFiles={kind === "file"}
          onOpenFile={(entry) => explorer.select(0, entry, false)}
          request={request}
        />
      </Box>
      <Paper variant="outlined" sx={{ flex: "0 0 auto", p: 1 }}>
        <Stack direction="row" alignItems="center" gap={1}>
          <Tooltip
            title={showHidden ? "Hide hidden files and folders" : "Show hidden files and folders"}
          >
            <IconButton
              aria-label={
                showHidden ? "Hide hidden files and folders" : "Show hidden files and folders"
              }
              aria-pressed={showHidden}
              onClick={() => setShowHidden((visible) => !visible)}
              color={showHidden ? "primary" : "default"}
            >
              {showHidden ? <VisibilityIcon /> : <VisibilityOffIcon />}
            </IconButton>
          </Tooltip>
          <Typography noWrap title={chosenPath ?? "Choose a file"} sx={{ flex: 1, minWidth: 0 }}>
            {chosenPath ?? "Choose a file"}
          </Typography>
          <Button
            variant="contained"
            disabled={chosenPath === null}
            onClick={() => {
              if (chosenPath !== null) onSelect(hostPathSchema.parse(chosenPath));
            }}
          >
            {label}
          </Button>
        </Stack>
      </Paper>
    </Stack>
  );
}
