// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: album membership loading, mutation feedback, and menu state remain together.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: this menu owns its loaded choices and action states.
import { useLanguage } from "@lite-nas/language";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import Alert from "@mui/material/Alert";
import IconButton from "@mui/material/IconButton";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import type { MouseEvent, ReactElement } from "react";
import { useId, useRef, useState } from "react";
import {
  addCollectionDirectory,
  listGalleries,
  listReferences,
  removeCollectionDirectory,
} from "./moments-data";
import type { GalleryRecord } from "./schemas";

/** One owned collection and whether it already contains this album directory. */
type CollectionChoice = { collection: GalleryRecord; included: boolean };

/** Loads every owned collection so the menu is not limited to the first page. */
async function ownedCollections(userId: string): Promise<GalleryRecord[]> {
  const collections: GalleryRecord[] = [];
  let page = 1;
  let totalPages = 1;
  do {
    const result = await listGalleries(page, userId, { type: "COLLECTION" });
    collections.push(
      ...result.items.filter((item) => item.type === "COLLECTION" && item.ownerId === userId),
    );
    totalPages = result.metadata.total_pages;
    page += 1;
  } while (page <= totalPages);
  return collections;
}

/** Changes collection membership for an album's source directory. */
export const GalleryAlbumMenu = ({
  gallery,
  userId,
}: {
  gallery: GalleryRecord;
  userId: string;
}): ReactElement | null => {
  const menuId = useId();
  const { getLabel } = useLanguage();
  const loadVersion = useRef(0);
  const [anchor, setAnchor] = useState<HTMLElement | null>(null);
  const [choices, setChoices] = useState<CollectionChoice[]>([]);
  const [loading, setLoading] = useState(false);
  const [changingId, setChangingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const path = gallery.uploadDirectory;

  const close = (): void => {
    loadVersion.current += 1;
    setAnchor(null);
  };

  const open = async (event: MouseEvent<HTMLButtonElement>): Promise<void> => {
    if (!path) return;
    const version = ++loadVersion.current;
    setAnchor(event.currentTarget);
    setChoices([]);
    setError(null);
    setLoading(true);
    try {
      const collections = await ownedCollections(userId);
      const loaded = await Promise.all(
        collections.map(async (collection): Promise<CollectionChoice> => {
          const references = await listReferences(collection.id);
          return {
            collection,
            included: references.some(
              (reference) => reference.type === "DIRECTORY" && reference.path === path,
            ),
          };
        }),
      );
      if (version === loadVersion.current) setChoices(loaded);
    } catch {
      if (version === loadVersion.current) setError(getLabel("moments.albumMenu.loadFailed"));
    } finally {
      if (version === loadVersion.current) setLoading(false);
    }
  };

  const change = async (choice: CollectionChoice): Promise<void> => {
    if (!path) return;
    setChangingId(choice.collection.id);
    setError(null);
    try {
      if (choice.included) await removeCollectionDirectory(choice.collection.id, path);
      else await addCollectionDirectory(choice.collection.id, path);
      setChoices((current) =>
        current.map((item) =>
          item.collection.id === choice.collection.id
            ? { ...item, included: !choice.included }
            : item,
        ),
      );
    } catch {
      setError(getLabel("moments.albumMenu.updateFailed"));
    } finally {
      setChangingId(null);
    }
  };

  if (!path) return null;
  return (
    <>
      <IconButton
        aria-label={getLabel("moments.albumMenu.options").replace("{name}", gallery.name)}
        aria-haspopup="menu"
        aria-controls={anchor ? menuId : undefined}
        aria-expanded={anchor ? "true" : undefined}
        className="moments-card-action"
        onClick={(event) => void open(event)}
        size="small"
        sx={{
          position: "absolute",
          top: 8,
          right: gallery.ownerId === userId ? 48 : 8,
          zIndex: 1,
          opacity: anchor ? 1 : 0,
          color: "text.primary",
          bgcolor: "background.paper",
          boxShadow: 2,
          transition: "opacity 150ms ease",
          "&:hover": { bgcolor: "background.paper" },
        }}
      >
        <MoreVertIcon fontSize="small" />
      </IconButton>
      <Menu id={menuId} anchorEl={anchor} open={anchor !== null} onClose={close}>
        {loading && <MenuItem disabled>{getLabel("moments.albumMenu.loading")}</MenuItem>}
        {!loading && choices.length === 0 && !error && (
          <MenuItem disabled>{getLabel("moments.albumMenu.empty")}</MenuItem>
        )}
        {choices.map((choice) => (
          <MenuItem
            key={choice.collection.id}
            disabled={changingId !== null}
            onClick={() => void change(choice)}
          >
            {choice.included
              ? getLabel("moments.albumMenu.removeFrom")
              : getLabel("moments.albumMenu.addTo")}{" "}
            {choice.collection.name}
          </MenuItem>
        ))}
        {error && (
          <Alert severity="error" sx={{ mx: 1 }}>
            {error}
          </Alert>
        )}
      </Menu>
    </>
  );
};
