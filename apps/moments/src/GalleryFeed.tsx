// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: the feed keeps pagination, filters, and create/edit modal transitions together.
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import ListItemText from "@mui/material/ListItemText";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import { GalleryEditorDialog } from "./GalleryEditorDialog";
import { GalleryFeedCard } from "./GalleryFeedCard";
import { listGalleries } from "./moments-data";
import {
  type GalleryMediaType,
  type GalleryRecord,
  galleryMediaTypeSchema,
  galleryMediaTypeSelectionSchema,
} from "./schemas";

/** Feed scope selected in the Moments header. */
export type GalleryFeedScope = "home" | "mine" | "shared";

/** Labels used by the multi-select while the API keeps stable media values. */
const mediaTypeLabels: Record<GalleryMediaType, string> = {
  IMAGE: "Photos",
  VIDEO: "Videos",
  AUDIO: "Audio",
  LIVE_PHOTO: "Live Photos",
  OTHER: "Other",
};

/** Lists real, paged gallery metadata and opens an album or collection. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: feed cards, pagination, and create entry point form one screen.
export function GalleryFeed({
  scope,
  userId,
  workingDir,
  onOpen,
}: {
  scope: GalleryFeedScope;
  userId: string;
  workingDir: string;
  onOpen: (gallery: GalleryRecord) => void;
}): ReactElement {
  const [items, setItems] = useState<GalleryRecord[]>([]);
  const [page, setPage] = useState(1);
  const [pages, setPages] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<GalleryRecord | null>(null);
  const [editorOpen, setEditorOpen] = useState(false);
  const [searchDraft, setSearchDraft] = useState("");
  const [search, setSearch] = useState("");
  const [kind, setKind] = useState<"" | "ALBUM" | "COLLECTION">("");
  const [visibility, setVisibility] = useState<"" | "PRIVATE" | "PUBLIC">("");
  const [mediaTypes, setMediaTypes] = useState<GalleryMediaType[]>([]);
  const [location, setLocation] = useState("");
  const sentinel = useRef<HTMLDivElement>(null);
  const resetFeed = (): void => {
    setPage(1);
    setPages(1);
    setItems([]);
  };
  useEffect(() => {
    const timer = window.setTimeout(() => {
      const value = searchDraft.trim();
      if (value !== search) {
        setSearch(value);
        setPage(1);
        setPages(1);
        setItems([]);
      }
    }, 250);
    return () => window.clearTimeout(timer);
  }, [searchDraft, search]);
  useEffect(() => {
    let active = true;
    setBusy(true);
    setError(null);
    void listGalleries(page, scope === "mine" ? userId : undefined, {
      q: search || undefined,
      type: kind || undefined,
      visibility: visibility || undefined,
      mediaType: mediaTypes.length > 0 ? mediaTypes : undefined,
      location: location.trim() || undefined,
    })
      .then((result) => {
        if (!active) return;
        setPages(result.metadata.total_pages);
        setItems((previous) => (page === 1 ? result.items : [...previous, ...result.items]));
      })
      .catch(() => {
        if (active) setError("Could not load galleries.");
      })
      .finally(() => {
        if (active) setBusy(false);
      });
    return () => {
      active = false;
    };
  }, [page, scope, userId, search, kind, visibility, mediaTypes, location]);
  useEffect(() => {
    if (busy || page >= pages || !sentinel.current || typeof IntersectionObserver === "undefined")
      return;
    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) setPage(page + 1);
    });
    observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [busy, page, pages]);
  const shown = scope === "shared" ? items.filter((item) => item.ownerId !== userId) : items;
  const title = { home: "Home", mine: "My collections", shared: "Shared collections" }[scope];
  return (
    <Stack spacing={2} data-testid="moments-gallery-feed">
      <Stack direction="row" justifyContent="space-between" alignItems="center">
        <Typography variant="h5">{title}</Typography>
        <Button
          onClick={() => {
            setEditing(null);
            setCreating(true);
          }}
        >
          Create album or collection
        </Button>
      </Stack>
      <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" alignItems="center">
        <TextField
          label="Search galleries"
          value={searchDraft}
          onChange={(event) => setSearchDraft(event.target.value)}
          size="small"
          sx={{ minWidth: 230, flexGrow: 1 }}
        />
        <TextField
          select
          label="Type"
          value={kind}
          onChange={(event) => {
            setKind(event.target.value as typeof kind);
            resetFeed();
          }}
          size="small"
          sx={{ minWidth: 145 }}
        >
          <MenuItem value="">All types</MenuItem>
          <MenuItem value="ALBUM">Albums</MenuItem>
          <MenuItem value="COLLECTION">Collections</MenuItem>
        </TextField>
        <TextField
          select
          label="Visibility"
          value={visibility}
          onChange={(event) => {
            setVisibility(event.target.value as typeof visibility);
            resetFeed();
          }}
          size="small"
          sx={{ minWidth: 150 }}
        >
          <MenuItem value="">Any visibility</MenuItem>
          <MenuItem value="PRIVATE">Private</MenuItem>
          <MenuItem value="PUBLIC">Public</MenuItem>
        </TextField>
        <FormControl size="small" sx={{ minWidth: 180, maxWidth: 300 }}>
          <InputLabel id="moments-media-type-label" shrink>
            Media type
          </InputLabel>
          <Select
            labelId="moments-media-type-label"
            label="Media type"
            multiple
            displayEmpty
            value={mediaTypes}
            renderValue={(selected) =>
              selected.length === 0
                ? "Any media"
                : selected.map((value) => mediaTypeLabels[value]).join(", ")
            }
            onChange={(event) => {
              const parsed = galleryMediaTypeSelectionSchema.safeParse(event.target.value);
              if (!parsed.success) return;
              setMediaTypes(parsed.data);
              resetFeed();
            }}
          >
            {galleryMediaTypeSchema.options.map((value) => (
              <MenuItem key={value} value={value}>
                <Checkbox checked={mediaTypes.includes(value)} size="small" />
                <ListItemText primary={mediaTypeLabels[value]} />
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField
          label="Location"
          value={location}
          onChange={(event) => {
            setLocation(event.target.value);
            resetFeed();
          }}
          size="small"
          sx={{ minWidth: 150 }}
        />
        <Button
          onClick={() => {
            setSearchDraft("");
            setSearch("");
            setKind("");
            setVisibility("");
            setMediaTypes([]);
            setLocation("");
            resetFeed();
          }}
        >
          Clear filters
        </Button>
      </Stack>
      {error && <Alert severity="error">{error}</Alert>}
      {shown.length === 0 && !busy && !error && page >= pages && (
        <Typography>No galleries yet.</Typography>
      )}
      <Box display="grid" gridTemplateColumns="repeat(auto-fill, minmax(230px, 1fr))" gap={2}>
        {shown.map((gallery) => (
          <GalleryFeedCard
            key={gallery.id}
            gallery={gallery}
            userId={userId}
            onOpen={onOpen}
            onEdit={(gallery) => {
              setEditing(gallery);
              setEditorOpen(true);
            }}
          />
        ))}
      </Box>
      <Box ref={sentinel} minHeight={8} />
      {typeof IntersectionObserver === "undefined" && page < pages && !busy && (
        <Button onClick={() => setPage((current) => current + 1)}>Load more</Button>
      )}
      {busy && <Typography>Loading galleries…</Typography>}
      <GalleryEditorDialog
        open={creating || editorOpen}
        gallery={editing ?? undefined}
        workingDir={workingDir}
        onClose={() => {
          setCreating(false);
          setEditorOpen(false);
        }}
        onCoverSaved={(saved) => {
          setEditing(saved);
          setItems((current) => current.map((item) => (item.id === saved.id ? saved : item)));
        }}
        onSaved={(gallery) => {
          if (editing) {
            setItems((current) => current.map((item) => (item.id === gallery.id ? gallery : item)));
            setEditing(gallery);
            setEditorOpen(false);
          } else {
            setCreating(false);
            onOpen(gallery);
          }
        }}
      />
    </Stack>
  );
}
