import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import Stack from "@mui/material/Stack";
import type { BackupSnapshotSummary } from "@schemas/backup-restore";
import type { ReactElement } from "react";

/** Shows recent snapshot labels and lets users expand the full snapshot list. */
export const BackupApplicationSnapshots = ({
  application,
  snapshots,
  expanded,
  getLabel,
  onToggle,
}: {
  application: string;
  snapshots: BackupSnapshotSummary[];
  expanded: boolean;
  getLabel: (key: string) => string;
  onToggle: () => void;
}): ReactElement | null => {
  if (snapshots.length === 0) return null;
  const visibleSnapshots = expanded ? snapshots : snapshots.slice(0, 4);
  return (
    <Stack direction="row" alignItems="center" flexWrap="wrap" gap={0.75} sx={{ mt: 1 }}>
      {visibleSnapshots.map((snapshot) => (
        <Chip
          key={snapshot.version}
          size="small"
          label={snapshot.name || snapshot.version}
          title={snapshot.name ? snapshot.version : undefined}
          data-testid={`backup-snapshot-${application}-${snapshot.version}`}
        />
      ))}
      {snapshots.length > 4 && (
        <Button
          data-testid={`backup-snapshots-toggle-${application}`}
          size="small"
          onClick={onToggle}
        >
          {getLabel(
            expanded ? "admin.backup.showFewerSnapshots" : "admin.backup.showAllSnapshots",
          ).replace("{count}", String(snapshots.length))}
        </Button>
      )}
    </Stack>
  );
};
