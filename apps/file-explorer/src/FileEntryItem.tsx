import { ListItemButton, ListItemIcon, ListItemText } from "@mui/material";
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
      <ListItemIcon>
        <FileEntryIcon kind={entry.kind} name={name} />
      </ListItemIcon>
      <ListItemText primary={name} secondary={typeLabel(entry)} />
    </ListItemButton>
  );
}
