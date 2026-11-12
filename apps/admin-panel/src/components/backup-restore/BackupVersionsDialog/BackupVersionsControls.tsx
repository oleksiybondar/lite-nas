import { useBackupRestore } from "@hooks/useBackupRestore";
import { useLanguage } from "@lite-nas/language";
import MenuItem from "@mui/material/MenuItem";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import { restoreModeSchema } from "@schemas/backup-restore";
import type { ReactElement } from "react";

/** Renders search and restore-mode controls for the backup version inventory. */
export const BackupVersionsControls = (): ReactElement => {
  const { busy, search, session, restoreMode, setSearch, setRestoreMode } = useBackupRestore();
  const { getLabel } = useLanguage();
  return (
    <Stack spacing={2}>
      <TextField
        data-testid="backup-name-search"
        label={getLabel("admin.backup.search")}
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        fullWidth
      />
      <TextField
        select
        data-testid="backup-restore-mode"
        label={getLabel("admin.backup.restoreMode")}
        value={restoreMode}
        onChange={(event) => setRestoreMode(restoreModeSchema.parse(event.target.value))}
        disabled={busy || Boolean(session)}
      >
        <MenuItem value="merge">{getLabel("admin.backup.mergeMode")}</MenuItem>
        <MenuItem value="replace">{getLabel("admin.backup.replaceMode")}</MenuItem>
      </TextField>
    </Stack>
  );
};
