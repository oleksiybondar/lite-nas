import { useLanguage } from "@lite-nas/language";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { z } from "zod";
import { GalleryDetail } from "./GalleryDetail";
import { getGallery } from "./moments-data";
import { defaultImageUiCooldownMs, type GalleryRecord } from "./schemas";

/** Loads a shared gallery URL through the caller's existing API permissions. */
export function GalleryRoute({
  userId,
  maxConcurrent,
  imageUiCooldownMs = defaultImageUiCooldownMs,
  workingDir,
}: {
  userId: string;
  maxConcurrent: number;
  imageUiCooldownMs?: number;
  workingDir: string;
}): ReactElement {
  const { getLabel } = useLanguage();
  const { id } = useParams();
  const navigate = useNavigate();
  const [gallery, setGallery] = useState<GalleryRecord | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const parsed = z.uuid().safeParse(id);
    if (!parsed.success) {
      setGallery(null);
      setError(getLabel("moments.gallery.invalidLink"));
      setLoading(false);
      return;
    }
    let active = true;
    setGallery(null);
    setError(null);
    setLoading(true);
    void getGallery(parsed.data)
      .then((result) => {
        if (active) setGallery(result);
      })
      .catch(() => {
        if (active) setError(getLabel("moments.gallery.openFailed"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [getLabel, id]);

  if (gallery)
    return (
      <GalleryDetail
        key={gallery.id}
        selected={gallery}
        userId={userId}
        maxConcurrent={maxConcurrent}
        imageUiCooldownMs={imageUiCooldownMs}
        workingDir={workingDir}
        onBack={() => navigate("/")}
      />
    );
  return <GalleryRouteLoading getLabel={getLabel} loading={loading} error={error} />;
}

/** Shows the route's loading, error, and return-to-galleries actions. */
function GalleryRouteLoading({
  getLabel,
  loading,
  error,
}: {
  getLabel: (key: string) => string;
  loading: boolean;
  error: string | null;
}): ReactElement {
  const navigate = useNavigate();
  return (
    <Stack spacing={2}>
      <Button onClick={() => navigate("/")} sx={{ alignSelf: "start" }}>
        Back to galleries
      </Button>
      {loading && <Typography>{getLabel("moments.gallery.loadingGallery")}</Typography>}
      {error && <Alert severity="error">{error}</Alert>}
    </Stack>
  );
}
