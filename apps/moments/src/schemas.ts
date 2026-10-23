import { z } from "zod";

/** Credentials accepted by the shared LiteNAS login contract. */
export const loginSchema = z.object({
  login: z
    .string()
    .min(1)
    .regex(/^(?:[A-Za-z0-9._-]+|[^@\s]+@[^@\s]+\.[^@\s]+)$/),
  password: z.string().min(4),
});

/** Verified current-user response from the Moments auth API. */
export const meSchema = z.object({
  data: z.object({
    authenticated: z.literal(true),
    auth_type: z.string(),
    user: z.object({
      id: z.string(),
      login: z.string(),
      full_name: z.string().optional(),
      avatar_url: z.string().optional(),
    }),
    roles: z.array(z.string()),
    scopes: z.array(z.string()),
  }),
});

/** Session envelope returned after login or refresh. */
export const sessionSchema = z.object({
  data: z.object({ authenticated: z.literal(true), user: z.object({ id: z.string() }) }),
});

/** Current user's validated browser profile. */
export type Me = z.infer<typeof meSchema>["data"];
