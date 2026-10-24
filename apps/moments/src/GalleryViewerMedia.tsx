// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: media lifecycle events and selected-media rendering are coupled.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: media lifecycle events and selected-media rendering are coupled.
// biome-ignore-all lint/a11y/useMediaCaption: source media has no caption track supplied by the gallery contract.
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement, SyntheticEvent } from "react";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { useCarousel } from "./CarouselContext";
import { useGallery } from "./GalleryContext";
import { attemptVideoAutoplay } from "./GalleryVideoPlayback";

/** Displays only the selected original and handles its image or video lifecycle. */
export const GalleryViewerMedia = (): ReactElement | null => {
  const { items, queue } = useGallery();
  useSyncExternalStore(queue.subscribe, queue.getSnapshot);
  const {
    current,
    selected,
    playing,
    setPlaying,
    videoAutoplay,
    videoError,
    setVideoError,
    navigate,
  } = useCarousel();
  const videoRef = useRef<HTMLVideoElement>(null);
  const [autoplayBlocked, setAutoplayBlocked] = useState(false);
  const [buffering, setBuffering] = useState(false);
  const activeVideoId = current?.kind === "video" ? current.id : null;
  useEffect(() => {
    const video = videoRef.current;
    return () => video?.pause();
  }, []);
  useEffect(() => {
    setAutoplayBlocked(false);
    setBuffering(false);
    const video = videoRef.current;
    if (!activeVideoId || !video || videoError) return;
    let active = true;
    void attemptVideoAutoplay(video, videoAutoplay).then((result) => {
      if (!active) return;
      if (result === "blocked") setAutoplayBlocked(true);
      if (result === "failed") setVideoError(true);
    });
    return () => {
      active = false;
    };
  }, [activeVideoId, setVideoError, videoAutoplay, videoError]);
  if (!current) return null;
  const status = current.kind === "image" ? queue.status(current.url) : null;
  const attempt = current.kind === "image" ? queue.attempt(current.url) : undefined;
  /** Advances after completed video playback only when slideshow is active. */
  const onVideoEnded = (event: SyntheticEvent<HTMLVideoElement>): void => {
    if (event.currentTarget !== videoRef.current) return;
    if (playing && selected !== null && selected < items.length - 1) navigate(1);
    else if (playing) setPlaying(false);
  };
  /** Advances past a failed slideshow video or exposes a manual file fallback. */
  const onVideoError = (event: SyntheticEvent<HTMLVideoElement>): void => {
    if (event.currentTarget !== videoRef.current) return;
    setBuffering(false);
    setVideoError(true);
  };
  return (
    <Stack
      alignItems="center"
      justifyContent="center"
      sx={{ flex: 1, minWidth: 0, height: "100%" }}
    >
      {current.kind === "image" && (status === "ready" || status === "loading") && (
        <Box
          component="img"
          src={current.url}
          alt={current.filename}
          onLoad={() => queue.complete(current.url, true, attempt)}
          onError={() => queue.fail(current.url, attempt)}
          sx={{
            maxWidth: "100%",
            maxHeight: "100%",
            objectFit: "contain",
            opacity: status === "ready" ? 1 : 0,
          }}
        />
      )}
      {current.kind === "image" && (status === "loading" || status === "queued") && (
        <CircularProgress />
      )}
      {current.kind === "video" && !videoError && (
        <Box
          sx={{
            position: "relative",
            maxWidth: "100%",
            maxHeight: "100%",
            display: "flex",
            alignItems: "center",
          }}
        >
          <video
            ref={videoRef}
            key={current.id}
            src={current.url}
            controls
            playsInline
            preload="metadata"
            onError={onVideoError}
            onEnded={onVideoEnded}
            onWaiting={(event) => {
              if (event.currentTarget === videoRef.current) setBuffering(true);
            }}
            onPlaying={(event) => {
              if (event.currentTarget !== videoRef.current) return;
              setBuffering(false);
              setAutoplayBlocked(false);
            }}
            onCanPlay={(event) => {
              if (event.currentTarget === videoRef.current) setBuffering(false);
            }}
            style={{ maxWidth: "100%", maxHeight: "100%" }}
          />
          {buffering && (
            <CircularProgress
              aria-label="Buffering video"
              sx={{
                position: "absolute",
                left: "50%",
                top: "50%",
                transform: "translate(-50%, -50%)",
                pointerEvents: "none",
              }}
            />
          )}
        </Box>
      )}
      {current.kind === "video" && autoplayBlocked && !videoError && (
        <Typography variant="body2" textAlign="center" role="status">
          Autoplay was blocked. Press play in the video controls.
        </Typography>
      )}
      {current.kind === "image" && status === "error" && (
        <Stack alignItems="center" gap={2}>
          <Typography textAlign="center">Could not load this image.</Typography>
          <Button onClick={() => queue.retry(current.url)}>Retry image</Button>
        </Stack>
      )}
      {(current.kind === "unsupported" || videoError) && (
        <Typography textAlign="center">This file cannot be previewed in this browser.</Typography>
      )}
    </Stack>
  );
};
