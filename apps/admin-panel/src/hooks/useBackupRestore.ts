import { BackupRestoreContext } from "@contexts/backup-restore-context";
import type { BackupRestoreContextValue } from "@dto/backup-restore";
import { useContext } from "react";

/** Reads backup and restore page state from its dedicated provider. */
export const useBackupRestore = (): BackupRestoreContextValue => {
  const context = useContext(BackupRestoreContext);

  if (context === undefined) {
    throw new Error("useBackupRestore must be used inside BackupRestoreProvider");
  }

  return context;
};
