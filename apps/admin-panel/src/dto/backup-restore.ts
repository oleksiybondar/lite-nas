import type { BackupDetails, BackupDifference, MergeSession } from "@schemas/backup-restore";
import type { Dispatch, SetStateAction } from "react";

/** Text values shown in the three panes of the merge editor. */
export type BackupFileContent = { backup: string; current: string; result: string };

/** A selected backup version used by restore and delete confirmation dialogs. */
export type BackupVersionTarget = { application: string; version: string };

/** State and operations exposed by the backup and restore page provider. */
export type BackupRestoreContextValue = {
  allowed: boolean;
  applications: string[];
  versionCounts: Record<string, number>;
  countsLoading: boolean;
  backupDetails: BackupDetails[];
  versionsLoading: boolean;
  detailsLoading: boolean;
  search: string;
  applicationSearch: string;
  backupsOpen: boolean;
  createOpen: boolean;
  backupTarget: string | "all" | null;
  application: string;
  versions: string[];
  version: string;
  name: string;
  description: string;
  restoreMode: "replace" | "merge";
  pendingRestore: BackupVersionTarget | null;
  pendingDelete: BackupVersionTarget | null;
  pendingApply: boolean;
  differences: BackupDifference[];
  session: MergeSession | null;
  selectedPath: string;
  contents: Record<string, BackupFileContent>;
  busy: boolean;
  error: string;
  message: string;
  setApplications: Dispatch<SetStateAction<string[]>>;
  setVersionCounts: Dispatch<SetStateAction<Record<string, number>>>;
  setBackupDetails: Dispatch<SetStateAction<BackupDetails[]>>;
  setSearch: Dispatch<SetStateAction<string>>;
  setApplicationSearch: Dispatch<SetStateAction<string>>;
  setBackupsOpen: Dispatch<SetStateAction<boolean>>;
  setCreateOpen: Dispatch<SetStateAction<boolean>>;
  setBackupTarget: Dispatch<SetStateAction<string | "all" | null>>;
  setApplication: Dispatch<SetStateAction<string>>;
  setVersions: Dispatch<SetStateAction<string[]>>;
  setVersion: Dispatch<SetStateAction<string>>;
  setName: Dispatch<SetStateAction<string>>;
  setDescription: Dispatch<SetStateAction<string>>;
  setRestoreMode: Dispatch<SetStateAction<"replace" | "merge">>;
  setPendingRestore: Dispatch<SetStateAction<BackupVersionTarget | null>>;
  setPendingDelete: Dispatch<SetStateAction<BackupVersionTarget | null>>;
  setPendingApply: Dispatch<SetStateAction<boolean>>;
  setSession: Dispatch<SetStateAction<MergeSession | null>>;
  setSelectedPath: Dispatch<SetStateAction<string>>;
  setContents: Dispatch<SetStateAction<Record<string, BackupFileContent>>>;
  openCreate: (target: string | "all") => void;
  createBackup: () => void;
  restoreBackup: (application: string, version: string) => Promise<void>;
  namedRestoreTarget: string | "all" | null;
  namedRestoreName: string;
  setNamedRestoreTarget: Dispatch<SetStateAction<string | "all" | null>>;
  setNamedRestoreName: Dispatch<SetStateAction<string>>;
  openNamedRestore: (target: string | "all") => void;
  restoreByName: () => Promise<void>;
  deleteBackup: (application: string, version: string) => Promise<void>;
  startMerge: (application: string, version: string) => void;
  finishMerge: (action: "apply" | "save" | "discard") => Promise<void>;
  refresh: () => Promise<void>;
};
