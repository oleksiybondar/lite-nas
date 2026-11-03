import { z } from "zod";

/** Validates the gateway's LiteNAS Auth session inventory. */
export const authSessionsResponseSchema = z.object({
  success: z.literal(true),
  timestamp: z.string(),
  data: z.array(
    z.object({
      session_id: z.string().regex(/^[A-Za-z0-9_-]{1,128}$/),
      subject: z.string().min(1),
      login: z.string().min(1),
      source: z.enum(["admin-panel", "moments"]).optional(),
      expires_at: z.iso.datetime({ offset: true }),
    }),
  ),
});

/** Validates confirmation of an authentication session revocation. */
export const authSessionRevokeResponseSchema = z.object({
  success: z.literal(true),
  timestamp: z.string(),
  revoked: z.boolean(),
});

/** One validated LiteNAS authentication session. */
export type AuthSession = z.infer<typeof authSessionsResponseSchema>["data"][number];
