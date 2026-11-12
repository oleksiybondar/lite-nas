import { z } from "zod";

/** Validates the gateway's host session inventory before display or actions. */
export const sessionsResponseSchema = z.object({
  success: z.literal(true),
  timestamp: z.string(),
  data: z.array(
    z.object({
      id: z.string().regex(/^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$/),
      uid: z.string().regex(/^[0-9]+$/),
      user: z.string().min(1),
      seat: z.string(),
      leader: z.string().regex(/^(?:[0-9]+|-)$/),
      class: z.string().min(1),
      tty: z.string(),
      type: z.string(),
      service: z.string(),
      remote: z.boolean(),
      remote_host: z.string(),
      display: z.string(),
      desktop: z.string(),
    }),
  ),
});

/** One validated login session returned by the gateway. */
export type HostSession = z.infer<typeof sessionsResponseSchema>["data"][number];
