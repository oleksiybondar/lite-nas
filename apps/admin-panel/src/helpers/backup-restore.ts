import type { ApiContextValue } from "@dto/api/api";
import { mergeReadResponseSchema, mergeWriteResponseSchema } from "@schemas/backup-restore";

const chunkSize = 64 * 1024;
const fileLimit = 1024 * 1024;

/** Builds a fixed gateway path from validated inventory identifiers. */
export const backupPath = (application: string, version?: string): string => {
  const base = `/api/backup-restore/applications/${encodeURIComponent(application)}`;
  return version === undefined ? base : `${base}/versions/${encodeURIComponent(version)}`;
};

/** Parses a successful gateway envelope and rejects failed responses. */
export const readData = async <T>(
  response: Response,
  schema: { parse: (value: unknown) => { data: T } },
): Promise<T> => {
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return schema.parse(await response.json()).data;
};

/** Reads and verifies one complete session side through bounded chunks. */
export const readMergeText = async (
  get: ApiContextValue["get"],
  sessionID: string,
  path: string,
  side: "snapshot" | "current" | "result",
): Promise<string> => {
  const parts: Uint8Array[] = [];
  let offset = 0;
  let checksum = "";
  for (;;) {
    const chunk = await readMergeChunk({ get, sessionID, path, side, offset });
    parts.push(chunk.bytes);
    offset = chunk.nextOffset;
    checksum = chunk.checksum;
    if (chunk.done) break;
  }
  const bytes = new Uint8Array(offset);
  let position = 0;
  for (const part of parts) {
    bytes.set(part, position);
    position += part.length;
  }
  if ((await sha256(bytes)) !== checksum.toLowerCase())
    throw new Error("Merge content checksum mismatch");
  return new TextDecoder("utf-8", { fatal: true }).decode(bytes);
};

type ReadMergeChunk = { bytes: Uint8Array; nextOffset: number; checksum: string; done: boolean };

/** Fetches one bounded merge-content chunk and validates its offset and size. */
const readMergeChunk = async (request: {
  get: ApiContextValue["get"];
  sessionID: string;
  path: string;
  side: "snapshot" | "current" | "result";
  offset: number;
}): Promise<ReadMergeChunk> => {
  const { get, sessionID, path, side, offset } = request;
  const params = new URLSearchParams({ path, side, offset: String(offset) });
  const response = await get(
    `/api/backup-restore/merge-sessions/${sessionID}/content?${params}`,
  ).execute();
  const chunk = await readData(response, mergeReadResponseSchema);
  const binary = atob(chunk.data);
  const nextOffset = offset + binary.length;
  if (
    binary.length > chunkSize ||
    nextOffset > fileLimit ||
    chunk.next_offset !== nextOffset ||
    (!chunk.done && binary.length === 0)
  ) {
    throw new Error("Invalid merge content chunk");
  }
  return {
    bytes: Uint8Array.from(binary, (character) => character.charCodeAt(0)),
    nextOffset,
    checksum: chunk.sha256,
    done: chunk.done,
  };
};

/** Uploads a validated editor result in the service's sequential chunk format. */
export const writeMergeText = async (
  put: ApiContextValue["put"],
  sessionID: string,
  path: string,
  content: string,
): Promise<void> => {
  const bytes = new TextEncoder().encode(content);
  if (bytes.length > fileLimit) throw new Error("Merge result exceeds the file limit");
  const digest = await sha256(bytes);
  let offset = 0;
  do {
    offset = await writeMergeChunk({ put, sessionID, path, bytes, offset, digest });
  } while (offset < bytes.length);
};

/** Sends one validated sequential content chunk and returns the next offset. */
const writeMergeChunk = async (request: {
  put: ApiContextValue["put"];
  sessionID: string;
  path: string;
  bytes: Uint8Array;
  offset: number;
  digest: string;
}): Promise<number> => {
  const { put, sessionID, path, bytes, offset, digest } = request;
  const end = Math.min(offset + chunkSize, bytes.length);
  const final = end === bytes.length;
  const binary = Array.from(bytes.subarray(offset, end), (value) =>
    String.fromCharCode(value),
  ).join("");
  const response = await put(`/api/backup-restore/merge-sessions/${sessionID}/content`, {
    path,
    offset,
    data: btoa(binary),
    final,
    sha256: final ? digest : "",
  }).execute();
  const result = await readData(response, mergeWriteResponseSchema);
  if (result.next_offset !== end) throw new Error("Invalid merge write offset");
  return end;
};

/** Computes the full-content checksum required by the service. */
const sha256 = async (bytes: Uint8Array): Promise<string> => {
  const digest = await crypto.subtle.digest("SHA-256", bytes as BufferSource);
  return Array.from(new Uint8Array(digest), (value) => value.toString(16).padStart(2, "0")).join(
    "",
  );
};

/** Returns line numbers changed between two text panes, including insertions. */
export const changedLines = (source: string, target: string): Set<number> => {
  const left = source.split("\n");
  const right = target.split("\n");
  const changed = new Set<number>();
  let sourceIndex = 0;
  let targetIndex = 0;
  while (targetIndex < right.length) {
    switch (diffStep(left, right, sourceIndex, targetIndex)) {
      case "same":
        sourceIndex++;
        targetIndex++;
        break;
      case "source-only":
        changed.add(targetIndex);
        sourceIndex++;
        break;
      case "target-only":
        changed.add(targetIndex++);
        break;
      case "both":
        changed.add(targetIndex++);
        sourceIndex++;
        break;
    }
  }
  return changed;
};

/** Chooses an edit step with one-line lookahead to localize insertions and deletions. */
const diffStep = (
  left: string[],
  right: string[],
  sourceIndex: number,
  targetIndex: number,
): "same" | "source-only" | "target-only" | "both" => {
  if (left[sourceIndex] === right[targetIndex]) return "same";
  if (left[sourceIndex + 1] === right[targetIndex]) return "source-only";
  if (left[sourceIndex] === right[targetIndex + 1]) return "target-only";
  return "both";
};
