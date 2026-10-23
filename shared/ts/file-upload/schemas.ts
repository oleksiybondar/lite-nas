import { z } from "zod";

/** Absolute, clean host paths accepted at the shared upload boundary. */
export const hostPathSchema = z
  .string()
  .min(1)
  .max(4096)
  .refine((value) => {
    if (!value.startsWith("/") || value.includes("\\") || value.includes("\0")) {
      return false;
    }
    if (value === "/") {
      return true;
    }
    return (
      !value.endsWith("/") &&
      value
        .slice(1)
        .split("/")
        .every((segment) => {
          return segment !== "" && segment !== "." && segment !== "..";
        })
    );
  });

/** File name accepted by File Service's create-upload DTO. */
export const uploadFilenameSchema = z
  .string()
  .min(1)
  .max(255)
  .refine((value) => {
    return value !== "." && value !== ".." && !/[\\/\r\n\0]/.test(value);
  });

/** Validated resource state returned by upload creation and progress reads. */
export const uploadProgressSchema = z.object({
  upload_id: z.string().min(1),
  state: z.enum(["active", "completed"]),
  total_size: z.number().int().nonnegative(),
  acknowledged_bytes: z.number().int().nonnegative(),
  next_chunk_index: z.number().int().nonnegative(),
  chunk_size: z
    .number()
    .int()
    .min(1)
    .max(16 * 1024 * 1024),
  maximum_chunk_size: z.number().int().min(1),
});

/** Validated final response from File Service. */
export const uploadCompletionSchema = z.object({
  upload_id: z.string().min(1),
  state: z.literal("completed"),
  sha256: z.string().regex(/^[0-9a-f]{64}$/i),
  resource: z.string().startsWith("/f/"),
});

/** Stable File Service error envelope. */
export const uploadErrorSchema = z.object({
  error: z.object({ code: z.string().min(1) }),
});

/** Current server capacity used after a retryable chunk admission failure. */
export const uploadWindowSchema = z.object({
  available_buffer_bytes: z.number().int().nonnegative(),
  maximum_chunk_size: z.number().int().min(1),
});

/** Request metadata checked before the first upload call. */
export const createUploadSchema = z.object({
  destination: hostPathSchema,
  filename: uploadFilenameSchema,
  size: z.number().int().nonnegative(),
  sha256: z.string().regex(/^[0-9a-f]{64}$/i),
  content_type: z.string().max(255),
  chunk_size: z
    .number()
    .int()
    .min(1024 * 1024)
    .max(16 * 1024 * 1024),
});
