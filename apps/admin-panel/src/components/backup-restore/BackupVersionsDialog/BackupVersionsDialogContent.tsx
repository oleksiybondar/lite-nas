import { BackupVersionDiff } from "@components/backup-restore/BackupVersionsDialog/BackupVersionDiff";
import { BackupVersionList } from "@components/backup-restore/BackupVersionsDialog/BackupVersionList";
import { BackupVersionsControls } from "@components/backup-restore/BackupVersionsDialog/BackupVersionsControls";
import { useBackupRestore } from "@hooks/useBackupRestore";
import { StandardDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/** Composes the backup search controls, inventory, and selected-version differences. */
export const BackupVersionsDialogContent = (): ReactElement => {
  const { application, backupsOpen, busy, setBackupsOpen } = useBackupRestore();
  const { getLabel } = useLanguage();
  return (
    <StandardDialog
      open={backupsOpen}
      fullWidth
      maxWidth="lg"
      title={application}
      closeLabel={getLabel("dialogs.close")}
      closeDisabled={busy}
      onClose={() => !busy && setBackupsOpen(false)}
      contentProps={{ dividers: true }}
    >
      <Stack spacing={2}>
        <BackupVersionsControls />
        <BackupVersionList />
        <BackupVersionDiff />
      </Stack>
    </StandardDialog>
  );
};
