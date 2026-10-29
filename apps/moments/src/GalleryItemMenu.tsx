// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: this one item menu combines target collections and owner-only removal.
import { useLanguage } from "@lite-nas/language";
import CloseIcon from "@mui/icons-material/Close";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
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
          <MenuItem onClick={() => void run(() => collectionActions.remove(path))}>
            {getLabel("moments.itemMenu.removeFrom")}
          </MenuItem>
        )}
      </Menu>
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)}>
        <DialogTitle sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <Box sx={{ flex: 1 }}>{getLabel("moments.itemMenu.newTitle")}</Box>
          <IconButton
            aria-label={getLabel("moments.itemMenu.close")}
            onClick={() => setCreateOpen(false)}
            edge="end"
            size="small"
          >
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent>
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
        </DialogContent>
        <DialogActions>
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
        </DialogActions>
      </Dialog>
      {error && !createOpen && <span role="alert">{error}</span>}
    </>
  );
};
