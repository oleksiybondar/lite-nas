import type { PropsWithChildren, ReactElement } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

/** Stable, service-owned loader and poller for one application task shape. */
export interface BackgroundTaskAdapter<Task> {
  /** Loads the most recent task for the provider's current task key. */
  loadLatest: (taskKey: string) => Promise<Task | null>;
  /** Reads one task by its service-owned identifier. */
  poll: (taskId: string) => Promise<Task>;
  /** Extracts the identifier used by `poll`. */
  getId: (task: Task) => string;
  /** Reports whether the task should continue polling. */
  isActive: (task: Task) => boolean;
}

/** State and lifecycle actions shared by background-task controls. */
export interface BackgroundTaskContextValue<Task> {
  /** Identity of the current service-owned task scope. */
  taskKey: string;
  /** Most recently loaded or action-returned task. */
  task: Task | null;
  /** True while the provider loads the most recent task. */
  loading: boolean;
  /** True while a caller-supplied task action runs. */
  busy: boolean;
  /** Localized message for the most recent load, poll, or action failure. */
  error: string | null;
  /** Runs a service action and stores the returned task or localized failure. */
  execute: (action: () => Promise<Task>, failureMessage: string) => Promise<void>;
  /** Clears the currently displayed action or polling error. */
  clearError: () => void;
}

/** Props for the service-agnostic polling provider. */
export interface BackgroundTaskProviderProps<Task> extends PropsWithChildren {
  /** Stable identity of the task being monitored, such as gallery ID plus task kind. */
  taskKey: string;
  /** Adapter mapping the owning service's API and task type. */
  adapter: BackgroundTaskAdapter<Task>;
  /** Enables latest-task loading, normally when the containing view is open. */
  enabled?: boolean;
  /** Poll interval in milliseconds. */
  pollIntervalMs?: number;
  /** Localized message used when loading the latest task fails. */
  loadFailedMessage: string;
  /** Localized message used when refreshing an active task fails. */
  refreshFailedMessage: string;
}

/** Internal React context carrying state for the nearest task scope. */
const BackgroundTaskContext = createContext<BackgroundTaskContextValue<unknown> | null>(null);

/** Owns task loading, polling, action state, and errors for its child controls. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider coordinates one task scope and its React lifecycle.
export function BackgroundTaskProvider<Task>({
  taskKey,
  adapter,
  enabled = true,
  pollIntervalMs = 3000,
  loadFailedMessage,
  refreshFailedMessage,
  children,
}: BackgroundTaskProviderProps<Task>): ReactElement {
  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const actionRunning = useRef(false);

  useEffect(() => {
    if (!enabled) return;
    let mounted = true;
    setLoading(true);
    setError(null);
    void adapter
      .loadLatest(taskKey)
      .then((latest) => {
        if (mounted) setTask(latest);
      })
      .catch(() => {
        if (mounted) setError(loadFailedMessage);
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, [taskKey, enabled, adapter, loadFailedMessage]);

  useEffect(() => {
    if (!task || !adapter.isActive(task)) return;
    let active = true;
    let timer: number;
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Keep one poll attempt and its retry scheduling together.
    const refresh = async (): Promise<void> => {
      try {
        const current = await adapter.poll(adapter.getId(task));
        if (active) {
          setTask(current);
          setError(null);
        }
      } catch {
        if (active) setError(refreshFailedMessage);
      } finally {
        if (active) timer = window.setTimeout(() => void refresh(), pollIntervalMs);
      }
    };
    timer = window.setTimeout(() => void refresh(), pollIntervalMs);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [task, adapter, pollIntervalMs, refreshFailedMessage]);

  const execute = useCallback(
    async (action: () => Promise<Task>, failureMessage: string): Promise<void> => {
      if (actionRunning.current) return;
      actionRunning.current = true;
      setBusy(true);
      setError(null);
      try {
        setTask(await action());
      } catch {
        setError(failureMessage);
      } finally {
        actionRunning.current = false;
        setBusy(false);
      }
    },
    [],
  );
  const clearError = useCallback(() => setError(null), []);
  const value = useMemo(
    () => ({ taskKey, task, loading, busy, error, execute, clearError }),
    [taskKey, task, loading, busy, error, execute, clearError],
  );

  return (
    <BackgroundTaskContext.Provider value={value as BackgroundTaskContextValue<unknown>}>
      {children}
    </BackgroundTaskContext.Provider>
  );
}

/** Reads the nearest task provider using the caller's service-owned task type. */
export function useBackgroundTask<Task>(): BackgroundTaskContextValue<Task> {
  const value = useContext(BackgroundTaskContext);
  if (!value) throw new Error("useBackgroundTask must be used inside BackgroundTaskProvider.");
  return value as BackgroundTaskContextValue<Task>;
}
