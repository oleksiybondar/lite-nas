import { fileServicePath } from "@lite-nas/file-browser";
import PhotoLibraryOutlinedIcon from "@mui/icons-material/PhotoLibraryOutlined";
import Box from "@mui/material/Box";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import { momentsApiPath } from "./api";
import { galleryItemsFromPaths } from "./gallery-model";
import { galleryFiles } from "./moments-data";
import { streamProtocol } from "./StreamContext";
import type { GalleryRecord } from "./schemas";

/** Shows the saved cover, the first accessible image, or an empty-gallery symbol. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: visibility, source resolution, and image failure share one card lifecycle.
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: the selected, uploaded, and first-image cover priorities stay in one card.
export function GalleryCover({ gallery }: { gallery: GalleryRecord }): ReactElement {
  const root = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);
  const [fallback, setFallback] = useState<string | null>(null);
  const [selectedPathFailed, setSelectedPathFailed] = useState(false);
  const [savedCoverFailed, setSavedCoverFailed] = useState(false);
  const [fallbackFailed, setFallbackFailed] = useState(false);
  const selectedUrl =
    gallery.thumbnailPath && !selectedPathFailed
      ? streamProtocol.image(fileServicePath("/f", gallery.thumbnailPath))
      : null;
  const savedUrl =
    (!gallery.thumbnailPath || selectedPathFailed) && gallery.thumbnailImageUrl && !savedCoverFailed
      ? momentsApiPath(gallery.thumbnailImageUrl)
      : null;
  useEffect(() => {
    if (!root.current) return;
    if (typeof IntersectionObserver === "undefined") {
      setVisible(true);
      return;
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true);
          observer.disconnect();
        }
      },
      { rootMargin: "300px" },
    );
    observer.observe(root.current);
    return () => observer.disconnect();
  }, []);
  useEffect(() => {
    if (!visible || selectedUrl || savedUrl) return;
    const controller = new AbortController();
    void galleryFiles(gallery, controller.signal)
      // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: first-image resolution handles gallery ordering and unmount cancellation.
      .then((paths) => {
        if (controller.signal.aborted) return;
        const first = galleryItemsFromPaths(paths.map((path) => fileServicePath("/f", path))).find(
          (item) => item.kind === "image",
        );
        setFallback(first ? streamProtocol.image(first.url) : null);
      })
      .catch(() => {
        if (!controller.signal.aborted) setFallback(null);
      });
    return () => controller.abort();
  }, [gallery, visible, selectedUrl, savedUrl]);
  const imageUrl = selectedUrl || savedUrl || (!fallbackFailed ? fallback : null);
  return (
    <Box
      ref={root}
      sx={{
        height: 150,
        bgcolor: "action.hover",
        display: "grid",
        placeItems: "center",
        overflow: "hidden",
      }}
    >
      {imageUrl ? (
        <Box
          component="img"
          src={imageUrl}
          alt=""
          loading="lazy"
          // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: one image error advances the cover priority order.
          onError={() => {
            if (selectedUrl) setSelectedPathFailed(true);
            else if (savedUrl) setSavedCoverFailed(true);
            else setFallbackFailed(true);
          }}
          sx={{ width: "100%", height: "100%", objectFit: "cover" }}
        />
      ) : (
        <PhotoLibraryOutlinedIcon color="disabled" fontSize="large" aria-hidden />
      )}
    </Box>
  );
}
