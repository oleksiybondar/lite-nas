import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Slider from "@mui/material/Slider";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { PointerEvent, ReactElement } from "react";
import { useRef, useState } from "react";

type ProfileAvatarCropDialogProps = {
  source: string | null;
  onClose: () => void;
  onSave: (file: File) => void;
};

/** Lets users position and zoom a selected image before saving their avatar. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The dialog owns one cohesive crop interaction and its pointer/canvas state.
export const ProfileAvatarCropDialog = ({
  source,
  onClose,
  onSave,
}: ProfileAvatarCropDialogProps): ReactElement => {
  const { getLabel } = useLanguage();
  const [zoom, setZoom] = useState(1);
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const drag = useRef<{ x: number; y: number; left: number; top: number } | null>(null);
  const image = useRef<HTMLImageElement>(null);
  const cropSize = 300;

  /** Creates a centered square JPEG from the visible crop region. */
  const handleCrop = (): void => {
    const element = image.current;
    if (!element || !source) return;
    const scale =
      Math.max(cropSize / element.naturalWidth, cropSize / element.naturalHeight) * zoom;
    const width = element.naturalWidth * scale;
    const height = element.naturalHeight * scale;
    const canvas = document.createElement("canvas");
    canvas.width = 512;
    canvas.height = 512;
    const context = canvas.getContext("2d");
    if (!context) return;
    context.drawImage(
      element,
      (cropSize / 2 - offset.x - width / 2) / scale,
      (cropSize / 2 - offset.y - height / 2) / scale,
      cropSize / scale,
      cropSize / scale,
      0,
      0,
      canvas.width,
      canvas.height,
    );
    canvas.toBlob(
      (blob) => {
        if (blob) onSave(new File([blob], "avatar.jpg", { type: "image/jpeg" }));
      },
      "image/jpeg",
      0.92,
    );
  };

  /** Keeps the dragged image within the circular crop boundary. */
  const moveCrop = (event: PointerEvent<HTMLDivElement>): void => {
    if (!drag.current || !image.current) return;
    const scale =
      Math.max(cropSize / image.current.naturalWidth, cropSize / image.current.naturalHeight) *
      zoom;
    const halfOverflowX = (image.current.naturalWidth * scale - cropSize) / 2;
    const halfOverflowY = (image.current.naturalHeight * scale - cropSize) / 2;
    setOffset({
      x: Math.max(
        -halfOverflowX,
        Math.min(halfOverflowX, drag.current.left + event.clientX - drag.current.x),
      ),
      y: Math.max(
        -halfOverflowY,
        Math.min(halfOverflowY, drag.current.top + event.clientY - drag.current.y),
      ),
    });
  };

  return (
    <Dialog open={source !== null} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>{getLabel("profile.cropAvatar")}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} alignItems="center" sx={{ pt: 1 }}>
          <div
            role="presentation"
            data-testid="profile-avatar-crop-area"
            onPointerDown={(event) => {
              event.currentTarget.setPointerCapture(event.pointerId);
              drag.current = {
                x: event.clientX,
                y: event.clientY,
                left: offset.x,
                top: offset.y,
              };
            }}
            onPointerMove={moveCrop}
            onPointerUp={() => {
              drag.current = null;
            }}
            style={{
              width: cropSize,
              height: cropSize,
              overflow: "hidden",
              position: "relative",
              cursor: "grab",
              touchAction: "none",
              borderRadius: 12,
              background: "#222",
            }}
          >
            {source ? (
              <img
                ref={image}
                src={source}
                alt=""
                onLoad={() => setOffset({ x: 0, y: 0 })}
                style={{
                  position: "absolute",
                  maxWidth: "none",
                  width:
                    Math.max(
                      cropSize / (image.current?.naturalWidth ?? cropSize),
                      cropSize / (image.current?.naturalHeight ?? cropSize),
                    ) *
                    (image.current?.naturalWidth ?? cropSize) *
                    zoom,
                  height: "auto",
                  left: "50%",
                  top: "50%",
                  transform: `translate(calc(-50% + ${offset.x}px), calc(-50% + ${offset.y}px))`,
                  pointerEvents: "none",
                }}
              />
            ) : null}
            <div
              aria-hidden="true"
              style={{
                position: "absolute",
                inset: 0,
                border: "2px solid rgba(255,255,255,.85)",
                borderRadius: "50%",
                boxShadow: "0 0 0 999px rgba(0,0,0,.35)",
                pointerEvents: "none",
              }}
            />
          </div>
          <Typography variant="body2">{getLabel("profile.cropHint")}</Typography>
          <Slider
            aria-label={getLabel("profile.zoomAvatar")}
            data-testid="profile-avatar-zoom"
            min={1}
            max={3}
            step={0.05}
            value={zoom}
            onChange={(_, value) => setZoom(value as number)}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{getLabel("profile.cancel")}</Button>
        <Button variant="contained" onClick={handleCrop} data-testid="profile-avatar-crop-confirm">
          {getLabel("profile.saveCrop")}
        </Button>
      </DialogActions>
    </Dialog>
  );
};
