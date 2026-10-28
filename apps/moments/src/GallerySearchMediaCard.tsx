import { fileServicePath } from "@lite-nas/file-browser";
import BrokenImageOutlinedIcon from "@mui/icons-material/BrokenImageOutlined";
import Box from "@mui/material/Box";
import Card from "@mui/material/Card";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useState } from "react";
import { galleryLink } from "./GalleryFeedCard";
import { streamProtocol } from "./StreamContext";
import type { MatchingMedia } from "./schemas";

/** Shows one matching photo and reveals its accessible gallery links on hover or focus. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: preview and gallery links form one accessible tile.
export function GallerySearchMediaCard({
  item,
  links,
}: {
  item: MatchingMedia;
  links: MatchingMedia["linkedGalleries"];
}): ReactElement {
  const [failed, setFailed] = useState(false);
  const filename = item.path.slice(item.path.lastIndexOf("/") + 1);
  const preview = streamProtocol.image(fileServicePath("/f", item.path));
  return (
    <Card
      data-testid="moments-media-search-card"
      sx={{
        position: "relative",
        height: 240,
        overflow: "hidden",
        bgcolor: "action.hover",
        "&:hover .moments-media-links, &:focus-within .moments-media-links": {
          opacity: 1,
          pointerEvents: "auto",
        },
        "@media (hover: none)": {
          "& .moments-media-links": { opacity: 1, pointerEvents: "auto" },
        },
      }}
    >
      {preview && !failed ? (
        <Box
          component="img"
          src={preview}
          alt={item.caption || filename}
          loading="lazy"
          onError={() => setFailed(true)}
          sx={{ width: "100%", height: "100%", objectFit: "cover" }}
        />
      ) : (
        <Box display="grid" height="100%" sx={{ placeItems: "center" }}>
          <BrokenImageOutlinedIcon color="disabled" fontSize="large" aria-hidden />
        </Box>
      )}
      <Stack
        className="moments-media-links"
        spacing={0.5}
        sx={{
          position: "absolute",
          inset: 0,
          justifyContent: "end",
          p: 1.5,
          color: "common.white",
          bgcolor: "rgba(0, 0, 0, 0.68)",
          opacity: 0,
          pointerEvents: "none",
          transition: "opacity 150ms ease",
          overflow: "auto",
        }}
      >
        <Typography variant="body2" fontWeight="bold">
          {item.caption || filename}
        </Typography>
        <Typography variant="caption">Found in</Typography>
        {links.map((gallery) => (
          <Box
            key={gallery.id}
            component="a"
            href={galleryLink(gallery)}
            sx={{ color: "common.white", textDecoration: "underline" }}
          >
            {gallery.name} · {gallery.type === "ALBUM" ? "Album" : "Collection"}
          </Box>
        ))}
      </Stack>
    </Card>
  );
}
