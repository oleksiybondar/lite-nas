import { sha256 } from "@noble/hashes/sha2.js";

import {
  createUploadSchema,
  hostPathSchema,
  uploadCompletionSchema,
  uploadErrorSchema,
  uploadFilenameSchema,
  uploadProgressSchema,
  uploadWindowSchema,
} from "./schemas";

const hashSliceSize = 1024 * 1024;
const uploadRoot = "/f/proprietary/upload";
type UploadProgress = ReturnType<typeof uploadProgressSchema.parse>;

/** Inputs needed to transfer and reconcile one file chunk. */
interface UploadChunkOptions {
  file: File;
  resource: string;
  progress: UploadProgress;
  signal: AbortSignal;
  onProgress: (progress: FileUploadProgress) => void;
}

/** Inputs for transferring all unacknowledged chunks of one file. */
interface SendRemainingChunksOptions extends Omit<UploadChunkOptions, "progress"> {
  initialProgress: UploadProgress;
}

/** Keeps small uploads cheap while reducing request count for large files. */
function chunkSizeFor(fileSize: number): number {
  const mib = 1024 * 1024;
  if (fileSize <= 8 * mib) {
    return mib;
  }
  if (fileSize <= 256 * mib) {
    return 4 * mib;
  }
  if (fileSize <= 2 * 1024 * mib) {
    return 8 * mib;
  }
  return 16 * mib;
}

/** Progress phase and byte count reported by one file transfer. */
export interface FileUploadProgress {
  phase: "hashing" | "uploading" | "completing";
  completed: number;
  total: number;
}

/** File Service failure with its stable server error code when available. */
export class FileUploadError extends Error {
  constructor(
    message: string,
    readonly code?: string,
  ) {
    super(message);
    this.name = "FileUploadError";
  }
}

/** Joins a validated directory and browser file name as one upload target. */
function destinationFor(directory: string, filename: string): string {
  hostPathSchema.parse(directory);
  uploadFilenameSchema.parse(filename);
  return createUploadSchema.shape.destination.parse(
    directory === "/" ? `/${filename}` : `${directory}/${filename}`,
  );
}

/** Hashes a browser file in bounded slices before creating an upload resource. */
async function hashFile(
  file: File,
  signal: AbortSignal,
  onProgress: (progress: FileUploadProgress) => void,
): Promise<string> {
  const hash = sha256.create();
  for (let offset = 0; offset < file.size; offset += hashSliceSize) {
    if (signal.aborted) {
      throw signal.reason;
    }
    const bytes = new Uint8Array(await file.slice(offset, offset + hashSliceSize).arrayBuffer());
    hash.update(bytes);
    onProgress({
      phase: "hashing",
      completed: Math.min(offset + bytes.length, file.size),
      total: file.size,
    });
  }
  return Array.from(hash.digest(), (byte) => byte.toString(16).padStart(2, "0")).join("");
}

/** Converts an unsuccessful File Service response into a visible transfer error. */
async function responseError(response: Response): Promise<FileUploadError> {
  const body: unknown = await response.json().catch(() => null);
  const code = uploadErrorSchema.safeParse(body);
  const errorCode = code.success ? code.data.error.code : undefined;
  return new FileUploadError(
    errorCode
      ? `Upload failed: ${errorCode.replaceAll("_", " ")}.`
      : `Upload failed (${response.status}).`,
    errorCode,
  );
}

/** Reads and validates a JSON response from the File Service boundary. */
async function readProgress(
  response: Response,
): Promise<ReturnType<typeof uploadProgressSchema.parse>> {
  if (!response.ok) {
    throw await responseError(response);
  }
  return uploadProgressSchema.parse(await response.json());
}

/** Reconciles uncertain chunk delivery with the server's committed position. */
async function currentProgress(resource: string, size: number, signal: AbortSignal) {
  const progress = await readProgress(
    await fetch(resource, { credentials: "same-origin", signal }),
  );
  if (
    progress.total_size !== size ||
    progress.acknowledged_bytes > size ||
    progress.state !== "active"
  ) {
    throw new FileUploadError("Upload state no longer matches the selected file.");
  }
  return progress;
}

/** Sends one chunk request and treats a lost response as uncertain delivery. */
async function requestChunk(
  resource: string,
  chunk: Blob,
  offset: number,
  signal: AbortSignal,
): Promise<Response | null> {
  try {
    return await fetch(resource, {
      method: "PATCH",
      credentials: "same-origin",
      headers: { "Content-Type": "application/octet-stream", "Upload-Offset": String(offset) },
      body: chunk,
      signal,
    });
  } catch (error) {
    if (signal.aborted) {
      throw error;
    }
    return null;
  }
}

/** Checks whether a response contains one recognized upload retry code. */
async function hasUploadErrorCode(
  response: Response,
  status: number,
  code: string,
): Promise<boolean> {
  if (response.status !== status) {
    return false;
  }
  const body: unknown = await response
    .clone()
    .json()
    .catch(() => null);
  const parsed = uploadErrorSchema.safeParse(body);
  return parsed.success && parsed.data.error.code === code;
}

/** Sends one chunk, reconciles retryable responses, and returns current progress. */
async function uploadChunk({
  file,
  resource,
  progress,
  signal,
  onProgress,
}: UploadChunkOptions): Promise<UploadProgress> {
  const offset = progress.acknowledged_bytes;
  const chunk = file.slice(offset, offset + progress.chunk_size);
  const response = await requestChunk(resource, chunk, offset, signal);
  if (response === null) {
    return currentProgress(resource, file.size, signal);
  }
  if (await hasUploadErrorCode(response, 503, "upload_window_full")) {
    await waitForWindow(progress.chunk_size, response, signal);
    return currentProgress(resource, file.size, signal);
  }
  if (await hasUploadErrorCode(response, 409, "chunk_conflict")) {
    return currentProgress(resource, file.size, signal);
  }
  if (!response.ok) {
    throw await responseError(response);
  }
  const next = await currentProgress(resource, file.size, signal);
  if (next.acknowledged_bytes <= offset) {
    throw new FileUploadError("Upload did not advance after a successful chunk.");
  }
  onProgress({ phase: "uploading", completed: next.acknowledged_bytes, total: file.size });
  return next;
}

/** Validates the created upload resource against the selected file. */
function validateCreatedProgress(progress: UploadProgress, file: File): void {
  if (
    progress.total_size !== file.size ||
    progress.acknowledged_bytes > file.size ||
    progress.state !== "active"
  ) {
    throw new FileUploadError("Upload creation returned an invalid state.");
  }
}

/** Completes an upload after every file byte has been acknowledged. */
async function completeUpload(
  resource: string,
  file: File,
  signal: AbortSignal,
  onProgress: (progress: FileUploadProgress) => void,
): Promise<void> {
  onProgress({ phase: "completing", completed: file.size, total: file.size });
  const response = await fetch(`${resource}/completion`, {
    method: "PUT",
    credentials: "same-origin",
    signal,
  });
  if (!response.ok) {
    throw await responseError(response);
  }
  uploadCompletionSchema.parse(await response.json());
}

/** Deletes a partially uploaded resource after explicit cancellation. */
async function cleanUpAbortedUpload(resource: string): Promise<void> {
  await fetch(resource, { method: "DELETE", credentials: "same-origin" }).catch(() => undefined);
}

/** Sends the remaining file chunks and reports each acknowledged byte count. */
async function sendRemainingChunks(options: SendRemainingChunksOptions): Promise<void> {
  let progress = options.initialProgress;
  while (progress.acknowledged_bytes < options.file.size) {
    if (options.signal.aborted) {
      throw options.signal.reason;
    }
    progress = await uploadChunk({ ...options, progress });
  }
}

/** Waits for the documented upload window to admit one configured chunk. */
async function waitForWindow(
  chunkSize: number,
  retryResponse: Response,
  signal: AbortSignal,
): Promise<void> {
  await waitForRetry(retryResponse, signal);
  while (true) {
    const response = await fetch(`${uploadRoot}-window`, { credentials: "same-origin", signal });
    if (!response.ok) {
      throw await responseError(response);
    }
    const windowState = uploadWindowSchema.parse(await response.json());
    if (windowState.available_buffer_bytes >= chunkSize) {
      return;
    }
    await waitForRetry(retryResponse, signal);
  }
}

/** Waits for an upload window retry while respecting cancellation. */
async function waitForRetry(response: Response, signal: AbortSignal): Promise<void> {
  if (signal.aborted) {
    throw signal.reason;
  }
  const seconds = Number(response.headers.get("Retry-After"));
  const duration = Number.isFinite(seconds) && seconds > 0 ? Math.min(seconds, 60) : 2;
  await new Promise<void>((resolve, reject) => {
    const timer = window.setTimeout(() => {
      signal.removeEventListener("abort", onAbort);
      resolve();
    }, duration * 1000);
    const onAbort = () => {
      window.clearTimeout(timer);
      reject(signal.reason);
    };
    signal.addEventListener("abort", onAbort, { once: true });
    if (signal.aborted) {
      onAbort();
    }
  });
}

/** Uploads one file with File Service's create, chunk, and completion protocol. */
export async function uploadFile(
  file: File,
  directory: string,
  signal: AbortSignal,
  onProgress: (progress: FileUploadProgress) => void,
): Promise<void> {
  const destination = destinationFor(directory, file.name);
  const checksum = await hashFile(file, signal, onProgress);
  const metadata = createUploadSchema.parse({
    destination,
    filename: file.name,
    size: file.size,
    sha256: checksum,
    content_type: file.type,
    chunk_size: chunkSizeFor(file.size),
  });
  const created = await fetch(uploadRoot, {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", "Idempotency-Key": crypto.randomUUID() },
    body: JSON.stringify(metadata),
    signal,
  });
  const progress = await readProgress(created);
  const resource = `${uploadRoot}/${encodeURIComponent(progress.upload_id)}`;
  validateCreatedProgress(progress, file);

  try {
    await sendRemainingChunks({ file, resource, initialProgress: progress, signal, onProgress });
    await completeUpload(resource, file, signal, onProgress);
  } catch (error) {
    if (signal.aborted) {
      await cleanUpAbortedUpload(resource);
    }
    throw error;
  }
}
