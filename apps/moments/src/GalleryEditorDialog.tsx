import { FileSelectionDialog, FolderSelectionDialog } from "@lite-nas/file-browser";
import CloseIcon from "@mui/icons-material/Close";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import IconButton from "@mui/material/IconButton";
import MenuItem from "@mui/material/MenuItem";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import { request } from "./api";
import { type GalleryLayoutMode, galleryLayoutModeSchema } from "./GalleryLayout";
import { GalleryScanControls } from "./GalleryScanControls";
import type { GalleryItem } from "./gallery-model";
import {
  addUserTag,
  chooseGalleryThumbnailPath,
  createGallery,
  listUserTags,
  removeUserTag,
  saveGalleryThumbnail,
  shareGalleryByUsername,
  updateGallery,
} from "./moments-data";
import { type GalleryRecord, shareUsernameSchema, userTagInputSchema } from "./schemas";
import { ThumbnailMaker } from "./ThumbnailMaker";

/** Creates or edits an owned gallery and its separately stored user tags. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: one modal owns the shared create and edit fields.
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: form state and conditional album fields stay together.
export function GalleryEditorDialog({
  open,
  gallery,
  items = [],
  workingDir,
  onClose,
  onSaved,
  onCoverSaved,
}: {
  open: boolean;
  gallery?: GalleryRecord;
  items?: readonly GalleryItem[];
  workingDir: string;
  onClose: () => void;
  onSaved: (gallery: GalleryRecord) => void;
  onCoverSaved?: (gallery: GalleryRecord) => void;
}): ReactElement {
  const [kind, setKind] = useState<"ALBUM" | "COLLECTION">(gallery?.type ?? "ALBUM");
  const [name, setName] = useState(gallery?.name ?? "");
  const [description, setDescription] = useState(gallery?.description ?? "");
  const [visibility, setVisibility] = useState<"PRIVATE" | "PUBLIC">(
    gallery?.visibility ?? "PUBLIC",
  );
  const [layout, setLayout] = useState<GalleryLayoutMode>(gallery?.wallRenderStrategy ?? "mosaic");
  const [directory, setDirectory] = useState(gallery?.uploadDirectory ?? workingDir);
  const [tags, setTags] = useState("");
  const [username, setUsername] = useState("");
  const [sharing, setSharing] = useState<string[]>([]);
  const [tagsReady, setTagsReady] = useState(!gallery);
  const [createdGallery, setCreatedGallery] = useState<GalleryRecord | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [coverPickerOpen, setCoverPickerOpen] = useState(false);
  const [generatorOpen, setGeneratorOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const initializedFor = useRef<string | null>(null);
  let coverLabel = "Cover: first gallery image";
  if (gallery?.thumbnailImageUrl) coverLabel = "Cover: uploaded image";
  if (gallery?.thumbnailPath) coverLabel = `Cover file: ${gallery.thumbnailPath}`;
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: initializing create and edit drafts belongs to this modal lifecycle.
  useEffect(() => {
    if (!open) {
      initializedFor.current = null;
      return;
    }
    const key = gallery?.id ?? "create";
    if (initializedFor.current === key) return;
    initializedFor.current = key;
    setKind(gallery?.type ?? "ALBUM");
    setName(gallery?.name ?? "");
    setDescription(gallery?.description ?? "");
    setVisibility(gallery?.visibility ?? "PUBLIC");
    setLayout(gallery?.wallRenderStrategy ?? "mosaic");
    setDirectory(gallery?.uploadDirectory ?? workingDir);
    setTags("");
    setUsername("");
    setSharing([]);
    setTagsReady(!gallery);
    setCreatedGallery(null);
    setCoverPickerOpen(false);
    setGeneratorOpen(false);
    setError(null);
    if (gallery) {
      void listUserTags(gallery.id)
        .then((items) => {
          setTags(items.map((item) => item.value).join(", "));
          setTagsReady(true);
        })
        .catch(() => setError("Could not load gallery tags."));
    }
  }, [open, gallery, workingDir]);

  /** Persists a selected NAS image without discarding other modal drafts. */
  const chooseCover = async (path: string): Promise<void> => {
    if (!gallery) return;
    setBusy(true);
    try {
      const saved = await chooseGalleryThumbnailPath(gallery.id, path);
      onCoverSaved?.(saved);
      setError(null);
    } catch {
      setError("Could not use the selected image as the cover.");
    } finally {
      setBusy(false);
    }
  };

  /** Uploads a browser image or a generated image as the gallery cover. */
  const uploadCover = async (image: Blob): Promise<void> => {
    if (!gallery) return;
    setBusy(true);
    try {
      const saved = await saveGalleryThumbnail(gallery.id, image);
      onCoverSaved?.(saved);
      setGeneratorOpen(false);
      setError(null);
    } catch {
      setError("Could not upload the cover image.");
      throw new Error("Could not upload the cover image.");
    } finally {
      setBusy(false);
    }
  };

  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: metadata and user tags are separate backend routes for one save action.
  // biome-ignore lint/complexity/noExcessiveLinesPerFunction: one submission coordinates the existing metadata, tag, and share routes.
  const submit = async (): Promise<void> => {
    setBusy(true);
    setError(null);
    let metadataSaved = false;
    try {
      const desired = [
        ...new Set(
          tags
            .split(",")
            .map((value) => value.trim())
            .filter(Boolean),
        ),
      ];
      for (const value of desired) userTagInputSchema.parse({ value });
      const usernames =
        visibility === "PRIVATE" ? [...new Set([...sharing, username.trim()].filter(Boolean))] : [];
      for (const login of usernames) shareUsernameSchema.parse({ username: login });
      const saved =
        gallery || createdGallery
          ? await updateGallery((gallery ?? createdGallery)?.id ?? "", {
              name,
              description,
              visibility,
              wallRenderStrategy: layout,
            })
          : await createGallery(
              kind === "ALBUM"
                ? {
                    type: kind,
                    name,
                    description,
                    visibility,
                    wallRenderStrategy: layout,
                    uploadDirectory: directory,
                  }
                : { type: kind, name, description, visibility, wallRenderStrategy: layout },
            );
      metadataSaved = true;
      if (!gallery) setCreatedGallery(saved);
      const currentTags = await listUserTags(saved.id);
      for (const tag of currentTags) {
        if (!desired.includes(tag.value)) await removeUserTag(saved.id, tag.id);
      }
      for (const value of desired) {
        if (!currentTags.some((tag) => tag.value === value)) await addUserTag(saved.id, value);
      }
      for (const login of usernames) await shareGalleryByUsername(saved.id, login);
      onSaved(saved);
    } catch {
      setError(
        metadataSaved
          ? "Gallery details saved, but tags or sharing failed. Fix the entry and retry."
          : "Could not save gallery. Check its details and try again.",
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <Dialog open={open} onClose={busy ? undefined : onClose} fullWidth keepMounted>
        <DialogTitle sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <Box sx={{ flex: 1 }}>
            {gallery ? "Edit album or collection" : "Create album or collection"}
          </Box>
          <IconButton
            aria-label={gallery ? "Close gallery editor" : "Close gallery creator"}
            disabled={busy}
            onClick={onClose}
            edge="end"
            size="small"
          >
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            {!gallery && (
              <TextField
                select
                label="Type"
                value={kind}
                onChange={(event) => setKind(event.target.value as "ALBUM" | "COLLECTION")}
              >
                <MenuItem value="ALBUM">Album</MenuItem>
                <MenuItem value="COLLECTION">Collection</MenuItem>
              </TextField>
            )}
            <TextField
              label="Name"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
            <TextField
              label="Description"
              multiline
              minRows={3}
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
            <TextField
              select
              label="Visibility"
              value={visibility}
              onChange={(event) => setVisibility(event.target.value as "PRIVATE" | "PUBLIC")}
            >
              <MenuItem value="PRIVATE">Private</MenuItem>
              <MenuItem value="PUBLIC">Public</MenuItem>
            </TextField>
            <TextField
              select
              label="Gallery layout"
              value={layout}
              onChange={(event) => {
                const parsed = galleryLayoutModeSchema.safeParse(event.target.value);
                if (parsed.success) setLayout(parsed.data);
              }}
            >
              <MenuItem value="mosaic">Mosaic</MenuItem>
              <MenuItem value="justified">Justified rows</MenuItem>
              <MenuItem value="masonry">Masonry</MenuItem>
              <MenuItem value="grid">Even grid</MenuItem>
            </TextField>
            <TextField
              label="Tags (comma separated)"
              value={tags}
              onChange={(event) => setTags(event.target.value)}
            />
            {visibility === "PRIVATE" && (
              <>
                <Stack direction="row" spacing={1}>
                  <TextField
                    label="Share with username"
                    value={username}
                    onChange={(event) => setUsername(event.target.value)}
                    fullWidth
                  />
                  <Button
                    disabled={!username.trim()}
                    onClick={() => {
                      const login = username.trim();
                      setSharing((current) => [...new Set([...current, login])]);
                      setUsername("");
                    }}
                  >
                    Add
                  </Button>
                </Stack>
                {sharing.length > 0 && (
                  <Stack direction="row" spacing={1} flexWrap="wrap">
                    {sharing.map((login) => (
                      <Chip
                        key={login}
                        label={login}
                        onDelete={() =>
                          setSharing((current) => current.filter((item) => item !== login))
                        }
                      />
                    ))}
                  </Stack>
                )}
              </>
            )}
            {kind === "ALBUM" && !gallery && (
              <Button onClick={() => setPickerOpen(true)}>Source folder: {directory}</Button>
            )}
            {gallery && (
              <Stack spacing={1}>
                <Typography variant="body2" color="text.secondary">
                  {coverLabel}
                </Typography>
                <Button onClick={() => setCoverPickerOpen(true)} disabled={busy}>
                  Choose NAS cover file
                </Button>
                <Button component="label" disabled={busy}>
                  Upload cover
                  <input
                    hidden
                    type="file"
                    accept="image/png,image/jpeg"
                    onChange={(event) => {
                      const file = event.target.files?.[0];
                      if (file) void uploadCover(file).catch(() => undefined);
                      event.target.value = "";
                    }}
                  />
                </Button>
                <Button onClick={() => setGeneratorOpen(true)} disabled={busy}>
                  Generate cover
                </Button>
              </Stack>
            )}
            {gallery && <GalleryScanControls key={gallery.id} galleryId={gallery.id} open={open} />}
            {error && <Alert severity="error">{error}</Alert>}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={onClose}>
            Cancel
          </Button>
          <Button disabled={busy || !tagsReady || !name.trim()} onClick={() => void submit()}>
            {gallery ? "Save" : "Create"}
          </Button>
        </DialogActions>
      </Dialog>
      <FolderSelectionDialog
        open={pickerOpen}
        initialDirectory={directory}
        onClose={() => setPickerOpen(false)}
        onSelect={setDirectory}
        request={request}
      />
      {gallery && (
        <>
          <FileSelectionDialog
            open={coverPickerOpen}
            initialDirectory={workingDir}
            extensionFilter={{
              include: ["jpg", "jpeg", "png", "gif", "webp", "avif", "heic", "heif"],
            }}
            onClose={() => setCoverPickerOpen(false)}
            onSelect={(path) => {
              setCoverPickerOpen(false);
              void chooseCover(path);
            }}
            request={request}
          />
          <Dialog
            open={generatorOpen}
            onClose={() => setGeneratorOpen(false)}
            fullWidth
            maxWidth="xl"
          >
            <DialogTitle sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Box sx={{ flex: 1 }}>Generate cover</Box>
              <IconButton
                aria-label="Close thumbnail generator"
                onClick={() => setGeneratorOpen(false)}
                edge="end"
                size="small"
              >
                <CloseIcon />
              </IconButton>
            </DialogTitle>
            <DialogContent>
              {generatorOpen && <ThumbnailMaker items={items} onSave={uploadCover} />}
            </DialogContent>
          </Dialog>
        </>
      )}
    </>
  );
}
