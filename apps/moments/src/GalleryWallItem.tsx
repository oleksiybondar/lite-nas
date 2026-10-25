// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: each wall item owns its focus and hover lifecycle.
// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: still and video fallback share one tile's controls.
import OpenInFullIcon from "@mui/icons-material/OpenInFull";
import Box from "@mui/material/Box";
import IconButton from "@mui/material/IconButton";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement, ReactNode } from "react";
import { useState } from "react";
import { useGallery } from "./GalleryContext";
import { GalleryItemMenu } from "./GalleryItemMenu";
import type { GalleryTile } from "./GalleryLayout";
import { GalleryPhotoTile } from "./GalleryPhotoTile";
import { GalleryVideoTile } from "./GalleryVideoTile";
import type { GalleryItem } from "./gallery-model";

/** Owns one wall item's geometry, hover/focus lifecycle, and optional action area. */
export const GalleryWallItem = ({
  item,
  tile,
  onOpen,
  hoverActions,
}: {
  item: GalleryItem;
  tile: GalleryTile;
  onOpen: () => void;
  hoverActions?: ReactNode;
}): ReactElement => {
  const [hovered, setHovered] = useState(false);
  const [focused, setFocused] = useState(false);
  const { queue } = useGallery();
  const interested = hovered || focused;
  const motionUrl =
    item.kind === "image" && queue.status(item.url) === "error" ? item.motionUrl : undefined;
  const videoItem: GalleryItem | null = motionUrl
    ? {
        ...item,
        kind: "video",
        url: motionUrl,
        filename: decodeURIComponent(motionUrl.split("/").at(-1) ?? ""),
      }
    : null;
  return (
    <Box
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocusCapture={() => setFocused(true)}
      onBlurCapture={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) setFocused(false);
      }}
      sx={{
        position: "absolute",
        left: tile.x,
        top: tile.y,
        width: tile.width,
        height: tile.height,
        borderRadius: 1,
        overflow: "hidden",
        bgcolor: "background.paper",
        "&:hover .gallery-wall-hover, &:focus-within .gallery-wall-hover": {
          opacity: 1,
          transform: "translateY(0)",
        },
      }}
    >
      {item.kind === "image" && !videoItem && (
        <GalleryPhotoTile item={item} onOpen={onOpen} interested={interested} />
      )}
      {(item.kind === "video" || videoItem) && <GalleryVideoTile item={videoItem ?? item} />}
      {item.kind === "unsupported" && (
        <Box
          component="button"
          type="button"
          aria-label={`Open ${item.filename}`}
          onClick={onOpen}
          sx={{
            width: "100%",
            height: "100%",
            p: 0,
            border: 0,
            cursor: "pointer",
            bgcolor: "background.paper",
            color: "text.primary",
            "&:focus-visible": {
              outline: "3px solid",
              outlineColor: "primary.main",
              outlineOffset: -3,
            },
          }}
        >
          <Typography variant="body2" textAlign="center" noWrap sx={{ px: 1 }}>
            Preview unavailable
          </Typography>
        </Box>
      )}
      <Stack
        className="gallery-wall-hover"
        direction="row"
        gap={0.5}
        sx={{
          position: "absolute",
          top: 4,
          right: 4,
          zIndex: 1,
          opacity: 0,
          transform: "translateY(-8px)",
          transition: "opacity 180ms ease, transform 180ms ease",
          "@media (hover: none), (prefers-reduced-motion: reduce)": {
            opacity: 1,
            transform: "none",
            transition: "none",
          },
        }}
      >
        {(item.kind === "video" || videoItem) && (
          <IconButton
            onClick={onOpen}
            size="small"
            aria-label={`Open ${item.filename} in viewer`}
            sx={{
              color: "common.white",
              bgcolor: "rgba(0,0,0,0.55)",
              "&:hover": { bgcolor: "rgba(0,0,0,0.8)" },
            }}
          >
            <OpenInFullIcon fontSize="small" />
          </IconButton>
        )}
        {hoverActions}
        <GalleryItemMenu item={item} wall />
      </Stack>
    </Box>
  );
};
