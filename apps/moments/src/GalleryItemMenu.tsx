// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: this one item menu combines target collections and owner-only removal.

import { ConfirmActionDialog, StandardDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";
import { useId, useState } from "react";
import { useGallery } from "./GalleryContext";
import type { GalleryItem } from "./gallery-model";

/** Gives wall and carousel items the same collection membership actions. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: menu and creation dialog share the selected media item.
export const GalleryItemMenu = ({
  item,
  wall = false,
}: {
  item: GalleryItem;
  wall?: boolean;
}): ReactElement => {
  const menuId = useId();
  const { getLabel } = useLanguage();
  const [anchor, setAnchor] = useState<HTMLElement | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [removeOpen, setRemoveOpen] = useState(false);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const { collectionActions } = useGallery();
  const path = decodeURIComponent(item.url.slice(2));
  const run = async (action: () => Promise<void>): Promise<boolean> => {
    setAnchor(null);
    setError(null);
    try {
      await action();
      return true;
    } catch {
      setError(getLabel("moments.itemMenu.updateFailed"));
      return false;
    }
  };
  const open = anchor !== null;
  return (
    <>
      <IconButton
        size={wall ? "small" : "medium"}
        color="inherit"
        aria-label={getLabel("moments.itemMenu.moreOptions").replace("{filename}", item.filename)}
        aria-haspopup="menu"
        aria-controls={open ? menuId : undefined}
        aria-expanded={open ? "true" : undefined}
        onClick={(event) => setAnchor(event.currentTarget)}
        sx={
          wall
            ? {
                color: "common.white",
                bgcolor: "rgba(0,0,0,0.55)",
                "&:hover": { bgcolor: "rgba(0,0,0,0.8)" },
              }
            : undefined
        }
      >
        <MoreVertIcon fontSize={wall ? "small" : "medium"} />
      </IconButton>
      <Menu
        id={menuId}
        anchorEl={anchor}
        open={open}
        onClose={() => setAnchor(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
        transformOrigin={{ vertical: "top", horizontal: "right" }}
      >
        {collectionActions?.collections.map((collection) => (
          <MenuItem
            key={collection.id}
            onClick={() => void run(() => collectionActions.add(collection.id, path))}
          >
            {getLabel("moments.itemMenu.addTo").replace("{collection}", collection.name)}
          </MenuItem>
        ))}
        <MenuItem
          disabled={!collectionActions}
          onClick={() => {
            setAnchor(null);
            setCreateOpen(true);
          }}
        >
          {getLabel("moments.itemMenu.addToNew")}
        </MenuItem>
        {collectionActions?.canRemove(path) && (
          <MenuItem
            onClick={() => {
              setAnchor(null);
              setRemoveOpen(true);
            }}
          >
            {getLabel("moments.itemMenu.removeFrom")}
          </MenuItem>
        )}
      </Menu>
      <ConfirmActionDialog
        open={removeOpen}
        title={getLabel("moments.itemMenu.removeFrom")}
        description={getLabel("moments.itemMenu.confirmRemove").replace(
          "{filename}",
          item.filename,
        )}
        confirmLabel={getLabel("moments.itemMenu.removeFrom")}
        onClose={() => setRemoveOpen(false)}
        onConfirm={async () => {
          if (!collectionActions || !(await run(() => collectionActions.remove(path))))
            throw new Error(getLabel("moments.itemMenu.updateFailed"));
        }}
      />
      <StandardDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title={getLabel("moments.itemMenu.newTitle")}
        closeLabel={getLabel("moments.itemMenu.close")}
        actions={
          <>
            <Button onClick={() => setCreateOpen(false)}>
              {getLabel("moments.itemMenu.cancel")}
            </Button>
            <Button
              disabled={!name.trim()}
              onClick={() => {
                if (!collectionActions) return;
                void run(() => collectionActions.createAndAdd(name, path)).then((success) => {
                  if (success) {
                    setCreateOpen(false);
                    setName("");
                  }
                });
              }}
            >
              {getLabel("moments.itemMenu.createAndAdd")}
            </Button>
          </>
        }
      >
        <TextField
          label={getLabel("moments.itemMenu.name")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          inputProps={{ maxLength: 160 }}
          sx={{ mt: 1 }}
        />
        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}
      </StandardDialog>
      {error && !createOpen && <span role="alert">{error}</span>}
    </>
  );
};
