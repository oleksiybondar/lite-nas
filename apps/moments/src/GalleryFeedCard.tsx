import EditIcon from "@mui/icons-material/Edit";
import Card from "@mui/material/Card";
import CardActionArea from "@mui/material/CardActionArea";
import CardContent from "@mui/material/CardContent";
import IconButton from "@mui/material/IconButton";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { MouseEvent, ReactElement } from "react";
import { GalleryAlbumMenu } from "./GalleryAlbumMenu";
import { GalleryCover } from "./GalleryCover";
import { type GalleryRecord, galleryLocationNames } from "./schemas";

/** Opens ordinary card clicks in the SPA while keeping native modified-link behavior. */
const openGalleryLink = (
  event: MouseEvent<HTMLAnchorElement>,
  gallery: GalleryRecord,
  onOpen: (gallery: GalleryRecord) => void,
): void => {
  const modified = [event.altKey, event.ctrlKey, event.metaKey, event.shiftKey].some(Boolean);
  if (event.button !== 0 || modified) return;
  event.preventDefault();
  onOpen(gallery);
};

/** Builds a shareable link for the current standalone or embedded mount. */
export const galleryLink = (gallery: Pick<GalleryRecord, "id">): string => {
  const prefix = window.location.pathname.startsWith("/apps/moments/") ? "/apps/moments" : "";
  return `${prefix}/galleries/${gallery.id}`;
};

/** Opens a feed gallery and gives its owner a direct edit action. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: the fixed card content and owner action form one card.
export const GalleryFeedCard = ({
  gallery,
  userId,
  onOpen,
  onEdit,
}: {
  gallery: GalleryRecord;
  userId: string;
  onOpen: (gallery: GalleryRecord) => void;
  onEdit: (gallery: GalleryRecord) => void;
}): ReactElement => {
  const locations = galleryLocationNames(gallery);
  const locationLabel = locations.join(" · ");
  return (
    <Card
      data-test-class="moments-gallery-card"
      data-test-name={gallery.name}
      sx={{
        position: "relative",
        minWidth: 0,
        "&:hover .moments-card-action, &:focus-within .moments-card-action": { opacity: 1 },
        "@media (hover: none)": { "& .moments-card-action": { opacity: 1 } },
      }}
    >
      <CardActionArea
        component="a"
        href={galleryLink(gallery)}
        onClick={(event) => openGalleryLink(event, gallery, onOpen)}
        sx={{ minWidth: 0 }}
      >
        <GalleryCover gallery={gallery} />
        <CardContent
          sx={{
            height: 128,
            boxSizing: "border-box",
            overflow: "hidden",
            "&:last-child": { pb: 2 },
          }}
        >
          <Typography variant="h6" noWrap>
            {gallery.name}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {gallery.type === "ALBUM" ? "Album" : "Collection"} · {gallery.visibility.toLowerCase()}
          </Typography>
          {gallery.description && (
            <Typography variant="body2" noWrap>
              {gallery.description}
            </Typography>
          )}
          <Tooltip title={locationLabel} arrow>
            <Typography variant="body2" color="text.secondary" noWrap>
              {locationLabel}
            </Typography>
          </Tooltip>
        </CardContent>
      </CardActionArea>
      {gallery.type === "ALBUM" && <GalleryAlbumMenu gallery={gallery} userId={userId} />}
      {gallery.ownerId === userId && (
        <IconButton
          aria-label={`Edit ${gallery.type === "ALBUM" ? "album" : "collection"} ${gallery.name}`}
          className="moments-card-action"
          onClick={() => onEdit(gallery)}
          size="small"
          sx={{
            position: "absolute",
            top: 8,
            right: 8,
            zIndex: 1,
            opacity: 0,
            color: "text.primary",
            bgcolor: "background.paper",
            boxShadow: 2,
            transition: "opacity 150ms ease",
            "&:hover": { bgcolor: "background.paper" },
          }}
        >
          <EditIcon fontSize="small" />
        </IconButton>
      )}
    </Card>
  );
};
