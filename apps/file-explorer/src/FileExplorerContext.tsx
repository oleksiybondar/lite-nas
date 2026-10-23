import type { Dispatch, ReactElement, ReactNode, SetStateAction } from "react";
import { createContext, useContext, useEffect, useState } from "react";
import type { FileEntry, OperationJob } from "./file-service";
import { getOperationJob } from "./file-service";
import { cachedJobsSchema, hostPathSchema } from "./schemas";

/** Navigation and selection owned by one of the two persistent panels. */
export interface ExplorerPanelState {
  directory: string;
  history: string[];
  selected: FileEntry[];
}

/** One visible action, optionally backed by a server job. */
export interface ExplorerQueuedAction {
  id: string;
  label: string;
  status: "pending" | "running" | "done" | "error";
  progress: number;
  error?: string;
  remoteId?: string;
}

/** Work submitted by a confirmed File Explorer action. */
export interface ExplorerActionTask {
  label: string;
  run: () => Promise<OperationJob | null>;
}

/** Shared state used by the panels and their bottom action bar. */
export interface FileExplorerContextValue {
  panels: [ExplorerPanelState, ExplorerPanelState];
  activePanel: 0 | 1;
  doublePanel: boolean;
  queue: ExplorerQueuedAction[];
  enqueueActions: (tasks: ExplorerActionTask[]) => Promise<void>;
  setActivePanel: (panel: 0 | 1) => void;
  toggleView: () => void;
  navigate: (panel: 0 | 1, directory: string) => void;
  back: (panel: 0 | 1) => void;
  select: (panel: 0 | 1, entry: FileEntry, multiple: boolean) => void;
  clearSelection: (panel: 0 | 1) => void;
}

const FileExplorerContext = createContext<FileExplorerContextValue | null>(null);
const cacheKey = "lite-nas-file-explorer-jobs";

/** Restores server job identities without trusting arbitrary local storage data. */
function initialQueue(): ExplorerQueuedAction[] {
  try {
    const parsed = cachedJobsSchema.safeParse(JSON.parse(localStorage.getItem(cacheKey) ?? "[]"));
    return parsed.success
      ? parsed.data.map(({ id, label }) => ({
          id,
          remoteId: id,
          label,
          status: "pending",
          progress: 0,
        }))
      : [];
  } catch {
    return [];
  }
}

/** Changes a queue record selected by its local id. */
function updateQueueItem(
  current: ExplorerQueuedAction[],
  id: string,
  change: (item: ExplorerQueuedAction) => ExplorerQueuedAction,
): ExplorerQueuedAction[] {
  return current.map((item) => (item.id === id ? change(item) : item));
}

/** Inputs needed to update one panel's selection and active focus. */
interface SelectPanelEntryOptions {
  setPanels: Dispatch<SetStateAction<[ExplorerPanelState, ExplorerPanelState]>>;
  setActivePanel: Dispatch<SetStateAction<0 | 1>>;
  panel: 0 | 1;
  entry: FileEntry;
  multiple: boolean;
}

/** Stores the result of one completed action in the visible queue. */
function actionResult(item: ExplorerQueuedAction, job: OperationJob | null): ExplorerQueuedAction {
  if (job === null) {
    return { ...item, status: "done", progress: 100 };
  }
  return {
    ...item,
    id: job.job_id,
    remoteId: job.job_id,
    status: job.state,
    progress: job.percent,
  };
}

/** Runs one action and records either its server job or its failure. */
async function enqueueAction(
  task: ExplorerActionTask,
  setQueue: Dispatch<SetStateAction<ExplorerQueuedAction[]>>,
): Promise<void> {
  const id = crypto.randomUUID();
  setQueue((current) => [...current, { id, label: task.label, status: "pending", progress: 0 }]);
  try {
    const job = await task.run();
    setQueue((current) => updateQueueItem(current, id, (item) => actionResult(item, job)));
  } catch (caught) {
    const message = caught instanceof Error ? caught.message : "Action failed.";
    setQueue((current) =>
      updateQueueItem(current, id, (item) => ({ ...item, status: "error", error: message })),
    );
  }
}

/** Polls one persisted File Service job and updates its visible queue record. */
function pollOperationJob(
  item: ExplorerQueuedAction,
  setQueue: Dispatch<SetStateAction<ExplorerQueuedAction[]>>,
): void {
  if (!item.remoteId) {
    return;
  }
  void getOperationJob(item.remoteId)
    .then((job) => {
      setQueue((current) =>
        current.map((entry) =>
          entry.remoteId === job.job_id ? mergeOperationJob(entry, job) : entry,
        ),
      );
    })
    .catch((caught: unknown) => {
      const message = caught instanceof Error ? caught.message : "Job unavailable.";
      setQueue((current) =>
        current.map((entry) =>
          entry.remoteId === item.remoteId ? { ...entry, status: "error", error: message } : entry,
        ),
      );
    });
}

/** Merges current server progress into one local queue record. */
function mergeOperationJob(entry: ExplorerQueuedAction, job: OperationJob): ExplorerQueuedAction {
  return {
    ...entry,
    status: job.state,
    progress: job.percent,
    error: job.error_code ? job.error_code.replaceAll("_", " ") : undefined,
  };
}

/** Maintains the local display cache containing only remote job ids and labels. */
function useQueueCache(queue: ExplorerQueuedAction[]): void {
  useEffect(() => {
    const cached = queue
      .filter((item) => item.remoteId)
      .map((item) => ({ id: item.remoteId as string, label: item.label }));
    localStorage.setItem(cacheKey, JSON.stringify(cached.slice(-128)));
  }, [queue]);
}

/** Polls active jobs until they reach a completed or failed state. */
function useQueuePolling(
  queue: ExplorerQueuedAction[],
  setQueue: Dispatch<SetStateAction<ExplorerQueuedAction[]>>,
): void {
  useEffect(() => {
    const pending = queue.filter(
      (item) => item.remoteId && item.status !== "done" && item.status !== "error",
    );
    if (pending.length === 0) {
      return;
    }
    const timer = window.setInterval(() => {
      pending.forEach((item) => {
        pollOperationJob(item, setQueue);
      });
    }, 1000);
    return () => window.clearInterval(timer);
  }, [queue, setQueue]);
}

/** Owns queue state, persistence, polling, and sequential action submission. */
function useActionQueue(): Pick<FileExplorerContextValue, "queue" | "enqueueActions"> {
  const [queue, setQueue] = useState<ExplorerQueuedAction[]>(initialQueue);
  useQueueCache(queue);
  useQueuePolling(queue, setQueue);
  const enqueueActions = async (tasks: ExplorerActionTask[]): Promise<void> => {
    for (const task of tasks) {
      await enqueueAction(task, setQueue);
    }
  };
  return { queue, enqueueActions };
}

/** Replaces one panel while retaining the other panel's independent state. */
function updatePanel(
  setPanels: Dispatch<SetStateAction<[ExplorerPanelState, ExplorerPanelState]>>,
  panel: 0 | 1,
  change: (current: ExplorerPanelState) => ExplorerPanelState,
): void {
  setPanels((current) => {
    const next: [ExplorerPanelState, ExplorerPanelState] = [...current];
    next[panel] = change(current[panel]);
    return next;
  });
}

/** Enters a directory and records the previous location for Back navigation. */
function navigatePanel(
  setPanels: Dispatch<SetStateAction<[ExplorerPanelState, ExplorerPanelState]>>,
  setActivePanel: Dispatch<SetStateAction<0 | 1>>,
  panel: 0 | 1,
  directory: string,
): void {
  hostPathSchema.parse(directory);
  updatePanel(setPanels, panel, (current) => ({
    directory,
    history: [...current.history, current.directory],
    selected: [],
  }));
  setActivePanel(panel);
}

/** Restores the previous panel directory without requiring a successful listing. */
function backPanel(
  setPanels: Dispatch<SetStateAction<[ExplorerPanelState, ExplorerPanelState]>>,
  setActivePanel: Dispatch<SetStateAction<0 | 1>>,
  panel: 0 | 1,
): void {
  updatePanel(setPanels, panel, (current) => {
    if (current.history.length === 0) {
      return current;
    }
    return {
      directory: current.history[current.history.length - 1],
      history: current.history.slice(0, -1),
      selected: [],
    };
  });
  setActivePanel(panel);
}

/** Selects one entry or toggles it in the panel's multi-selection. */
function selectPanelEntry({
  setPanels,
  setActivePanel,
  panel,
  entry,
  multiple,
}: SelectPanelEntryOptions): void {
  hostPathSchema.parse(entry.path);
  updatePanel(setPanels, panel, (current) => ({
    ...current,
    selected: multiple ? toggleSelectedEntry(current.selected, entry) : [entry],
  }));
  setActivePanel(panel);
}

/** Toggles an entry in a selection while preserving the order of selected rows. */
function toggleSelectedEntry(selected: FileEntry[], entry: FileEntry): FileEntry[] {
  const alreadySelected = selected.some((item) => item.path === entry.path);
  return alreadySelected
    ? selected.filter((item) => item.path !== entry.path)
    : [...selected, entry];
}

/** Owns the two panels and exposes navigation and selection actions. */
function useExplorerPanels(): Pick<
  FileExplorerContextValue,
  | "panels"
  | "activePanel"
  | "doublePanel"
  | "setActivePanel"
  | "toggleView"
  | "navigate"
  | "back"
  | "select"
  | "clearSelection"
> {
  const [panels, setPanels] = useState<[ExplorerPanelState, ExplorerPanelState]>([
    { directory: "/", history: [], selected: [] },
    { directory: "/", history: [], selected: [] },
  ]);
  const [activePanel, setActivePanel] = useState<0 | 1>(0);
  const [doublePanel, setDoublePanel] = useState(true);
  const navigate = (panel: 0 | 1, directory: string): void =>
    navigatePanel(setPanels, setActivePanel, panel, directory);
  const back = (panel: 0 | 1): void => backPanel(setPanels, setActivePanel, panel);
  const select = (panel: 0 | 1, entry: FileEntry, multiple: boolean): void =>
    selectPanelEntry({ setPanels, setActivePanel, panel, entry, multiple });
  const clearSelection = (panel: 0 | 1): void =>
    updatePanel(setPanels, panel, (current) => ({ ...current, selected: [] }));
  const toggleView = (): void => {
    if (doublePanel) {
      setActivePanel(0);
    }
    setDoublePanel(!doublePanel);
  };
  return {
    panels,
    activePanel,
    doublePanel,
    setActivePanel,
    toggleView,
    navigate,
    back,
    select,
    clearSelection,
  };
}

/** Provides two independent directory histories even in single-panel view. */
export function FileExplorerProvider({ children }: { children: ReactNode }): ReactElement {
  const panelsState = useExplorerPanels();
  const actionQueue = useActionQueue();

  return (
    <FileExplorerContext.Provider
      value={{
        ...panelsState,
        ...actionQueue,
      }}
    >
      {children}
    </FileExplorerContext.Provider>
  );
}

/** Reads the File Explorer state from its required provider. */
export function useFileExplorer(): FileExplorerContextValue {
  const value = useContext(FileExplorerContext);
  if (value === null) {
    throw new Error("File Explorer context is unavailable.");
  }
  return value;
}
