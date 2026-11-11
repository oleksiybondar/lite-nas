import { BackupVersionsDialog } from "@components/backup-restore/BackupVersionsDialog";
import { MergeDiffViewer } from "@components/backup-restore/MergeDiffViewer";
import { useBackupRestore } from "@hooks/useBackupRestore";
import { ConfirmActionDialog, StandardDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import BackupIcon from "@mui/icons-material/Backup";
import DeleteSweepIcon from "@mui/icons-material/DeleteSweep";
import RefreshIcon from "@mui/icons-material/Refresh";
import RestoreIcon from "@mui/icons-material/Restore";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import LinearProgress from "@mui/material/LinearProgress";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import Snackbar from "@mui/material/Snackbar";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import { BackupApplicationCard } from "@pages/BackupApplicationCard";
import { restoreModeSchema } from "@schemas/backup-restore";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import { Navigate } from "react-router-dom";

type BackupNotification = { id: string; text: string; failed: boolean };

/** Resolves the label for a completed backup task result. */
const backupNotificationLabel = (status: string): string => {
  if (status === "failed") return "admin.backup.notificationFailed";
  if (status === "skipped") return "admin.backup.notificationSkipped";
  return "admin.backup.notificationCompleted";
};

/** Returns newly completed results and records which applications were shown. */
const newBackupNotifications = (
  results: ReturnType<typeof useBackupRestore>["backupTaskResults"],
  seen: Set<string>,
  getLabel: (key: string) => string,
): BackupNotification[] => {
  const fresh: BackupNotification[] = [];
  for (const result of results) {
    if (seen.has(result.application)) continue;
    seen.add(result.application);
    fresh.push({
      id: result.application,
      text: `${result.application}: ${getLabel(backupNotificationLabel(result.status))}`,
      failed: result.status === "failed",
    });
  }
  return fresh;
};

/** Administrator page for snapshots, restore, and three-pane text merging. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page keeps backup and merge controls with their shared selection state.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The page keeps backup and merge controls with their shared selection state.
export const BackupRestorePageContent = (): ReactElement => {
  const { getLabel } = useLanguage();
  const [discardConfirmationOpen, setDiscardConfirmationOpen] = useState(false);
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [bulkDeleteConfirmationOpen, setBulkDeleteConfirmationOpen] = useState(false);
  const [bulkDeleteName, setBulkDeleteName] = useState("");
  const [expandedSnapshotApplications, setExpandedSnapshotApplications] = useState<Set<string>>(
    () => new Set(),
  );
  const {
    allowed,
    applications,
    visibleApplications,
    versionCounts,
    snapshotsByApplication,
    availableBackupNames,
    availableRestoreTargets,
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
    backupProgress,
    backupTaskResults,
    error,
    message,
    openCreate,
    createBackup,
    restoreBackup,
    namedRestoreTarget,
    namedRestoreName,
    setNamedRestoreTarget,
    setNamedRestoreName,
    restoreMode,
    setRestoreMode,
    openNamedRestore,
    restoreByName,
    deleteBackup,
    deleteBackupsByName,
    finishMerge,
    refresh,
  } = useBackupRestore();
  const [notifications, setNotifications] = useState<BackupNotification[]>([]);
  const notifiedResults = useRef(new Set<string>());
  useEffect(() => {
    const fresh = newBackupNotifications(backupTaskResults, notifiedResults.current, getLabel);
    if (fresh.length > 0) setNotifications((current) => [...current, ...fresh]);
    if (!busy && backupTaskResults.length === 0) notifiedResults.current.clear();
  }, [backupTaskResults, busy, getLabel]);
  const openBackups = (app: string): void => {
    setApplication(app);
    setVersion("");
    setVersions([]);
    setBackupDetails([]);
    setSearch("");
    setBackupsOpen(true);
  };
  const openBulkDelete = (): void => {
    setBulkDeleteName((previous) =>
      availableBackupNames.includes(previous) ? previous : availableBackupNames[0] || "",
    );
    setBulkDeleteOpen(true);
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
          {backupTaskResults.length > 0 && applications.length > 0 && (
            <Typography data-testid="backup-task-counts">
              {getLabel("admin.backup.liveCounts")
                .replace("{processed}", String(backupTaskResults.length))
                .replace("{total}", String(applications.length))
                .replace(
                  "{completed}",
                  String(backupTaskResults.filter((item) => item.status === "completed").length),
                )
                .replace(
                  "{failed}",
                  String(backupTaskResults.filter((item) => item.status === "failed").length),
                )}
            </Typography>
          )}
          <LinearProgress
            variant={backupProgress === null ? "indeterminate" : "determinate"}
            value={backupProgress ?? undefined}
          />
        </Box>
      )}
      {error && (
        <Typography role="alert" color="error" sx={{ whiteSpace: "pre-line" }}>
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
        <Tooltip title={getLabel("admin.backup.bulkDeleteByName")}>
          <span>
            <IconButton
              data-testid="backup-delete-all-by-name"
              aria-label={getLabel("admin.backup.bulkDeleteByName")}
              disabled={busy || !!session || availableBackupNames.length === 0}
              onClick={openBulkDelete}
              color="error"
            >
              <DeleteSweepIcon />
            </IconButton>
          </span>
        </Tooltip>
        <Tooltip title={getLabel("admin.backup.restoreAllByName")}>
          <span>
            <IconButton
              data-testid="backup-restore-all-by-name"
              aria-label={getLabel("admin.backup.restoreAllByName")}
              disabled={busy || !!session || applications.length === 0}
              onClick={() => openNamedRestore("all")}
              color="primary"
            >
              <RestoreIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Stack>
      <Stack
        sx={{ position: "fixed", right: 16, bottom: 16, zIndex: (theme) => theme.zIndex.snackbar }}
        spacing={1}
      >
        {notifications.map((notification) => (
          <Snackbar
            key={notification.id}
            open
            autoHideDuration={4000}
            onClose={() =>
              setNotifications((current) => current.filter((item) => item.id !== notification.id))
            }
          >
            <Alert
              severity={notification.failed ? "error" : "success"}
              onClose={() =>
                setNotifications((current) => current.filter((item) => item.id !== notification.id))
              }
            >
              {notification.text}
            </Alert>
          </Snackbar>
        ))}
      </Stack>
      {applications.length === 0 && (
        <Typography>{getLabel("admin.backup.noApplications")}</Typography>
      )}
      {applications.length > 0 && visibleApplications.length === 0 && (
        <Typography>{getLabel("admin.backup.noApplicationsMatch")}</Typography>
      )}
      <Stack spacing={1}>
        {visibleApplications.map((app) => (
          <BackupApplicationCard
            key={app}
            application={app}
            snapshots={snapshotsByApplication[app] ?? []}
            expanded={expandedSnapshotApplications.has(app)}
            count={versionCounts[app] ?? 0}
            countsLoading={countsLoading}
            busy={busy}
            sessionActive={session !== null}
            getLabel={getLabel}
            onOpen={() => openBackups(app)}
            onCreate={() => openCreate(app)}
            onRestore={() => openNamedRestore(app)}
            onToggle={() =>
              setExpandedSnapshotApplications((previous) => {
                const next = new Set(previous);
                if (next.has(app)) next.delete(app);
                else next.add(app);
                return next;
              })
            }
          />
        ))}
      </Stack>
      <BackupVersionsDialog />
      <StandardDialog
        open={namedRestoreTarget !== null}
        fullWidth
        maxWidth="sm"
        title={
          namedRestoreTarget === "all"
            ? getLabel("admin.backup.restoreAllByName")
            : getLabel("admin.backup.restoreByName")
        }
        closeLabel={getLabel("dialogs.close")}
        closeDisabled={busy}
        onClose={() => !busy && setNamedRestoreTarget(null)}
      >
        <Stack spacing={2}>
          <Typography>
            {namedRestoreTarget === "all"
              ? getLabel("admin.backup.restoreAllByNameHint")
              : namedRestoreTarget}
          </Typography>
          {namedRestoreTarget === "all" ? (
            <TextField
              select
              data-testid="backup-restore-name"
              label={getLabel("admin.backup.name")}
              value={namedRestoreName}
              onChange={(event) => setNamedRestoreName(event.target.value)}
              disabled={busy || availableRestoreTargets.length === 0}
              fullWidth
            >
              {availableRestoreTargets.map((restoreTarget) => (
                <MenuItem key={restoreTarget} value={restoreTarget}>
                  {restoreTarget}
                </MenuItem>
              ))}
            </TextField>
          ) : (
            <TextField
              data-testid="backup-restore-name"
              label={getLabel("admin.backup.name")}
              value={namedRestoreName}
              onChange={(event) => setNamedRestoreName(event.target.value)}
              inputProps={{ maxLength: 128 }}
              fullWidth
            />
          )}
          <TextField
            select
            data-testid="backup-named-restore-mode"
            label={getLabel("admin.backup.restoreMode")}
            value={restoreMode}
            onChange={(event) => setRestoreMode(restoreModeSchema.parse(event.target.value))}
          >
            <MenuItem value="merge">{getLabel("admin.backup.mergeMode")}</MenuItem>
            <MenuItem value="replace">{getLabel("admin.backup.replaceMode")}</MenuItem>
          </TextField>
          <Button
            data-testid="backup-restore-name-confirm"
            variant="contained"
            disabled={busy || namedRestoreName.length === 0}
            onClick={() => void restoreByName()}
          >
            {getLabel("admin.backup.restore")}
          </Button>
        </Stack>
      </StandardDialog>
      <StandardDialog
        open={bulkDeleteOpen}
        fullWidth
        maxWidth="sm"
        title={getLabel("admin.backup.bulkDeleteByName")}
        closeLabel={getLabel("dialogs.close")}
        closeDisabled={busy}
        onClose={() => !busy && setBulkDeleteOpen(false)}
      >
        <Stack spacing={2}>
          <Typography>{getLabel("admin.backup.bulkDeleteHint")}</Typography>
          <TextField
            select
            data-testid="backup-delete-name"
            label={getLabel("admin.backup.name")}
            value={bulkDeleteName}
            onChange={(event) => setBulkDeleteName(event.target.value)}
            disabled={busy || availableBackupNames.length === 0}
          >
            {availableBackupNames.map((backupName) => (
              <MenuItem key={backupName} value={backupName}>
                {backupName}
              </MenuItem>
            ))}
          </TextField>
          <Button
            data-testid="backup-delete-name-continue"
            variant="contained"
            color="error"
            disabled={busy || !bulkDeleteName}
            onClick={() => setBulkDeleteConfirmationOpen(true)}
          >
            {getLabel("admin.backup.bulkDeleteAction")}
          </Button>
        </Stack>
      </StandardDialog>
      <ConfirmActionDialog
        open={bulkDeleteConfirmationOpen}
        title={getLabel("admin.backup.bulkDeleteAction")}
        description={getLabel("admin.backup.bulkDeleteConfirm").replace("{name}", bulkDeleteName)}
        confirmLabel={getLabel("admin.backup.bulkDeleteAction")}
        onClose={() => setBulkDeleteConfirmationOpen(false)}
        onConfirm={async () => {
          if (!bulkDeleteName) return;
          await deleteBackupsByName(bulkDeleteName);
          setBulkDeleteConfirmationOpen(false);
          setBulkDeleteOpen(false);
          setBulkDeleteName("");
        }}
      />
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
