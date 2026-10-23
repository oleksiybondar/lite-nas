import type { z } from "zod";
import { fileServicePath, parentPath } from "./paths";
import {
  createDirectorySchema,
  directoryListingSchema,
  fileServiceErrorSchema,
  hostPathSchema,
  operationJobSchema,
  transferEntrySchema,
} from "./schemas";

/** Validated server-side operation status. */
export type OperationJob = z.infer<typeof operationJobSchema>;

/** One validated direct child presented by the File Explorer. */
export interface FileEntry {
  kind: "directory" | "file";
  path: string;
}

/** True when File Service serves this file type inline for a modal preview. */
export function canPreviewFile(entry: FileEntry): boolean {
  if (entry.kind !== "file") {
    return false;
  }
  return /\.(txt|jpe?g|png|gif|webp|pdf)$/i.test(entry.path);
}

const nameCollator = new Intl.Collator(undefined, { numeric: true, sensitivity: "base" });

/** Sorts direct entries by directory first, then by their natural file name. */
function compareEntries(left: FileEntry, right: FileEntry): number {
  if (left.kind !== right.kind) {
    return left.kind === "directory" ? -1 : 1;
  }
  const leftName = left.path.slice(left.path.lastIndexOf("/") + 1);
  const rightName = right.path.slice(right.path.lastIndexOf("/") + 1);
  return nameCollator.compare(leftName, rightName);
}

/** Converts one failed File Service response into a user-visible error. */
async function mutationError(response: Response): Promise<Error> {
  const body: unknown = await response.json().catch(() => null);
  const parsed = fileServiceErrorSchema.safeParse(body);
  const code = parsed.success ? parsed.data.error.code : `HTTP ${response.status}`;
  return new Error(`File operation failed: ${code.replaceAll("_", " ")}.`);
}

/** Creates one directory under the File Service's user-scoped policy. */
export async function createDirectory(path: string): Promise<void> {
  const body = createDirectorySchema.parse({ path });
  const response = await fetch("/f/mkdir", {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw await mutationError(response);
  }
}

/** Starts one server-side copy or move to a complete destination path. */
export async function transferEntry(
  kind: "copy" | "move",
  source: string,
  destination: string,
): Promise<OperationJob> {
  const body = transferEntrySchema.parse({ source, destination });
  const response = await fetch(`/f/${kind}`, {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!response.ok) throw await mutationError(response);
  return operationJobSchema.parse(await response.json());
}

/** Deletes a file immediately or starts a server job for a directory. */
export async function deleteEntry(entry: FileEntry): Promise<OperationJob | null> {
  hostPathSchema.parse(entry.path);
  const response = await fetch(fileServicePath("/f", entry.path), {
    method: "DELETE",
    credentials: "same-origin",
  });
  if (!response.ok) throw await mutationError(response);
  return response.status === 202 ? operationJobSchema.parse(await response.json()) : null;
}

/** Reads progress for a previously submitted server job. */
export async function getOperationJob(id: string): Promise<OperationJob> {
  operationJobSchema.shape.job_id.parse(id);
  const response = await fetch(`/f/jobs/${id}`, { credentials: "same-origin" });
  if (!response.ok) throw await mutationError(response);
  return operationJobSchema.parse(await response.json());
}

/** Reads one authorized directory and sorts its direct children. */
export async function listDirectory(path: string, signal: AbortSignal): Promise<FileEntry[]> {
  hostPathSchema.parse(path);
  const response = await fetch(fileServicePath("/ls", path), {
    credentials: "same-origin",
    signal,
  });
  if (!response.ok) {
    const body: unknown = await response.json().catch(() => null);
    const parsed = fileServiceErrorSchema.safeParse(body);
    const code = parsed.success ? parsed.data.error.code : `HTTP ${response.status}`;
    throw new Error(`Cannot open ${path}: ${code.replaceAll("_", " ")}.`);
  }

  const listing = directoryListingSchema.parse(await response.json());
  const entries: FileEntry[] = [
    ...listing.directories.map((entry) => ({ kind: "directory" as const, path: entry })),
    ...listing.files.map((entry) => ({ kind: "file" as const, path: entry })),
  ];
  if (entries.some((entry) => parentPath(entry.path) !== path)) {
    throw new Error("File Service returned an invalid directory listing.");
  }
  return entries.sort(compareEntries);
}
