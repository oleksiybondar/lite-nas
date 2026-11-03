import { Stack, Typography } from "@mui/material";
import type { ReactElement } from "react";

/** Embeds the same-origin Nginx Moments mount inside the admin panel. */
export const MomentsEmbeddedPage = (): ReactElement => {
  return (
    <Stack spacing={2} data-testid="moments-embedded-page">
      <Typography variant="h1">Moments</Typography>
      <iframe
        title="Moments"
        allowFullScreen
        src="/apps/moments/"
        data-testid="moments-frame"
        style={{ width: "100%", minHeight: "70vh", border: 0 }}
      />
    </Stack>
  );
};
