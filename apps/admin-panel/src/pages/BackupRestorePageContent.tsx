import { BackupVersionsDialog } from "@components/backup-restore/BackupVersionsDialog";
import { MergeDiffViewer } from "@components/backup-restore/MergeDiffViewer";
import { useBackupRestore } from "@hooks/useBackupRestore";
import { ConfirmActionDialog, StandardDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import BackupIcon from "@mui/icons-material/Backup";
import ListAltIcon from "@mui/icons-material/ListAlt";
import RefreshIcon from "@mui/icons-material/Refresh";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CardActionArea from "@mui/material/CardActionArea";
import IconButton from "@mui/material/IconButton";
import LinearProgress from "@mui/material/LinearProgress";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useState } from "react";
import { Navigate } from "react-router-dom";

/** Administrator page for snapshots, restore, and three-pane text merging. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page keeps backup and merge controls with their shared selection state.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The page keeps backup and merge controls with their shared selection state.
export const BackupRestorePageContent = (): ReactElement => {
  const { getLabel } = useLanguage();
  const [discardConfirmationOpen, setDiscardConfirmationOpen] = useState(false);
  const {
    allowed,
    applications,
    versionCounts,
    countsLoading,
    setBackupDetails,
    setSearch,
    applicationSearch,
    setApplicationSearch,
    setBackupsOpen,
    createOpen,
    setCreateOpen,
    backupTarget,
    setApplication,
    setVersions,
    setVersion,
    name,
    setName,
    description,
    setDescription,
    pendingApply,
    setPendingApply,
    pendingRestore,
    setPendingRestore,
    pendingDelete,
    setPendingDelete,
    session,
    selectedPath,
    setSelectedPath,
    contents,
    setContents,
    busy,
    error,
    message,
    openCreate,
    createBackup,
    restoreBackup,
    deleteBackup,
    finishMerge,
    refresh,
  } = useBackupRestore();
  const visibleApplications = applications.filter((app) =>
    app.toLocaleLowerCase().includes(applicationSearch.toLocaleLowerCase()),
  );
  const openBackups = (app: string): void => {
    setApplication(app);
    setVersion("");
    setVersions([]);
    setBackupDetails([]);
    setSearch("");
    setBackupsOpen(true);
  };

  if (!allowed) return <Navigate replace to="/" />;

  const content = contents[selectedPath];
  return (
    <Stack data-testid="backup-restore-page" spacing={3}>
      <Typography variant="h1" data-testid="backup-restore-title">
        {getLabel("admin.backup.title")}
      </Typography>
      <Typography>{getLabel("admin.backup.summary")}</Typography>
      {busy && (
        <Box aria-live="polite" data-testid="backup-operation-progress">
          <Typography>{getLabel("admin.backup.operationInProgress")}</Typography>
          <LinearProgress />
        </Box>
      )}
      {error && (
        <Typography role="alert" color="error">
          {error}
        </Typography>
      )}
      {message && <Typography role="status">{message}</Typography>}
      <Stack direction={{ xs: "column", sm: "row" }} alignItems="center" spacing={1}>
        <TextField
          data-testid="backup-application-search"
          label={getLabel("admin.backup.searchApplications")}
          value={applicationSearch}
          onChange={(event) => setApplicationSearch(event.target.value)}
          fullWidth
          sx={{ flex: 1 }}
        />
        <Tooltip title={getLabel("admin.backup.refresh")}>
          <span>
            <IconButton
              data-testid="backup-refresh"
              aria-label={getLabel("admin.backup.refresh")}
              disabled={busy || !!session}
              onClick={() => void refresh()}
              color="primary"
            >
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
        <Tooltip title={getLabel("admin.backup.backupAll")}>
          <span>
            <IconButton
              data-testid="backup-all"
              aria-label={getLabel("admin.backup.backupAll")}
              disabled={busy || !!session || applications.length === 0}
              onClick={() => openCreate("all")}
              color="primary"
            >
              <BackupIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Stack>
      {applications.length === 0 && (
        <Typography>{getLabel("admin.backup.noApplications")}</Typography>
      )}
      {applications.length > 0 && visibleApplications.length === 0 && (
        <Typography>{getLabel("admin.backup.noApplicationsMatch")}</Typography>
      )}
      <Stack spacing={1}>
        {visibleApplications.map((app) => (
          <Paper
            key={app}
            variant="outlined"
            sx={{ p: 2 }}
            data-test-class="backup-application-card"
            data-test-name={app}
          >
            <Stack direction="row" alignItems="center" spacing={1}>
              <CardActionArea
                data-testid={`backup-open-${app}`}
                onClick={() => openBackups(app)}
                sx={{ flex: 1, borderRadius: 1, px: 1, py: 0.5 }}
              >
                <Stack direction="row" alignItems="baseline" flexWrap="wrap" columnGap={1}>
                  <Typography variant="body1" fontWeight={600} color="primary.main">
                    {app}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {countsLoading
                      ? getLabel("admin.backup.countLoading")
                      : getLabel("admin.backup.backupCount").replace(
                          "{count}",
                          String(versionCounts[app] ?? 0),
                        )}
                  </Typography>
                </Stack>
              </CardActionArea>
              <Stack direction="row" spacing={0.5}>
                <Tooltip title={getLabel("admin.backup.create")}>
                  <span>
                    <IconButton
                      data-testid={`backup-create-${app}`}
                      aria-label={getLabel("admin.backup.create")}
                      disabled={busy || !!session}
                      onClick={() => openCreate(app)}
                      color="primary"
                    >
                      <BackupIcon />
                    </IconButton>
                  </span>
                </Tooltip>
                <Tooltip title={getLabel("admin.backup.showBackups")}>
                  <span>
                    <IconButton
                      data-testid={`backup-show-${app}`}
                      aria-label={getLabel("admin.backup.showBackups")}
                      disabled={busy}
                      onClick={() => openBackups(app)}
                    >
                      <ListAltIcon />
                    </IconButton>
                  </span>
                </Tooltip>
              </Stack>
            </Stack>
          </Paper>
        ))}
      </Stack>
      <BackupVersionsDialog />
      <ConfirmActionDialog
        open={pendingRestore !== null}
        title={getLabel("admin.backup.restore")}
        description={getLabel("admin.backup.confirmRestore")}
        confirmLabel={getLabel("admin.backup.restore")}
        onClose={() => setPendingRestore(null)}
        onConfirm={() =>
          pendingRestore
            ? restoreBackup(pendingRestore.application, pendingRestore.version)
            : Promise.resolve()
        }
      />
      <ConfirmActionDialog
        open={pendingDelete !== null}
        title={getLabel("admin.backup.delete")}
        description={getLabel("admin.backup.confirmDelete")}
        confirmLabel={getLabel("admin.backup.delete")}
        onClose={() => setPendingDelete(null)}
        onConfirm={() =>
          pendingDelete
            ? deleteBackup(pendingDelete.application, pendingDelete.version)
            : Promise.resolve()
        }
      />
      <ConfirmActionDialog
        open={pendingApply}
        title={getLabel("admin.backup.apply")}
        description={getLabel("admin.backup.confirmApply")}
        confirmLabel={getLabel("admin.backup.apply")}
        onClose={() => setPendingApply(false)}
        onConfirm={() => finishMerge("apply")}
      />
      <StandardDialog
        open={createOpen}
        fullWidth
        maxWidth="sm"
        title={getLabel("admin.backup.create")}
        closeLabel={getLabel("dialogs.close")}
        onClose={() => !busy && setCreateOpen(false)}
        closeDisabled={busy}
      >
        <Stack spacing={2}>
          <Typography>
            {backupTarget === "all" ? getLabel("admin.backup.backupAllTarget") : backupTarget}
          </Typography>
          <TextField
            data-testid="backup-name"
            label={getLabel("admin.backup.name")}
            value={name}
            onChange={(event) => setName(event.target.value)}
            inputProps={{ maxLength: 128 }}
          />
          <TextField
            data-testid="backup-description"
            label={getLabel("admin.backup.description")}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            inputProps={{ maxLength: 1024 }}
            multiline
          />
          <Button
            data-testid="backup-create-confirm"
            variant="contained"
            disabled={busy || !backupTarget}
            onClick={createBackup}
          >
            {getLabel("admin.backup.create")}
          </Button>
        </Stack>
      </StandardDialog>
      <StandardDialog
        open={Boolean(session)}
        fullWidth
        maxWidth="xl"
        title={getLabel("admin.backup.editor")}
        closeLabel={getLabel("dialogs.close")}
        closeDisabled={busy}
        onClose={() => !busy && setDiscardConfirmationOpen(true)}
        contentProps={{ dividers: true }}
      >
        {session && (
          <Paper sx={{ p: 3 }}>
            <Stack spacing={2}>
              <Typography variant="h2">{getLabel("admin.backup.editor")}</Typography>
              <TextField
                select
                data-testid="merge-file"
                label={getLabel("admin.backup.file")}
                value={selectedPath}
                onChange={(event) => setSelectedPath(event.target.value)}
                disabled={busy}
              >
                {session.files.map((file) => (
                  <MenuItem key={file.path} value={file.path}>
                    {file.path} ({file.status})
                  </MenuItem>
                ))}
              </TextField>
              {content ? (
                <Stack spacing={2}>
                  <MergeDiffViewer
                    oldValue={content.backup}
                    newValue={content.result}
                    currentValue={content.current}
                    disabled={busy}
                    onChange={(result) =>
                      setContents((previous) => ({
                        ...previous,
                        [selectedPath]: { ...content, result },
                      }))
                    }
                    summary={selectedPath}
                    leftTitle={getLabel("admin.backup.backupPane")}
                    resultTitle={getLabel("admin.backup.resultPane")}
                    rightTitle={getLabel("admin.backup.currentPane")}
                  />
                </Stack>
              ) : (
                <Typography>{getLabel("common.loading")}</Typography>
              )}
              <Typography>{getLabel("admin.backup.diffLegend")}</Typography>
              <Stack direction="row" spacing={1}>
                <Button
                  data-testid="merge-use-backup"
                  disabled={!content || busy}
                  onClick={() =>
                    content &&
                    setContents((previous) => ({
                      ...previous,
                      [selectedPath]: { ...content, result: content.backup },
                    }))
                  }
                >
                  {getLabel("admin.backup.useBackup")}
                </Button>
                <Button
                  data-testid="merge-use-current"
                  disabled={!content || busy}
                  onClick={() =>
                    content &&
                    setContents((previous) => ({
                      ...previous,
                      [selectedPath]: { ...content, result: content.current },
                    }))
                  }
                >
                  {getLabel("admin.backup.useCurrent")}
                </Button>
              </Stack>
              <Stack direction="row" spacing={1}>
                <Button
                  data-testid="merge-apply"
                  disabled={busy || !session.apply_available}
                  onClick={() => setPendingApply(true)}
                >
                  {getLabel("admin.backup.apply")}
                </Button>
                <Button
                  data-testid="merge-save"
                  disabled={busy}
                  onClick={() => void finishMerge("save")}
                >
                  {getLabel("admin.backup.save")}
                </Button>
                <Button
                  data-testid="merge-discard"
                  disabled={busy}
                  onClick={() => setDiscardConfirmationOpen(true)}
                >
                  {getLabel("admin.backup.discard")}
                </Button>
              </Stack>
            </Stack>
          </Paper>
        )}
      </StandardDialog>
      <ConfirmActionDialog
        open={discardConfirmationOpen}
        title={getLabel("dialogs.areYouSure")}
        description={getLabel("admin.backup.confirmDiscard")}
        confirmLabel={getLabel("admin.backup.discard")}
        requireYes={false}
        onClose={() => setDiscardConfirmationOpen(false)}
        onConfirm={() => finishMerge("discard")}
      />
    </Stack>
  );
};
