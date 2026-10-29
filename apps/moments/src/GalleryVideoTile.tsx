// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: playback state and its user-initiated controls compose one tile.
// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: fallback and time seeking share the tile's video lifecycle.
// biome-ignore-all lint/a11y/useMediaCaption: original files have no associated caption resource in the gallery contract.
import { useLanguage } from "@lite-nas/language";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import { useCarousel } from "./CarouselContext";
import type { GalleryItem } from "./gallery-model";
import { useStream } from "./StreamContext";
import { useWall } from "./WallContext";

/** Offers native wall playback without requesting a video until a user presses Play. */
export const GalleryVideoTile = ({ item }: { item: GalleryItem }): ReactElement => {
  const { getLabel } = useLanguage();
  const videoRef = useRef<HTMLVideoElement>(null);
  const [loaded, setLoaded] = useState(false);
  const [mode, setMode] = useState<"original" | "transcode" | "failed">("original");
  const [transcodeStart, setTranscodeStart] = useState(0);
  const stream = useStream();
  const { selected } = useCarousel();
  const { activeVideoId, setActiveVideoId } = useWall();
  useEffect(() => {
    if (selected !== null || (activeVideoId !== null && activeVideoId !== item.id))
      videoRef.current?.pause();
  }, [activeVideoId, item.id, selected]);
  useEffect(() => {
    if (loaded && mode === "transcode") {
      const video = videoRef.current;
      const url = stream.transcode(item.url, transcodeStart);
      if (video && url) {
        video.src = url;
        void video.play().catch(() => {
          // Native controls remain available if automatic fallback playback is blocked.
        });
      }
    }
  }, [item.url, loaded, mode, stream, transcodeStart]);
  /** Starts this original only after a direct user action. */
  const play = (): void => {
    const video = videoRef.current;
    if (!video) return;
    setActiveVideoId(item.id);
    setLoaded(true);
    setMode("original");
    setTranscodeStart(0);
    if (!video.src) video.src = stream.original(item.url) ?? "";
    void video.play().catch(() => {
      // Native controls remain available if the browser declines immediate playback.
    });
  };
  return (
    <Box
      sx={{
        position: "relative",
        width: "100%",
        height: "100%",
        bgcolor: "common.black",
        color: "common.white",
      }}
    >
      <video
        ref={videoRef}
        src={
          loaded
            ? ((mode === "transcode"
                ? stream.transcode(item.url, transcodeStart)
                : stream.original(item.url)) ?? undefined)
            : undefined
        }
        controls={loaded}
        playsInline
        preload="none"
        onError={() => setMode((current) => (current === "original" ? "transcode" : "failed"))}
        aria-label={`${getLabel("moments.viewer.playVideo")} ${item.filename}`}
        style={{ width: "100%", height: "100%", objectFit: "contain", display: "block" }}
      />
      {loaded && mode === "transcode" && (
        <Stack direction="row" gap={1} sx={{ position: "absolute", top: 8, left: 8 }}>
          <Button
            size="small"
            variant="contained"
            onClick={() =>
              setTranscodeStart(
                Math.min(
                  86400,
                  Math.max(0, transcodeStart + Math.floor(videoRef.current?.currentTime ?? 0) - 30),
                ),
              )
            }
          >
            {getLabel("moments.viewer.back30")}
          </Button>
          <Button
            size="small"
            variant="contained"
            onClick={() =>
              setTranscodeStart(
                Math.min(
                  86400,
                  transcodeStart + Math.floor(videoRef.current?.currentTime ?? 0) + 30,
                ),
              )
            }
          >
            {getLabel("moments.viewer.forward30")}
          </Button>
        </Stack>
      )}
      {!loaded && (
        <Stack
          alignItems="center"
          justifyContent="center"
          gap={0.5}
          sx={{ position: "absolute", inset: 0, px: 1, bgcolor: "rgba(0,0,0,0.55)" }}
        >
          <Typography variant="body2" textAlign="center" noWrap sx={{ maxWidth: "100%" }}>
            {item.filename}
          </Typography>
          <Button
            startIcon={<PlayArrowIcon />}
            onClick={play}
            variant="contained"
            size="small"
            aria-label={`${getLabel("moments.viewer.playVideo")} ${item.filename}`}
          >
            {getLabel("moments.viewer.playVideo")}
          </Button>
        </Stack>
      )}
      {mode === "failed" && (
        <Typography
          sx={{
            position: "absolute",
            left: 8,
            bottom: 8,
            bgcolor: "rgba(0,0,0,0.75)",
            px: 1,
            borderRadius: 1,
          }}
          variant="caption"
        >
          {getLabel("moments.viewer.videoUnavailable")}
        </Typography>
      )}
    </Box>
  );
};
