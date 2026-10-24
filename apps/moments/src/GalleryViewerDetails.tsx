// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: camera and location actions compose one metadata overlay.
import MapOutlinedIcon from "@mui/icons-material/MapOutlined";
import PhoneAndroidOutlinedIcon from "@mui/icons-material/PhoneAndroidOutlined";
import PhotoCameraOutlinedIcon from "@mui/icons-material/PhotoCameraOutlined";
import TabletAndroidOutlinedIcon from "@mui/icons-material/TabletAndroidOutlined";
import Box from "@mui/material/Box";
import IconButton from "@mui/material/IconButton";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import type { ReactElement } from "react";
import type { PhotoMetadata } from "./GalleryExif";
import { googleMapsUrl } from "./GalleryMaps";

/** Selects a device glyph based on EXIF camera text. */
const cameraIcon = (name: string): typeof PhotoCameraOutlinedIcon => {
  if (name.includes("phone")) return PhoneAndroidOutlinedIcon;
  if (name.includes("tablet")) return TabletAndroidOutlinedIcon;
  return PhotoCameraOutlinedIcon;
};

import type { GalleryItem } from "./gallery-model";

/** Places selected-item metadata actions over the lower-left of viewer media. */
export const GalleryViewerDetails = ({
  item,
  details,
}: {
  item: GalleryItem;
  details: PhotoMetadata | null | undefined;
}): ReactElement | null => {
  const mapUrl = details?.location ? googleMapsUrl(details.location) : null;
  const deviceName = details?.camera?.toLowerCase() ?? "";
  const DeviceIcon = cameraIcon(deviceName);
  if (!details?.camera && !mapUrl) return null;

  return (
    <Stack
      direction="row"
      alignItems="center"
      gap={0.5}
      sx={{ position: "absolute", left: 12, bottom: 12, zIndex: 1, pointerEvents: "none" }}
    >
      {details?.camera && (
        <Tooltip title={details.camera}>
          <Box
            component="span"
            sx={{
              display: "flex",
              alignItems: "center",
              p: 1,
              borderRadius: "50%",
              color: "common.white",
              bgcolor: "rgba(0,0,0,0.58)",
              pointerEvents: "auto",
            }}
          >
            <DeviceIcon fontSize="small" aria-label={`Camera: ${details.camera}`} />
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
            aria-label={`Show ${item.filename} on map`}
            sx={{
              color: "common.white",
              bgcolor: "rgba(0,0,0,0.58)",
              pointerEvents: "auto",
              "&:hover": { bgcolor: "rgba(0,0,0,0.82)" },
            }}
          >
            <MapOutlinedIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      )}
    </Stack>
  );
};
