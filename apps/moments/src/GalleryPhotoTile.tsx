// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: image state and EXIF overlays compose one wall tile.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: image state and EXIF overlays compose one wall tile.
import MapOutlinedIcon from "@mui/icons-material/MapOutlined";
import PhoneAndroidOutlinedIcon from "@mui/icons-material/PhoneAndroidOutlined";
import PhotoCameraOutlinedIcon from "@mui/icons-material/PhotoCameraOutlined";
import TabletAndroidOutlinedIcon from "@mui/icons-material/TabletAndroidOutlined";
import Box from "@mui/material/Box";
import IconButton from "@mui/material/IconButton";
import Skeleton from "@mui/material/Skeleton";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";
import { useGallery } from "./GalleryContext";
import type { PhotoMetadata } from "./GalleryExif";
import { googleMapsUrl } from "./GalleryMaps";
import type { GalleryItem } from "./gallery-model";

/** Chooses a familiar camera glyph from the reported device name. */
const cameraIcon = (name: string): typeof PhotoCameraOutlinedIcon => {
  if (name.includes("phone")) return PhoneAndroidOutlinedIcon;
  if (name.includes("tablet")) return TabletAndroidOutlinedIcon;
  return PhotoCameraOutlinedIcon;
};

/** Shows one wall photo and reads EXIF only while its parent item has interest. */
export const GalleryPhotoTile = ({
  item,
  onOpen,
  interested,
}: {
  item: GalleryItem;
  onOpen: () => void;
  interested: boolean;
}): ReactElement => {
  const { queue, metadata: reader } = useGallery();
  const status = queue.status(item.url);
  const attempt = queue.attempt(item.url);
  const [details, setDetails] = useState<PhotoMetadata | null | undefined>(() =>
    reader.cached(item.url),
  );
  const mapUrl = details?.location ? googleMapsUrl(details.location) : null;
  const deviceName = details?.camera?.toLowerCase() ?? "";
  const DeviceIcon = cameraIcon(deviceName);
  useEffect(() => {
    if (!interested) return;
    const cached = reader.cached(item.url);
    if (cached !== undefined) {
      setDetails(cached);
      return;
    }
    let active = true;
    const timer = setTimeout(() => {
      void reader.load(item.url).then((value) => {
        if (active) setDetails(value);
      });
    }, 250);
    return () => {
      active = false;
      clearTimeout(timer);
      reader.cancel(item.url);
    };
  }, [interested, item.url, reader]);

  return (
    <Box
      sx={{
        position: "relative",
        width: "100%",
        height: "100%",
      }}
    >
      <Box
        component="button"
        type="button"
        aria-label={`Open ${item.filename}`}
        onClick={onOpen}
        sx={{
          position: "absolute",
          inset: 0,
          width: "100%",
          height: "100%",
          p: 0,
          border: 0,
          cursor: "pointer",
          bgcolor: "transparent",
          color: "text.primary",
          textAlign: "left",
          "&:focus-visible": {
            outline: "3px solid",
            outlineColor: "primary.main",
            outlineOffset: -3,
          },
        }}
      >
        {(status === "ready" || status === "loading") && (
          <Box
            component="img"
            src={item.url}
            alt={item.filename}
            onLoad={() => queue.complete(item.url, true, attempt)}
            onError={() => queue.fail(item.url, attempt)}
            sx={{
              position: "absolute",
              inset: 0,
              width: "100%",
              height: "100%",
              objectFit: "cover",
              opacity: status === "ready" ? 1 : 0,
              transform: status === "ready" ? "translateY(0)" : "translateY(35px)",
              transition: "opacity 320ms ease-out, transform 320ms ease-out",
              "@media (prefers-reduced-motion: reduce)": { transition: "none", transform: "none" },
            }}
          />
        )}
        {status !== "error" && status !== "ready" && (
          <Skeleton
            variant="rectangular"
            width="100%"
            height="100%"
            sx={{ position: "absolute", inset: 0, pointerEvents: "none" }}
          />
        )}
        {status === "error" && <Typography sx={{ p: 2 }}>Image unavailable</Typography>}
      </Box>
      <Stack
        className="gallery-wall-hover"
        gap={0.5}
        sx={{
          position: "absolute",
          left: 0,
          right: 0,
          bottom: 0,
          p: 1,
          color: "common.white",
          bgcolor: "rgba(0,0,0,0.76)",
          pointerEvents: "none",
          opacity: 0,
          transform: "translateY(8px)",
          transition: "opacity 180ms ease, transform 180ms ease",
          "@media (prefers-reduced-motion: reduce)": { transition: "none", transform: "none" },
        }}
      >
        <Stack direction="row" alignItems="center" gap={1} sx={{ minWidth: 0 }}>
          <Typography
            variant="body2"
            fontWeight={600}
            noWrap
            title={item.filename}
            sx={{ minWidth: 0, flex: "1 1 auto" }}
          >
            {item.filename}
          </Typography>
          {details?.capturedAt && (
            <Typography
              variant="caption"
              fontWeight={600}
              sx={{ flexShrink: 0, whiteSpace: "nowrap" }}
            >
              {details.capturedAt}
            </Typography>
          )}
        </Stack>
        {(details?.camera || mapUrl) && (
          <Stack direction="row" alignItems="center" gap={0.5}>
            {details?.camera && (
              <Tooltip title={details.camera}>
                <Box
                  sx={{
                    display: "flex",
                    alignItems: "center",
                    cursor: "help",
                    pointerEvents: "auto",
                    p: 0.5,
                  }}
                >
                  <DeviceIcon fontSize="small" aria-label={details.camera} />
                </Box>
              </Tooltip>
            )}
            {mapUrl && (
              <Tooltip title="Show on map">
                <IconButton
                  component="a"
                  href={mapUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  size="small"
                  aria-label="Show on map"
                  sx={{ color: "common.white", pointerEvents: "auto", p: 0.5 }}
                >
                  <MapOutlinedIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            )}
          </Stack>
        )}
      </Stack>
    </Box>
  );
};
