import { z } from "zod";

/** Wraps an operation's data schema in the gateway's response envelope. */
const envelope = <T extends z.ZodType>(data: T) =>
  z.object({ success: z.literal(true), timestamp: z.string(), data });

/** Validates application and version inventories returned by the gateway. */
export const applicationListResponseSchema = envelope(
  z.array(z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/)),
);

/** Validates backup version identifiers before using them as route operands. */
export const versionListResponseSchema = envelope(
  z.array(z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/)),
);

/** Validates a created snapshot's version. */
export const versionResponseSchema = envelope(
  z.object({ version: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/) }),
);

/** Validates metadata embedded in one stored application archive. */
export const backupDetailsResponseSchema = envelope(
  z.object({
    application: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/),
    version: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/),
    restart_units: z.array(z.string()).optional(),
    name: z.string().optional(),
    description: z.string().optional(),
    created_at: z.string().datetime({ offset: true }),
    created_by: z.string(),
  }),
);

/** Validates each outcome returned by the synchronous backup-all operation. */
export const backupAllResponseSchema = envelope(
  z.array(
    z.object({
      application: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/),
      status: z.enum(["completed", "skipped", "failed"]),
      version: z
        .string()
        .regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/)
        .optional(),
      error: z.string().optional(),
    }),
  ),
);

/** Validates confirmation that a stored archive was deleted. */
export const backupDeleteResponseSchema = envelope(z.object({ deleted: z.literal(true) }));

/** Validates a path-level comparison with live files. */
export const diffResponseSchema = envelope(
  z.array(z.object({ path: z.string(), status: z.string() })),
);

/** Validates one editable file in a merge session. */
export const mergeFileSchema = z.object({
  path: z.string().min(1),
  snapshot_size: z.number().int().nonnegative(),
  current_size: z.number().int().nonnegative(),
  status: z.string(),
});

/** Validates session creation and status responses. */
export const mergeSessionResponseSchema = envelope(
  z.object({
    session_id: z.string().regex(/^[a-fA-F0-9]{32}$/),
    files: z.array(mergeFileSchema),
    apply_available: z.boolean(),
  }),
);

/** Validates a bounded content transfer chunk. */
export const mergeReadResponseSchema = envelope(
  z.object({
    data: z.string(),
    next_offset: z.number().int().nonnegative(),
    done: z.boolean(),
    sha256: z.string().regex(/^[a-fA-F0-9]{64}$/),
  }),
);

/** Validates acknowledgment of a sequential result upload. */
export const mergeWriteResponseSchema = envelope(
  z.object({ next_offset: z.number().int().nonnegative() }),
);

/** Validates a completed restore. */
export const restoreResponseSchema = envelope(
  z.object({ restored: z.literal(true), warnings: z.array(z.string()).optional() }),
);

/** Validates the exact name and mode submitted for a named restore. */
export const restoreByNameRequestSchema = z.object({
  name: z.string().min(1).max(128),
  mode: z.enum(["replace", "merge"]),
});

/** Validates the selected version returned by a named restore. */
export const restoreByNameResponseSchema = envelope(
  z.object({
    version: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/),
    restored: z.literal(true),
    warnings: z.array(z.string()).optional(),
  }),
);

/** Validates every outcome of a restore-all-by-name request. */
export const restoreAllByNameResponseSchema = envelope(
  z.object({
    results: z.array(
      z.object({
        application: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/),
        status: z.enum(["completed", "skipped", "failed"]),
        version: z
          .string()
          .regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/)
          .optional(),
        warnings: z.array(z.string()).optional(),
        error: z.string().optional(),
      }),
    ),
  }),
);

/** Validates a completed apply or discard. */
export const mergeActionResponseSchema = envelope(
  z.object({ completed: z.literal(true), warnings: z.array(z.string()).optional() }),
);

/** One editable file described by the active session. */
export type MergeFile = z.infer<typeof mergeFileSchema>;

/** One active merge session. */
export type MergeSession = z.infer<typeof mergeSessionResponseSchema>["data"];

/** One path-level change in a backup. */
export type BackupDifference = z.infer<typeof diffResponseSchema>["data"][number];
/** One backup card's stored metadata. */
export type BackupDetails = z.infer<typeof backupDetailsResponseSchema>["data"];

/** Validates metadata entered for a new snapshot. */
export const backupMetadataSchema = z.object({
  name: z.string().max(128),
  description: z.string().max(1024),
});

/** Validates the selected restore mode. */
export const restoreModeSchema = z.enum(["replace", "merge"]);

/** Validates the editable UFW rule export before the merge provider uploads it. */
export const ufwStateSchema = z
  .object({
    schemaVersion: z.literal(1),
    enabled: z.boolean(),
    logging: z.enum(["off", "low", "medium", "high", "full"]),
    defaults: z
      .object({
        incoming: z.enum(["allow", "deny", "reject"]),
        outgoing: z.enum(["allow", "deny", "reject"]),
        routed: z.enum(["allow", "deny", "reject"]),
      })
      .strict(),
    rules: z
      .array(
        z
          .object({
            kind: z.enum(["rule", "route"]),
            action: z.enum(["allow", "deny", "reject", "limit"]),
            arguments: z
              .array(
                z
                  .string()
                  .min(1)
                  .max(256)
                  .refine(
                    (value) =>
                      !value.startsWith("-") &&
                      ![...value].some((character) => character.charCodeAt(0) < 32),
                  ),
              )
              .min(1)
              .max(64),
          })
          .strict(),
      )
      .max(4096),
  })
  .strict();

/** Parses an edited UFW JSON document before any merge upload request. */
export const validateUFWMergeText = (content: string): void => {
  ufwStateSchema.parse(JSON.parse(content));
};
