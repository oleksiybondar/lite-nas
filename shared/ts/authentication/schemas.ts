import { z } from "zod";

/** One server-controlled PAM prompt and its non-secret messages. */
export const passwordConversationSchema = z.object({
  id: z.string().min(1).max(128),
  sequence: z.number().int().positive(),
  prompt: z.string(),
  secret: z.boolean(),
  expires_at: z.string(),
  messages: z
    .array(z.object({ level: z.enum(["info", "warn", "error"]), text: z.string() }))
    .optional(),
});

/** Runtime response contract shared by login and password change endpoints. */
export const authFlowSchema = z.object({
  data: z.object({
    authenticated: z.boolean(),
    status: z
      .enum(["authenticated", "password_change_required", "denied", "service_unavailable"])
      .default("authenticated"),
    user: z.object({ id: z.string() }),
    redirect_to: z.string().max(2048).optional(),
    conversation: passwordConversationSchema.optional(),
    messages: z
      .array(z.object({ level: z.enum(["info", "warn", "error"]), text: z.string() }))
      .optional(),
  }),
});

/** Validates an answer before it leaves the browser. */
export const passwordReplySchema = z.object({
  conversation_id: z.string().min(1).max(128),
  sequence: z.number().int().positive(),
  answer: z.string().min(1).max(4096),
});

export type AuthFlow = z.infer<typeof authFlowSchema>["data"];
export type PasswordConversation = z.infer<typeof passwordConversationSchema>;
