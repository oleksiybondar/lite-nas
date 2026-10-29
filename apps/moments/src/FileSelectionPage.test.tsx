// @vitest-environment jsdom

import {
  FileSelectionDialog,
  FileSelectionPage,
  FolderSelectionDialog,
} from "@lite-nas/file-browser";
import { cleanup, fireEvent, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { renderWithLanguage as render } from "./test-utils/language";

/** Returns one valid File Service listing entry for picker interaction tests. */
function entry(name: string, kind: "file" | "directory") {
  return {
    path: `/home/oleksiybondar/${name}`,
    name,
    kind,
    size: null,
    owner: null,
    group: null,
    permissions: null,
    last_modified: null,
  };
}

/** Supplies one directory with visible and hidden choices. */
function stubListing(): ReturnType<typeof vi.fn> {
  const fetchMock = vi.fn(
    async () =>
      new Response(
        JSON.stringify([
          entry("pictures", "directory"),
          entry(".private", "directory"),
          entry("photo.JPG", "file"),
          entry("notes.txt", "file"),
          entry(".secret.jpg", "file"),
        ]),
        { status: 200 },
      ),
  );
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

test("moments/FR-004 selects the starting folder", async () => {
  stubListing();
  const onSelect = vi.fn();
  render(
    <FileSelectionPage
      kind="directory"
      initialDirectory="/home/oleksiybondar"
      onSelect={onSelect}
    />,
  );
  await screen.findByText("pictures");
  fireEvent.click(screen.getByRole("button", { name: "Select folder" }));
  expect(onSelect).toHaveBeenCalledWith("/home/oleksiybondar");
});

test("moments/FR-004 picker listings use the supplied renewing request", async () => {
  const request = vi.fn(
    async (_path: string, _init?: RequestInit) =>
      new Response(JSON.stringify([entry("pictures", "directory")]), { status: 200 }),
  );
  render(
    <FileSelectionPage
      kind="directory"
      initialDirectory="/home/oleksiybondar"
      onSelect={vi.fn()}
      request={request}
    />,
  );
  await screen.findByText("pictures");
  expect(request.mock.calls[0][0]).toBe("/ls/home/oleksiybondar");
});

test("moments/FR-004 hides dot-prefixed files and folders by default", async () => {
  stubListing();
  render(
    <FileSelectionPage kind="file" initialDirectory="/home/oleksiybondar" onSelect={vi.fn()} />,
  );
  await screen.findByText("photo.JPG");
  expect(screen.queryByText(".secret.jpg")).toBeNull();
});

test("moments/FR-004 selects an allowed file", async () => {
  stubListing();
  const onSelect = vi.fn();
  render(
    <FileSelectionPage
      kind="file"
      initialDirectory="/home/oleksiybondar"
      extensionFilter={{ include: ["jpg"] }}
      onSelect={onSelect}
    />,
  );
  const row = await screen.findByText("photo.JPG");
  fireEvent.click(row, { detail: 1 });
  fireEvent.click(screen.getByRole("button", { name: "Select file" }));
  expect(onSelect).toHaveBeenCalledWith("/home/oleksiybondar/photo.JPG");
});

test("moments/FR-004 disables file selection after a folder row is clicked", async () => {
  stubListing();
  render(
    <FileSelectionPage
      kind="file"
      initialDirectory="/home/oleksiybondar"
      extensionFilter={{ include: ["jpg"] }}
      onSelect={vi.fn()}
    />,
  );
  fireEvent.click(await screen.findByText("photo.JPG"), { detail: 1 });
  fireEvent.click(screen.getByText("pictures"), { detail: 1 });
  expect((screen.getByRole("button", { name: "Select file" }) as HTMLButtonElement).disabled).toBe(
    true,
  );
});

test("moments/FR-004 excludes unlisted file extensions", async () => {
  stubListing();
  render(
    <FileSelectionPage
      kind="file"
      initialDirectory="/home/oleksiybondar"
      extensionFilter={{ include: ["jpg"] }}
      onSelect={vi.fn()}
    />,
  );
  await screen.findByText("photo.JPG");
  expect(screen.queryByText("notes.txt")).toBeNull();
});

test("moments/FR-004 closes the folder dialog from its title control", async () => {
  stubListing();
  const onClose = vi.fn();
  render(
    <FolderSelectionDialog
      open
      initialDirectory="/home/oleksiybondar"
      onSelect={vi.fn()}
      onClose={onClose}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Close folder selector" }));
  expect(onClose).toHaveBeenCalledOnce();
});

test("moments/FR-004 confirms the file selection and closes its dialog", async () => {
  stubListing();
  const onSelect = vi.fn();
  const onClose = vi.fn();
  render(
    <FileSelectionDialog
      open
      initialDirectory="/home/oleksiybondar"
      onSelect={onSelect}
      onClose={onClose}
    />,
  );
  fireEvent.click(await screen.findByText("photo.JPG"), { detail: 1 });
  fireEvent.click(screen.getByRole("button", { name: "Select file" }));
  expect({ selected: onSelect.mock.calls, closed: onClose.mock.calls.length }).toEqual({
    selected: [["/home/oleksiybondar/photo.JPG"]],
    closed: 1,
  });
});
