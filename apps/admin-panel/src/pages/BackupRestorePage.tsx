import { BackupRestorePageContent } from "@pages/BackupRestorePageContent";
import { BackupRestoreProvider } from "@providers/BackupRestoreProvider";
import type { ReactElement } from "react";

/** Administrator page for snapshots, restore, and three-pane text merging. */
export const BackupRestorePage = (): ReactElement => (
  <BackupRestoreProvider>
    <BackupRestorePageContent />
  </BackupRestoreProvider>
);
