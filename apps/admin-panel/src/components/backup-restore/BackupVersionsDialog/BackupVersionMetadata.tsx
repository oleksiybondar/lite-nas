import { useLanguage } from "@lite-nas/language";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { BackupDetails } from "@schemas/backup-restore";
import type { ReactElement } from "react";

/** Presents the descriptive and timestamp metadata for one backup version. */
export const BackupVersionMetadata = ({ item }: { item: BackupDetails }): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack direction="row" useFlexGap flexWrap="wrap" alignItems="center" gap={2} sx={{ flex: 1 }}>
      <Typography variant="h3" color="primary.main" sx={{ overflowWrap: "anywhere" }}>
        {item.name || item.version}
      </Typography>
      <Typography variant="body2" sx={{ overflowWrap: "anywhere" }}>
        {item.version}
      </Typography>
      {item.description && <Typography variant="body2">{item.description}</Typography>}
      <Typography variant="body2">
        {getLabel("admin.backup.createdAt")}:{" "}
        <time dateTime={item.created_at}>{new Date(item.created_at).toLocaleString()}</time>
      </Typography>
      <Typography variant="body2">
        {getLabel("admin.backup.createdBy")}: {item.created_by}
      </Typography>
    </Stack>
  );
};
