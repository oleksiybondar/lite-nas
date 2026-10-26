import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import ArrowUpwardIcon from "@mui/icons-material/ArrowUpward";
import RefreshIcon from "@mui/icons-material/Refresh";
import {
  Alert,
  Box,
  IconButton,
  LinearProgress,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import type { KeyboardEvent, ReactElement } from "react";
import { useEffect, useState } from "react";

import { FileEntryItem } from "./FileEntryItem";
import { useFileExplorer } from "./FileExplorerContext";
import type { FileEntry, FileServiceRequest } from "./file-service";
import {
  type FileExtensionFilter,
  filterFileEntries,
  listDirectory,
  visibleFileEntries,
} from "./file-service";
import { fileServicePath, parentPath } from "./paths";

interface DirectoryListing {
  entries: FileEntry[];
  loading: boolean;
  error: string;
}

/** Loads one directory and ignores responses from an obsolete panel request. */
function useDirectoryListing(directory: string, request?: FileServiceRequest): DirectoryListing {
  const [entries, setEntries] = useState<FileEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setError("");
    void listDirectory(directory, controller.signal, request)
      .then(setEntries)
      .catch((caught: unknown) => {
        if (!controller.signal.aborted) {
          setEntries([]);
          setError(errorMessage(caught));
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      });
    return () => controller.abort();
  }, [directory, request]);

  return { entries, loading, error };
}

/** Converts a listing failure into the message shown in the panel. */
function errorMessage(caught: unknown): string {
  return caught instanceof Error ? caught.message : "Could not load this directory.";
}

/** Moves keyboard focus to the adjacent entry row. */
function moveEntryFocus(event: KeyboardEvent<HTMLUListElement>): void {
  if (event.key !== "ArrowDown" && event.key !== "ArrowUp") {
    return;
  }
  const items = Array.from(
    event.currentTarget.querySelectorAll<HTMLElement>('[data-test-class="file-explorer-entry"]'),
  );
  const currentIndex = items.indexOf(document.activeElement as HTMLElement);
  const nextIndex =
    event.key === "ArrowDown"
      ? Math.min(currentIndex + 1, items.length - 1)
      : Math.max(currentIndex - 1, 0);
  items[nextIndex]?.focus();
  event.preventDefault();
}

/** Renders loading and error feedback before a panel's directory entries. */
function DirectoryStatus({
  loading,
  error,
}: {
  loading: boolean;
  error: string;
}): ReactElement | null {
  if (loading) {
    return <LinearProgress />;
  }
  if (error) {
    return (
      <Alert severity="error" data-testid="file-explorer-error">
        {error}
      </Alert>
    );
  }
  return null;
}

/** Renders the parent shortcut and selectable entries for one directory. */
function DirectoryRows({
  panel,
  directory,
  parent,
  entries,
  selected,
  navigate,
  select,
  clearSelection,
  open,
  selectableKind,
  singleSelection,
}: {
  panel: 0 | 1;
  directory: string;
  parent: string | null;
  entries: FileEntry[];
  selected: FileEntry[];
  navigate: (panel: 0 | 1, directory: string) => void;
  select: (panel: 0 | 1, entry: FileEntry, multiple: boolean) => void;
  clearSelection: (panel: 0 | 1) => void;
  open: (entry: FileEntry) => void;
  selectableKind?: FileEntry["kind"];
  singleSelection?: boolean;
}): ReactElement {
  return (
    <List
      aria-label={`Files and directories in ${directory}`}
      onKeyDown={moveEntryFocus}
      sx={{ py: 0 }}
    >
      {parent !== null && (
        <ListItemButton onClick={() => navigate(panel, parent)} sx={{ py: 0.5 }}>
          <ListItemIcon sx={{ minWidth: 40 }}>
            <ArrowUpwardIcon />
          </ListItemIcon>
          <ListItemText primary=".." secondary="Parent directory" sx={{ my: 0.25 }} />
        </ListItemButton>
      )}
      {entries.map((entry) => (
        <FileEntryItem
          key={entry.path}
          entry={entry}
          selected={selected.some((item) => item.path === entry.path)}
          onSelect={(multiple) => {
            if (selectableKind !== undefined && entry.kind !== selectableKind) {
              clearSelection(panel);
              return;
            }
            select(panel, entry, !singleSelection && multiple);
          }}
          onOpen={() => open(entry)}
        />
      ))}
    </List>
  );
}

/** Independently browsable panel with navigation controls outside its error state. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: panel options and rendering are kept together for callers.
export function ExplorerPanel({
  panel,
  refresh,
  showHidden,
  extensionFilter,
  selectableKind,
  singleSelection,
  onOpenFile,
  showFiles = true,
  request,
}: {
  panel: 0 | 1;
  refresh: number;
  showHidden: boolean;
  extensionFilter?: FileExtensionFilter;
  selectableKind?: FileEntry["kind"];
  singleSelection?: boolean;
  onOpenFile?: (entry: FileEntry) => void;
  showFiles?: boolean;
  request?: FileServiceRequest;
}): ReactElement {
  const { panels, activePanel, setActivePanel, navigate, back, select, clearSelection } =
    useFileExplorer();
  const { directory, history, selected } = panels[panel];
  const visibleSelected = visibleFileEntries(selected, showHidden);
  const [localRefresh, setLocalRefresh] = useState(0);
  const parent = parentPath(directory);

  /** Opens a folder in this panel or a file in a new tab. */
  const open = (entry: FileEntry): void => {
    if (entry.kind === "directory") {
      navigate(panel, entry.path);
      return;
    }
    if (onOpenFile) {
      onOpenFile(entry);
    } else {
      window.open(fileServicePath("/f", entry.path), "_blank", "noopener,noreferrer");
    }
  };

  return (
    <Paper
      variant="outlined"
      onClick={() => setActivePanel(panel)}
      data-test-class="file-explorer-panel"
      data-test-name={panel === 0 ? "left" : "right"}
      sx={{
        borderColor: activePanel === panel ? "primary.main" : "divider",
        borderWidth: 2,
        flex: 1,
        height: "100%",
        minWidth: 0,
        overflow: "hidden",
      }}
    >
      <Stack spacing={0.5} sx={{ p: 1, height: "100%", minHeight: 0 }}>
        <PanelHeader
          directory={directory}
          selectedCount={visibleSelected.length}
          canGoBack={history.length > 0}
          onBack={() => back(panel)}
          onRefresh={() => setLocalRefresh((value) => value + 1)}
        />
        <PanelDirectory
          key={`${refresh}-${localRefresh}`}
          panel={panel}
          directory={directory}
          parent={parent}
          selected={visibleSelected}
          showHidden={showHidden}
          extensionFilter={extensionFilter}
          selectableKind={selectableKind}
          singleSelection={singleSelection}
          showFiles={showFiles}
          navigate={navigate}
          select={select}
          clearSelection={clearSelection}
          open={open}
          request={request}
        />
      </Stack>
    </Paper>
  );
}

/** Displays the path, selection count, and panel navigation controls. */
function PanelHeader({
  directory,
  selectedCount,
  canGoBack,
  onBack,
  onRefresh,
}: {
  directory: string;
  selectedCount: number;
  canGoBack: boolean;
  onBack: () => void;
  onRefresh: () => void;
}): ReactElement {
  return (
    <Stack direction="row" alignItems="center" gap={1}>
      <IconButton aria-label="Back" disabled={!canGoBack} onClick={onBack}>
        <ArrowBackIcon />
      </IconButton>
      <Box sx={{ flex: 1, minWidth: 0 }}>
        <Typography variant="subtitle1" noWrap title={directory}>
          {directory}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          {selectedCount} selected
        </Typography>
      </Box>
      <IconButton aria-label="Refresh directory" onClick={onRefresh}>
        <RefreshIcon />
      </IconButton>
    </Stack>
  );
}

/** Loads and displays the content area for one mounted panel directory. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: listing viewport and its status are kept together.
function PanelDirectory({
  panel,
  directory,
  parent,
  selected,
  showHidden,
  extensionFilter,
  selectableKind,
  singleSelection,
  showFiles,
  navigate,
  select,
  clearSelection,
  open,
  request,
}: {
  panel: 0 | 1;
  directory: string;
  parent: string | null;
  selected: FileEntry[];
  showHidden: boolean;
  extensionFilter?: FileExtensionFilter;
  selectableKind?: FileEntry["kind"];
  singleSelection?: boolean;
  showFiles: boolean;
  navigate: (panel: 0 | 1, directory: string) => void;
  select: (panel: 0 | 1, entry: FileEntry, multiple: boolean) => void;
  clearSelection: (panel: 0 | 1) => void;
  open: (entry: FileEntry) => void;
  request?: FileServiceRequest;
}): ReactElement {
  const { entries, loading, error } = useDirectoryListing(directory, request);
  const visibleEntries = filterFileEntries(
    visibleFileEntries(entries, showHidden),
    extensionFilter,
  ).filter((entry) => showFiles || entry.kind === "directory");
  return (
    <Box
      sx={{
        flex: 1,
        minWidth: 0,
        minHeight: 0,
        overflowX: "auto",
        overflowY: "auto",
        overscrollBehavior: "contain",
        scrollbarGutter: "stable",
      }}
    >
      <DirectoryStatus loading={loading} error={error} />
      {!loading && !error && (
        <Box sx={{ minWidth: 1040 }}>
          <Box
            role="row"
            aria-label="File details"
            sx={{
              display: "grid",
              gridTemplateColumns: "minmax(260px, 2fr) 100px 130px 100px 130px 130px 190px",
              px: 2,
              py: 0.75,
              color: "text.secondary",
              borderBottom: 1,
              borderColor: "divider",
            }}
          >
            {["Name", "Size", "File type", "Permissions", "Owner", "Group", "Last modified"].map(
              (label) => (
                <Typography key={label} variant="caption" fontWeight="bold">
                  {label}
                </Typography>
              ),
            )}
          </Box>
          <DirectoryRows
            panel={panel}
            directory={directory}
            parent={parent}
            entries={visibleEntries}
            selected={selected}
            navigate={navigate}
            select={select}
            clearSelection={clearSelection}
            open={open}
            selectableKind={selectableKind}
            singleSelection={singleSelection}
          />
          {visibleEntries.length === 0 && (
            <Typography color="text.secondary">This directory is empty.</Typography>
          )}
        </Box>
      )}
    </Box>
  );
}
