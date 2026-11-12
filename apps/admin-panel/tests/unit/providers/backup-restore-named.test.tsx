import { useApi } from "@hooks/useApi";
import { useBackupRestore } from "@hooks/useBackupRestore";
import { useRbac } from "@hooks/useRbac";
import { BackupRestoreProvider } from "@providers/BackupRestoreProvider";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

/** Drives the named restore operation through the real provider state. */
const NamedRestoreConsumer = ({ target }: { target: string }) => {
  const { openNamedRestore, setNamedRestoreName, restoreByName } = useBackupRestore();
  return (
    <>
      <button onClick={() => openNamedRestore(target)} type="button">
        Open
      </button>
      <button onClick={() => setNamedRestoreName("release")} type="button">
        Name
      </button>
      <button onClick={() => void restoreByName()} type="button">
        Restore
      </button>
    </>
  );
};

// Requirements: backup-restore/FR-012, backup-restore/IR-004.
describe("named restore requests", () => {
  test.each([
    [
      "one application",
      "example",
      "/api/backup-restore/applications/example/restore-by-name",
      { version: "v1", restored: true },
    ],
    ["all applications", "all", "/api/backup-restore/restore-by-name", { results: [] }],
  ])("sends %s with a validated JSON payload", async (_label, target, path, data) => {
    window.sessionStorage.clear();
    const get = vi.fn().mockReturnValue({
      execute: () =>
        Promise.resolve(
          new Response(
            JSON.stringify({ success: true, timestamp: new Date().toISOString(), data: [] }),
          ),
        ),
    });
    const post = vi.fn().mockReturnValue({
      execute: () =>
        Promise.resolve(
          new Response(
            JSON.stringify({ success: true, timestamp: new Date().toISOString(), data }),
          ),
        ),
    });
    vi.mocked(useApi).mockReturnValue({ get, post } as unknown as ReturnType<typeof useApi>);
    vi.mocked(useRbac).mockReturnValue({ requireAdmin: () => true } as ReturnType<typeof useRbac>);
    renderWithLanguage(
      <BackupRestoreProvider>
        <NamedRestoreConsumer target={target} />
      </BackupRestoreProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    fireEvent.click(screen.getByRole("button", { name: "Name" }));
    fireEvent.click(screen.getByRole("button", { name: "Restore" }));
    await waitFor(() =>
      expect(post).toHaveBeenCalledWith(path, { name: "release", mode: "replace" }),
    );
  });
});
