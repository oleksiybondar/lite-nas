import type { z } from "zod";
import { fileServicePath, parentPath } from "./paths";
import {
  createDirectorySchema,
  type directoryEntrySchema,
  directoryListingSchema,
  fileServiceErrorSchema,
  hostPathSchema,
  operationJobSchema,
  transferEntrySchema,
} from "./schemas";

/** Validated server-side operation status. */
export type OperationJob = z.infer<typeof operationJobSchema>;

/** One validated direct child and its nullable filesystem metadata. */
export type FileEntry = z.infer<typeof directoryEntrySchema>;

/** Preview renderer selected from the allowlisted file extension. */
export type FilePreviewKind = "image" | "pdf" | "media" | "text";

const previewKindByExtension: Record<string, FilePreviewKind> = {
  jpg: "image",
  jpeg: "image",
  png: "image",
  gif: "image",
  webp: "image",
  pdf: "pdf",
  mp4: "media",
  m4v: "media",
  mov: "media",
  avi: "media",
  mkv: "media",
  webm: "media",
  ogg: "media",
  ogv: "media",
  mp3: "media",
  wav: "media",
  m4a: "media",
  aac: "media",
  flac: "media",
  txt: "text",
  xml: "text",
  md: "text",
  json: "text",
  yaml: "text",
  yml: "text",
  csv: "text",
  log: "text",
  html: "text",
  css: "text",
  js: "text",
  jsx: "text",
  ts: "text",
  tsx: "text",
  mjs: "text",
  cjs: "text",
  py: "text",
  go: "text",
  rs: "text",
  java: "text",
  c: "text",
  h: "text",
  cpp: "text",
  hpp: "text",
  sh: "text",
  bash: "text",
  sql: "text",
  toml: "text",
  ini: "text",
  conf: "text",
  diff: "text",
};

/** Returns the safe built-in preview renderer for a supported file entry. */
export function filePreviewKind(entry: Pick<FileEntry, "kind" | "path">): FilePreviewKind | null {
  if (entry.kind !== "file") return null;
  const extension = entry.path.slice(entry.path.lastIndexOf(".") + 1).toLowerCase();
  return previewKindByExtension[extension] ?? null;
}

/** True when a file has a safe built-in inline preview renderer. */
export function canPreviewFile(entry: FileEntry): boolean {
  return filePreviewKind(entry) !== null;
}

/** Identifies dot-prefixed direct children that the UI can hide on request. */
export function isHiddenFileEntry(entry: Pick<FileEntry, "name">): boolean {
  return entry.name.startsWith(".");
}

/** Filters dot-prefixed entries when Show hidden files is disabled. */
export function visibleFileEntries(entries: FileEntry[], showHidden: boolean): FileEntry[] {
  return showHidden ? entries : entries.filter((entry) => !isHiddenFileEntry(entry));
}

/** Optional file suffix rules; directories remain visible for navigation. */
export interface FileExtensionFilter {
  include?: readonly string[];
  exclude?: readonly string[];
}

/** Applies case-insensitive extension rules to files, including extensionless files. */
export function filterFileEntries(
  entries: FileEntry[],
  filter: FileExtensionFilter = {},
): FileEntry[] {
  const normalize = (value: string): string => value.replace(/^\./, "").toLowerCase();
  const included = filter.include?.map(normalize);
  const excluded = filter.exclude?.map(normalize);
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: include and exclude checks stay explicit at the filtering boundary.
  return entries.filter((entry) => {
    if (entry.kind === "directory") return true;
    const dot = entry.name.lastIndexOf(".");
    const extension =
      dot > 0 && dot < entry.name.length - 1 ? entry.name.slice(dot + 1).toLowerCase() : "";
    return (
      (included === undefined || included.includes(extension)) && !excluded?.includes(extension)
    );
  });
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
export async function createDirectory(path: string, privateDirectory = false): Promise<void> {
  const body = createDirectorySchema.parse({ path, private: privateDirectory });
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
export async function deleteEntry(
  entry: Pick<FileEntry, "path" | "kind">,
): Promise<OperationJob | null> {
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
  if (
    listing.some(
      (entry) =>
        parentPath(entry.path) !== path ||
        entry.name !== entry.path.slice(entry.path.lastIndexOf("/") + 1),
    )
  ) {
    throw new Error("File Service returned an invalid directory listing.");
  }
  return listing.sort(compareEntries);
}
