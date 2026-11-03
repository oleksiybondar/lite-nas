import { useBackupRestore } from "@hooks/useBackupRestore";
import { useLanguage } from "@lite-nas/language";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Shows the selected backup's file differences, or its empty comparison state. */
export const BackupVersionDiff = (): ReactElement | null => {
  const { differences, version } = useBackupRestore();
  const { getLabel } = useLanguage();
  if (!version) return null;
  return (
    <Paper sx={{ p: 2 }}>
      <Typography variant="h3">{getLabel("admin.backup.changes")}</Typography>
      {differences.length === 0 && <Typography>{getLabel("admin.backup.noChanges")}</Typography>}
      {differences.map((item) => (
        <Typography key={item.path} data-test-class="backup-difference" data-test-name={item.path}>
          {item.status}: {item.path}
        </Typography>
      ))}
    </Paper>
  );
};
