import type { ApiContextValue } from "@dto/api/api";
import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import {
  BackgroundTaskActions,
  BackgroundTaskProgress,
  BackgroundTaskProvider,
  useBackgroundTask,
} from "@lite-nas/background-task";
import { FolderSelectionDialog, hostPathSchema } from "@lite-nas/file-browser";
import { useLanguage } from "@lite-nas/language";
import { Alert, Button, Stack, Typography } from "@mui/material";
import type { ReactElement } from "react";
import { useEffect, useMemo, useState } from "react";
import { z } from "zod";

const taskSchema = z.object({
  task_id: z.string().regex(/^[a-f0-9]{32}$/),
  state: z.enum(["queued", "running", "succeeded", "failed", "cancelled"]),
  output: z.string(),
  error: z.string().optional(),
  progress: z
    .object({ percent: z.number().min(0).max(100).optional(), message: z.string().optional() })
    .optional(),
});
type Task = z.infer<typeof taskSchema>;
const antivirusInfoSchema = z.object({
  state: z.enum(["ready", "missing_signatures"]),
  update_state: z.enum(["idle", "queued", "updating", "failed"]),
  update_details: z.string().optional(),
  engine_version: z.string().optional(),
  signature_updated_at: z.string().optional(),
  signature_database_count: z.number(),
  details: z.string().optional(),
});
const integrityInfoSchema = z.object({
  baseline_available: z.boolean(),
  baseline_path: z.string(),
  baseline_updated_at: z.string().optional(),
  baseline_size: z.number().optional(),
  details: z.string().optional(),
});
type TaskKind = "integrity" | "antivirus" | "baseline";

/** Maps a task kind to the localized page title. */
const pageLabel = (kind: TaskKind): string => {
  if (kind === "antivirus") return "navigation.Antivirus";
  if (kind === "baseline") return "admin.securityTasks.baselineTitle";
  return "navigation.Integrity";
};

/** Selects the task creation endpoint for one supported security task. */
const endpointForKind = (kind: TaskKind): string => {
  if (kind === "antivirus") return "antivirus/scans";
  if (kind === "baseline") return "integrity/baselines";
  return "integrity/checks";
};

/** Selects which service owns a task snapshot. */
const serviceForKind = (kind: TaskKind): "antivirus" | "integrity" =>
  kind === "antivirus" ? "antivirus" : "integrity";

/** Renders one security scan page using its service-owned gateway task API. */
export function TaskPage({ kind }: { kind: TaskKind }): ReactElement | null {
  const { requireSecurity } = useRbac();
  const allowed = requireSecurity();
  const { getLabel } = useLanguage();
  const { get, post } = useApi();
  const adapter = useMemo(
    () => ({
      // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The task flow keeps request status, validation, and feedback together.
      loadLatest: async (key: string): Promise<Task | null> => {
        const id = window.sessionStorage.getItem(`security-task:${key}`);
        if (!id || !/^[a-f0-9]{32}$/.test(id)) return null;
        const service = key === "antivirus" ? "antivirus" : "integrity";
        try {
          return await readTask(
            get,
            `/api/security/tasks/${service}/tasks/${encodeURIComponent(id)}`,
          );
        } catch {
          // Task state is in memory, so a saved ID can outlive its service process.
          window.sessionStorage.removeItem(`security-task:${key}`);
          return null;
        }
      },
      poll: async (id: string): Promise<Task> =>
        readTask(
          get,
          `/api/security/tasks/${serviceForKind(kind)}/tasks/${encodeURIComponent(id)}`,
        ),
      getId: (task: Task) => task.task_id,
      isActive: (task: Task) => task.state === "queued" || task.state === "running",
    }),
    [get, kind],
  );
  if (!allowed) return null;
  return (
    <BackgroundTaskProvider
      taskKey={kind}
      adapter={adapter}
      loadFailedMessage={getLabel("admin.securityTasks.loadFailed")}
      refreshFailedMessage={getLabel("admin.securityTasks.refreshFailed")}
    >
      <TaskControls kind={kind} get={get} post={post} />
    </BackgroundTaskProvider>
  );
}

/** Owns the localized start action and folder selection for one scan page. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The task flow keeps request status, validation, and feedback together.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The task flow keeps request status, validation, and feedback together.
function TaskControls({
  kind,
  get,
  post,
}: {
  kind: TaskKind;
  get: ApiContextValue["get"];
  post: ApiContextValue["post"];
}): ReactElement {
  const { getLabel } = useLanguage();
  const { task, loading, error, execute } = useBackgroundTask<Task>();
  const [pickerOpen, setPickerOpen] = useState(false);
  const [startDetail, setStartDetail] = useState<string | null>(null);
  const [info, setInfo] = useState<
    z.infer<typeof antivirusInfoSchema> | z.infer<typeof integrityInfoSchema> | null
  >(null);
  const [infoError, setInfoError] = useState(false);
  const [updateError, setUpdateError] = useState<string | null>(null);
  const active = task?.state === "queued" || task?.state === "running";
  const label = getLabel(pageLabel(kind));
  useEffect(() => {
    let mounted = true;
    const service = kind === "antivirus" ? "antivirus" : "integrity";
    void get(`/api/security/tasks/${service}/info`)
      .execute()
      // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The task flow keeps request status, validation, and feedback together.
      .then(async (response) => {
        if (!response.ok) throw new Error("info read failed");
        const envelope = await response.json();
        const data = z
          .object({ data: kind === "antivirus" ? antivirusInfoSchema : integrityInfoSchema })
          .parse(envelope).data;
        if (mounted) {
          setInfo(data);
          setInfoError(false);
        }
      })
      .catch(() => {
        if (mounted) {
          setInfo(null);
          setInfoError(true);
        }
      });
    return () => {
      mounted = false;
    };
  }, [get, kind]);
  useEffect(() => {
    if (
      kind !== "antivirus" ||
      !info ||
      !("update_state" in info) ||
      !["queued", "updating"].includes(info.update_state)
    )
      return;
    const timer = window.setInterval(() => {
      void get("/api/security/tasks/antivirus/info")
        .execute()
        .then(async (response) => {
          if (!response.ok) throw new Error("signature status read failed");
          const envelope = z.object({ data: antivirusInfoSchema }).parse(await response.json());
          setInfo(envelope.data);
        })
        .catch(() => setInfoError(true));
    }, 3000);
    return () => window.clearInterval(timer);
  }, [get, info, kind]);
  const updateSignatures = (): void => {
    setUpdateError(null);
    void post("/api/security/tasks/antivirus/signatures/update")
      .execute()
      .then(async (response) => {
        if (!response.ok) throw new Error(await readResponseError(response));
        const envelope = z
          .object({ data: z.object({ state: z.enum(["queued", "updating", "already_updating"]) }) })
          .parse(await response.json());
        // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The task flow keeps request status, validation, and feedback together.
        setInfo((current) =>
          current && "update_state" in current
            ? {
                ...current,
                update_state:
                  envelope.data.state === "already_updating" ? "updating" : envelope.data.state,
              }
            : current,
        );
      })
      .catch((error: unknown) =>
        setUpdateError(
          error instanceof Error ? error.message : getLabel("admin.securityTasks.updateFailed"),
        ),
      );
  };
  const start = (folder?: string): void => {
    setStartDetail(null);
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The task flow keeps request status, validation, and feedback together.
    void execute(async () => {
      try {
        return await startTask(post, kind, folder);
      } catch (error) {
        setStartDetail(error instanceof Error ? error.message : null);
        throw error;
      }
    }, getLabel("admin.securityTasks.startFailed"));
  };
  let statusAlert: ReactElement | null = null;
  if (info && kind === "antivirus" && "state" in info) {
    statusAlert = (
      <Alert severity={info.state === "ready" ? "success" : "warning"}>
        {getLabel(
          info.state === "ready"
            ? "admin.securityTasks.signaturesReady"
            : "admin.securityTasks.signaturesMissing",
        )}
        {info.signature_updated_at
          ? ` · ${new Date(info.signature_updated_at).toLocaleString()}`
          : ""}
      </Alert>
    );
  } else if (info && "baseline_available" in info) {
    statusAlert = (
      <Alert severity={info.baseline_available ? "success" : "warning"}>
        {getLabel(
          info.baseline_available
            ? "admin.securityTasks.baselineReady"
            : "admin.securityTasks.baselineMissing",
        )}
        {info.baseline_updated_at
          ? ` · ${new Date(info.baseline_updated_at).toLocaleString()}`
          : ""}
      </Alert>
    );
  }
  return (
    <Stack spacing={2} data-testid={`${kind}-page`}>
      <Typography variant="h1">{label}</Typography>
      {statusAlert}
      {infoError && <Alert severity="error">{getLabel("admin.securityTasks.infoFailed")}</Alert>}
      {info && "details" in info && info.details && (
        <Alert severity="warning">{info.details}</Alert>
      )}
      {kind === "antivirus" && info && "update_state" in info && (
        <Stack spacing={1}>
          <Button
            variant="outlined"
            onClick={updateSignatures}
            disabled={info.update_state === "queued" || info.update_state === "updating"}
          >
            {info.update_state === "queued" || info.update_state === "updating"
              ? getLabel("admin.securityTasks.updatingSignatures")
              : getLabel("admin.securityTasks.updateSignatures")}
          </Button>
          {info.update_details && (
            <Typography role={info.update_state === "failed" ? "alert" : undefined}>
              {info.update_details}
            </Typography>
          )}
          {updateError && <Alert severity="error">{updateError}</Alert>}
        </Stack>
      )}
      <BackgroundTaskActions
        startLabel={getLabel(
          kind === "baseline"
            ? "admin.securityTasks.generateBaseline"
            : "admin.securityTasks.start",
        )}
        active={Boolean(active)}
        disabled={loading}
        onStart={() => (kind === "antivirus" ? setPickerOpen(true) : start())}
      />
      {kind === "antivirus" && (
        <FolderSelectionDialog
          open={pickerOpen}
          onClose={() => setPickerOpen(false)}
          initialDirectory="/home"
          onSelect={(folder) => {
            setPickerOpen(false);
            start(folder);
          }}
        />
      )}
      {error && <Alert severity="error">{error}</Alert>}
      {error && startDetail && <Alert severity="error">{startDetail}</Alert>}
      {task && (
        <BackgroundTaskProgress
          active={Boolean(active)}
          label={getLabel("admin.securityTasks.progress")}
          value={task.progress?.percent ?? null}
        >
          <Typography>{getLabel(`admin.securityTasks.state.${task.state}`)}</Typography>
          {task.progress?.message && <Typography>{task.progress.message}</Typography>}
          {task.error && <Typography role="alert">{task.error}</Typography>}
          {task.output && <pre>{task.output}</pre>}
        </BackgroundTaskProgress>
      )}
    </Stack>
  );
}

/** Starts a scan and validates the returned task snapshot before displaying it. */
async function startTask(
  post: ApiContextValue["post"],
  kind: TaskKind,
  folder?: string,
): Promise<Task> {
  const endpoint = endpointForKind(kind);
  const validatedFolder = folder === undefined ? undefined : hostPathSchema.parse(folder);
  const url =
    kind === "antivirus" && validatedFolder !== undefined
      ? `/api/security/tasks/${endpoint}?${new URLSearchParams({ folder: validatedFolder })}`
      : `/api/security/tasks/${endpoint}`;
  const response = await post(url).execute();
  if (!response.ok) throw new Error(await readResponseError(response));
  const envelope = z
    .object({
      success: z.literal(true),
      data: z.object({
        task_id: z.string().regex(/^[a-f0-9]{32}$/),
        state: z.enum(["queued", "running", "succeeded", "failed", "cancelled"]),
      }),
    })
    .parse(await response.json());
  window.sessionStorage.setItem(`security-task:${kind}`, envelope.data.task_id);
  return { task_id: envelope.data.task_id, state: envelope.data.state, output: "" };
}

/** Extracts a safe, server-provided detail from an unsuccessful API response. */
async function readResponseError(response: Response): Promise<string> {
  try {
    const body = z
      .object({ detail: z.string().optional(), message: z.string().optional() })
      .passthrough()
      .parse(await response.json());
    return body.detail ?? body.message ?? `Request failed with status ${response.status}.`;
  } catch {
    return `Request failed with status ${response.status}.`;
  }
}

/** Reads and validates one task snapshot through the authenticated API client. */
async function readTask(get: ApiContextValue["get"], path: string): Promise<Task> {
  const response = await get(path).execute();
  if (!response.ok) throw new Error("task read failed");
  const envelope = z.object({ data: taskSchema }).parse(await response.json());
  return envelope.data;
}
