import { RbacContext } from "@contexts/rbac-context";
import type { RbacContextValue } from "@dto/rbac/rbac";
import { useContext } from "react";

/**
 * Reads RBAC state and guard helpers from React context.
 *
 * Feature components should destructure only the values they need from this
 * hook, for example `const { requireSecurity } = useRbac();`. The hook exposes
 * domain-level guards instead of generic role-matching primitives so feature
 * code stays aligned with the app's authorization vocabulary.
 */
export const useRbac = (): RbacContextValue => {
  const context = useContext(RbacContext);

  if (context === undefined) {
    throw new Error("useRbac must be used inside RbacProvider");
  }

  return context;
};
