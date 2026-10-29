// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: this component composes the existing viewer regions and their controls.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: this component composes the existing viewer regions and their controls.
// biome-ignore-all lint/style/noNestedTernary: media kind selection is localized to the preview renderer.

import { useLanguage } from "@lite-nas/language";
import ArrowBackIosNewIcon from "@mui/icons-material/ArrowBackIosNew";
import ArrowForwardIosIcon from "@mui/icons-material/ArrowForwardIos";
import CloseIcon from "@mui/icons-material/Close";
import DownloadIcon from "@mui/icons-material/Download";
import FullscreenIcon from "@mui/icons-material/Fullscreen";
import FullscreenExitIcon from "@mui/icons-material/FullscreenExit";
import InsertDriveFileOutlinedIcon from "@mui/icons-material/InsertDriveFileOutlined";
import PauseIcon from "@mui/icons-material/Pause";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import VideoFileOutlinedIcon from "@mui/icons-material/VideoFileOutlined";
import Box from "@mui/material/Box";
import ButtonBase from "@mui/material/ButtonBase";
import Dialog from "@mui/material/Dialog";
import IconButton from "@mui/material/IconButton";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import Switch from "@mui/material/Switch";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { KeyboardEvent, ReactElement } from "react";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { useCarousel } from "./CarouselContext";
import { useGallery } from "./GalleryContext";
import type { PhotoMetadata } from "./GalleryExif";
import { GalleryItemMenu } from "./GalleryItemMenu";
import { GalleryViewerDetails } from "./GalleryViewerDetails";
import { GalleryViewerMedia } from "./GalleryViewerMedia";
import { galleryDownloadUrl, imageDelaySchema, videoAutoplaySchema } from "./gallery-model";
import { useStream } from "./StreamContext";
import { useWall } from "./WallContext";

/** Presents the selected original in a modal while leaving the wall mounted. */
export const GalleryViewer = (): ReactElement => {
  const { getLabel } = useLanguage();
  const { items, queue, metadata } = useGallery();
  const stream = useStream();
  const {
    window: { visible: visibleWallItems },
  } = useWall();
  useSyncExternalStore(queue.subscribe, queue.getSnapshot);
  const {
    current,
    selected,
    playing,
    countdown,
    setPlaying,
    videoAutoplay,
    setVideoAutoplay,
    delay,
    setDelay,
    close,
    navigate,
  } = useCarousel();
  const fullscreenRef = useRef<HTMLDivElement>(null);
  const selectedPreviewRef = useRef<HTMLButtonElement>(null);
  const controlsTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [fullscreen, setFullscreen] = useState(false);
  const [controlsVisible, setControlsVisible] = useState(true);
  const [details, setDetails] = useState<PhotoMetadata | null | undefined>(undefined);
  const downloadUrl = current ? galleryDownloadUrl(current.url) : null;
  const previewOffset = visibleWallItems.length;
  const previewItems = items
    .map((item, index) => ({ item, index }))
    .filter(({ item, index }) => {
      if (selected === null || Math.abs(index - selected) > previewOffset) return false;
      if (item.kind === "image")
        return (
          queue.status(item.url) === "ready" ||
          (!!item.motionUrl && queue.status(item.url) === "error")
        );
      return true;
    });
  useEffect(() => {
    if (!current) {
      setDetails(undefined);
      return;
    }
    const cached = metadata.cached(current.url);
    setDetails(cached);
    if (current.kind !== "image") return;
    let active = true;
    void metadata.load(current.url).then((value) => {
      if (active) setDetails(value);
    });
    return () => {
      active = false;
      metadata.cancel(current.url);
    };
  }, [current?.id, current?.kind, current?.url, metadata, current]);
  useEffect(() => {
    const sync = (): void => {
      const isFullscreen = document.fullscreenElement === fullscreenRef.current;
      setFullscreen(isFullscreen);
      setControlsVisible(!isFullscreen);
    };
    document.addEventListener("fullscreenchange", sync);
    return () => document.removeEventListener("fullscreenchange", sync);
  }, []);
  // Selection changes update the preview ref, so this effect must follow selection.
  // biome-ignore lint/correctness/useExhaustiveDependencies: selection changes the target referenced by the preview ref.
  useEffect(() => {
    if (!fullscreen)
      selectedPreviewRef.current?.scrollIntoView({ block: "nearest", inline: "center" });
  }, [fullscreen, previewItems.length, selected]);
  useEffect(
    () => () => {
      if (controlsTimerRef.current !== null) clearTimeout(controlsTimerRef.current);
    },
    [],
  );
  /** Reveals fullscreen controls briefly after pointer movement. */
  const revealFullscreenControls = (): void => {
    if (!fullscreen) return;
    setControlsVisible(true);
    if (controlsTimerRef.current !== null) clearTimeout(controlsTimerRef.current);
    controlsTimerRef.current = setTimeout(() => setControlsVisible(false), 1800);
  };
  /** Toggles browser fullscreen for the active gallery viewer. */
  const toggleFullscreen = async (): Promise<void> => {
    if (document.fullscreenElement === fullscreenRef.current) await document.exitFullscreen();
    else await fullscreenRef.current?.requestFullscreen();
  };
  /** Leaves browser fullscreen and closes the modal. */
  const closeViewer = (): void => {
    if (document.fullscreenElement === fullscreenRef.current)
      void document.exitFullscreen().catch(() => {});
    close();
  };
  /** Applies arrow navigation without taking those keys from form controls. */
  const onKeyDown = (event: KeyboardEvent): void => {
    if (
      event.target instanceof HTMLElement &&
      event.target.closest('input, select, video, [role="combobox"], [role="menu"]')
    )
      return;
    if (event.key === "ArrowRight") {
      event.preventDefault();
      navigate(1);
    }
    if (event.key === "ArrowLeft") {
      event.preventDefault();
      navigate(-1);
    }
  };
  return (
    <Dialog
      fullScreen
      open={selected !== null}
      onClose={closeViewer}
      onKeyDown={onKeyDown}
      slotProps={{ paper: { sx: { bgcolor: "#101014", color: "white" } } }}
    >
      {current && (
        <Stack
          ref={fullscreenRef}
          onMouseMove={revealFullscreenControls}
          sx={{
            height: "100%",
            minHeight: 0,
            position: "relative",
            "&:fullscreen": { width: "100vw", height: "100dvh", bgcolor: "#101014" },
          }}
        >
          <Box
            px={2}
            py={1}
            sx={{
              ...(fullscreen && {
                position: "absolute",
                top: 0,
                left: 0,
                right: 0,
                zIndex: 2,
                opacity: controlsVisible ? 1 : 0,
                visibility: controlsVisible ? "visible" : "hidden",
                pointerEvents: controlsVisible ? "auto" : "none",
                transition: "opacity 180ms ease",
                bgcolor: "rgba(16,16,20,0.68)",
              }),
            }}
          >
            <Box
              sx={{
                display: "grid",
                gridTemplateColumns: "minmax(0, 1fr) minmax(100px, 36%) minmax(0, 1fr)",
                alignItems: "center",
                gap: 1,
                minHeight: 40,
              }}
            >
              <Stack direction="row" alignItems="center" gap={1} sx={{ minWidth: 0 }}>
                <Tooltip title={getLabel("moments.viewer.download")}>
                  <span>
                    <IconButton
                      size="small"
                      component="a"
                      href={downloadUrl ?? undefined}
                      download={current.filename}
                      color="inherit"
                      aria-label={getLabel("moments.viewer.downloadNamed").replace(
                        "{filename}",
                        current.filename,
                      )}
                      disabled={!downloadUrl}
                    >
                      <DownloadIcon fontSize="small" />
                    </IconButton>
                  </span>
                </Tooltip>
                <Tooltip
                  title={getLabel(
                    fullscreen ? "moments.viewer.exitFullscreen" : "moments.viewer.showFullscreen",
                  )}
                >
                  <span>
                    <IconButton
                      size="small"
                      color="inherit"
                      aria-label={getLabel(
                        fullscreen
                          ? "moments.viewer.exitFullscreen"
                          : "moments.viewer.showFullscreen",
                      )}
                      disabled={!document.fullscreenEnabled}
                      onClick={() => void toggleFullscreen().catch(() => {})}
                    >
                      {fullscreen ? (
                        <FullscreenExitIcon fontSize="small" />
                      ) : (
                        <FullscreenIcon fontSize="small" />
                      )}
                    </IconButton>
                  </span>
                </Tooltip>
              </Stack>
              <Stack alignItems="center" sx={{ minWidth: 0, textAlign: "center" }}>
                <Typography variant="body2" fontWeight={600} noWrap title={current.filename}>
                  {current.filename}
                </Typography>
                {details?.capturedAt && (
                  <Typography variant="caption" noWrap sx={{ color: "rgba(255,255,255,0.72)" }}>
                    {details.capturedAt}
                  </Typography>
                )}
              </Stack>
              <Stack direction="row" alignItems="center" justifyContent="flex-end" gap={1.25}>
                {current.kind !== "video" ? (
                  <>
                    <Tooltip
                      title={getLabel(
                        playing ? "moments.viewer.pauseSlideshow" : "moments.viewer.playSlideshow",
                      )}
                    >
                      <IconButton
                        size="small"
                        aria-label={getLabel(
                          playing
                            ? "moments.viewer.pauseSlideshow"
                            : "moments.viewer.playSlideshow",
                        )}
                        onClick={() => setPlaying(!playing)}
                        sx={{
                          bgcolor: "primary.main",
                          color: "primary.contrastText",
                          "&:hover": { bgcolor: "primary.dark" },
                        }}
                      >
                        {playing ? (
                          <PauseIcon fontSize="small" />
                        ) : (
                          <PlayArrowIcon fontSize="small" />
                        )}
                      </IconButton>
                    </Tooltip>
                    <Select
                      size="small"
                      value={delay}
                      onChange={(event) => {
                        const result = imageDelaySchema.safeParse(Number(event.target.value));
                        if (result.success) setDelay(result.data);
                      }}
                      aria-label={getLabel("moments.viewer.slideshowDelay")}
                      sx={{
                        color: "white",
                        minWidth: 64,
                        ".MuiOutlinedInput-notchedOutline": { borderColor: "#777" },
                        ".MuiSvgIcon-root": { color: "white" },
                      }}
                    >
                      <MenuItem value={1000}>1 s</MenuItem>
                      <MenuItem value={3000}>3 s</MenuItem>
                      <MenuItem value={5000}>5 s</MenuItem>
                    </Select>
                  </>
                ) : (
                  <Stack direction="row" alignItems="center" gap={0.25}>
                    <Typography variant="caption">{getLabel("moments.viewer.autoplay")}</Typography>
                    <Tooltip title={getLabel("moments.viewer.autoplayVideos")}>
                      <Switch
                        size="small"
                        checked={videoAutoplay}
                        onChange={(event) => {
                          const result = videoAutoplaySchema.safeParse(event.target.checked);
                          if (result.success) setVideoAutoplay(result.data);
                        }}
                        inputProps={{ "aria-label": getLabel("moments.viewer.autoplayVideos") }}
                      />
                    </Tooltip>
                  </Stack>
                )}
                <Tooltip title={getLabel("moments.viewer.close")}>
                  <IconButton
                    size="small"
                    aria-label={getLabel("moments.viewer.close")}
                    onClick={closeViewer}
                    sx={{
                      bgcolor: "error.main",
                      color: "common.white",
                      "&:hover": { bgcolor: "error.dark" },
                    }}
                  >
                    <CloseIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Stack>
            </Box>
          </Box>
          <Stack
            alignItems="center"
            justifyContent="center"
            sx={{ position: "relative", flex: 1, minHeight: 0 }}
          >
            <GalleryViewerMedia />
            {playing && current.kind !== "video" && countdown !== null && (
              <Typography
                component="output"
                variant="h2"
                aria-label={getLabel("moments.viewer.nextSlide").replace(
                  "{countdown}",
                  String(countdown),
                )}
                sx={{
                  position: "absolute",
                  top: 12,
                  left: 12,
                  color: "primary.main",
                  bgcolor: "transparent",
                  fontWeight: 700,
                  lineHeight: 1,
                  pointerEvents: "none",
                  textShadow: "0 2px 12px rgba(0,0,0,0.65)",
                }}
              >
                {String(countdown).padStart(2, "0")}
              </Typography>
            )}
            <GalleryViewerDetails item={current} details={details} />
            <Box sx={{ position: "absolute", right: 12, bottom: 12, zIndex: 1 }}>
              <GalleryItemMenu item={current} wall />
            </Box>
          </Stack>
          {!fullscreen && (
            <Box sx={{ position: "relative", flexShrink: 0 }}>
              <Typography
                variant="caption"
                role="status"
                aria-live="polite"
                sx={{
                  position: "absolute",
                  top: -22,
                  left: "50%",
                  transform: "translateX(-50%)",
                  zIndex: 1,
                  color: "common.white",
                  bgcolor: "transparent",
                  textShadow: "0 1px 5px rgba(0,0,0,0.85)",
                  pointerEvents: "none",
                }}
              >
                {selected === null ? 0 : selected + 1} / {items.length}
              </Typography>
              <Stack
                component="nav"
                direction="row"
                alignItems="center"
                justifyContent="center"
                gap={1}
                px={2}
                py={1}
                sx={{ borderTop: "1px solid rgba(255,255,255,0.16)" }}
                aria-label={getLabel("moments.viewer.previews")}
              >
                <IconButton
                  color="inherit"
                  aria-label={getLabel("moments.viewer.previous")}
                  disabled={selected === 0}
                  onClick={() => navigate(-1)}
                >
                  <ArrowBackIosNewIcon />
                </IconButton>
                <Stack
                  direction="row"
                  gap={1}
                  sx={{
                    flex: "0 1 auto",
                    minWidth: 0,
                    maxWidth: "calc(100% - 112px)",
                    overflowX: "auto",
                    mx: "auto",
                  }}
                >
                  {previewItems.map(({ item, index }) => (
                    <ButtonBase
                      key={item.id}
                      ref={index === selected ? selectedPreviewRef : undefined}
                      onClick={() => {
                        if (selected !== null) navigate(index - selected);
                      }}
                      aria-label={getLabel("moments.viewer.show").replace(
                        "{filename}",
                        item.filename,
                      )}
                      aria-pressed={selected === index}
                      sx={{
                        width: 76,
                        height: 58,
                        flex: "0 0 auto",
                        overflow: "hidden",
                        borderRadius: 1,
                        border: selected === index ? "2px solid #90caf9" : "2px solid transparent",
                        bgcolor: "rgba(255,255,255,0.1)",
                        color: "white",
                      }}
                    >
                      {item.kind === "image" ? (
                        <Box
                          component="img"
                          src={stream.image(item.url) ?? undefined}
                          alt=""
                          loading="lazy"
                          decoding="async"
                          sx={{ width: "100%", height: "100%", objectFit: "cover" }}
                        />
                      ) : item.kind === "video" ? (
                        <VideoFileOutlinedIcon />
                      ) : (
                        <InsertDriveFileOutlinedIcon />
                      )}
                    </ButtonBase>
                  ))}
                </Stack>
                <IconButton
                  color="inherit"
                  aria-label={getLabel("moments.viewer.next")}
                  disabled={selected === items.length - 1}
                  onClick={() => navigate(1)}
                >
                  <ArrowForwardIosIcon />
                </IconButton>
              </Stack>
            </Box>
          )}
        </Stack>
      )}
    </Dialog>
  );
};
