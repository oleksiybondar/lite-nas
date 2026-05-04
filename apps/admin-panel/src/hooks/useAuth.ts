import { AuthContext } from "@contexts/auth-context";
import type { AuthContextValue } from "@dto/auth/auth";
import { useContext } from "react";

/**
 * Reads browser auth state and commands from React context.
 *
 * Feature components should destructure only the values they need from this
 * hook, for example `const { isAuthenticated, login } = useAuth();`.
 */
export const useAuth = (): AuthContextValue => {
  const context = useContext(AuthContext);

  if (context === undefined) {
    throw new Error("useAuth must be used inside AuthProvider");
  }

  return context;
};
