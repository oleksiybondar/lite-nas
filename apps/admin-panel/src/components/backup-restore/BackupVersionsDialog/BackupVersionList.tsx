import { BackupVersionCard } from "@components/backup-restore/BackupVersionsDialog/BackupVersionCard";
import { useBackupRestore } from "@hooks/useBackupRestore";
import { useLanguage } from "@lite-nas/language";
import LinearProgress from "@mui/material/LinearProgress";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Filters and renders backup version cards with inventory loading and empty states. */
export const BackupVersionList = (): ReactElement => {
  const {
    application,
    backupDetails,
    busy,
    detailsLoading,
    search,
    session,
    setPendingDelete,
    setPendingRestore,
    setVersion,
    startMerge,
    version,
    versions,
    versionsLoading,
  } = useBackupRestore();
  const { getLabel } = useLanguage();
  const visible = backupDetails.filter((item) =>
    `${item.name ?? ""} ${item.version}`.toLocaleLowerCase().includes(search.toLocaleLowerCase()),
  );
  return (
    <Stack spacing={2}>
      {visible.map((item) => (
        <BackupVersionCard
          key={item.version}
          item={item}
          selected={item.version === version}
          busy={busy}
          sessionActive={Boolean(session)}
          onCompare={() => setVersion(item.version)}
          onRestore={() => setPendingRestore({ application, version: item.version })}
          onMerge={() => startMerge(application, item.version)}
          onDelete={() => setPendingDelete({ application, version: item.version })}
        />
      ))}
      {backupDetails.length === 0 && versions.length === 0 && !versionsLoading && (
        <Typography>{getLabel("admin.backup.noBackups")}</Typography>
      )}
      {backupDetails.length > 0 && visible.length === 0 && (
        <Typography>{getLabel("admin.backup.noBackupsMatch")}</Typography>
      )}
      {(versionsLoading || detailsLoading) && <LinearProgress />}
    </Stack>
  );
};
