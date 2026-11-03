import type { BackupRestoreContextValue } from "@dto/backup-restore";
import { createContext } from "react";

/** Context for backup inventory, restore, and merge state scoped to its page. */
export const BackupRestoreContext = createContext<BackupRestoreContextValue | undefined>(undefined);
