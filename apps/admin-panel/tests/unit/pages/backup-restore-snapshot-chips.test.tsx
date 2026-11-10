import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { BackupRestorePage } from "@pages/BackupRestorePage";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import {
  backupResponse,
  exactApiRoute,
  mockBackupGet,
} from "@tests/unit/test-utils/backup-restore-api";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

/** Renders the backup page with five named snapshots for the application card. */
const renderSnapshotCard = (): ReturnType<typeof vi.fn> => {
  const get = mockBackupGet([
    exactApiRoute("/api/backup-restore/applications", () => backupResponse(["example"])),
    exactApiRoute("/api/backup-restore/applications/example/versions", () =>
      backupResponse({
        versions: ["20261110T120000.000000000Z", "v5", "v4", "v3", "v2", "v1"],
        snapshots: [
          { version: "20261110T120000.000000000Z" },
          { version: "v5", name: "snapshot-five" },
          { version: "v4", name: "snapshot-four" },
          { version: "v3", name: "snapshot-three" },
          { version: "v2", name: "snapshot-two" },
          { version: "v1", name: "snapshot-one" },
        ],
      }),
    ),
  ]);
  const post = vi.fn(() => ({ execute: async () => backupResponse({ results: [] }) }));
  vi.mocked(useApi).mockReturnValue({
    get,
    post,
    put: vi.fn(),
    delete: vi.fn(),
  } as unknown as ReturnType<typeof useApi>);
  vi.mocked(useRbac).mockReturnValue({ requireAdmin: () => true } as ReturnType<typeof useRbac>);
  renderWithLanguage(<BackupRestorePage />);
  return post;
};

/** Renders the bulk-delete controls with a duplicate name shared by two applications. */
const renderBulkDeletePage = (deleteRequest: ReturnType<typeof vi.fn>): void => {
  const get = mockBackupGet([
    exactApiRoute("/api/backup-restore/applications", () => backupResponse(["alpha", "beta"])),
    exactApiRoute("/api/backup-restore/applications/alpha/versions", () =>
      backupResponse({ versions: ["v1"], snapshots: [{ version: "v1", name: "nightly" }] }),
    ),
    exactApiRoute("/api/backup-restore/applications/beta/versions", () =>
      backupResponse({ versions: ["v2"], snapshots: [{ version: "v2", name: "nightly" }] }),
    ),
  ]);
  vi.mocked(useApi).mockReturnValue({
    get,
    post: vi.fn(),
    put: vi.fn(),
    delete: deleteRequest,
  } as unknown as ReturnType<typeof useApi>);
  vi.mocked(useRbac).mockReturnValue({ requireAdmin: () => true } as ReturnType<typeof useRbac>);
  renderWithLanguage(<BackupRestorePage />);
};

describe("backup application snapshot display", () => {
  it("shows a snapshot name on the application card", async () => {
    renderSnapshotCard();

    expect(await screen.findByTestId("backup-snapshot-example-v5")).toHaveTextContent(
      "snapshot-five",
    );
  });
});

describe("backup application snapshot expansion", () => {
  it("expands the card to show snapshots beyond the first four", async () => {
    renderSnapshotCard();
    fireEvent.click(await screen.findByRole("button", { name: "Show all 6" }));

    expect(await screen.findByTestId("backup-snapshot-example-v1")).toHaveTextContent(
      "snapshot-one",
    );
  });
});

describe("bulk delete by snapshot name", () => {
  it("requires typing yes before deleting snapshots with the selected name", async () => {
    const deleteRequest = vi.fn(() => ({
      execute: async () =>
        backupResponse({
          results: [
            { application: "alpha", version: "v1", status: "deleted" },
            { application: "beta", version: "v2", status: "deleted" },
          ],
        }),
    }));
    renderBulkDeletePage(deleteRequest);

    const openDelete = await screen.findByTestId("backup-delete-all-by-name");
    await waitFor(() => expect(openDelete).toBeEnabled());
    fireEvent.click(openDelete);
    fireEvent.mouseDown(screen.getByRole("combobox", { name: "Name" }));
    const options = await screen.findAllByRole("option", { name: "nightly" });
    const firstOption = options[0];
    if (!firstOption) throw new Error("Expected a nightly snapshot option");
    fireEvent.click(firstOption);
    fireEvent.click(screen.getByTestId("backup-delete-name-continue"));
    const confirmation = await screen.findByTestId("confirm-action-input");
    const confirm = screen.getByTestId("confirm-action-submit");
    const confirmWasDisabled = (confirm as HTMLButtonElement).disabled;
    const confirmationInput = confirmation.querySelector("input");
    if (!confirmationInput) throw new Error("Expected the delete confirmation input");
    fireEvent.change(confirmationInput, { target: { value: "yes" } });
    fireEvent.click(confirm);

    await waitFor(() =>
      expect({
        uniqueNameOptions: options.length,
        confirmWasDisabled,
        request: deleteRequest.mock.calls[0],
      }).toEqual({
        uniqueNameOptions: 1,
        confirmWasDisabled: true,
        request: ["/api/backup-restore/backups/by-name/nightly"],
      }),
    );
  });
});

describe("restore all by snapshot name", () => {
  it("uses the available snapshot names for restore all by name", async () => {
    const post = renderSnapshotCard();
    fireEvent.click(await screen.findByTestId("backup-restore-all-by-name"));
    fireEvent.mouseDown(screen.getByRole("combobox", { name: "Name" }));
    fireEvent.click(await screen.findByRole("option", { name: "snapshot-five" }));
    fireEvent.click(screen.getByTestId("backup-restore-name-confirm"));

    await waitFor(() =>
      expect(post.mock.calls[0]).toEqual([
        "/api/backup-restore/restore-by-name",
        { name: "snapshot-five", mode: "replace" },
      ]),
    );
  });
});

describe("restore all by unnamed snapshot timestamp", () => {
  it("offers an unnamed snapshot timestamp in restore all and sends it as the restore target", async () => {
    const post = renderSnapshotCard();
    fireEvent.click(await screen.findByTestId("backup-restore-all-by-name"));
    fireEvent.mouseDown(screen.getByRole("combobox", { name: "Name" }));
    const timestamp = "20261110T120000.000000000Z";
    fireEvent.click(await screen.findByRole("option", { name: timestamp }));
    fireEvent.click(screen.getByTestId("backup-restore-name-confirm"));

    await waitFor(() =>
      expect(post.mock.calls[0]).toEqual([
        "/api/backup-restore/restore-by-name",
        { name: timestamp, mode: "replace" },
      ]),
    );
  });
});
