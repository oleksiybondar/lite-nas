import { z } from "zod";

/** Validates the gateway's authenticated current-user response. */
export const authMeResponseSchema = z.object({
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

/** Validates login and refresh responses before using an issued session. */
export const authSessionResponseSchema = z.object({
  data: z.object({
    authenticated: z.literal(true),
    user: z.object({ id: z.string() }),
    redirect_to: z.string().max(2048).optional(),
  }),
});
