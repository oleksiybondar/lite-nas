import { formatByteSize } from "@lite-nas/formatting";
import { Box, ListItemButton, ListItemIcon, ListItemText, Typography } from "@mui/material";
import type { MouseEvent, ReactElement } from "react";

import { FileEntryIcon } from "./FileEntryIcon";
import type { FileEntry } from "./file-service";
import { displayName } from "./paths";

/** Human-readable type detail shown under one item name. */
function typeLabel(entry: FileEntry): string {
  if (entry.kind === "directory") {
    return "Folder";
  }
  const name = displayName(entry.path);
  const dot = name.lastIndexOf(".");
  return dot > 0 && dot < name.length - 1 ? `${name.slice(dot + 1).toUpperCase()} file` : "File";
}

/** Converts File Service's rwx permission string to a four-digit octal mode. */
function formatPermissions(permissions: string | null): string {
  if (permissions === null) return "—";
  const bits = [permissions.slice(0, 3), permissions.slice(3, 6), permissions.slice(6, 9)].map(
    (triplet) =>
      Number(triplet[0] === "r") * 4 +
      Number(triplet[1] === "w") * 2 +
      Number("xs".includes(triplet[2])),
  );
  const special =
    Number("sS".includes(permissions[2])) * 4 +
    Number("sS".includes(permissions[5])) * 2 +
    Number("tT".includes(permissions[8]));
  return `${special}${bits.join("")}`;
}

/** Formats a nullable ISO timestamp for the listing's local display. */
function formatModified(value: string | null): string {
  if (value === null) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : date.toLocaleString();
}

/** Selectable file or folder row with a separate open interaction. */
export function FileEntryItem({
  entry,
  selected,
  onSelect,
  onOpen,
}: {
  entry: FileEntry;
  selected: boolean;
  onSelect: (multiple: boolean) => void;
  onOpen: () => void;
}): ReactElement {
  const name = displayName(entry.path);

  /** Uses Ctrl or Command to toggle an item in a multi-selection. */
  const handleClick = (event: MouseEvent<HTMLElement>): void => {
    if (event.detail === 0) {
      return;
    }
    onSelect(event.ctrlKey || event.metaKey);
  };

  return (
    <ListItemButton
      selected={selected}
      sx={{
        py: 0.5,
        display: "grid",
        gridTemplateColumns: "minmax(260px, 2fr) 100px 130px 100px 130px 130px 190px",
        minWidth: 1040,
      }}
      onClick={handleClick}
      onDoubleClick={onOpen}
      onKeyDown={(event) => {
        if (event.key === "Enter") {
          onOpen();
        } else if (event.key === " " || event.key === "Spacebar") {
          event.preventDefault();
          onSelect(true);
        }
      }}
      data-test-class="file-explorer-entry"
      data-test-path={entry.path}
    >
      <Box sx={{ display: "flex", alignItems: "center", minWidth: 0 }}>
        <ListItemIcon sx={{ minWidth: 40 }}>
          <FileEntryIcon kind={entry.kind} name={name} />
        </ListItemIcon>
        <ListItemText primary={name} sx={{ my: 0.25, minWidth: 0 }} />
      </Box>
      <Typography variant="body2" color="text.secondary">
        {formatByteSize(entry.size)}
      </Typography>
      <Typography variant="body2" color="text.secondary">
        {typeLabel(entry)}
      </Typography>
      <Typography variant="body2" color="text.secondary">
        {formatPermissions(entry.permissions)}
      </Typography>
      <Typography variant="body2" color="text.secondary" noWrap title={entry.owner ?? ""}>
        {entry.owner ?? "—"}
      </Typography>
      <Typography variant="body2" color="text.secondary" noWrap title={entry.group ?? ""}>
        {entry.group ?? "—"}
      </Typography>
      <Typography variant="body2" color="text.secondary" noWrap>
        {formatModified(entry.last_modified)}
      </Typography>
    </ListItemButton>
  );
}
