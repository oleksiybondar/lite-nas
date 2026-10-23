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
import type { FileEntry } from "./file-service";
import { listDirectory } from "./file-service";
import { fileServicePath, parentPath } from "./paths";

interface DirectoryListing {
  entries: FileEntry[];
  loading: boolean;
  error: string;
}

/** Loads one directory and ignores responses from an obsolete panel request. */
function useDirectoryListing(directory: string): DirectoryListing {
  const [entries, setEntries] = useState<FileEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setError("");
    void listDirectory(directory, controller.signal)
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
  }, [directory]);

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
  open,
}: {
  panel: 0 | 1;
  directory: string;
  parent: string | null;
  entries: FileEntry[];
  selected: FileEntry[];
  navigate: (panel: 0 | 1, directory: string) => void;
  select: (panel: 0 | 1, entry: FileEntry, multiple: boolean) => void;
  open: (entry: FileEntry) => void;
}): ReactElement {
  return (
    <List aria-label={`Files and directories in ${directory}`} onKeyDown={moveEntryFocus}>
      {parent !== null && (
        <ListItemButton onClick={() => navigate(panel, parent)}>
          <ListItemIcon>
            <ArrowUpwardIcon />
          </ListItemIcon>
          <ListItemText primary=".." secondary="Parent directory" />
        </ListItemButton>
      )}
      {entries.map((entry) => (
        <FileEntryItem
          key={entry.path}
          entry={entry}
          selected={selected.some((item) => item.path === entry.path)}
          onSelect={(multiple) => select(panel, entry, multiple)}
          onOpen={() => open(entry)}
        />
      ))}
    </List>
  );
}

/** Independently browsable panel with navigation controls outside its error state. */
export function ExplorerPanel({ panel, refresh }: { panel: 0 | 1; refresh: number }): ReactElement {
  const { panels, activePanel, setActivePanel, navigate, back, select } = useFileExplorer();
  const { directory, history, selected } = panels[panel];
  const [localRefresh, setLocalRefresh] = useState(0);
  const parent = parentPath(directory);

  /** Opens a folder in this panel or a file in a new tab. */
  const open = (entry: FileEntry): void => {
    if (entry.kind === "directory") {
      navigate(panel, entry.path);
      return;
    }
    window.open(fileServicePath("/f", entry.path), "_blank", "noopener,noreferrer");
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
        minWidth: 0,
        overflow: "hidden",
      }}
    >
      <Stack spacing={1} sx={{ p: 2 }}>
        <PanelHeader
          directory={directory}
          selectedCount={selected.length}
          canGoBack={history.length > 0}
          onBack={() => back(panel)}
          onRefresh={() => setLocalRefresh((value) => value + 1)}
        />
        <PanelDirectory
          key={`${refresh}-${localRefresh}`}
          panel={panel}
          directory={directory}
          parent={parent}
          selected={selected}
          navigate={navigate}
          select={select}
          open={open}
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
function PanelDirectory({
  panel,
  directory,
  parent,
  selected,
  navigate,
  select,
  open,
}: {
  panel: 0 | 1;
  directory: string;
  parent: string | null;
  selected: FileEntry[];
  navigate: (panel: 0 | 1, directory: string) => void;
  select: (panel: 0 | 1, entry: FileEntry, multiple: boolean) => void;
  open: (entry: FileEntry) => void;
}): ReactElement {
  const { entries, loading, error } = useDirectoryListing(directory);
  return (
    <>
      <DirectoryStatus loading={loading} error={error} />
      {!loading && !error && (
        <>
          <DirectoryRows
            panel={panel}
            directory={directory}
            parent={parent}
            entries={entries}
            selected={selected}
            navigate={navigate}
            select={select}
            open={open}
          />
          {entries.length === 0 && (
            <Typography color="text.secondary">This directory is empty.</Typography>
          )}
        </>
      )}
    </>
  );
}
