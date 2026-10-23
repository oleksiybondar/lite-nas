import { z } from "zod";

/** Absolute host path accepted from File Service and navigation state. */
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

/** Direct children returned by GET /ls/{absolute-path...}. */
export const directoryListingSchema = z.object({
  files: z.array(hostPathSchema),
  directories: z.array(hostPathSchema),
});

/** Stable error envelope returned by File Service. */
export const fileServiceErrorSchema = z.object({
  error: z.object({ code: z.string().min(1) }),
});

/** One new directory name, excluding path separators and special entries. */
export const newFolderNameSchema = z
  .string()
  .min(1)
  .max(255)
  .refine((value) => value !== "." && value !== ".." && !/[\\/\r\n\0]/.test(value));

/** Body sent to the File Service create-directory route. */
export const createDirectorySchema = z.object({ path: hostPathSchema });

/** Body sent to the File Service copy and move routes. */
export const transferEntrySchema = z.object({
  source: hostPathSchema,
  destination: hostPathSchema,
});

/** Server-owned copy, move, or directory-delete job state. */
export const operationJobSchema = z.object({
  job_id: z.string().regex(/^[0-9a-f]{48}$/),
  state: z.enum(["pending", "running", "done", "error"]),
  percent: z.number().int().min(0).max(100),
  completed_bytes: z.number().int().nonnegative(),
  total_bytes: z.number().int().nonnegative(),
  completed_items: z.number().int().nonnegative(),
  total_items: z.number().int().nonnegative(),
  error_code: z.string().optional(),
});

/** Browser cache contains only job IDs and display labels. */
export const cachedJobsSchema = z
  .array(z.object({ id: operationJobSchema.shape.job_id, label: z.string().min(1).max(512) }))
  .max(128);
