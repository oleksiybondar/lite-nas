import { createTheme, ThemeProvider } from "@mui/material/styles";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  waitForElementToBeRemoved,
  within,
} from "@testing-library/react";
import { afterEach, describe, expect, test, vi } from "vitest";
import { ConfirmActionDialog } from "../src/ConfirmActionDialog";
import { ExplorerActionBar } from "../src/ExplorerActionBar";
import { FileEntryIcon } from "../src/FileEntryIcon";
import { FileEntryItem } from "../src/FileEntryItem";
import { FileExplorerProvider } from "../src/FileExplorerContext";
import { FileExplorerWorkspace } from "../src/FileExplorerWorkspace";
import { NewFolderDialog } from "../src/NewFolderDialog";
import { listingEntry } from "./listing-fixture";

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

/** Opens the new-folder dialog with observable creation and close callbacks. */
function renderNewFolderDialog(onCreate: (name: string) => Promise<void> = vi.fn()) {
  const onClose = vi.fn();
  renderThemed(<NewFolderDialog open directory="/media" onClose={onClose} onCreate={onCreate} />);
  return { onCreate, onClose };
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
      new Response(
        JSON.stringify([
          listingEntry("/a.txt", "file"),
          listingEntry("/movie.mp4", "file"),
          listingEntry("/sound.mp3", "file"),
          listingEntry("/poster.png", "file"),
          listingEntry("/document.pdf", "file"),
          listingEntry("/.profile", "file"),
          listingEntry("/.config", "directory"),
          listingEntry("/dest", "directory"),
        ]),
        { status: 200 },
      ),
    "/ls/dest": () => new Response(JSON.stringify([]), { status: 200 }),
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

/** Loads both panels at root with hidden file and folder entries. */
async function renderHiddenFilesWorkspace() {
  vi.stubGlobal(
    "fetch",
    vi.fn(
      async () =>
        new Response(
          JSON.stringify([
            listingEntry("/.profile", "file"),
            listingEntry("/.config", "directory"),
          ]),
          { status: 200 },
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
  return {
    left: document.querySelector<HTMLElement>(
      "[data-test-class='file-explorer-panel'][data-test-name='left']",
    ) as HTMLElement,
    right: document.querySelector<HTMLElement>(
      "[data-test-class='file-explorer-panel'][data-test-name='right']",
    ) as HTMLElement,
  };
}

/** Opens root in both panels and loads one directory that will deny access. */
async function renderFailedListingWorkspace() {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockImplementation(
        async (url: string) =>
          new Response(
            JSON.stringify(
              url === "/ls/"
                ? [listingEntry("/blocked", "directory")]
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
  return {
    left: document.querySelector<HTMLElement>(
      "[data-test-class='file-explorer-panel'][data-test-name='left']",
    ) as HTMLElement,
    right: document.querySelector<HTMLElement>(
      "[data-test-class='file-explorer-panel'][data-test-name='right']",
    ) as HTMLElement,
  };
}

/** Selects the source file in the left panel. */
async function selectSourceFile(panel: HTMLElement) {
  fireEvent.click(await within(panel).findByText("a.txt"), { detail: 1 });
}

/** Confirms a selected file operation after the dialog's exact confirmation phrase. */
async function confirmFileAction(buttonTestId: string) {
  fireEvent.click(screen.getByTestId(buttonTestId));
  fireEvent.change(screen.getByLabelText("Type yes to confirm"), { target: { value: "yes" } });
  await act(async () => {
    fireEvent.click(screen.getByTestId("confirm-action-submit"));
  });
}

/** Opens a text preview and waits until its content has loaded. */
async function openTextPreview(panel: HTMLElement) {
  vi.mocked(fetch).mockImplementationOnce(
    async () => new Response("const value = 1;", { status: 200 }),
  );
  await selectSourceFile(panel);
  fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
  await screen.findByText("TXT source");
  return screen.getByRole("dialog");
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
    const disabledInitially = submit.disabled;
    fireEvent.change(screen.getByLabelText("Type yes to confirm"), { target: { value: "YES" } });
    const disabledForUppercase = submit.disabled;
    fireEvent.change(screen.getByLabelText("Type yes to confirm"), { target: { value: "yes" } });
    await act(async () => {
      fireEvent.click(submit);
    });
    expect({
      disabledInitially,
      disabledForUppercase,
      confirmedOnce: onConfirm.mock.calls.length === 1,
      closedOnce: onClose.mock.calls.length === 1,
    }).toEqual({
      disabledInitially: true,
      disabledForUppercase: true,
      confirmedOnce: true,
      closedOnce: true,
    });
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
    const errorAppeared = await screen.findByText("permission denied");
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    });
    expect({
      errorShown: errorAppeared.textContent,
      errorRemovedOnCancel: screen.queryByText("permission denied") === null,
    }).toEqual({ errorShown: "permission denied", errorRemovedOnCancel: true });
  });

  test("file-explorer-app/FR-005 rejects a folder name with a path separator", async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);
    renderNewFolderDialog(onCreate);
    fireEvent.change(screen.getByLabelText("Folder name"), { target: { value: "../bad" } });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-submit"));
    const validationMessage = await screen.findByText(
      "Enter a folder name without path separators.",
    );
    expect({
      message: validationMessage.textContent,
      creationCalls: onCreate.mock.calls.length,
    }).toEqual({ message: "Enter a folder name without path separators.", creationCalls: 0 });
  });

  test("file-explorer-app/FR-005 submits a valid folder name", async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);
    renderNewFolderDialog(onCreate);
    fireEvent.change(screen.getByLabelText("Folder name"), { target: { value: "photos" } });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-submit"));
    await waitFor(() => expect(onCreate.mock.calls).toEqual([["photos"]]));
  });

  test("file-explorer-app/FR-003 selects and toggles an entry with click and Space", () => {
    const onSelect = vi.fn();
    const onOpen = vi.fn();
    renderThemed(
      <FileEntryItem
        entry={listingEntry("/report.pdf", "file")}
        selected={false}
        onSelect={onSelect}
        onOpen={onOpen}
      />,
    );
    const row = screen
      .getByText("report.pdf")
      .closest("[data-test-class='file-explorer-entry']") as HTMLElement;
    fireEvent.click(row, { detail: 1 });
    fireEvent.click(row, { ctrlKey: true, detail: 1 });
    fireEvent.keyDown(row, { key: " " });
    expect({
      selectionModes: onSelect.mock.calls.map(([multiple]) => multiple),
      openCalls: onOpen.mock.calls.length,
    }).toEqual({ selectionModes: [false, true, true], openCalls: 0 });
  });

  test("file-explorer-app/FR-003 displays size, octal permissions, ownership, and modification time", () => {
    const entry = {
      ...listingEntry("/report.pdf", "file"),
      size: 1_572_864,
      owner: "alice",
      group: "media",
      permissions: "rwxr-x--x",
      last_modified: "2026-10-23T07:00:00Z",
    };
    renderThemed(
      <FileEntryItem entry={entry} selected={false} onSelect={vi.fn()} onOpen={vi.fn()} />,
    );
    const row = screen
      .getByText("report.pdf")
      .closest("[data-test-class='file-explorer-entry']") as HTMLElement;
    const rowText = row.textContent ?? "";
    expect({
      sizeTypePermissionsOwnerGroup: rowText.includes("1.5 MBPDF file0751alicemedia"),
      modificationYear: rowText.includes("2026"),
    }).toEqual({ sizeTypePermissionsOwnerGroup: true, modificationYear: true });
  });

  test("file-explorer-app/FR-003 opens an entry on Enter and double-click", () => {
    const onSelect = vi.fn();
    const onOpen = vi.fn();
    renderThemed(
      <FileEntryItem
        entry={listingEntry("/report.pdf", "file")}
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
    ["app.js", "JavascriptIcon"],
    ["view.html", "HtmlIcon"],
    ["config.json", "DataObjectIcon"],
    ["service.go", "CodeIcon"],
    ["track.mp3", "MusicNoteIcon"],
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
      showHidden: true,
      onToggleHidden: vi.fn(),
    };
    renderThemed(
      <ExplorerActionBar doublePanel hasSelection canPreview canDelete {...callbacks} />,
    );
    const visibleInDoublePanel = [
      screen.queryByRole("button", { name: "Copy" }) !== null,
      screen.queryByRole("button", { name: "Move" }) !== null,
    ];
    cleanup();
    renderThemed(
      <ExplorerActionBar doublePanel={false} hasSelection canPreview canDelete {...callbacks} />,
    );
    const visibleInSinglePanel = [
      screen.queryByRole("button", { name: "Copy" }) !== null,
      screen.queryByRole("button", { name: "Move" }) !== null,
    ];
    expect({ visibleInDoublePanel, visibleInSinglePanel }).toEqual({
      visibleInDoublePanel: [true, true],
      visibleInSinglePanel: [false, false],
    });
  });
});

// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The workspace tests cover one connected interaction surface.
describe("File Explorer workspace", () => {
  test("file-explorer-app/FR-001 keeps Back available when a listing fails", async () => {
    const { left } = await renderFailedListingWorkspace();
    fireEvent.doubleClick(await within(left).findByText("blocked"));
    const listingError = await within(left).findByTestId("file-explorer-error");
    const backButton = within(left).getByRole("button", { name: "Back" });
    const backAvailableAfterFailure = backButton.disabled === false;
    fireEvent.click(backButton);
    await within(left).findByText("blocked");
    expect({
      deniedMessage: listingError.textContent,
      backAvailableAfterFailure,
      backDisabledAtParent: backButton.disabled,
      returnedToParent: within(left).getByText("blocked") !== null,
    }).toEqual({
      deniedMessage: expect.stringContaining("permission denied"),
      backAvailableAfterFailure: true,
      backDisabledAtParent: true,
      returnedToParent: true,
    });
  });

  test("file-explorer-app/FR-002 preserves panel locations when switching views", async () => {
    const { left, right } = await renderFailedListingWorkspace();
    fireEvent.doubleClick(await within(left).findByText("blocked"));
    await within(left).findByTestId("file-explorer-error");
    fireEvent.click(screen.getByRole("button", { name: "Show one panel" }));
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Show two panels" }));
    });
    expect({
      leftStillAtDeniedDirectory: within(left).queryByTestId("file-explorer-error") !== null,
      rightStillAtRoot: within(right).queryByText("blocked") !== null,
    }).toEqual({ leftStillAtDeniedDirectory: true, rightStillAtRoot: true });
  });

  test("file-explorer-app/FR-002 keeps two independent panel selections", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(
        async () =>
          new Response(
            JSON.stringify([listingEntry("/a.txt", "file"), listingEntry("/b.pdf", "file")]),
            {
              status: 200,
            },
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
    const panels = document.querySelectorAll<HTMLElement>(
      "[data-test-class='file-explorer-panel']",
    );
    fireEvent.click(await within(panels[0]).findByText("a.txt"), { detail: 1 });
    fireEvent.click(within(panels[0]).getByText("b.pdf"), { ctrlKey: true, detail: 1 });
    fireEvent.click(await within(panels[1]).findByText("a.txt"), { detail: 1 });
    expect({
      leftSelection: within(panels[0]).getByText("2 selected").textContent,
      rightSelection: within(panels[1]).getByText("1 selected").textContent,
    }).toEqual({ leftSelection: "2 selected", rightSelection: "1 selected" });
  });

  test("file-explorer-app/FR-003 toggles hidden files and folders, enabled by default", async () => {
    const { left, right } = await renderHiddenFilesWorkspace();
    await Promise.all([
      within(left).findByText(".profile"),
      within(left).findByText(".config"),
      within(right).findByText(".profile"),
      within(right).findByText(".config"),
    ]);
    const toggle = screen.getByTestId("file-explorer-hidden-toggle");
    const enabledByDefault = toggle.getAttribute("aria-pressed");
    fireEvent.click(screen.getByRole("button", { name: "Hide hidden files" }));
    const hiddenWhenDisabled =
      within(left).queryByText(".profile") === null &&
      within(left).queryByText(".config") === null &&
      within(right).queryByText(".profile") === null &&
      within(right).queryByText(".config") === null;
    const iconWhenDisabled = toggle.querySelector("svg")?.getAttribute("data-testid");
    fireEvent.click(screen.getByRole("button", { name: "Show hidden files" }));
    const visibleWhenEnabled =
      within(left).queryByText(".profile") !== null &&
      within(left).queryByText(".config") !== null &&
      within(right).queryByText(".profile") !== null &&
      within(right).queryByText(".config") !== null;
    const iconWhenEnabled = toggle.querySelector("svg")?.getAttribute("data-testid");
    expect({
      enabledByDefault,
      hiddenWhenDisabled,
      visibleWhenEnabled,
      iconWhenDisabled,
      iconWhenEnabled,
    }).toEqual({
      enabledByDefault: "true",
      hiddenWhenDisabled: true,
      visibleWhenEnabled: true,
      iconWhenDisabled: "VisibilityOffIcon",
      iconWhenEnabled: "VisibilityIcon",
    });
  });

  test("file-explorer-app/FR-003 excludes hidden selections from actions when hidden", async () => {
    const { left } = await renderHiddenFilesWorkspace();
    fireEvent.click(await within(left).findByText(".profile"), { detail: 1 });
    fireEvent.click(screen.getByRole("button", { name: "Hide hidden files" }));
    expect({
      selectedCount: within(left).getByText("0 selected").textContent,
      deleteDisabled: screen.getByRole("button", { name: "Delete" }).hasAttribute("disabled"),
    }).toEqual({ selectedCount: "0 selected", deleteDisabled: true });
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
    vi.mocked(fetch).mockImplementationOnce(
      async () => new Response("<script>alert('unsafe')</script>", { status: 200 }),
    );
    await selectSourceFile(left);
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    await waitFor(() =>
      expect({
        sourceText: document.querySelector("code")?.textContent,
        hasExecutableScript: document.querySelector("code script") !== null,
      }).toEqual({
        sourceText: "<script>alert('unsafe')</script>",
        hasExecutableScript: false,
      }),
    );
  });

  test("file-explorer-app/FR-003 reports source preview request errors", async () => {
    const { left } = await renderTransferWorkspace();
    vi.mocked(fetch).mockImplementationOnce(async () => new Response(null, { status: 403 }));
    await selectSourceFile(left);
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    expect(await screen.findByText("Could not load file (HTTP 403).")).toBeInTheDocument();
  });

  test("file-explorer-app/FR-003 reports source preview network errors", async () => {
    const { left } = await renderTransferWorkspace();
    vi.mocked(fetch).mockImplementationOnce(() => Promise.reject("network failure"));
    await selectSourceFile(left);
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    expect(await screen.findByText("Could not load this file.")).toBeInTheDocument();
  });

  test("file-explorer-app/FR-003 ignores a source request failure after preview closes", async () => {
    const { left } = await renderTransferWorkspace();
    let rejectPreviewRequest: (reason: unknown) => void = () => undefined;
    const pendingPreviewRequest = new Promise<Response>((_resolve, reject) => {
      rejectPreviewRequest = reject;
    });
    vi.mocked(fetch).mockImplementationOnce(() => pendingPreviewRequest);
    await selectSourceFile(left);
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    await screen.findByLabelText("Loading text preview");
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    await waitForElementToBeRemoved(() => screen.getByRole("dialog"));
    await act(async () => {
      rejectPreviewRequest(new Error("request closed"));
      await Promise.resolve();
    });
    expect(screen.queryByText("request closed")).not.toBeInTheDocument();
  });

  test("file-explorer-app/FR-003 previews supported video with HTML media controls", async () => {
    const { left } = await renderTransferWorkspace();
    fireEvent.click(await within(left).findByText("movie.mp4"), { detail: 1 });
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    const video = await waitFor(() => {
      const current = document.querySelector("video");
      if (!current) throw new Error("video preview has not rendered");
      return current;
    });
    expect({
      source: video.getAttribute("src"),
      controlsEnabled: video.controls,
      preload: video.preload,
    }).toEqual({ source: "/f/movie.mp4", controlsEnabled: true, preload: "metadata" });
  });

  test("file-explorer-app/FR-003 previews supported audio with HTML media controls", async () => {
    const { left } = await renderTransferWorkspace();
    fireEvent.click(await within(left).findByText("sound.mp3"), { detail: 1 });
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    const audio = await waitFor(() => {
      const current = document.querySelector("audio");
      if (!current) throw new Error("audio preview has not rendered");
      return current;
    });
    expect({
      source: audio.getAttribute("src"),
      controlsEnabled: audio.controls,
      preload: audio.preload,
    }).toEqual({ source: "/f/sound.mp3", controlsEnabled: true, preload: "metadata" });
  });

  test("file-explorer-app/FR-003 previews supported images directly", async () => {
    const { left } = await renderTransferWorkspace();
    fireEvent.click(await within(left).findByText("poster.png"), { detail: 1 });
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    expect(await screen.findByAltText("poster.png")).toHaveAttribute("src", "/f/poster.png");
  });

  test("file-explorer-app/FR-003 previews PDFs with the browser document viewer", async () => {
    const { left } = await renderTransferWorkspace();
    fireEvent.click(await within(left).findByText("document.pdf"), { detail: 1 });
    fireEvent.click(screen.getByTestId("file-explorer-preview-button"));
    expect(await screen.findByTitle("PDF preview")).toHaveAttribute("src", "/f/document.pdf");
  });

  test("file-explorer-app/FR-005 creates a folder in the active panel", async () => {
    const { requests } = await renderTransferWorkspace();
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockImplementationOnce(async (url: string, init?: RequestInit) => {
      requests.push({ url, method: init?.method ?? "GET", body: init?.body as string | undefined });
      if (url === "/f/mkdir") return new Response(null, { status: 204 });
      return new Response(
        JSON.stringify([listingEntry("/a.txt", "file"), listingEntry("/dest", "directory")]),
        {
          status: 200,
        },
      );
    });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-button"));
    fireEvent.change(screen.getByLabelText("Folder name"), { target: { value: "new" } });
    fireEvent.click(screen.getByTestId("file-explorer-new-folder-submit"));
    await waitForElementToBeRemoved(() => screen.getByTestId("file-explorer-new-folder-dialog"));
    const mkdirRequest = requests.find((request) => request.url === "/f/mkdir");
    expect({
      requestSubmitted: mkdirRequest !== undefined,
      requestBody: mkdirRequest?.body,
    }).toEqual({ requestSubmitted: true, requestBody: JSON.stringify({ path: "/dest/new" }) });
  });
});

test("file-explorer-app/FR-003 shows a close control in the large preview dialog", async () => {
  const { left } = await renderTransferWorkspace();
  const dialog = await openTextPreview(left);
  expect({
    closeButtonVisible: screen.getByRole("button", { name: "Close preview" }) !== null,
    largeWidth: dialog.classList.contains("MuiDialog-paperWidthLg"),
  }).toEqual({ closeButtonVisible: true, largeWidth: true });
});

test("file-explorer-app/FR-003 closes a preview with its close control", async () => {
  const { left } = await renderTransferWorkspace();
  await openTextPreview(left);
  fireEvent.click(screen.getByRole("button", { name: "Close preview" }));
  await waitForElementToBeRemoved(() => screen.getByRole("dialog"));
  expect(screen.queryByRole("dialog")).toBeNull();
});
