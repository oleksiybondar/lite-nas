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
  backupAllResponseSchema,
  backupDeleteResponseSchema,
  backupDetailsResponseSchema,
  backupMetadataSchema,
  diffResponseSchema,
  type MergeSession,
  mergeActionResponseSchema,
  mergeSessionResponseSchema,
  restoreModeSchema,
  restoreResponseSchema,
  validateUFWMergeText,
  versionListResponseSchema,
  versionResponseSchema,
} from "@schemas/backup-restore";
import type { PropsWithChildren, ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import { z } from "zod";

const storedSessionKey = "lite-nas-backup-merge-session";
const storedSessionSchema = z.object({
  id: z.string().regex(/^[a-fA-F0-9]{32}$/),
  application: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/),
  version: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/),
});

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
  const [pendingRestore, setPendingRestore] = useState<BackupVersionTarget | null>(null);
  const [pendingDelete, setPendingDelete] = useState<BackupVersionTarget | null>(null);
  const [pendingApply, setPendingApply] = useState(false);
  const [differences, setDifferences] = useState<BackupDifference[]>([]);
  const [session, setSession] = useState<MergeSession | null>(null);
  const [selectedPath, setSelectedPath] = useState("");
  const [contents, setContents] = useState<Record<string, BackupFileContent>>({});
  const [busy, setBusy] = useState(false);
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
        return [app, list.length] as const;
      }),
    )
      .then((counts) => {
        if (active) {
          setVersionCounts(Object.fromEntries(counts));
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
          setVersions(list);
          // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps ordered backup requests and their React state updates together.
          setVersion((previous) => (list.includes(previous) ? previous : list[0] || ""));
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
    }
  };

  const reloadVersions = async (): Promise<void> => {
    const list = await readData(
      await get(`${backupPath(application)}/versions`).execute(),
      versionListResponseSchema,
    );
    setVersions(list);
    setVersionCounts((previous) => ({ ...previous, [application]: list.length }));
    setVersion(list[0] || "");
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
      if (backupTarget === "all") {
        const results = await readData(
          await post("/api/backup-restore/backups", payload).execute(),
          backupAllResponseSchema,
        );
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
            return [app, list.length] as const;
          }),
        );
        setVersionCounts(Object.fromEntries(counts));
        if (backupsOpen && application) {
          const list = await readData(
            await get(`${backupPath(application)}/versions`).execute(),
            versionListResponseSchema,
          );
          setVersions(list);
        }
        return;
      }
      const created = await readData(
        await post(`${backupPath(backupTarget)}/backups`, payload).execute(),
        versionResponseSchema,
      );
      const list = await readData(
        await get(`${backupPath(backupTarget)}/versions`).execute(),
        versionListResponseSchema,
      );
      setVersionCounts((previous) => ({ ...previous, [backupTarget]: list.length }));
      if (application === backupTarget) setVersions(list);
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
        await post(
          `${backupPath(targetApplication, targetVersion)}/restore?mode=${encodeURIComponent(mode)}`,
        ).execute(),
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
        setVersions(list);
        setVersion((previous) => (previous === targetVersion ? list[0] || "" : previous));
        setVersionCounts((previous) => ({ ...previous, [targetApplication]: list.length }));
        setBackupDetails((previous) => previous.filter((item) => item.version !== targetVersion));
        setMessage(getLabel("admin.backup.deleted"));
        setPendingDelete(null);
      },
      "admin.backup.deleteFailed",
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
          return [app, list.length] as const;
        }),
      );
      setVersionCounts(Object.fromEntries(counts));

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
      setVersions(refreshedVersions);
      const activeVersion = refreshedVersions.includes(version)
        ? version
        : refreshedVersions[0] || "";
      setVersion(activeVersion);
      if (refreshedVersions.length === 0) {
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
    deleteBackup,
    startMerge,
    finishMerge,
    refresh,
  };

  return <BackupRestoreContext.Provider value={value}>{children}</BackupRestoreContext.Provider>;
};
