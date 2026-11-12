import { useApi } from "@hooks/useApi";
import { useBackupRestore } from "@hooks/useBackupRestore";
import { useRbac } from "@hooks/useRbac";
import { BackupRestoreProvider } from "@providers/BackupRestoreProvider";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import {
  backupResponse,
  exactApiRoute,
  mockBackupGet,
  suffixApiRoute,
} from "@tests/unit/test-utils/backup-restore-api";
import { renderWithLanguage } from "@tests/unit/test-utils/language";
import { useEffect } from "react";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

const taskID = "0123456789abcdef0123456789abcdef";
const timestamp = "2026-11-10T12:00:00Z";

/** Builds a terminal status for one application's completed backup task. */
const completedApplicationTask = (): Response =>
  backupResponse({
    task_id: taskID,
    application: "example",
    state: "succeeded",
    started_at: timestamp,
    updated_at: timestamp,
    finished_at: timestamp,
    progress: { percent: 100 },
    version: "v2",
  });

/** Builds an in-progress aggregate status with one completed target. */
const runningAggregateTask = (): Response =>
  backupResponse({
    task_id: taskID,
    application: "all",
    state: "running",
    started_at: timestamp,
    updated_at: timestamp,
    progress: { percent: 50, message: "Processed example" },
    results: [{ application: "example", status: "completed", version: "v2" }],
  });

/** Builds a terminal aggregate status with one completed and one skipped target. */
const completedAggregateTask = (): Response =>
  backupResponse({
    task_id: taskID,
    application: "all",
    state: "succeeded",
    started_at: timestamp,
    updated_at: timestamp,
    finished_at: timestamp,
    progress: { percent: 100 },
    results: [
      { application: "example", status: "completed", version: "v2" },
      { application: "other", status: "skipped" },
    ],
  });

/** Supplies the version inventory returned after the background task completes. */
const exampleInventory = (): Response =>
  backupResponse({ versions: ["v2"], snapshots: [{ version: "v2", name: "nightly" }] });

/** Exposes background backup actions and observable provider state to the test. */
const BackupTaskConsumer = ({
  target,
  onProgress,
}: {
  target: string;
  onProgress?: (value: number) => void;
}) => {
  const { openCreate, createBackup, busy, backupProgress, error, message } = useBackupRestore();
  useEffect(() => {
    if (backupProgress !== null) onProgress?.(backupProgress);
  }, [backupProgress, onProgress]);
  return (
    <>
      <button onClick={() => openCreate(target)} type="button">
        Select target
      </button>
      <button onClick={createBackup} type="button">
        Start backup
      </button>
      <div data-testid="busy">{String(busy)}</div>
      <div data-testid="progress">{backupProgress ?? "indeterminate"}</div>
      <div data-testid="error">{error}</div>
      <div data-testid="message">{message}</div>
    </>
  );
};

/** Exposes provider-level application search results for snapshot-name matching. */
const SnapshotSearchConsumer = () => {
  const { setApplicationSearch, visibleApplications } = useBackupRestore();
  return (
    <>
      <button onClick={() => setApplicationSearch("nightly-release")} type="button">
        Search snapshot
      </button>
      <div data-testid="visible-applications">{visibleApplications.join(",")}</div>
    </>
  );
};

/** Installs API and administrator mocks shared by the background task flows. */
const mockApi = (get: unknown, post: unknown): void => {
  vi.mocked(useApi).mockReturnValue({
    get,
    post,
    put: vi.fn(),
    delete: vi.fn(),
  } as unknown as ReturnType<typeof useApi>);
  vi.mocked(useRbac).mockReturnValue({ requireAdmin: () => true } as ReturnType<typeof useRbac>);
};

// Requirements: backup-restore/FR-013, backup-restore/IR-004.
beforeEach(() => {
  window.localStorage.clear();
});

describe("single-application background backup", () => {
  it("starts and polls one application backup before refreshing its inventory", async () => {
    const get = mockBackupGet([
      exactApiRoute("/api/backup-restore/applications", () => backupResponse(["example"])),
      suffixApiRoute(`/backup-tasks/${taskID}`, completedApplicationTask),
      suffixApiRoute("/versions", exampleInventory),
    ]);
    const post = vi.fn(() => ({
      execute: async () => backupResponse({ task_id: taskID, state: "queued" }),
    }));
    mockApi(get, post);

    renderWithLanguage(
      <BackupRestoreProvider>
        <BackupTaskConsumer target="example" />
      </BackupRestoreProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Select target" }));
    fireEvent.click(screen.getByRole("button", { name: "Start backup" }));

    await waitFor(() =>
      expect({
        request: post.mock.calls[0],
        message: screen.getByTestId("message").textContent,
        storedTask: window.localStorage.getItem("lite-nas-backup-task"),
      }).toEqual({
        request: [
          "/api/backup-restore/applications/example/backup-tasks",
          { name: "", description: "" },
        ],
        message: "Backup created.",
        storedTask: null,
      }),
    );
  });
});

describe("resuming a stored background backup", () => {
  it("resumes a stored task ID after a browser refresh", async () => {
    window.localStorage.setItem("lite-nas-backup-task", taskID);
    const get = mockBackupGet([
      exactApiRoute("/api/backup-restore/applications", () => backupResponse(["example"])),
      suffixApiRoute(`/backup-tasks/${taskID}`, completedApplicationTask),
      suffixApiRoute("/versions", exampleInventory),
    ]);
    const post = vi.fn(() => ({ execute: async () => backupResponse({}) }));
    mockApi(get, post);

    renderWithLanguage(
      <BackupRestoreProvider>
        <BackupTaskConsumer target="example" />
      </BackupRestoreProvider>,
    );

    await waitFor(() =>
      expect({
        statusRequested: get.mock.calls.some(
          ([path]) => path === `/api/backup-restore/backup-tasks/${taskID}`,
        ),
        message: screen.getByTestId("message").textContent,
        storedTask: window.localStorage.getItem("lite-nas-backup-task"),
        postCalls: post.mock.calls.length,
      }).toEqual({
        statusRequested: true,
        message: "Backup created.",
        storedTask: null,
        postCalls: 0,
      }),
    );
  });
});

describe("all-application background backup progress", () => {
  it("shows per-target progress for an all-applications backup", async () => {
    let statusReads = 0;
    const progress = vi.fn();
    const get = mockBackupGet([
      exactApiRoute("/api/backup-restore/applications", () => backupResponse(["example"])),
      exactApiRoute(`/api/backup-restore/backup-tasks/${taskID}`, () => {
        statusReads += 1;
        return statusReads === 1 ? runningAggregateTask() : completedAggregateTask();
      }),
      suffixApiRoute("/versions", exampleInventory),
    ]);
    const post = vi.fn(() => ({
      execute: async () => backupResponse({ task_id: taskID, state: "queued" }),
    }));
    mockApi(get, post);

    renderWithLanguage(
      <BackupRestoreProvider>
        <BackupTaskConsumer target="all" onProgress={progress} />
      </BackupRestoreProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Select target" }));
    fireEvent.click(screen.getByRole("button", { name: "Start backup" }));

    await waitFor(() =>
      expect({
        request: post.mock.calls[0],
        observedProgress: progress.mock.calls.map(([value]) => value),
        message: screen.getByTestId("message").textContent,
      }).toEqual({
        request: ["/api/backup-restore/backup-tasks", { name: "", description: "" }],
        observedProgress: expect.arrayContaining([50]),
        message: "Backups finished: 1 created, 1 skipped, 0 failed.",
      }),
    );
  });
});

describe("filtering applications by snapshot name", () => {
  it("filters applications by snapshot names through provider context", async () => {
    const get = mockBackupGet([
      exactApiRoute("/api/backup-restore/applications", () => backupResponse(["alpha", "beta"])),
      exactApiRoute("/api/backup-restore/applications/alpha/versions", () =>
        backupResponse({ versions: ["v1"], snapshots: [{ version: "v1", name: "weekly" }] }),
      ),
      exactApiRoute("/api/backup-restore/applications/beta/versions", () =>
        backupResponse({
          versions: ["v2"],
          snapshots: [{ version: "v2", name: "nightly-release" }],
        }),
      ),
    ]);
    mockApi(
      get,
      vi.fn(() => ({ execute: async () => backupResponse({}) })),
    );

    renderWithLanguage(
      <BackupRestoreProvider>
        <SnapshotSearchConsumer />
      </BackupRestoreProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Search snapshot" }));

    await waitFor(() =>
      expect(screen.getByTestId("visible-applications").textContent).toBe("beta"),
    );
  });
});
