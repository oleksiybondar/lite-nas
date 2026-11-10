import BackupIcon from "@mui/icons-material/Backup";
import ListAltIcon from "@mui/icons-material/ListAlt";
import RestoreIcon from "@mui/icons-material/Restore";
import CardActionArea from "@mui/material/CardActionArea";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { BackupSnapshotSummary } from "@schemas/backup-restore";
import type { ReactElement } from "react";
import { BackupApplicationSnapshots } from "./BackupApplicationSnapshots";

/** Displays one application with its backup actions and recent snapshot chips. */
export const BackupApplicationCard = ({
  application,
  snapshots,
  expanded,
  count,
  countsLoading,
  busy,
  sessionActive,
  getLabel,
  onOpen,
  onCreate,
  onRestore,
  onToggle,
}: {
  application: string;
  snapshots: BackupSnapshotSummary[];
  expanded: boolean;
  count: number;
  countsLoading: boolean;
  busy: boolean;
  sessionActive: boolean;
  getLabel: (key: string) => string;
  onOpen: () => void;
  onCreate: () => void;
  onRestore: () => void;
  onToggle: () => void;
}): ReactElement => {
  return (
    <Paper
      variant="outlined"
      sx={{ p: 2 }}
      data-test-class="backup-application-card"
      data-test-name={application}
    >
      <Stack direction="row" alignItems="center" spacing={1}>
        <CardActionArea
          data-testid={`backup-open-${application}`}
          onClick={onOpen}
          sx={{ flex: 1, borderRadius: 1, px: 1, py: 0.5 }}
        >
          <Stack direction="row" alignItems="baseline" flexWrap="wrap" columnGap={1}>
            <Typography variant="body1" fontWeight={600} color="primary.main">
              {application}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {countsLoading
                ? getLabel("admin.backup.countLoading")
                : getLabel("admin.backup.backupCount").replace("{count}", String(count))}
            </Typography>
          </Stack>
        </CardActionArea>
        <Stack direction="row" spacing={0.5}>
          <Tooltip title={getLabel("admin.backup.create")}>
            <span>
              <IconButton
                data-testid={`backup-create-${application}`}
                aria-label={getLabel("admin.backup.create")}
                disabled={busy || sessionActive}
                onClick={onCreate}
                color="primary"
              >
                <BackupIcon />
              </IconButton>
            </span>
          </Tooltip>
          <Tooltip title={getLabel("admin.backup.showBackups")}>
            <span>
              <IconButton
                data-testid={`backup-show-${application}`}
                aria-label={getLabel("admin.backup.showBackups")}
                disabled={busy}
                onClick={onOpen}
              >
                <ListAltIcon />
              </IconButton>
            </span>
          </Tooltip>
          <Tooltip title={getLabel("admin.backup.restoreByName")}>
            <span>
              <IconButton
                data-testid={`backup-restore-by-name-${application}`}
                aria-label={getLabel("admin.backup.restoreByName")}
                disabled={busy || sessionActive}
                onClick={onRestore}
              >
                <RestoreIcon />
              </IconButton>
            </span>
          </Tooltip>
        </Stack>
      </Stack>
      <BackupApplicationSnapshots
        application={application}
        snapshots={snapshots}
        expanded={expanded}
        getLabel={getLabel}
        onToggle={onToggle}
      />
    </Paper>
  );
};
