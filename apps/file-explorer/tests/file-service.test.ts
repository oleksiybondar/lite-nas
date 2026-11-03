import { afterEach, expect, test, vi } from "vitest";
import {
  canPreviewFile,
  createDirectory,
  deleteEntry,
  filterFileEntries,
  getOperationJob,
  isHiddenFileEntry,
  listDirectory,
  transferEntry,
} from "../src/file-service";
import { displayName, fileServicePath, parentPath } from "../src/paths";
import {
  cachedJobsSchema,
  hostPathSchema,
  newFolderNameSchema,
  operationJobSchema,
} from "../src/schemas";
import { listingEntry } from "./listing-fixture";

/** Valid pending operation used by File Service response fixtures. */
const job = {
  job_id: "a".repeat(48),
  state: "pending",
  percent: 0,
  completed_bytes: 0,
  total_bytes: 10,
  completed_items: 0,
  total_items: 1,
};

afterEach(() => vi.unstubAllGlobals());

/** Stubs one File Service response and returns the observable fetch mock. */
function stubFetchResponse(response: Response) {
  const fetchMock = vi.fn().mockResolvedValue(response);
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

test.each([
  ["root", "/", "/"],
  ["nested file", "/some/thing.txt", "thing.txt"],
] as const)("file-explorer-app/IR-001 displays the %s name", (_case, path, expected) => {
  expect(displayName(path)).toBe(expected);
});

test.each([
  ["root", "/", null],
  ["top-level child", "/some", "/"],
  ["nested child", "/some/thing.txt", "/some"],
] as const)("file-explorer-app/IR-001 resolves the parent of a %s path", (_case, path, expected) => {
  expect(parentPath(path)).toBe(expected);
});

test("file-explorer-app/IR-001 escapes a File Service resource path", () => {
  expect(fileServicePath("/f", "/some/a #.txt")).toBe("/f/some/a%20%23.txt");
});

test("file-explorer-app/IR-001 accepts an absolute host path", () => {
  expect(hostPathSchema.safeParse("/some/thing").success).toBe(true);
});

test.each([
  [{ include: [".JPG"] }, ["/root/photo.jpg", "/root/folder"]],
  [{ exclude: ["jpg"] }, ["/root/notes.txt", "/root/folder"]],
  [{ include: ["jpg", "txt"], exclude: ["txt"] }, ["/root/photo.jpg", "/root/folder"]],
])("file-explorer-app/FR-001 filters file extensions while retaining directories", (filter, expected) => {
  const entries = [
    listingEntry("/root/photo.jpg", "file"),
    listingEntry("/root/notes.txt", "file"),
    listingEntry("/root/folder", "directory"),
  ];
  expect(filterFileEntries(entries, filter).map((entry) => entry.path)).toEqual(expected);
});

test.each([
  "",
  "relative",
  "/trailing/",
  "/a//b",
  "/a/../b",
  "/a/./b",
  "/a\\b",
  "/a\0b",
])("file-explorer-app/IR-001 rejects malformed host path %j", (path) =>
  expect(hostPathSchema.safeParse(path).success).toBe(false));

test("file-explorer-app/FR-001 sorts listing children and retains metadata", async () => {
  const fileWithMetadata = {
    ...listingEntry("/root/file2.txt", "file"),
    size: 42,
    owner: "alice",
    group: "users",
    permissions: "rw-r-----",
    last_modified: "2026-10-23T07:00:00Z",
  };
  const signal = new AbortController().signal;
  const fetchMock = stubFetchResponse(
    new Response(
      JSON.stringify([
        listingEntry("/root/file10.txt", "file"),
        fileWithMetadata,
        listingEntry("/root/dir10", "directory"),
        listingEntry("/root/dir2", "directory"),
      ]),
      { status: 200 },
    ),
  );
  const entries = await listDirectory("/root", signal);
  expect({ entries, request: fetchMock.mock.calls[0] }).toEqual({
    entries: [
      listingEntry("/root/dir2", "directory"),
      listingEntry("/root/dir10", "directory"),
      fileWithMetadata,
      listingEntry("/root/file10.txt", "file"),
    ],
    request: ["/ls/root", { credentials: "same-origin", signal }],
  });
});

test.each([
  ["unrelated child", [listingEntry("/elsewhere/file", "file")], 200, "invalid directory listing"],
  ["service denial", { error: { code: "permission_denied" } }, 403, "permission denied"],
] as const)("file-explorer-app/FR-001 reports a %s listing failure", async (_case, body, status, message) => {
  stubFetchResponse(new Response(JSON.stringify(body), { status }));
  await expect(listDirectory("/root", new AbortController().signal)).rejects.toThrow(message);
});

test("file-explorer-app/IR-001 rejects malformed listing metadata", async () => {
  stubFetchResponse(
    new Response(JSON.stringify([{ ...listingEntry("/root/a", "file"), size: -1 }]), {
      status: 200,
    }),
  );
  await expect(listDirectory("/root", new AbortController().signal)).rejects.toThrow();
});

test("file-explorer-app/FR-005 creates a directory at the requested path", async () => {
  const fetchMock = stubFetchResponse(new Response(null, { status: 204 }));
  await createDirectory("/root/new");
  expect({
    url: fetchMock.mock.calls[0][0],
    body: JSON.parse(fetchMock.mock.calls[0][1].body),
  }).toEqual({ url: "/f/mkdir", body: { path: "/root/new", private: false } });
});

test.each([
  ["folder", true],
  ["", false],
  [".", false],
  ["..", false],
  ["a/b", false],
  ["a\\b", false],
  ["a\nb", false],
])("file-explorer-app/FR-005 validates folder name %j", (name, expected) => {
  expect(newFolderNameSchema.safeParse(name).success).toBe(expected);
});

test("file-explorer-app/FR-006 submits a transfer with source and destination", async () => {
  const fetchMock = stubFetchResponse(new Response(JSON.stringify(job), { status: 202 }));
  const result = await transferEntry("copy", "/root/a", "/root/b");
  expect({
    result,
    url: fetchMock.mock.calls[0][0],
    body: JSON.parse(fetchMock.mock.calls[0][1].body),
  }).toEqual({
    result: job,
    url: "/f/copy",
    body: { source: "/root/a", destination: "/root/b" },
  });
});

test.each([
  ["file", new Response(null, { status: 204 }), null],
  ["directory", new Response(JSON.stringify(job), { status: 202 }), job],
] as const)("file-explorer-app/FR-007 deletes a %s and returns its completion result", async (kind, response, result) => {
  const fetchMock = stubFetchResponse(response);
  const deletion = await deleteEntry({ kind, path: "/root/a" });
  expect({ result: deletion, url: fetchMock.mock.calls[0][0] }).toEqual({
    result,
    url: "/f/root/a",
  });
});

test("file-explorer-app/FR-009 retrieves a server operation job", async () => {
  const fetchMock = stubFetchResponse(new Response(JSON.stringify(job), { status: 200 }));
  const result = await getOperationJob(job.job_id);
  expect({ result, url: fetchMock.mock.calls[0][0] }).toEqual({
    result: job,
    url: `/f/jobs/${job.job_id}`,
  });
});

test("file-explorer-app/IR-001 rejects a relative directory before requesting the service", async () => {
  const fetchMock = stubFetchResponse(new Response(null, { status: 500 }));
  const error = await createDirectory("relative").then(
    () => null,
    (caught: unknown) => caught,
  );
  expect({ rejected: error instanceof Error, requestCount: fetchMock.mock.calls.length }).toEqual({
    rejected: true,
    requestCount: 0,
  });
});

test("file-explorer-app/IR-001 rejects an invalid job ID before requesting the service", async () => {
  const fetchMock = stubFetchResponse(new Response(null, { status: 500 }));
  const error = await getOperationJob("bad-id").then(
    () => null,
    (caught: unknown) => caught,
  );
  expect({ rejected: error instanceof Error, requestCount: fetchMock.mock.calls.length }).toEqual({
    rejected: true,
    requestCount: 0,
  });
});

test.each([
  {
    caseName: "transfer",
    action: () => transferEntry("move", "/a", "/b"),
    status: 403,
    body: { error: { code: "permission_denied" } },
    message: "permission denied",
  },
  {
    caseName: "deletion",
    action: () => deleteEntry({ kind: "file", path: "/a" }),
    status: 500,
    body: null,
    message: "HTTP 500",
  },
])("file-explorer-app/IR-001 reports a failed $caseName response", async ({
  action,
  status,
  body,
  message,
}) => {
  stubFetchResponse(new Response(JSON.stringify(body), { status }));
  await expect(action()).rejects.toThrow(message);
});

test.each([
  ["/a.PDF", "file", true],
  ["/a.png", "file", true],
  ["/a.mp4", "file", true],
  ["/a.mkv", "file", true],
  ["/a.mp3", "file", true],
  ["/a.rs", "file", true],
  ["/a.zip", "file", false],
  ["/a.pdf", "directory", false],
] as const)("file-explorer-app/FR-003 preview support for %s", (path, kind, expected) => {
  expect(canPreviewFile({ kind, path })).toBe(expected);
});

test.each([
  [".profile", true],
  [".config", true],
  ["notes.txt", false],
])("file-explorer-app/FR-003 identifies hidden entry names %s", (name, expected) => {
  expect(isHiddenFileEntry({ name })).toBe(expected);
});

test.each([
  ["valid job", job, true],
  ["invalid percentage", { ...job, percent: 101 }, false],
])("file-explorer-app/FR-009 validates an %s response", (_case, value, expected) => {
  expect(operationJobSchema.safeParse(value).success).toBe(expected);
});

test.each([
  ["valid cached job", [{ id: job.job_id, label: "copy a" }], true],
  ["invalid cached job", [{ id: "bad", label: "copy a" }], false],
])("file-explorer-app/FR-009 validates an %s reference", (_case, value, expected) => {
  expect(cachedJobsSchema.safeParse(value).success).toBe(expected);
});
