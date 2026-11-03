// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: the feed keeps pagination, filters, and create/edit modal transitions together.

import { useLanguage } from "@lite-nas/language";
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
import { GallerySearchMediaCard } from "./GallerySearchMediaCard";
import { listGalleries } from "./moments-data";
import {
  type GalleryMediaType,
  type GalleryRecord,
  galleryMediaTypeSchema,
  galleryMediaTypeSelectionSchema,
  type MatchingMedia,
} from "./schemas";

/** Feed scope selected in the Moments header. */
export type GalleryFeedScope = "home" | "mine" | "shared";

/** Labels used by the multi-select while the API keeps stable media values. */
const mediaTypeLabelKeys: Record<GalleryMediaType, string> = {
  IMAGE: "image",
  VIDEO: "video",
  AUDIO: "audio",
  LIVE_PHOTO: "livePhoto",
  OTHER: "other",
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
  const { getLabel } = useLanguage();
  const [items, setItems] = useState<GalleryRecord[]>([]);
  const [mediaMatches, setMediaMatches] = useState<MatchingMedia[]>([]);
  const [mediaTotal, setMediaTotal] = useState(0);
  const [mediaBusy, setMediaBusy] = useState(false);
  const [mediaError, setMediaError] = useState<string | null>(null);
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
  const requestVersion = useRef(0);
  const resetFeed = (): void => {
    requestVersion.current += 1;
    setPage(1);
    setPages(1);
    setItems([]);
    setMediaMatches([]);
    setMediaTotal(0);
    setMediaBusy(false);
    setMediaError(null);
  };
  useEffect(() => {
    const timer = window.setTimeout(() => {
      const value = searchDraft.trim();
      if (value !== search) {
        requestVersion.current += 1;
        setSearch(value);
        setPage(1);
        setPages(1);
        setItems([]);
        setMediaMatches([]);
        setMediaTotal(0);
        setMediaBusy(false);
        setMediaError(null);
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
        if (page === 1) {
          setMediaMatches(result.matchingMedia?.items ?? []);
          setMediaTotal(result.matchingMedia?.total ?? 0);
        }
      })
      .catch(() => {
        if (active) setError(getLabel("moments.feed.loadFailed"));
      })
      .finally(() => {
        if (active) setBusy(false);
      });
    return () => {
      active = false;
    };
  }, [page, scope, userId, search, kind, visibility, mediaTypes, location, getLabel]);
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
  const shownMedia = mediaMatches
    .map((item) => ({
      ...item,
      linkedGalleries:
        scope === "shared"
          ? item.linkedGalleries.filter((gallery) => gallery.ownerId !== userId)
          : item.linkedGalleries,
    }))
    .filter((item) => item.linkedGalleries.length > 0);
  const hasMediaSearch = Boolean(search || location.trim());
  const title = getLabel(`moments.feed.${scope}`);
  /** Loads the next image page without disturbing gallery pagination. */
  const loadMoreMedia = async (): Promise<void> => {
    const version = requestVersion.current;
    setMediaBusy(true);
    setMediaError(null);
    try {
      const result = await listGalleries(
        1,
        scope === "mine" ? userId : undefined,
        {
          q: search || undefined,
          type: kind || undefined,
          visibility: visibility || undefined,
          mediaType: mediaTypes.length > 0 ? mediaTypes : undefined,
          location: location.trim() || undefined,
        },
        mediaMatches.length,
      );
      if (version !== requestVersion.current) return;
      setMediaMatches((previous) => [...previous, ...(result.matchingMedia?.items ?? [])]);
      setMediaTotal(result.matchingMedia?.total ?? 0);
    } catch {
      if (version === requestVersion.current)
        setMediaError(getLabel("moments.feed.morePhotosFailed"));
    } finally {
      if (version === requestVersion.current) setMediaBusy(false);
    }
  };
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
          {getLabel("moments.editor.createTitle")}
        </Button>
      </Stack>
      <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" alignItems="center">
        <TextField
          label={getLabel("moments.feed.search")}
          value={searchDraft}
          onChange={(event) => setSearchDraft(event.target.value)}
          size="small"
          sx={{ minWidth: 230, flexGrow: 1 }}
        />
        <TextField
          select
          label={getLabel("moments.feed.type")}
          value={kind}
          onChange={(event) => {
            setKind(event.target.value as typeof kind);
            resetFeed();
          }}
          size="small"
          sx={{ minWidth: 145 }}
        >
          <MenuItem value="">{getLabel("moments.feed.allTypes")}</MenuItem>
          <MenuItem value="ALBUM">{getLabel("moments.feed.albums")}</MenuItem>
          <MenuItem value="COLLECTION">{getLabel("moments.feed.collections")}</MenuItem>
        </TextField>
        <TextField
          select
          label={getLabel("moments.feed.visibility")}
          value={visibility}
          onChange={(event) => {
            setVisibility(event.target.value as typeof visibility);
            resetFeed();
          }}
          size="small"
          sx={{ minWidth: 150 }}
        >
          <MenuItem value="">{getLabel("moments.feed.anyVisibility")}</MenuItem>
          <MenuItem value="PRIVATE">{getLabel("moments.feed.private")}</MenuItem>
          <MenuItem value="PUBLIC">{getLabel("moments.feed.public")}</MenuItem>
        </TextField>
        <FormControl size="small" sx={{ minWidth: 180, maxWidth: 300 }}>
          <InputLabel id="moments-media-type-label" shrink>
            {getLabel("moments.feed.mediaType")}
          </InputLabel>
          <Select
            labelId="moments-media-type-label"
            label={getLabel("moments.feed.mediaType")}
            multiple
            displayEmpty
            value={mediaTypes}
            renderValue={(selected) =>
              selected.length === 0
                ? getLabel("moments.feed.anyMedia")
                : selected
                    .map((value) =>
                      getLabel(`moments.feed.mediaLabels.${mediaTypeLabelKeys[value]}`),
                    )
                    .join(", ")
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
                <ListItemText
                  primary={getLabel(`moments.feed.mediaLabels.${mediaTypeLabelKeys[value]}`)}
                />
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField
          label={getLabel("moments.feed.locationLabel")}
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
          {getLabel("moments.feed.clearFilters")}
        </Button>
      </Stack>
      {error && <Alert severity="error">{error}</Alert>}
      <Typography variant="h6">{getLabel("moments.feed.albumsAndCollections")}</Typography>
      {shown.length === 0 && !busy && !error && page >= pages && (
        <Typography>{getLabel("moments.feed.noGalleries")}</Typography>
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
      {hasMediaSearch && (
        <Stack spacing={1.5} data-testid="moments-media-search-results">
          <Typography variant="h6">
            {getLabel("moments.feed.photos")} {scope === "shared" ? "" : `(${mediaTotal})`}
          </Typography>
          {shownMedia.length === 0 && !busy && !mediaBusy && (
            <Typography>{getLabel("moments.feed.noPhotos")}</Typography>
          )}
          <Box display="grid" gridTemplateColumns="repeat(auto-fill, minmax(210px, 1fr))" gap={2}>
            {shownMedia.map((item) => (
              <GallerySearchMediaCard key={item.path} item={item} links={item.linkedGalleries} />
            ))}
          </Box>
          {mediaError && <Alert severity="error">{mediaError}</Alert>}
          {mediaMatches.length < mediaTotal && mediaMatches.length < 100000 && (
            <Button
              disabled={mediaBusy}
              onClick={() => void loadMoreMedia()}
              sx={{ alignSelf: "start" }}
            >
              {mediaBusy
                ? getLabel("moments.feed.loadingPhotos")
                : getLabel("moments.feed.loadMore")}
            </Button>
          )}
        </Stack>
      )}
      <Box ref={sentinel} minHeight={8} />
      {typeof IntersectionObserver === "undefined" && page < pages && !busy && (
        <Button onClick={() => setPage((current) => current + 1)}>
          {getLabel("moments.feed.loadMore")}
        </Button>
      )}
      {busy && <Typography>{getLabel("moments.feed.loadingGalleries")}</Typography>}
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
