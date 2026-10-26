// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: scan history, polling, and user actions belong to one status control.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: the scan control keeps its lifecycle and feedback together.
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import LinearProgress from "@mui/material/LinearProgress";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import {
  cancelGalleryScan,
  getGalleryScan,
  latestGalleryScan,
  startGalleryScan,
} from "./moments-data";
import type { GalleryScan } from "./schemas";

/** Active jobs need progress polling to renew the worker's access token. */
const scanIsActive = (scan: GalleryScan | null): boolean =>
  scan?.status === "PENDING" || scan?.status === "DISCOVERING" || scan?.status === "RUNNING";

/** Starts, monitors, and cancels a saved gallery's full media scan. */
export const GalleryScanControls = ({
  galleryId,
  open,
}: {
  galleryId: string;
  open: boolean;
}): ReactElement => {
  const [scan, setScan] = useState<GalleryScan | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionBusy, setActionBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
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
    void latestGalleryScan(galleryId)
      .then((job) => {
        if (active) setScan(job);
      })
      .catch(() => {
        if (active) setError("Could not load media scan status.");
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
  }, [galleryId, open]);

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
        if (active) setError("Could not refresh media scan progress.");
      } finally {
        if (active) timer = window.setTimeout(() => void poll(), 3000);
      }
    };
    timer = window.setTimeout(() => void poll(), 3000);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [scan]);

  const start = async (): Promise<void> => {
    setActionBusy(true);
    setError(null);
    try {
      setScan(await startGalleryScan(galleryId));
    } catch {
      setError("Could not start media scan.");
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
      setError("Could not cancel media scan.");
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
      <Typography variant="h6">Media scan</Typography>
      <Typography variant="body2" color="text.secondary">
        Scan the gallery's files to update searchable media types and tags.
      </Typography>
      <Stack direction="row" spacing={1}>
        <Button disabled={loading || actionBusy || active} onClick={() => void start()}>
          Scan media
        </Button>
        {active && (
          <Button disabled={actionBusy} onClick={() => void cancel()}>
            Cancel scan
          </Button>
        )}
      </Stack>
      {scan && (
        <Typography role="status" variant="body2">
          {scan.status.toLowerCase()} · {scan.processedItems}
          {scan.totalItems === null ? "" : ` / ${scan.totalItems}`} processed ·{" "}
          {scan.successfulItems} successful · {scan.failedItems} failed
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
