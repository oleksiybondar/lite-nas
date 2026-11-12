import type { LoginRequestDTO } from "@dto/auth/auth";
import { z } from "zod";

/**
 * Login value accepted by the gateway login boundary.
 *
 * The pattern mirrors the Huma `LoginRequestBody.Login` constraint in
 * `services/web-gateway/dto/auth/login.go`.
 */
const loginPattern = /^(?:[A-Za-z0-9._-]+|[^@\s]+@[^@\s]+\.[^@\s]+)$/;

/**
 * Runtime schema for `/api/auth/login` request payloads.
 */
export const loginRequestSchema = z.object({
  login: z.string().regex(loginPattern, "Enter a valid login or email address."),
  password: z.string().min(1).max(4096),
}) satisfies z.ZodType<LoginRequestDTO>;
