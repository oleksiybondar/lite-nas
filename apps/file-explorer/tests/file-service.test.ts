import { afterEach, expect, test, vi } from "vitest";
import {
  canPreviewFile,
  createDirectory,
  deleteEntry,
  getOperationJob,
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

test("file-explorer-app/IR-001 validates host paths", () => {
  expect(displayName("/")).toBe("/");
  expect(displayName("/some/thing.txt")).toBe("thing.txt");
  expect(parentPath("/")).toBeNull();
  expect(parentPath("/some")).toBe("/");
  expect(parentPath("/some/thing.txt")).toBe("/some");
  expect(fileServicePath("/f", "/some/a #.txt")).toBe("/f/some/a%20%23.txt");
  for (const path of [
    "",
    "relative",
    "/trailing/",
    "/a//b",
    "/a/../b",
    "/a/./b",
    "/a\\b",
    "/a\0b",
  ]) {
    expect(hostPathSchema.safeParse(path).success).toBe(false);
  }
});

test("file-explorer-app/FR-001 sorts listing children and retains metadata", async () => {
  const fileWithMetadata = {
    ...listingEntry("/root/file2.txt", "file"),
    size: 42,
    owner: "alice",
    group: "users",
    permissions: "rw-r-----",
    last_modified: "2026-10-23T07:00:00Z",
  };
  const fetchMock = vi
    .fn()
    .mockResolvedValue(
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
  vi.stubGlobal("fetch", fetchMock);
  const signal = new AbortController().signal;
  expect(await listDirectory("/root", signal)).toEqual([
    listingEntry("/root/dir2", "directory"),
    listingEntry("/root/dir10", "directory"),
    fileWithMetadata,
    listingEntry("/root/file10.txt", "file"),
  ]);
  expect(fetchMock).toHaveBeenCalledWith("/ls/root", { credentials: "same-origin", signal });
});

test("file-explorer-app/FR-001 rejects unrelated children and service errors", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify([listingEntry("/elsewhere/file", "file")]), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: { code: "permission_denied" },
          }),
          { status: 403 },
        ),
      ),
  );
  await expect(listDirectory("/root", new AbortController().signal)).rejects.toThrow(
    "invalid directory listing",
  );
  await expect(listDirectory("/root", new AbortController().signal)).rejects.toThrow(
    "permission denied",
  );
});

test("file-explorer-app/IR-001 rejects malformed listing metadata", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(JSON.stringify([{ ...listingEntry("/root/a", "file"), size: -1 }]), {
        status: 200,
      }),
    ),
  );
  await expect(listDirectory("/root", new AbortController().signal)).rejects.toThrow();
});

test("file-explorer-app/FR-005 creates a directory at the requested path", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
  vi.stubGlobal("fetch", fetchMock);
  await createDirectory("/root/new");
  expect(fetchMock.mock.calls[0][0]).toBe("/f/mkdir");
  expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ path: "/root/new" });
});

test("file-explorer-app/FR-005 accepts only single child folder names", () => {
  expect(newFolderNameSchema.safeParse("folder").success).toBe(true);
  for (const name of ["", ".", "..", "a/b", "a\\b", "a\nb"]) {
    expect(newFolderNameSchema.safeParse(name).success).toBe(false);
  }
});

test("file-explorer-app/FR-006 submits a transfer with source and destination", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(job), { status: 202 }));
  vi.stubGlobal("fetch", fetchMock);
  expect(await transferEntry("copy", "/root/a", "/root/b")).toEqual(job);
  expect(fetchMock.mock.calls[0][0]).toBe("/f/copy");
  expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({
    source: "/root/a",
    destination: "/root/b",
  });
});

test("file-explorer-app/FR-007 treats immediate file deletion as complete", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
  vi.stubGlobal("fetch", fetchMock);
  expect(await deleteEntry({ kind: "file", path: "/root/a" })).toBeNull();
  expect(fetchMock.mock.calls[0][0]).toBe("/f/root/a");
});

test("file-explorer-app/FR-007 returns the job for directory deletion", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(job), { status: 202 }));
  vi.stubGlobal("fetch", fetchMock);
  expect(await deleteEntry({ kind: "directory", path: "/root/a" })).toEqual(job);
  expect(fetchMock.mock.calls[0][0]).toBe("/f/root/a");
});

test("file-explorer-app/FR-009 retrieves a server operation job", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(job), { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  expect(await getOperationJob(job.job_id)).toEqual(job);
  expect(fetchMock.mock.calls[0][0]).toBe(`/f/jobs/${job.job_id}`);
});

test("file-explorer-app/IR-001 rejects invalid input and failed responses", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          error: { code: "permission_denied" },
        }),
        { status: 403 },
      ),
    )
    .mockResolvedValueOnce(new Response(null, { status: 500 }));
  vi.stubGlobal("fetch", fetchMock);
  await expect(createDirectory("relative")).rejects.toThrow();
  await expect(getOperationJob("bad-id")).rejects.toThrow();
  expect(fetchMock).not.toHaveBeenCalled();
  await expect(transferEntry("move", "/a", "/b")).rejects.toThrow("permission denied");
  await expect(deleteEntry({ kind: "file", path: "/a" })).rejects.toThrow("HTTP 500");
});

test("file-explorer-app/FR-003 accepts only supported previews", () => {
  expect(canPreviewFile({ kind: "file", path: "/a.PDF" })).toBe(true);
  expect(canPreviewFile({ kind: "file", path: "/a.zip" })).toBe(false);
  expect(canPreviewFile({ kind: "directory", path: "/a.pdf" })).toBe(false);
});

test("file-explorer-app/FR-009 validates operation jobs and cached job references", () => {
  expect(operationJobSchema.safeParse(job).success).toBe(true);
  expect(operationJobSchema.safeParse({ ...job, percent: 101 }).success).toBe(false);
  expect(cachedJobsSchema.safeParse([{ id: job.job_id, label: "copy a" }]).success).toBe(true);
  expect(cachedJobsSchema.safeParse([{ id: "bad", label: "copy a" }]).success).toBe(false);
});
