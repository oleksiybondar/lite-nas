import { createTheme, ThemeProvider } from "@mui/material/styles";
import { act, cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, test, vi } from "vitest";
import { ConfirmActionDialog } from "../src/ConfirmActionDialog";
import { ExplorerActionBar } from "../src/ExplorerActionBar";
import { FileEntryIcon } from "../src/FileEntryIcon";
import { FileEntryItem } from "../src/FileEntryItem";
import { FileExplorerProvider } from "../src/FileExplorerContext";
import { FileExplorerWorkspace } from "../src/FileExplorerWorkspace";
import { NewFolderDialog } from "../src/NewFolderDialog";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  localStorage.clear();
});

/** Renders a component with the same MUI theme support as the app shell. */
function renderThemed(element: React.ReactElement) {
  return render(<ThemeProvider theme={createTheme()}>{element}</ThemeProvider>);
}

const transferJob = {
  job_id: "b".repeat(48),
  state: "done",
  percent: 100,
  completed_bytes: 1,
  total_bytes: 1,
  completed_items: 1,
  total_items: 1,
};

/** Records one File Service request and returns its configured response. */
function recordTransferRequest(
  requests: Array<{ url: string; method: string; body?: string }>,
  responses: Record<string, () => Response>,
) {
  return async (url: string, init?: RequestInit): Promise<Response> => {
    requests.push({ url, method: init?.method ?? "GET", body: init?.body as string | undefined });
    const respond = responses[url];
    if (!respond) {
      throw new Error(`unexpected request ${url}`);
    }
    return respond();
  };
}

/** Loads the workspace with one source file and a destination directory in each panel. */
async function renderTransferWorkspace() {
  const requests: Array<{ url: string; method: string; body?: string }> = [];
  const responses: Record<string, () => Response> = {
    "/ls/": () =>
      new Response(JSON.stringify({ files: ["/a.txt"], directories: ["/dest"] }), {
        status: 200,
      }),
    "/ls/dest": () => new Response(JSON.stringify({ files: [], directories: [] }), { status: 200 }),
    "/f/copy": () => new Response(JSON.stringify(transferJob), { status: 202 }),
    "/f/move": () => new Response(JSON.stringify(transferJob), { status: 202 }),
    "/f/a.txt": () => new Response(null, { status: 204 }),
  };
  vi.stubGlobal("fetch", vi.fn().mockImplementation(recordTransferRequest(requests, responses)));
  await act(async () => {
    renderThemed(
      <FileExplorerProvider>
        <FileExplorerWorkspace />
      </FileExplorerProvider>,
    );
  });
  const left = document.querySelector<HTMLElement>(
    "[data-test-class='file-explorer-panel'][data-test-name='left']",
  ) as HTMLElement;
  const right = document.querySelector<HTMLElement>(
    "[data-test-class='file-explorer-panel'][data-test-name='right']",
  ) as HTMLElement;
  fireEvent.doubleClick(await within(right).findByText("dest"));
  await within(left).findByText("a.txt");
  return { left, requests };
}

/** Selects the source file in the left panel. */
async function selectSourceFile(panel: HTMLElement) {
  fireEvent.click(await within(panel).findByText("a.txt"), { detail: 1 });
}

/** Confirms a selected file operation after the dialog's exact confirmation phrase. */
async function confirmFileAction(buttonTestId: string) {
  fireEvent.click(screen.getByTestId(buttonTestId));
  fireEvent.change(screen.getByLabelText("Type yes to confirm"), { target: { value: "yes" } });
  fireEvent.click(screen.getByTestId("confirm-action-submit"));
  await waitFor(() =>
    expect(screen.queryByTestId("confirm-action-dialog")).not.toBeInTheDocument(),
  );
}

// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The controls are tested together as one user-facing region.
describe("File Explorer controls", () => {
  test("file-explorer-app/FR-008 requires yes before a consequential action", async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    const onClose = vi.fn();
    renderThemed(
      <ConfirmActionDialog
        open
        title="Delete files"
        description="Type yes"
        confirmLabel="Delete"
        onConfirm={onConfirm}
        onClose={onClose}
      />,
    );
    const submit = screen.getByTestId("confirm-action-submit");
    expect(submit).toBeDisabled();
    fireEvent.change(screen.getByLabelText("Type yes to confirm"), { target: { value: "YES" } });
    expect(submit).toBeDisabled();
    fireEvent.change(screen.getByLabelText("Type yes to confirm"), { target: { value: "yes" } });
    fireEvent.click(submit);
    await waitFor(() => expect(onConfirm).toHaveBeenCalledOnce());
    expect(onClose).toHaveBeenCalledOnce();
  });

  test("file-explorer-app/FR-008 surfaces confirmation failures", async () => {
    renderThemed(
      <ConfirmActionDialog
        open
        title="Move"
        description="Type yes"
        confirmLabel="Move"
        onConfirm={vi.fn().mockRejectedValue(new Error("permission denied"))}
        onClose={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByLabelText("Type yes to confirm"), { target: { value: "yes" } });
    fireEvent.click(screen.getByTestId("confirm-action-submit"));
    expect(await screen.findByText("permission denied")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByText("permission denied")).not.toBeInTheDocument();
  });

  test("file-explorer-app/FR-005 validates a folder name before creation", async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);
    renderThemed(<NewFolderDialog open directory="/media" onClose={vi.fn()} onCreate={onCreate} />);
    fireEvent.change(screen.getByLabelText("Folder name"), { target: { value: "../bad" } });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-submit"));
    expect(
      await screen.findByText("Enter a folder name without path separators."),
    ).toBeInTheDocument();
    expect(onCreate).not.toHaveBeenCalled();
    fireEvent.change(screen.getByLabelText("Folder name"), { target: { value: "photos" } });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-submit"));
    await waitFor(() => expect(onCreate).toHaveBeenCalledWith("photos"));
  });

  test("file-explorer-app/FR-003 selects and toggles an entry with click and Space", () => {
    const onSelect = vi.fn();
    const onOpen = vi.fn();
    renderThemed(
      <FileEntryItem
        entry={{ kind: "file", path: "/report.pdf" }}
        selected={false}
        onSelect={onSelect}
        onOpen={onOpen}
      />,
    );
    const row = screen
      .getByText("report.pdf")
      .closest("[data-test-class='file-explorer-entry']") as HTMLElement;
    expect(screen.getByText("PDF file")).toBeInTheDocument();
    fireEvent.click(row, { detail: 1 });
    fireEvent.click(row, { ctrlKey: true, detail: 1 });
    fireEvent.keyDown(row, { key: " " });
    expect(onSelect).toHaveBeenNthCalledWith(1, false);
    expect(onSelect).toHaveBeenNthCalledWith(2, true);
    expect(onSelect).toHaveBeenNthCalledWith(3, true);
    expect(onOpen).not.toHaveBeenCalled();
  });

  test("file-explorer-app/FR-003 opens an entry on Enter and double-click", () => {
    const onSelect = vi.fn();
    const onOpen = vi.fn();
    renderThemed(
      <FileEntryItem
        entry={{ kind: "file", path: "/report.pdf" }}
        selected={false}
        onSelect={onSelect}
        onOpen={onOpen}
      />,
    );
    const row = screen
      .getByText("report.pdf")
      .closest("[data-test-class='file-explorer-entry']") as HTMLElement;
    fireEvent.keyDown(row, { key: "Enter" });
    fireEvent.doubleClick(row);
    expect(onOpen).toHaveBeenCalledTimes(2);
  });

  test.each([
    ["movie.mp4", "MovieIcon"],
    ["image.jpg", "ImageIcon"],
    ["archive.zip", "ArchiveIcon"],
    ["notes.txt", "DescriptionIcon"],
    ["unknown.bin", "InsertDriveFileIcon"],
  ])("file-explorer-app/FR-003 uses the expected icon for %s", (name, iconTestId) => {
    const { container } = renderThemed(<FileEntryIcon kind="file" name={name} />);
    expect(container.querySelector("svg")).toHaveAttribute("data-testid", iconTestId);
  });

  test("file-explorer-app/FR-002 shows transfer actions only in two-panel view", () => {
    const callbacks = {
      onCopy: vi.fn(),
      onMove: vi.fn(),
      onPreview: vi.fn(),
      onDelete: vi.fn(),
      onUpload: vi.fn(),
      onNewFolder: vi.fn(),
    };
    renderThemed(
      <ExplorerActionBar doublePanel hasSelection canPreview canDelete {...callbacks} />,
    );
    expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Move" })).toBeInTheDocument();
    cleanup();
    renderThemed(
      <ExplorerActionBar doublePanel={false} hasSelection canPreview canDelete {...callbacks} />,
    );
    expect(screen.queryByRole("button", { name: "Copy" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Move" })).not.toBeInTheDocument();
  });
});

// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The workspace tests cover one connected interaction surface.
describe("File Explorer workspace", () => {
  test("file-explorer-app/FR-001 keeps Back available when a listing fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(
        // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: This response varies only by the requested test route.
        async (url: string) =>
          new Response(
            JSON.stringify(
              url === "/ls/"
                ? { files: [], directories: ["/blocked"] }
                : { error: { code: "permission_denied" } },
            ),
            { status: url === "/ls/" ? 200 : 403 },
          ),
      ),
    );
    await act(async () => {
      renderThemed(
        <FileExplorerProvider>
          <FileExplorerWorkspace />
        </FileExplorerProvider>,
      );
    });
    const left = screen.getAllByTestId("file-explorer-page")[0];
    expect(left).toBeInTheDocument();
    const panels = document.querySelectorAll<HTMLElement>(
      "[data-test-class='file-explorer-panel']",
    );
    expect(panels).toHaveLength(2);
    await act(async () => {
      fireEvent.doubleClick(await within(panels[0]).findByText("blocked"));
    });
    expect(await within(panels[0]).findByTestId("file-explorer-error")).toHaveTextContent(
      "permission denied",
    );
    await act(async () => {
      fireEvent.click(within(panels[0]).getByRole("button", { name: "Back" }));
    });
    expect(await within(panels[0]).findByText("blocked")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Show one panel" }));
    expect(document.querySelectorAll("[data-test-class='file-explorer-panel']")).toHaveLength(1);
    fireEvent.click(screen.getByRole("button", { name: "Show two panels" }));
    expect(document.querySelectorAll("[data-test-class='file-explorer-panel']")).toHaveLength(2);
    const right = document.querySelector<HTMLElement>(
      "[data-test-class='file-explorer-panel'][data-test-name='right']",
    ) as HTMLElement;
    expect(await within(right).findByText("blocked")).toBeInTheDocument();
  });

  test("file-explorer-app/FR-002 keeps two independent panel selections", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(
        async () =>
          new Response(JSON.stringify({ files: ["/a.txt", "/b.pdf"], directories: [] }), {
            status: 200,
          }),
      ),
    );
    await act(async () => {
      renderThemed(
        <FileExplorerProvider>
          <FileExplorerWorkspace />
        </FileExplorerProvider>,
      );
    });
    const panels = document.querySelectorAll<HTMLElement>(
      "[data-test-class='file-explorer-panel']",
    );
    fireEvent.click(await within(panels[0]).findByText("a.txt"), { detail: 1 });
    fireEvent.click(within(panels[0]).getByText("b.pdf"), { ctrlKey: true, detail: 1 });
    expect(within(panels[0]).getByText("2 selected")).toBeInTheDocument();
    fireEvent.click(await within(panels[1]).findByText("a.txt"), { detail: 1 });
    expect(within(panels[1]).getByText("1 selected")).toBeInTheDocument();
    expect(within(panels[0]).getByText("2 selected")).toBeInTheDocument();
    expect(screen.getByText(/Right panel selected/)).toBeInTheDocument();
  });

  test("file-explorer-app/FR-006 copies the selected file into the other panel directory", async () => {
    const { left, requests } = await renderTransferWorkspace();
    await selectSourceFile(left);
    await confirmFileAction("file-explorer-copy-button");
    expect(requests.find((request) => request.url === "/f/copy")?.body).toBe(
      JSON.stringify({ source: "/a.txt", destination: "/dest/a.txt" }),
    );
  });

  test("file-explorer-app/FR-006 moves the selected file into the other panel directory", async () => {
    const { left, requests } = await renderTransferWorkspace();
    await selectSourceFile(left);
    await confirmFileAction("file-explorer-move-button");
    expect(requests.find((request) => request.url === "/f/move")?.body).toBe(
      JSON.stringify({ source: "/a.txt", destination: "/dest/a.txt" }),
    );
  });

  test("file-explorer-app/FR-007 deletes the selected file after confirmation", async () => {
    const { left, requests } = await renderTransferWorkspace();
    await selectSourceFile(left);
    await confirmFileAction("file-explorer-delete-button");
    expect(
      requests.some((request) => request.url === "/f/a.txt" && request.method === "DELETE"),
    ).toBe(true);
  });

  test("file-explorer-app/FR-003 opens a preview for the selected file", async () => {
    const { left } = await renderTransferWorkspace();
    await selectSourceFile(left);
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    expect(screen.getByTitle("File preview")).toHaveAttribute("src", "/f/a.txt");
  });

  test("file-explorer-app/FR-005 creates a folder in the active panel", async () => {
    const { requests } = await renderTransferWorkspace();
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockImplementationOnce(async (url: string, init?: RequestInit) => {
      requests.push({ url, method: init?.method ?? "GET", body: init?.body as string | undefined });
      if (url === "/f/mkdir") return new Response(null, { status: 204 });
      return new Response(JSON.stringify({ files: ["/a.txt"], directories: ["/dest"] }), {
        status: 200,
      });
    });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-button"));
    fireEvent.change(screen.getByLabelText("Folder name"), { target: { value: "new" } });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-submit"));
    await waitFor(() => expect(requests.some((request) => request.url === "/f/mkdir")).toBe(true));
    expect(requests.find((request) => request.url === "/f/mkdir")?.body).toBe(
      JSON.stringify({ path: "/new" }),
    );
  });
});
