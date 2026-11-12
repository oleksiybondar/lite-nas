import { useLanguage } from "@lite-nas/language";
import { Stack, Typography } from "@mui/material";
import type { ReactElement } from "react";

/** Embeds the same-origin Nginx File Explorer mount inside the admin panel. */
export const FileExplorerEmbeddedPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack spacing={2} data-testid="file-explorer-embedded-page">
      <Typography variant="h1">{getLabel("navigation.File Explorer")}</Typography>
      <iframe
        title={getLabel("navigation.File Explorer")}
        src="/apps/file-explorer/"
        data-testid="file-explorer-frame"
        style={{ width: "100%", minHeight: "70vh", border: 0 }}
      />
    </Stack>
  );
};
