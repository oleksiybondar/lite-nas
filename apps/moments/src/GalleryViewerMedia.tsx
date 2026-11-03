// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: media lifecycle events and selected-media rendering are coupled.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: media lifecycle events and selected-media rendering are coupled.
// biome-ignore-all lint/a11y/useMediaCaption: source media has no caption track supplied by the gallery contract.

import { useLanguage } from "@lite-nas/language";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import IconButton from "@mui/material/IconButton";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement, SyntheticEvent } from "react";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { useCarousel } from "./CarouselContext";
import { useGallery } from "./GalleryContext";
import { attemptVideoAutoplay } from "./GalleryVideoPlayback";
import { useStream } from "./StreamContext";

/** Displays only the selected original and handles its image or video lifecycle. */
export const GalleryViewerMedia = (): ReactElement | null => {
  const { getLabel } = useLanguage();
  const { items, queue } = useGallery();
  const stream = useStream();
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
  const [videoMode, setVideoMode] = useState<"original" | "transcode">("original");
  const [transcodeStart, setTranscodeStart] = useState(0);
  const [playingLivePhotoId, setPlayingLivePhotoId] = useState<string | null>(null);
  const [motionErrorId, setMotionErrorId] = useState<string | null>(null);
  const showingMotion = current?.kind === "image" && current.id === playingLivePhotoId;
  const fallbackMotion =
    current?.kind === "image" && !!current.motionUrl && queue.status(current.url) === "error";
  const viewingMotion = showingMotion || fallbackMotion;
  const activeVideoId = current?.kind === "video" || viewingMotion ? current?.id : null;
  const currentUrl = viewingMotion ? current?.motionUrl : current?.url;
  const shouldAutoplay = showingMotion || (current?.kind === "video" && videoAutoplay);
  useEffect(() => {
    const video = videoRef.current;
    return () => video?.pause();
  }, []);
  useEffect(() => {
    if (activeVideoId === null) return;
    setVideoMode("original");
    setTranscodeStart(0);
  }, [activeVideoId]);
  useEffect(() => {
    setAutoplayBlocked(false);
    setBuffering(false);
    const video = videoRef.current;
    if (!activeVideoId || !video || videoError) return;
    if (videoMode === "transcode" && currentUrl) {
      const url = stream.transcode(currentUrl, transcodeStart);
      if (url) video.src = url;
    }
    let active = true;
    void attemptVideoAutoplay(video, shouldAutoplay).then((result) => {
      if (!active) return;
      if (result === "blocked") setAutoplayBlocked(true);
      if (result === "failed") {
        if (videoMode === "original") setVideoMode("transcode");
        else if (showingMotion) {
          setMotionErrorId(activeVideoId);
          setPlayingLivePhotoId(null);
        } else setVideoError(true);
      }
    });
    return () => {
      active = false;
    };
  }, [
    activeVideoId,
    currentUrl,
    setVideoError,
    stream,
    shouldAutoplay,
    showingMotion,
    videoError,
    videoMode,
    transcodeStart,
  ]);
  if (!current) return null;
  const status = current.kind === "image" ? queue.status(current.url) : null;
  const attempt = current.kind === "image" ? queue.attempt(current.url) : undefined;
  let videoSource: string | null = null;
  if (current.kind === "video" || viewingMotion)
    videoSource =
      videoMode === "original"
        ? stream.original(currentUrl ?? "")
        : stream.transcode(currentUrl ?? "", transcodeStart);
  /** Advances after completed video playback only when slideshow is active. */
  const onVideoEnded = (event: SyntheticEvent<HTMLVideoElement>): void => {
    if (event.currentTarget !== videoRef.current) return;
    if (showingMotion) {
      setPlayingLivePhotoId(null);
      return;
    }
    if (playing && selected !== null && selected < items.length - 1) navigate(1);
    else if (playing) setPlaying(false);
  };
  /** Advances past a failed slideshow video or exposes a manual file fallback. */
  const onVideoError = (event: SyntheticEvent<HTMLVideoElement>): void => {
    if (event.currentTarget !== videoRef.current) return;
    setBuffering(false);
    if (videoMode === "original") {
      setVideoMode("transcode");
      return;
    }
    if (showingMotion) {
      setMotionErrorId(current.id);
      setPlayingLivePhotoId(null);
    } else setVideoError(true);
  };
  /** Restarts converted playback at a validated source-video time. */
  const seekTranscode = (delta: number): void => {
    const elapsed = Math.floor(videoRef.current?.currentTime ?? 0);
    setTranscodeStart(Math.min(86400, Math.max(0, transcodeStart + elapsed + delta)));
  };
  return (
    <Stack
      alignItems="center"
      justifyContent="center"
      sx={{ flex: 1, minWidth: 0, height: "100%" }}
    >
      {current.kind === "image" &&
        !showingMotion &&
        (status === "ready" || status === "loading") && (
          <Box
            sx={{
              position: "relative",
              display: "inline-flex",
              maxWidth: "100%",
              maxHeight: "100%",
            }}
          >
            <Box
              component="img"
              src={stream.image(current.url) ?? undefined}
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
            {current.motionUrl && status === "ready" && (
              <IconButton
                size="small"
                aria-label={`${getLabel("moments.viewer.playLivePhoto")} ${current.filename}`}
                title={getLabel("moments.viewer.playLivePhoto")}
                onClick={() => {
                  setMotionErrorId(null);
                  setPlaying(false);
                  setVideoMode("original");
                  setTranscodeStart(0);
                  setPlayingLivePhotoId(current.id);
                }}
                sx={{
                  position: "absolute",
                  top: 8,
                  left: 8,
                  color: "common.white",
                  bgcolor: "rgba(0,0,0,0.7)",
                }}
              >
                <PlayArrowIcon fontSize="small" />
              </IconButton>
            )}
          </Box>
        )}
      {motionErrorId === current.id && !showingMotion && (
        <Typography variant="body2" role="status">
          Live Photo motion is unavailable.
        </Typography>
      )}
      {current.kind === "image" && (status === "loading" || status === "queued") && (
        <CircularProgress />
      )}
      {(current.kind === "video" || viewingMotion) && !videoError && (
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
            src={videoSource ?? undefined}
            controls
            autoPlay={shouldAutoplay}
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
          {videoMode === "transcode" && !showingMotion && (
            <Stack direction="row" gap={1} sx={{ position: "absolute", top: 8, left: 8 }}>
              <Button size="small" variant="contained" onClick={() => seekTranscode(-30)}>
                Back 30s
              </Button>
              <Button size="small" variant="contained" onClick={() => seekTranscode(30)}>
                Forward 30s
              </Button>
            </Stack>
          )}
          {buffering && (
            <CircularProgress
              aria-label={getLabel("moments.viewer.buffering")}
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
      {(current.kind === "video" || fallbackMotion) && autoplayBlocked && !videoError && (
        <Typography variant="body2" textAlign="center" role="status">
          Autoplay was blocked. Press play in the video controls.
        </Typography>
      )}
      {current.kind === "image" && status === "error" && !fallbackMotion && (
        <Stack alignItems="center" gap={2}>
          <Typography textAlign="center">{getLabel("moments.gallery.imageLoadFailed")}</Typography>
          <Button onClick={() => queue.retry(current.url)}>
            {getLabel("moments.gallery.retryImage")}
          </Button>
        </Stack>
      )}
      {(current.kind === "unsupported" || videoError) && (
        <Typography textAlign="center">{getLabel("moments.gallery.unsupportedPreview")}</Typography>
      )}
    </Stack>
  );
};
