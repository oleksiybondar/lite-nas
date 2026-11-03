import {
  BackgroundTaskActions,
  BackgroundTaskProgress,
  BackgroundTaskProvider,
  useBackgroundTask,
} from "@lite-nas/background-task";
import { useLanguage } from "@lite-nas/language";
import Alert from "@mui/material/Alert";
import FormControlLabel from "@mui/material/FormControlLabel";
import Stack from "@mui/material/Stack";
import Switch from "@mui/material/Switch";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useMemo, useState } from "react";
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
  // Load an existing scan even when the editor starts closed; later openings refresh it.
  const [hasOpened, setHasOpened] = useState(open);
  useEffect(() => {
    if (open) setHasOpened(true);
  }, [open]);
  const label = getLabel(kind === "CAPTION" ? "moments.scan.captions" : "moments.scan.media");
  const adapter = useMemo(
    () => ({
      loadLatest: () => latestGalleryScan(galleryId, kind),
      poll: getGalleryScan,
      getId: (scan: GalleryScan) => scan.id,
      isActive: scanIsActive,
    }),
    [galleryId, kind],
  );

  return (
    <BackgroundTaskProvider
      key={`${galleryId}-${kind}`}
      taskKey={`${galleryId}-${kind}`}
      adapter={adapter}
      enabled={open || !hasOpened}
      loadFailedMessage={getLabel("moments.scan.loadingStatus").replace("{kind}", label)}
      refreshFailedMessage={getLabel("moments.scan.refreshFailed").replace("{kind}", label)}
    >
      <GalleryScanControlsContent galleryId={galleryId} kind={kind} label={label} />
    </BackgroundTaskProvider>
  );
};

/** Renders gallery-specific actions and counters using the shared task lifecycle. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Domain-specific start, cancellation, and scan counters belong in this control.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The control keeps the scan summary beside its actions.
function GalleryScanControlsContent({
  galleryId,
  kind,
  label,
}: {
  galleryId: string;
  kind: GalleryScan["kind"];
  label: string;
}): ReactElement {
  const { getLabel } = useLanguage();
  const { task: scan, loading, error, execute } = useBackgroundTask<GalleryScan>();
  const [force, setForce] = useState(false);

  const start = async (): Promise<void> => {
    await execute(
      () => (kind === "CAPTION" ? startCaptionScan(galleryId, force) : startGalleryScan(galleryId)),
      getLabel("moments.scan.startFailed").replace("{kind}", label),
    );
  };

  const cancel = async (): Promise<void> => {
    if (!scan) return;
    await execute(
      () => cancelGalleryScan(scan.id),
      getLabel("moments.scan.cancelFailed").replace("{kind}", label),
    );
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
      <BackgroundTaskActions
        active={active}
        disabled={loading}
        startLabel={getLabel(
          kind === "CAPTION" ? "moments.scan.generateCaptions" : "moments.scan.scanMedia",
        )}
        cancelLabel={getLabel("moments.scan.cancel")}
        onStart={() => void start()}
        onCancel={() => void cancel()}
      />
      <BackgroundTaskProgress active={active} label={label} value={progress}>
        {scan && (
          <Typography role="status" variant="body2">
            {getLabel(`moments.scan.status.${scan.status.toLowerCase()}`)} · {scan.processedItems}
            {scan.totalItems === null ? "" : ` / ${scan.totalItems}`}{" "}
            {getLabel("moments.scan.processed")} · {scan.successfulItems}{" "}
            {getLabel("moments.scan.successful")} · {scan.failedItems}{" "}
            {getLabel("moments.scan.failed")}
          </Typography>
        )}
      </BackgroundTaskProgress>
      {scan?.errorSummary && <Alert severity="warning">{scan.errorSummary}</Alert>}
      {error && <Alert severity="error">{error}</Alert>}
    </Stack>
  );
}
