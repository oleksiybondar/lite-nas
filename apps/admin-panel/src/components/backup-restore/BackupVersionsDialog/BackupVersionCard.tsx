import { BackupVersionActions } from "@components/backup-restore/BackupVersionsDialog/BackupVersionActions";
import { BackupVersionMetadata } from "@components/backup-restore/BackupVersionsDialog/BackupVersionMetadata";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import type { BackupDetails } from "@schemas/backup-restore";
import type { ReactElement } from "react";

/** Displays one backup's metadata and the actions available for its version. */
export const BackupVersionCard = ({
  item,
  selected,
  busy,
  sessionActive,
  onCompare,
  onRestore,
  onMerge,
  onDelete,
}: {
  item: BackupDetails;
  selected: boolean;
  busy: boolean;
  sessionActive: boolean;
  onCompare: () => void;
  onRestore: () => void;
  onMerge: () => void;
  onDelete: () => void;
}): ReactElement => (
  <Paper
    variant="outlined"
    data-test-class="backup-version-card"
    data-test-name={item.name || item.version}
    sx={{ borderColor: selected ? "primary.main" : undefined, p: 2 }}
  >
    <Stack direction={{ xs: "column", sm: "row" }} justifyContent="space-between" spacing={2}>
      <BackupVersionMetadata item={item} />
      <BackupVersionActions
        version={item.version}
        busy={busy}
        sessionActive={sessionActive}
        onCompare={onCompare}
        onRestore={onRestore}
        onMerge={onMerge}
        onDelete={onDelete}
      />
    </Stack>
  </Paper>
);
