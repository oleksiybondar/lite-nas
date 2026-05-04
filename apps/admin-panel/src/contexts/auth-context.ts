import type { AuthContextValue } from "@dto/auth/auth";
import { createContext } from "react";

/**
 * Context for browser authentication state and auth commands.
 *
 * Consumers should use `useAuth` so missing provider wiring fails with a clear
 * local error instead of leaking `undefined` checks into feature components.
 */
export const AuthContext = createContext<AuthContextValue | undefined>(undefined);
