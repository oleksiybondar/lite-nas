import {
  FileSelectionDialog,
  FolderSelectionDialog,
  fileServicePath,
} from "@lite-nas/file-browser";
import { UploadDialog } from "@lite-nas/file-upload";
import EditIcon from "@mui/icons-material/Edit";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";
import { request } from "./api";
import { Gallery } from "./Gallery";
import { GalleryEditorDialog } from "./GalleryEditorDialog";
import { galleryItemsFromPaths } from "./gallery-model";
import {
  addCollectionDirectory,
  changeCollectionFile,
  createGallery,
  galleryFiles,
  getGallery,
  listGalleries,
  listReferences,
  scanUploadedMedia,
} from "./moments-data";
import { type GalleryRecord, galleryLocationNames } from "./schemas";

/** Loads authorized originals for an album or explicit collection membership. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: gallery source loading and collection actions form one detail screen.
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: the owner edit control belongs beside the detail heading.
export function GalleryDetail({
  selected,
  userId,
  maxConcurrent,
  workingDir,
  onBack,
}: {
  selected: GalleryRecord;
  userId: string;
  maxConcurrent: number;
  workingDir: string;
  onBack: () => void;
}): ReactElement {
  const [gallery, setGallery] = useState(selected);
  const [paths, setPaths] = useState<string[]>([]);
  const [memberPaths, setMemberPaths] = useState<string[]>([]);
  const [collections, setCollections] = useState<GalleryRecord[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [busy, setBusy] = useState(true);
  const [editing, setEditing] = useState(false);
  const [picker, setPicker] = useState<"file" | "directory" | null>(null);
  const [uploadOpen, setUploadOpen] = useState(false);
  useEffect(() => {
    const controller = new AbortController();
    setBusy(true);
    setError(null);
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: two source kinds and paged owned-collection lookup share one lifecycle.
    void (async () => {
      const current = await getGallery(selected.id);
      const [sourcePaths, references] = await Promise.all([
        galleryFiles(current, controller.signal),
        current.type === "COLLECTION" ? listReferences(current.id) : Promise.resolve([]),
      ]);
      const owned: GalleryRecord[] = [];
      let page = 1;
      let pages = 1;
      do {
        const result = await listGalleries(page, userId);
        owned.push(...result.items.filter((item) => item.type === "COLLECTION"));
        pages = result.metadata.total_pages;
        page += 1;
      } while (page <= pages);
      if (controller.signal.aborted) return;
      setGallery(current);
      setPaths(sourcePaths);
      setMemberPaths(
        references
          .filter((reference) => reference.type === "FILE")
          .map((reference) => reference.path),
      );
      setCollections(owned);
    })()
      .catch(() => {
        if (!controller.signal.aborted) setError("Could not load gallery media.");
      })
      .finally(() => {
        if (!controller.signal.aborted) setBusy(false);
      });
    return () => controller.abort();
  }, [selected.id, userId]);

  const actions = {
    collections,
    canRemove: (path: string): boolean =>
      gallery.type === "COLLECTION" && gallery.ownerId === userId && memberPaths.includes(path),
    add: async (id: string, path: string): Promise<void> => {
      await changeCollectionFile(id, path, true);
      if (id === gallery.id) setMemberPaths((current) => [...new Set([...current, path])]);
    },
    remove: async (path: string): Promise<void> => {
      await changeCollectionFile(gallery.id, path, false);
      setMemberPaths((current) => current.filter((member) => member !== path));
      setPaths((current) => current.filter((member) => member !== path));
    },
    createAndAdd: async (name: string, path: string): Promise<void> => {
      const created = await createGallery({ type: "COLLECTION", name });
      await changeCollectionFile(created.id, path, true);
      setCollections((current) => [...current, created]);
    },
  };
  const items = galleryItemsFromPaths(paths.map((path) => fileServicePath("/f", path)));
  const owned = gallery.ownerId === userId;
  const ownedCollection = gallery.type === "COLLECTION" && owned;
  const albumDirectory = gallery.type === "ALBUM" && owned ? gallery.uploadDirectory : null;
  const locations = galleryLocationNames(gallery);
  const locationLabel = locations.join(" · ");
  /** Scans only the uploaded path and refreshes the album listing. */
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: upload, metadata scan, and refresh have distinct failure messages.
  const onUploaded = async (file: File): Promise<void> => {
    if (!albumDirectory) return;
    try {
      await scanUploadedMedia(gallery.id, `${albumDirectory.replace(/\/$/, "")}/${file.name}`);
    } catch {
      setUploadError(
        `Could not scan metadata for ${file.name}. The upload completed; retry a media scan.`,
      );
      return;
    }
    try {
      const [updated, nextPaths] = await Promise.all([
        getGallery(gallery.id),
        galleryFiles(gallery, new AbortController().signal),
      ]);
      setGallery(updated);
      setPaths(nextPaths);
      setUploadError(null);
    } catch {
      setUploadError(`The upload and metadata scan completed, but the gallery could not refresh.`);
    }
  };
  /** Adds a selected source and refreshes this collection's resolved media. */
  const addSource = async (kind: "file" | "directory", path: string): Promise<void> => {
    try {
      if (kind === "file") await changeCollectionFile(gallery.id, path, true);
      else await addCollectionDirectory(gallery.id, path);
      const [nextPaths, references] = await Promise.all([
        galleryFiles(gallery, new AbortController().signal),
        listReferences(gallery.id),
      ]);
      setPaths(nextPaths);
      setMemberPaths(
        references
          .filter((reference) => reference.type === "FILE")
          .map((reference) => reference.path),
      );
      setError(null);
    } catch {
      setError("Could not add the selected source to this collection.");
    }
  };
  return (
    <Stack spacing={2} data-testid="moments-gallery-detail">
      <Button onClick={onBack} sx={{ alignSelf: "start" }}>
        Back to galleries
      </Button>
      <Stack direction="row" alignItems="center" justifyContent="space-between">
        <Typography variant="h4">{gallery.name}</Typography>
        {owned && (
          <Button startIcon={<EditIcon />} onClick={() => setEditing(true)}>
            Edit
          </Button>
        )}
      </Stack>
      {gallery.description && <Typography>{gallery.description}</Typography>}
      {locationLabel && (
        <Tooltip title={locationLabel} arrow>
          <Typography color="text.secondary" noWrap sx={{ maxWidth: "100%" }}>
            Locations: {locationLabel}
          </Typography>
        </Tooltip>
      )}
      {owned && (
        <Stack direction="row" spacing={1} flexWrap="wrap">
          {ownedCollection && <Button onClick={() => setPicker("file")}>Add file</Button>}
          {ownedCollection && <Button onClick={() => setPicker("directory")}>Add folder</Button>}
          {albumDirectory && <Button onClick={() => setUploadOpen(true)}>Upload media</Button>}
        </Stack>
      )}
      {error && <Alert severity="error">{error}</Alert>}
      {uploadError && <Alert severity="error">{uploadError}</Alert>}
      {busy && <Typography>Loading media…</Typography>}
      {!busy && !error && items.length === 0 && <Typography>No media in this gallery.</Typography>}
      {!busy && !error && items.length > 0 && (
        <Gallery
          items={items}
          layout={gallery.wallRenderStrategy}
          collectionActions={actions}
          maxConcurrent={maxConcurrent}
        />
      )}
      {owned && (
        <GalleryEditorDialog
          open={editing}
          gallery={gallery}
          items={items}
          workingDir={workingDir}
          onClose={() => setEditing(false)}
          onCoverSaved={setGallery}
          onSaved={(updated) => {
            setGallery(updated);
            setEditing(false);
          }}
        />
      )}
      {ownedCollection && (
        <>
          <FileSelectionDialog
            open={picker === "file"}
            initialDirectory={workingDir}
            onClose={() => setPicker(null)}
            onSelect={(path) => void addSource("file", path)}
            request={request}
          />
          <FolderSelectionDialog
            open={picker === "directory"}
            initialDirectory={workingDir}
            onClose={() => setPicker(null)}
            onSelect={(path) => void addSource("directory", path)}
            request={request}
          />
        </>
      )}
      {albumDirectory && (
        <UploadDialog
          open={uploadOpen}
          directory={albumDirectory}
          onClose={() => setUploadOpen(false)}
          onUploaded={onUploaded}
        />
      )}
    </Stack>
  );
}
