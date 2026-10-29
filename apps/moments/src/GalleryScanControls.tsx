// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: scan history, polling, and user actions belong to one status control.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: the scan control keeps its lifecycle and feedback together.
import { useLanguage } from "@lite-nas/language";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import FormControlLabel from "@mui/material/FormControlLabel";
import LinearProgress from "@mui/material/LinearProgress";
import Stack from "@mui/material/Stack";
import Switch from "@mui/material/Switch";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import {
  cancelGalleryScan,
  getGalleryScan,
  latestGalleryScan,
  startCaptionScan,
  startGalleryScan,
} from "./moments-data";
import type { GalleryScan } from "./schemas";

/** Active jobs need progress polling to renew the worker's access token. */
const scanIsActive = (scan: GalleryScan | null): boolean =>
  scan?.status === "PENDING" || scan?.status === "DISCOVERING" || scan?.status === "RUNNING";

/** Starts, monitors, and cancels one kind of saved gallery scan. */
export const GalleryScanControls = ({
  galleryId,
  open,
  kind = "MEDIA",
}: {
  galleryId: string;
  open: boolean;
  kind?: GalleryScan["kind"];
}): ReactElement => {
  const { getLabel } = useLanguage();
  const [scan, setScan] = useState<GalleryScan | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionBusy, setActionBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [force, setForce] = useState(false);
  const label = getLabel(kind === "CAPTION" ? "moments.scan.captions" : "moments.scan.media");
  const loadedGallery = useRef<string | null>(null);
  const historyLoaded = useRef(false);

  useEffect(() => {
    // Closing the editor leaves the active poll running; opening it refreshes history.
    if (!open && loadedGallery.current === galleryId && historyLoaded.current) return;
    loadedGallery.current = galleryId;
    historyLoaded.current = false;
    let active = true;
    setLoading(true);
    setError(null);
    void latestGalleryScan(galleryId, kind)
      .then((job) => {
        if (active) setScan(job);
      })
      .catch(() => {
        if (active) setError(getLabel("moments.scan.loadingStatus").replace("{kind}", label));
      })
      .finally(() => {
        if (active) {
          historyLoaded.current = true;
          setLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [galleryId, open, kind, label, getLabel]);

  useEffect(() => {
    if (!scanIsActive(scan) || !scan) return;
    let active = true;
    let timer: number;
    const poll = async (): Promise<void> => {
      try {
        const current = await getGalleryScan(scan.id);
        if (active) {
          setScan(current);
          setError(null);
        }
      } catch {
        if (active) setError(getLabel("moments.scan.refreshFailed").replace("{kind}", label));
      } finally {
        if (active) timer = window.setTimeout(() => void poll(), 3000);
      }
    };
    timer = window.setTimeout(() => void poll(), 3000);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [scan, label, getLabel]);

  const start = async (): Promise<void> => {
    setActionBusy(true);
    setError(null);
    try {
      setScan(
        await (kind === "CAPTION"
          ? startCaptionScan(galleryId, force)
          : startGalleryScan(galleryId)),
      );
    } catch {
      setError(getLabel("moments.scan.startFailed").replace("{kind}", label));
    } finally {
      setActionBusy(false);
    }
  };

  const cancel = async (): Promise<void> => {
    if (!scan) return;
    setActionBusy(true);
    setError(null);
    try {
      setScan(await cancelGalleryScan(scan.id));
    } catch {
      setError(getLabel("moments.scan.cancelFailed").replace("{kind}", label));
    } finally {
      setActionBusy(false);
    }
  };

  const active = scanIsActive(scan);
  const progress =
    scan?.totalItems && scan.totalItems > 0
      ? Math.min(100, (scan.processedItems / scan.totalItems) * 100)
      : null;
  return (
    <Stack
      spacing={1.5}
      data-testid="gallery-scan-controls"
      sx={{ borderTop: 1, borderColor: "divider", pt: 2 }}
    >
      <Typography variant="h6">
        {getLabel(kind === "CAPTION" ? "moments.scan.captions" : "moments.scan.media")}
      </Typography>
      <Typography variant="body2" color="text.secondary">
        {kind === "CAPTION"
          ? getLabel("moments.scan.descriptionCaptions")
          : getLabel("moments.scan.descriptionMedia")}
      </Typography>
      {kind === "CAPTION" && (
        <FormControlLabel
          control={<Switch checked={force} onChange={(event) => setForce(event.target.checked)} />}
          label={getLabel("moments.scan.regenerate")}
        />
      )}
      <Stack direction="row" spacing={1}>
        <Button disabled={loading || actionBusy || active} onClick={() => void start()}>
          {getLabel(
            kind === "CAPTION" ? "moments.scan.generateCaptions" : "moments.scan.scanMedia",
          )}
        </Button>
        {active && (
          <Button disabled={actionBusy} onClick={() => void cancel()}>
            {getLabel("moments.scan.cancel")}
          </Button>
        )}
      </Stack>
      {scan && (
        <Typography role="status" variant="body2">
          {getLabel(`moments.scan.status.${scan.status.toLowerCase()}`)} · {scan.processedItems}
          {scan.totalItems === null ? "" : ` / ${scan.totalItems}`}{" "}
          {getLabel("moments.scan.processed")} · {scan.successfulItems}{" "}
          {getLabel("moments.scan.successful")} · {scan.failedItems}{" "}
          {getLabel("moments.scan.failed")}
        </Typography>
      )}
      {active && (
        <LinearProgress
          variant={progress === null ? "indeterminate" : "determinate"}
          value={progress ?? undefined}
        />
      )}
      {scan?.errorSummary && <Alert severity="warning">{scan.errorSummary}</Alert>}
      {error && <Alert severity="error">{error}</Alert>}
    </Stack>
  );
};
