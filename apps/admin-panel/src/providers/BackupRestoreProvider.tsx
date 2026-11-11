import { BackupRestoreContext } from "@contexts/backup-restore-context";
import type {
  BackupFileContent,
  BackupRestoreContextValue,
  BackupVersionTarget,
} from "@dto/backup-restore";
import { backupPath, readData, readMergeText, writeMergeText } from "@helpers/backup-restore";
import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import {
  applicationListResponseSchema,
  type BackupDetails,
  type BackupDifference,
  type BackupSnapshotSummary,
  backupDeleteAllByNameResponseSchema,
  backupDeleteResponseSchema,
  backupDetailsResponseSchema,
  backupMetadataSchema,
  backupTaskStartResponseSchema,
  backupTaskStatusResponseSchema,
  diffResponseSchema,
  type MergeSession,
  mergeActionResponseSchema,
  mergeSessionResponseSchema,
  restoreAllByNameResponseSchema,
  restoreByNameRequestSchema,
  restoreByNameResponseSchema,
  restoreModeSchema,
  restoreResponseSchema,
  validateUFWMergeText,
  versionListResponseSchema,
  versionResponseSchema,
} from "@schemas/backup-restore";
import type { PropsWithChildren, ReactElement } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { z } from "zod";

const storedSessionKey = "lite-nas-backup-merge-session";
const storedBackupTaskKey = "lite-nas-backup-task";
const storedBackupTaskSchema = z.string().regex(/^[a-fA-F0-9]{32}$/);
const storedSessionSchema = z.object({
  id: z.string().regex(/^[a-fA-F0-9]{32}$/),
  application: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/),
  version: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/),
});

/** Identifies task states that no longer need polling. */
const isTerminalBackupTask = (state: string): boolean =>
  state === "succeeded" || state === "failed" || state === "cancelled";

/** Validates the resumable task identifier and clears malformed browser state. */
const readStoredBackupTaskID = (): string | null => {
  const raw = window.localStorage.getItem(storedBackupTaskKey);
  if (!raw) return null;
  const stored = storedBackupTaskSchema.safeParse(raw);
  if (stored.success) return stored.data;
  window.localStorage.removeItem(storedBackupTaskKey);
  return null;
};

type BackupTaskStatus = z.infer<typeof backupTaskStatusResponseSchema>["data"];

/** Polls a backup task until it finishes, publishing progress only while its owner remains active. */
const pollBackupTaskStatus = async (
  taskID: string,
  isActive: () => boolean,
  readStatus: (id: string) => Promise<BackupTaskStatus>,
  publish: {
    progress: (percent: number | null) => void;
    results: (results: NonNullable<BackupTaskStatus["results"]>) => void;
  },
): Promise<BackupTaskStatus> => {
  const status = await readStatus(taskID);
  if (!isActive()) return status;
  publish.progress(status.progress?.percent ?? null);
  publish.results(status.results ?? []);
  if (isTerminalBackupTask(status.state)) {
    window.localStorage.removeItem(storedBackupTaskKey);
    return status;
  }
  await new Promise((resolve) => window.setTimeout(resolve, 800));
  return pollBackupTaskStatus(taskID, isActive, readStatus, publish);
};

/** Summarizes terminal backup outcomes for the page's message and error areas. */
const describeBackupTaskStatus = (
  status: BackupTaskStatus,
  label: (key: string) => string,
): { message: string; error: string } => {
  if (status.state !== "succeeded")
    return { message: "", error: status.error || label("admin.backup.createFailed") };
  if (status.application !== "all") return { message: label("admin.backup.created"), error: "" };
  const results = status.results ?? [];
  const completed = results.filter((result) => result.status === "completed").length;
  const skipped = results.filter((result) => result.status === "skipped").length;
  const failures = results.filter((result) => result.status === "failed");
  const message = label("admin.backup.allCompleted")
    .replace("{completed}", String(completed))
    .replace("{skipped}", String(skipped))
    .replace("{failed}", String(failures.length));
  const error = failures.length
    ? [
        label("admin.backup.failedTargets"),
        ...failures.map((result) => `${result.application}: ${result.error || "—"}`),
      ].join("\n")
    : "";
  return { message, error };
};

/** Keeps the current version when present and otherwise selects the first available version. */
const selectExistingVersion = (current: string, versions: string[]): string =>
  versions.includes(current) ? current : versions[0] || "";

/** Owns backup inventory, restore actions, and merge editor state for its page. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider keeps ordered backup requests and their React state updates together.
export const BackupRestoreProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { get, post, put, delete: del } = useApi();
  const { requireAdmin } = useRbac();
  const { getLabel } = useLanguage();
  const getRef = useRef(get);
  const labelRef = useRef(getLabel);
  getRef.current = get;
  labelRef.current = getLabel;
  const [applications, setApplications] = useState<string[]>([]);
  const [versionCounts, setVersionCounts] = useState<Record<string, number>>({});
  const [snapshotsByApplication, setSnapshotsByApplication] = useState<
    Record<string, BackupSnapshotSummary[]>
  >({});
  const [countsLoading, setCountsLoading] = useState(true);
  const [backupDetails, setBackupDetails] = useState<BackupDetails[]>([]);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [applicationSearch, setApplicationSearch] = useState("");
  const [backupsOpen, setBackupsOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [backupTarget, setBackupTarget] = useState<string | "all" | null>(null);
  const [application, setApplication] = useState("");
  const [versions, setVersions] = useState<string[]>([]);
  const [version, setVersion] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [restoreMode, setRestoreMode] = useState<"replace" | "merge">("replace");
  const [namedRestoreTarget, setNamedRestoreTarget] = useState<string | "all" | null>(null);
  const [namedRestoreName, setNamedRestoreName] = useState("");
  const [pendingRestore, setPendingRestore] = useState<BackupVersionTarget | null>(null);
  const [pendingDelete, setPendingDelete] = useState<BackupVersionTarget | null>(null);
  const [pendingApply, setPendingApply] = useState(false);
  const [differences, setDifferences] = useState<BackupDifference[]>([]);
  const [session, setSession] = useState<MergeSession | null>(null);
  const [selectedPath, setSelectedPath] = useState("");
  const [contents, setContents] = useState<Record<string, BackupFileContent>>({});
  const [busy, setBusy] = useState(false);
  const [backupProgress, setBackupProgress] = useState<number | null>(null);
  const [backupTaskResults, setBackupTaskResults] = useState<
    NonNullable<BackupTaskStatus["results"]>
  >([]);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const allowed = requireAdmin();

  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Session restoration validates the saved state before reuse.
  useEffect(() => {
    if (!allowed) return;
    const raw = window.sessionStorage.getItem(storedSessionKey);
    if (!raw) return;
    let stored: ReturnType<typeof storedSessionSchema.safeParse>;
    try {
      stored = storedSessionSchema.safeParse(JSON.parse(raw));
    } catch {
      window.sessionStorage.removeItem(storedSessionKey);
      return;
    }
    if (!stored.success) return;
    let active = true;
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
    void (async () => {
      try {
        const resumed = await readData(
          await getRef.current(`/api/backup-restore/merge-sessions/${stored.data.id}`).execute(),
          mergeSessionResponseSchema,
        );
        if (active) {
          setSession(resumed);
          setApplication(stored.data.application);
          setVersion(stored.data.version);
          setSelectedPath(resumed.files[0]?.path || "");
        }
      } catch {
        window.sessionStorage.removeItem(storedSessionKey);
      }
    })();
    return () => {
      active = false;
    };
  }, [allowed]);

  useEffect(() => {
    if (!allowed || applications.length === 0) return;
    let active = true;
    setCountsLoading(true);
    void Promise.all(
      applications.map(async (app) => {
        const list = await readData(
          await getRef.current(`${backupPath(app)}/versions`).execute(),
          versionListResponseSchema,
        );
        return [app, list] as const;
      }),
    )
      .then((counts) => {
        if (active) {
          setVersionCounts(
            Object.fromEntries(counts.map(([app, list]) => [app, list.versions.length])),
          );
          setSnapshotsByApplication(
            Object.fromEntries(counts.map(([app, list]) => [app, list.snapshots])),
          );
          setCountsLoading(false);
        }
      })
      .catch(() => {
        if (active) {
          setError(labelRef.current("admin.backup.loadFailed"));
          setCountsLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [allowed, applications]);

  useEffect(() => {
    if (!allowed) return;
    let active = true;
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
    void (async () => {
      try {
        const apps = await readData(
          await getRef.current("/api/backup-restore/applications").execute(),
          applicationListResponseSchema,
        );
        if (active) {
          setApplications(apps);
          setApplication((previous) => previous || apps[0] || "");
        }
      } catch {
        if (active) setError(labelRef.current("admin.backup.loadFailed"));
      }
    })();
    return () => {
      active = false;
    };
  }, [allowed]);

  useEffect(() => {
    if (!allowed || !application || (!backupsOpen && !session)) {
      setVersionsLoading(false);
      return;
    }
    let active = true;
    setVersionsLoading(true);
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
    void (async () => {
      try {
        const list = await readData(
          await getRef.current(`${backupPath(application)}/versions`).execute(),
          versionListResponseSchema,
        );
        if (active) {
          setVersions(list.versions);
          setSnapshotsByApplication((previous) => ({ ...previous, [application]: list.snapshots }));
          // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
          setVersion((previous) =>
            list.versions.includes(previous) ? previous : list.versions[0] || "",
          );
        }
      } catch {
        if (active) setError(labelRef.current("admin.backup.loadFailed"));
      } finally {
        if (active) setVersionsLoading(false);
      }
    })();
    return () => {
      active = false;
    };
  }, [allowed, application, backupsOpen, session]);

  useEffect(() => {
    if (!allowed || !backupsOpen || !application || versions.length === 0) {
      setBackupDetails([]);
      setDetailsLoading(false);
      return;
    }
    let active = true;
    setDetailsLoading(true);
    void Promise.all(
      versions.map(async (item) =>
        readData(
          await getRef.current(`${backupPath(application, item)}`).execute(),
          backupDetailsResponseSchema,
        ),
      ),
    )
      .then((details) => {
        if (active) {
          setBackupDetails(details);
          setDetailsLoading(false);
        }
      })
      .catch(() => {
        if (active) {
          setError(labelRef.current("admin.backup.loadFailed"));
          setDetailsLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [allowed, backupsOpen, application, versions]);

  useEffect(() => {
    if (!allowed || !application || !version) {
      setDifferences([]);
      return;
    }
    let active = true;
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
    void (async () => {
      try {
        const list = await readData(
          await getRef.current(`${backupPath(application, version)}/diff`).execute(),
          diffResponseSchema,
        );
        if (active) setDifferences(list);
      } catch {
        if (active) setError(labelRef.current("admin.backup.diffFailed"));
      }
    })();
    return () => {
      active = false;
    };
  }, [allowed, application, version]);

  useEffect(() => {
    if (!session || !selectedPath || contents[selectedPath]) return;
    let active = true;
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
    void (async () => {
      try {
        const [backup, current, result] = await Promise.all([
          readMergeText(getRef.current, session.session_id, selectedPath, "snapshot"),
          readMergeText(getRef.current, session.session_id, selectedPath, "current"),
          readMergeText(getRef.current, session.session_id, selectedPath, "result"),
        ]);
        if (active)
          setContents((previous) => ({ ...previous, [selectedPath]: { backup, current, result } }));
      } catch {
        if (active) setError(labelRef.current("admin.backup.readFailed"));
      }
    })();
    return () => {
      active = false;
    };
  }, [session, selectedPath, contents]);

  const run = async (
    action: () => Promise<void>,
    failure: string,
    propagateError = false,
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
  ): Promise<void> => {
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await action();
    } catch (caught) {
      setError(getLabel(failure));
      if (propagateError) throw caught;
    } finally {
      setBusy(false);
      setBackupProgress(null);
    }
  };

  /** Polls a durable backup task and mirrors its latest completion percentage into page state. */
  const pollBackupTask = useCallback(
    (taskID: string, isActive: () => boolean = () => true): Promise<BackupTaskStatus> =>
      pollBackupTaskStatus(
        taskID,
        isActive,
        async (id) =>
          readData(
            await getRef.current(`/api/backup-restore/backup-tasks/${id}`).execute(),
            backupTaskStatusResponseSchema,
          ),
        { progress: setBackupProgress, results: setBackupTaskResults },
      ),
    [],
  );

  useEffect(() => {
    if (!allowed) return;
    const taskID = readStoredBackupTaskID();
    if (!taskID) return;
    let active = true;
    setBusy(true);
    void pollBackupTask(taskID, () => active)
      .then((status) => {
        if (!active) return;
        const result = describeBackupTaskStatus(status, labelRef.current);
        setMessage(result.message);
        setError(result.error);
      })
      .catch(() => {
        if (active) setError(labelRef.current("admin.backup.createFailed"));
      })
      .finally(() => {
        if (active) {
          setBusy(false);
          setBackupProgress(null);
          setBackupTaskResults([]);
        }
      });
    return () => {
      active = false;
    };
  }, [allowed, pollBackupTask]);

  const reloadVersions = async (): Promise<void> => {
    const list = await readData(
      await get(`${backupPath(application)}/versions`).execute(),
      versionListResponseSchema,
    );
    setVersions(list.versions);
    setVersionCounts((previous) => ({ ...previous, [application]: list.versions.length }));
    setSnapshotsByApplication((previous) => ({ ...previous, [application]: list.snapshots }));
    setVersion(list.versions[0] || "");
  };

  const reloadDifferences = async (): Promise<void> => {
    if (!version) return;
    const list = await readData(
      await get(`${backupPath(application, version)}/diff`).execute(),
      diffResponseSchema,
    );
    setDifferences(list);
  };

  const openCreate = (target: string | "all"): void => {
    setBackupTarget(target);
    if (target !== "all") setApplication(target);
    setName("");
    setDescription("");
    setCreateOpen(true);
  };

  // biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider keeps ordered backup requests and their React state updates together.
  const createBackup = (): void => {
    if (!backupTarget) return;
    setCreateOpen(false);
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
    // biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider keeps ordered backup requests and their React state updates together.
    void run(async () => {
      const payload = backupMetadataSchema.parse({ name, description });
      const start = await readData(
        await post(
          backupTarget === "all"
            ? "/api/backup-restore/backup-tasks"
            : `${backupPath(backupTarget)}/backup-tasks`,
          { ...payload },
        ).execute(),
        backupTaskStartResponseSchema,
      );
      window.localStorage.setItem(storedBackupTaskKey, start.task_id);
      setBackupProgress(backupTarget === "all" ? 0 : null);
      setBackupTaskResults([]);
      const task = await pollBackupTask(start.task_id);
      if (task.state !== "succeeded") throw new Error(task.error || "Backup failed");
      if (backupTarget === "all") {
        const results = task.results ?? [];
        const completed = results.filter((result) => result.status === "completed").length;
        const skipped = results.filter((result) => result.status === "skipped").length;
        const failures = results.filter((result) => result.status === "failed");
        setMessage(
          getLabel("admin.backup.allCompleted")
            .replace("{completed}", String(completed))
            .replace("{skipped}", String(skipped))
            .replace("{failed}", String(failures.length)),
        );
        if (failures.length > 0) {
          setError(
            [
              getLabel("admin.backup.failedTargets"),
              ...failures.map((result) => `${result.application}: ${result.error || "—"}`),
            ].join("\n"),
          );
        }
        const counts = await Promise.all(
          applications.map(async (app) => {
            const list = await readData(
              await get(`${backupPath(app)}/versions`).execute(),
              versionListResponseSchema,
            );
            return [app, list] as const;
          }),
        );
        setVersionCounts(
          Object.fromEntries(counts.map(([app, list]) => [app, list.versions.length])),
        );
        setSnapshotsByApplication(
          Object.fromEntries(counts.map(([app, list]) => [app, list.snapshots])),
        );
        if (backupsOpen && application) {
          const list = await readData(
            await get(`${backupPath(application)}/versions`).execute(),
            versionListResponseSchema,
          );
          setVersions(list.versions);
          setSnapshotsByApplication((previous) => ({ ...previous, [application]: list.snapshots }));
        }
        return;
      }
      const created = { version: task.version ?? "" };
      const list = await readData(
        await get(`${backupPath(backupTarget)}/versions`).execute(),
        versionListResponseSchema,
      );
      setVersionCounts((previous) => ({ ...previous, [backupTarget]: list.versions.length }));
      setSnapshotsByApplication((previous) => ({ ...previous, [backupTarget]: list.snapshots }));
      if (application === backupTarget) setVersions(list.versions);
      setVersion(created.version);
      setMessage(getLabel("admin.backup.created"));
    }, "admin.backup.createFailed");
  };

  const restoreBackup = async (targetApplication: string, targetVersion: string): Promise<void> => {
    setApplication(targetApplication);
    setVersion(targetVersion);
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const mode = restoreModeSchema.parse(restoreMode);
      const outcome = await readData(
        await post(`${backupPath(targetApplication, targetVersion)}/restore`, { mode }).execute(),
        restoreResponseSchema,
      );
      setMessage([getLabel("admin.backup.restored"), ...(outcome.warnings ?? [])].join(" "));
      setPendingRestore(null);
    } catch (caught) {
      setError(getLabel("admin.backup.restoreFailed"));
      throw caught;
    } finally {
      setBusy(false);
    }
  };

  /** Opens a named restore for one application or every configured application. */
  const openNamedRestore = (target: string | "all"): void => {
    setNamedRestoreName(target === "all" ? availableRestoreTargets[0] || "" : "");
    setNamedRestoreTarget(target);
  };

  /** Sends the exact backup name and selected mode to the matching restore RPC. */
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps the single-target and all-target request outcomes together.
  // biome-ignore lint/complexity/noExcessiveLinesPerFunction: The named restore flow handles both request results in one user action.
  const restoreByName = async (): Promise<void> => {
    if (!namedRestoreTarget) return;
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const payload = restoreByNameRequestSchema.parse({
        name: namedRestoreName,
        mode: restoreMode,
      });
      if (namedRestoreTarget === "all") {
        const outcome = await readData(
          await post("/api/backup-restore/restore-by-name", payload).execute(),
          restoreAllByNameResponseSchema,
        );
        const completed = outcome.results.filter((item) => item.status === "completed");
        const skipped = outcome.results.filter((item) => item.status === "skipped");
        const failed = outcome.results.filter((item) => item.status === "failed");
        setMessage(
          getLabel("admin.backup.namedRestoreAllDone")
            .replace("{completed}", String(completed.length))
            .replace("{skipped}", String(skipped.length))
            .replace("{failed}", String(failed.length)),
        );
        if (failed.length > 0 || completed.some((item) => (item.warnings?.length ?? 0) > 0)) {
          setError(
            outcome.results
              // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Each result can carry both an operation error and restart warnings.
              .flatMap((item) => [
                ...(item.error ? [`${item.application}: ${item.error}`] : []),
                ...(item.warnings ?? []).map((warning) => `${item.application}: ${warning}`),
              ])
              .join("\n"),
          );
        }
      } else {
        const outcome = await readData(
          await post(`${backupPath(namedRestoreTarget)}/restore-by-name`, payload).execute(),
          restoreByNameResponseSchema,
        );
        setMessage(getLabel("admin.backup.namedRestoreDone").replace("{version}", outcome.version));
        if (outcome.warnings?.length) setError(outcome.warnings.join("\n"));
      }
      setNamedRestoreTarget(null);
    } catch (caught) {
      setError(getLabel("admin.backup.restoreFailed"));
      throw caught;
    } finally {
      setBusy(false);
    }
  };

  const deleteBackup = async (targetApplication: string, targetVersion: string): Promise<void> => {
    await run(
      async () => {
        await readData(
          await del(`${backupPath(targetApplication, targetVersion)}`).execute(),
          backupDeleteResponseSchema,
        );
        const list = await readData(
          await get(`${backupPath(targetApplication)}/versions`).execute(),
          versionListResponseSchema,
        );
        setVersions(list.versions);
        setVersion((previous) => (previous === targetVersion ? list.versions[0] || "" : previous));
        setVersionCounts((previous) => ({
          ...previous,
          [targetApplication]: list.versions.length,
        }));
        setSnapshotsByApplication((previous) => ({
          ...previous,
          [targetApplication]: list.snapshots,
        }));
        setBackupDetails((previous) => previous.filter((item) => item.version !== targetVersion));
        setMessage(getLabel("admin.backup.deleted"));
        setPendingDelete(null);
      },
      "admin.backup.deleteFailed",
      true,
    );
  };

  const reloadAllInventories = async () => {
    const inventories = await Promise.all(
      applications.map(async (app) => {
        const inventory = await readData(
          await get(`${backupPath(app)}/versions`).execute(),
          versionListResponseSchema,
        );
        return [app, inventory] as const;
      }),
    );
    setVersionCounts(
      Object.fromEntries(inventories.map(([app, inventory]) => [app, inventory.versions.length])),
    );
    setSnapshotsByApplication(
      Object.fromEntries(inventories.map(([app, inventory]) => [app, inventory.snapshots])),
    );
    const selectedInventory = inventories.find(([app]) => app === application)?.[1];
    if (!selectedInventory) return;
    setVersions(selectedInventory.versions);
    setVersion((previous) => selectExistingVersion(previous, selectedInventory.versions));
  };

  const reportBulkDelete = (
    deleted: Array<{ application: string; version: string }>,
    failures: Array<{ application: string; version: string; error?: string | undefined }>,
  ): void => {
    const deletedVersions = new Set(
      deleted.map((result) => `${result.application}/${result.version}`),
    );
    setBackupDetails((previous) =>
      previous.filter((detail) => !deletedVersions.has(`${detail.application}/${detail.version}`)),
    );
    setMessage(
      getLabel("admin.backup.bulkDeleteCompleted")
        .replace("{deleted}", String(deleted.length))
        .replace("{failed}", String(failures.length)),
    );
    if (failures.length === 0) return;
    setError(
      [
        getLabel("admin.backup.bulkDeleteFailed"),
        ...failures.map(
          (result) => `${result.application} (${result.version}): ${result.error || "—"}`,
        ),
      ].join("\n"),
    );
  };

  const deleteBackupsByName = async (targetName: string): Promise<void> => {
    await run(
      async () => {
        const outcome = await readData(
          await del(
            `/api/backup-restore/backups/by-name/${encodeURIComponent(targetName)}`,
          ).execute(),
          backupDeleteAllByNameResponseSchema,
        );
        const deleted = outcome.results.filter((result) => result.status === "deleted");
        const failures = outcome.results.filter((result) => result.status === "failed");
        await reloadAllInventories();
        reportBulkDelete(deleted, failures);
      },
      "admin.backup.bulkDeleteFailed",
      true,
    );
  };

  const startMerge = (targetApplication: string, targetVersion: string): void => {
    void run(async () => {
      const started = await readData(
        await post(`${backupPath(targetApplication, targetVersion)}/merge`).execute(),
        mergeSessionResponseSchema,
      );
      setSession(started);
      window.sessionStorage.setItem(
        storedSessionKey,
        JSON.stringify({
          id: started.session_id,
          application: targetApplication,
          version: targetVersion,
        }),
      );
      setApplication(targetApplication);
      setVersion(targetVersion);
      setBackupsOpen(false);
      setSelectedPath(started.files[0]?.path || "");
      setContents({});
    }, "admin.backup.mergeFailed");
  };

  const finishMerge = async (action: "apply" | "save" | "discard"): Promise<void> => {
    if (!session) return;
    await run(
      // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
      async () => {
        if (action !== "discard") {
          for (const [path, content] of Object.entries(contents)) {
            if (content.result !== content.current) {
              if (application === "ufw-rules" && path === "/ufw.json")
                validateUFWMergeText(content.result);
              await writeMergeText(put, session.session_id, path, content.result);
            }
          }
        }
        const path = `/api/backup-restore/merge-sessions/${session.session_id}`;
        const response =
          action === "discard"
            ? await del(path).execute()
            : await post(
                `${path}/${action}`,
                action === "save" ? backupMetadataSchema.parse({ name, description }) : undefined,
              ).execute();
        if (action === "save") await readData(response, versionResponseSchema);
        else {
          const outcome = await readData(response, mergeActionResponseSchema);
          setMessage(
            [getLabel(`admin.backup.${action}Done`), ...(outcome.warnings ?? [])].join(" "),
          );
        }
        setSession(null);
        window.sessionStorage.removeItem(storedSessionKey);
        setSelectedPath("");
        setContents({});
        if (action === "save") {
          try {
            await reloadVersions();
          } catch {
            setError(getLabel("admin.backup.loadFailed"));
          }
          setMessage(getLabel("admin.backup.saveDone"));
        }
        if (action === "apply") {
          try {
            await reloadDifferences();
          } catch {
            setError(getLabel("admin.backup.diffFailed"));
          }
        }
        if (action === "apply") setPendingApply(false);
      },
      "admin.backup.actionFailed",
      action === "apply",
    );
  };

  const normalizedApplicationSearch = applicationSearch.toLocaleLowerCase();
  const visibleApplications = applications.filter(
    (app) =>
      app.toLocaleLowerCase().includes(normalizedApplicationSearch) ||
      (snapshotsByApplication[app] ?? []).some((snapshot) =>
        (snapshot.name ?? "").toLocaleLowerCase().includes(normalizedApplicationSearch),
      ),
  );
  const availableBackupNames = useMemo(
    () =>
      Array.from(
        new Set(
          Object.values(snapshotsByApplication)
            .flatMap((snapshots) => snapshots.map((snapshot) => snapshot.name ?? ""))
            .filter((snapshotName) => snapshotName.trim().length > 0),
        ),
      ).sort((left, right) => left.localeCompare(right)),
    [snapshotsByApplication],
  );
  const availableRestoreTargets = useMemo(
    () =>
      Array.from(
        new Set(
          Object.values(snapshotsByApplication).flatMap((snapshots) =>
            snapshots.map((snapshot) => snapshot.name || snapshot.version),
          ),
        ),
      ).sort((left, right) => left.localeCompare(right)),
    [snapshotsByApplication],
  );

  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
  // biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider keeps ordered backup requests and their React state updates together.
  const refresh = async (): Promise<void> => {
    if (busy || session) return;
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const apps = await readData(
        await get("/api/backup-restore/applications").execute(),
        applicationListResponseSchema,
      );
      setApplications(apps);
      const counts = await Promise.all(
        apps.map(async (app) => {
          const list = await readData(
            await get(`${backupPath(app)}/versions`).execute(),
            versionListResponseSchema,
          );
          return [app, list] as const;
        }),
      );
      setVersionCounts(
        Object.fromEntries(counts.map(([app, list]) => [app, list.versions.length])),
      );
      setSnapshotsByApplication(
        Object.fromEntries(counts.map(([app, list]) => [app, list.snapshots])),
      );

      const activeApplication = apps.includes(application) ? application : apps[0] || "";
      if (activeApplication !== application) setApplication(activeApplication);
      if (!activeApplication) {
        setVersions([]);
        setBackupDetails([]);
        setDifferences([]);
        return;
      }
      if (!backupsOpen) return;

      const refreshedVersions = await readData(
        await get(`${backupPath(activeApplication)}/versions`).execute(),
        versionListResponseSchema,
      );
      setVersions(refreshedVersions.versions);
      const activeVersion = refreshedVersions.versions.includes(version)
        ? version
        : refreshedVersions.versions[0] || "";
      setVersion(activeVersion);
      if (refreshedVersions.versions.length === 0) {
        setBackupDetails([]);
      }
      if (activeVersion) {
        const refreshedDifferences = await readData(
          await get(`${backupPath(activeApplication, activeVersion)}/diff`).execute(),
          diffResponseSchema,
        );
        setDifferences(refreshedDifferences);
      } else {
        setDifferences([]);
      }
    } catch {
      setError(getLabel("admin.backup.loadFailed"));
    } finally {
      setBusy(false);
    }
  };

  const value: BackupRestoreContextValue = {
    allowed,
    applications,
    versionCounts,
    snapshotsByApplication,
    availableBackupNames,
    availableRestoreTargets,
    visibleApplications,
    countsLoading,
    backupDetails,
    versionsLoading,
    detailsLoading,
    search,
    applicationSearch,
    backupsOpen,
    createOpen,
    backupTarget,
    application,
    versions,
    version,
    name,
    description,
    restoreMode,
    pendingRestore,
    pendingDelete,
    pendingApply,
    differences,
    session,
    selectedPath,
    contents,
    busy,
    backupProgress,
    backupTaskResults,
    error,
    message,
    setApplications,
    setVersionCounts,
    setBackupDetails,
    setSearch,
    setApplicationSearch,
    setBackupsOpen,
    setCreateOpen,
    setBackupTarget,
    setApplication,
    setVersions,
    setVersion,
    setName,
    setDescription,
    setRestoreMode,
    setPendingRestore,
    setPendingDelete,
    setPendingApply,
    setSession,
    setSelectedPath,
    setContents,
    openCreate,
    createBackup,
    restoreBackup,
    namedRestoreTarget,
    namedRestoreName,
    setNamedRestoreTarget,
    setNamedRestoreName,
    openNamedRestore,
    restoreByName,
    deleteBackup,
    deleteBackupsByName,
    startMerge,
    finishMerge,
    refresh,
  };

  return <BackupRestoreContext.Provider value={value}>{children}</BackupRestoreContext.Provider>;
};
