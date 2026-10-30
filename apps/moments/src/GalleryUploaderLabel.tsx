import { useLanguage } from "@lite-nas/language";
import AccountCircleOutlinedIcon from "@mui/icons-material/AccountCircleOutlined";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Identifies the member whose ID is encoded in an album upload folder. */
export function GalleryUploaderLabel({ uploaderId }: { uploaderId: string }): ReactElement {
  const { getLabel } = useLanguage();
  return (
    <Stack direction="row" alignItems="center" gap={0.5}>
      <AccountCircleOutlinedIcon fontSize="small" aria-hidden="true" data-testid="uploader-icon" />
      <Typography variant="caption" noWrap title={uploaderId}>
        {getLabel("moments.gallery.uploadedBy").replace("{userId}", uploaderId)}
      </Typography>
    </Stack>
  );
}
